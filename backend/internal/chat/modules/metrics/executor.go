package metrics

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sachin-sivadasan/ledgerguard/internal/chat"
)

type executor struct {
	gql *chat.GraphQLExecutor
}

func (e *executor) execute(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	switch call.Name {
	case "get_latest_metrics":
		return e.getLatestMetrics(ctx, call)
	case "get_metrics_trend":
		return e.getMetricsTrend(ctx, call)
	case "get_aggregate_metrics":
		return e.getAggregateMetrics(ctx, call)
	default:
		return errorResult(call.ID, fmt.Sprintf("unknown metrics tool: %s", call.Name))
	}
}

type appIDArgs struct {
	AppID string `json:"app_id"`
}

func (e *executor) getLatestMetrics(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	var args appIDArgs
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		return errorResult(call.ID, "invalid arguments: "+err.Error())
	}

	query := `query($appId: ID!) {
		metrics(appId: $appId) {
			activeMrrCents revenueAtRiskCents usageRevenueCents totalRevenueCents
			renewalSuccessRate safeCount oneCycleMissedCount twoCyclesMissedCount churnedCount
		}
	}`

	data, err := e.gql.Execute(ctx, query, map[string]any{"appId": args.AppID})
	if err != nil {
		return errorResult(call.ID, err.Error())
	}
	return chat.ToolResult{ToolCallID: call.ID, Content: string(data)}
}

type trendArgs struct {
	AppID  string `json:"app_id"`
	Months *int   `json:"months,omitempty"`
}

func (e *executor) getMetricsTrend(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	var args trendArgs
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		return errorResult(call.ID, "invalid arguments: "+err.Error())
	}

	vars := map[string]any{"appId": args.AppID}
	if args.Months != nil {
		vars["months"] = *args.Months
	}

	query := `query($appId: ID!, $months: Int) {
		metricsTrend(appId: $appId, months: $months) {
			date activeMrrCents renewalSuccessRate revenueAtRiskCents safeCount churnedCount
		}
	}`

	data, err := e.gql.Execute(ctx, query, vars)
	if err != nil {
		return errorResult(call.ID, err.Error())
	}
	return chat.ToolResult{ToolCallID: call.ID, Content: string(data)}
}

func (e *executor) getAggregateMetrics(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	// Aggregate across all apps: first get apps, then get metrics for each
	query := `query {
		apps {
			id name
		}
	}`

	data, err := e.gql.Execute(ctx, query, nil)
	if err != nil {
		return errorResult(call.ID, err.Error())
	}

	var appsResult struct {
		Apps []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"apps"`
	}
	if err := json.Unmarshal(data, &appsResult); err != nil {
		return errorResult(call.ID, "failed to parse apps: "+err.Error())
	}

	type appMetrics struct {
		AppID   string `json:"app_id"`
		AppName string `json:"app_name"`
		Metrics any    `json:"metrics"`
	}

	results := make([]appMetrics, 0, len(appsResult.Apps))
	for _, app := range appsResult.Apps {
		metricsQuery := `query($appId: ID!) {
			metrics(appId: $appId) {
				activeMrrCents revenueAtRiskCents totalRevenueCents renewalSuccessRate
			}
		}`
		mData, err := e.gql.Execute(ctx, metricsQuery, map[string]any{"appId": app.ID})
		if err != nil {
			continue // skip apps with no metrics
		}
		var m struct {
			Metrics any `json:"metrics"`
		}
		json.Unmarshal(mData, &m)
		results = append(results, appMetrics{AppID: app.ID, AppName: app.Name, Metrics: m.Metrics})
	}

	out, _ := json.Marshal(map[string]any{"apps": results, "app_count": len(results)})
	return chat.ToolResult{ToolCallID: call.ID, Content: string(out)}
}

func errorResult(toolCallID, msg string) chat.ToolResult {
	return chat.ToolResult{
		ToolCallID: toolCallID,
		Content:    fmt.Sprintf(`{"error": %q}`, msg),
		IsError:    true,
	}
}
