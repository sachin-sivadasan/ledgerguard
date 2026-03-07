package subscriptions

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
	case "list_subscriptions":
		return e.listSubscriptions(ctx, call)
	case "get_subscription_detail":
		return e.getSubscriptionDetail(ctx, call)
	case "get_subscription_summary":
		return e.getSubscriptionSummary(ctx, call)
	case "search_subscriptions":
		return e.searchSubscriptions(ctx, call)
	default:
		return errorResult(call.ID, fmt.Sprintf("unknown subscriptions tool: %s", call.Name))
	}
}

type listArgs struct {
	AppID     string  `json:"app_id"`
	RiskState *string `json:"risk_state,omitempty"`
	Status    *string `json:"status,omitempty"`
	Domain    *string `json:"domain,omitempty"`
	Limit     *int    `json:"limit,omitempty"`
	Offset    *int    `json:"offset,omitempty"`
}

func (e *executor) listSubscriptions(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	var args listArgs
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		return errorResult(call.ID, "invalid arguments: "+err.Error())
	}

	vars := map[string]any{"appId": args.AppID}
	if args.RiskState != nil {
		vars["riskState"] = *args.RiskState
	}
	if args.Status != nil {
		vars["status"] = *args.Status
	}
	if args.Domain != nil {
		vars["domain"] = *args.Domain
	}
	if args.Limit != nil {
		vars["limit"] = *args.Limit
	}
	if args.Offset != nil {
		vars["offset"] = *args.Offset
	}

	query := `query($appId: ID!, $riskState: RiskState, $status: AppSubscriptionStatus, $domain: String, $limit: Int, $offset: Int) {
		subscriptions(appId: $appId, riskState: $riskState, status: $status, domain: $domain, limit: $limit, offset: $offset) {
			nodes {
				id domain shopName plan status riskState mrrCents currency
				daysPastDue lastPaymentDate expectedNextCharge billingInterval
			}
			totalCount
			pageInfo { hasNextPage hasPreviousPage }
		}
	}`

	data, err := e.gql.Execute(ctx, query, vars)
	if err != nil {
		return errorResult(call.ID, err.Error())
	}
	return chat.ToolResult{ToolCallID: call.ID, Content: string(data)}
}

type detailArgs struct {
	SubscriptionID string `json:"subscription_id"`
}

func (e *executor) getSubscriptionDetail(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	var args detailArgs
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		return errorResult(call.ID, "invalid arguments: "+err.Error())
	}

	query := `query($id: ID!) {
		subscription(id: $id) {
			id domain shopName plan status riskState mrrCents currency
			daysPastDue lastPaymentDate expectedNextCharge billingInterval
			events {
				id eventType fromStatus toStatus fromRiskState toRiskState reason occurredAt
			}
		}
	}`

	data, err := e.gql.Execute(ctx, query, map[string]any{"id": args.SubscriptionID})
	if err != nil {
		return errorResult(call.ID, err.Error())
	}
	return chat.ToolResult{ToolCallID: call.ID, Content: string(data)}
}

type appIDArgs struct {
	AppID string `json:"app_id"`
}

func (e *executor) getSubscriptionSummary(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	var args appIDArgs
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		return errorResult(call.ID, "invalid arguments: "+err.Error())
	}

	query := `query($appId: ID!) {
		subscriptions(appId: $appId, limit: 1000) {
			nodes { status riskState mrrCents }
			totalCount
		}
	}`

	data, err := e.gql.Execute(ctx, query, map[string]any{"appId": args.AppID})
	if err != nil {
		return errorResult(call.ID, err.Error())
	}
	return chat.ToolResult{ToolCallID: call.ID, Content: string(data)}
}

type searchArgs struct {
	AppID string `json:"app_id"`
	Query string `json:"query"`
}

func (e *executor) searchSubscriptions(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	var args searchArgs
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		return errorResult(call.ID, "invalid arguments: "+err.Error())
	}

	query := `query($appId: ID!, $domain: String) {
		subscriptions(appId: $appId, domain: $domain) {
			nodes {
				id domain shopName plan status riskState mrrCents currency
				daysPastDue lastPaymentDate expectedNextCharge billingInterval
			}
			totalCount
		}
	}`

	data, err := e.gql.Execute(ctx, query, map[string]any{"appId": args.AppID, "domain": args.Query})
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
