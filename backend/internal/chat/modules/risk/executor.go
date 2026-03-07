package risk

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
	case "get_risk_summary":
		return e.getRiskSummary(ctx, call)
	case "list_at_risk":
		return e.listAtRisk(ctx, call)
	case "get_risk_timeline":
		return e.getRiskTimeline(ctx, call)
	default:
		return chat.ToolResult{
			ToolCallID: call.ID,
			Content:    fmt.Sprintf(`{"error": "unknown risk tool: %s"}`, call.Name),
			IsError:    true,
		}
	}
}

type appIDArgs struct {
	AppID string `json:"app_id"`
}

func (e *executor) getRiskSummary(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	var args appIDArgs
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		return errorResult(call.ID, "invalid arguments: "+err.Error())
	}

	query := `query($appId: ID!) {
		riskSummary(appId: $appId) {
			totalSubscriptions
			safe
			oneCycleMissed
			twoCyclesMissed
			churned
			revenueAtRiskCents
			atRiskSubscriptions {
				domain
				shopName
				riskState
				daysPastDue
				mrrCents
			}
		}
	}`

	data, err := e.gql.Execute(ctx, query, map[string]any{"appId": args.AppID})
	if err != nil {
		return errorResult(call.ID, err.Error())
	}

	return chat.ToolResult{ToolCallID: call.ID, Content: string(data)}
}

type listAtRiskArgs struct {
	AppID     string  `json:"app_id"`
	RiskState *string `json:"risk_state,omitempty"`
}

func (e *executor) listAtRisk(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	var args listAtRiskArgs
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		return errorResult(call.ID, "invalid arguments: "+err.Error())
	}

	query := `query($appId: ID!) {
		riskSummary(appId: $appId) {
			atRiskSubscriptions {
				domain
				shopName
				riskState
				daysPastDue
				mrrCents
				lastPaymentDate
				expectedNextCharge
			}
		}
	}`

	data, err := e.gql.Execute(ctx, query, map[string]any{"appId": args.AppID})
	if err != nil {
		return errorResult(call.ID, err.Error())
	}

	// If a specific risk_state filter was requested, filter the results
	if args.RiskState != nil {
		var result struct {
			RiskSummary struct {
				AtRiskSubscriptions []json.RawMessage `json:"atRiskSubscriptions"`
			} `json:"riskSummary"`
		}
		if err := json.Unmarshal(data, &result); err == nil {
			filtered := make([]json.RawMessage, 0)
			for _, sub := range result.RiskSummary.AtRiskSubscriptions {
				var s struct {
					RiskState string `json:"riskState"`
				}
				if err := json.Unmarshal(sub, &s); err == nil && s.RiskState == *args.RiskState {
					filtered = append(filtered, sub)
				}
			}
			filteredData, _ := json.Marshal(map[string]any{
				"riskSummary": map[string]any{
					"atRiskSubscriptions": filtered,
				},
			})
			return chat.ToolResult{ToolCallID: call.ID, Content: string(filteredData)}
		}
	}

	return chat.ToolResult{ToolCallID: call.ID, Content: string(data)}
}

type riskTimelineArgs struct {
	AppID  string `json:"app_id"`
	Months *int   `json:"months,omitempty"`
}

func (e *executor) getRiskTimeline(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	var args riskTimelineArgs
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		return errorResult(call.ID, "invalid arguments: "+err.Error())
	}

	vars := map[string]any{"appId": args.AppID}
	if args.Months != nil {
		vars["months"] = *args.Months
	}

	query := `query($appId: ID!, $months: Int) {
		metricsTrend(appId: $appId, months: $months) {
			date
			safeCount
			churnedCount
			revenueAtRiskCents
		}
	}`

	data, err := e.gql.Execute(ctx, query, vars)
	if err != nil {
		return errorResult(call.ID, err.Error())
	}

	return chat.ToolResult{ToolCallID: call.ID, Content: string(data)}
}

func errorResult(toolCallID, msg string) chat.ToolResult {
	return chat.ToolResult{
		ToolCallID: toolCallID,
		Content:    fmt.Sprintf(`{"error": %q}`, msg),
		IsError:    true,
	}
}
