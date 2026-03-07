package earnings

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
	case "get_breakdown":
		return e.getBreakdown(ctx, call)
	case "get_transactions":
		return e.getTransactions(ctx, call)
	default:
		return errorResult(call.ID, fmt.Sprintf("unknown earnings tool: %s", call.Name))
	}
}

type appIDArgs struct {
	AppID string `json:"app_id"`
}

func (e *executor) getBreakdown(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	var args appIDArgs
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		return errorResult(call.ID, "invalid arguments: "+err.Error())
	}

	query := `query($appId: ID!) {
		earnings(appId: $appId) {
			recurringCents usageCents oneTimeCents refundCents totalCents
		}
	}`

	data, err := e.gql.Execute(ctx, query, map[string]any{"appId": args.AppID})
	if err != nil {
		return errorResult(call.ID, err.Error())
	}
	return chat.ToolResult{ToolCallID: call.ID, Content: string(data)}
}

type transactionsArgs struct {
	AppID      string  `json:"app_id"`
	ChargeType *string `json:"charge_type,omitempty"`
	Domain     *string `json:"domain,omitempty"`
	StartDate  *string `json:"start_date,omitempty"`
	EndDate    *string `json:"end_date,omitempty"`
	Limit      *int    `json:"limit,omitempty"`
	Offset     *int    `json:"offset,omitempty"`
}

func (e *executor) getTransactions(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	var args transactionsArgs
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		return errorResult(call.ID, "invalid arguments: "+err.Error())
	}

	vars := map[string]any{"appId": args.AppID}
	if args.ChargeType != nil {
		vars["chargeType"] = *args.ChargeType
	}
	if args.Domain != nil {
		vars["domain"] = *args.Domain
	}
	if args.StartDate != nil {
		vars["startDate"] = *args.StartDate
	}
	if args.EndDate != nil {
		vars["endDate"] = *args.EndDate
	}
	if args.Limit != nil {
		vars["limit"] = *args.Limit
	}
	if args.Offset != nil {
		vars["offset"] = *args.Offset
	}

	query := `query($appId: ID!, $chargeType: ChargeType, $domain: String, $startDate: DateTime, $endDate: DateTime, $limit: Int, $offset: Int) {
		transactions(appId: $appId, chargeType: $chargeType, domain: $domain, startDate: $startDate, endDate: $endDate, limit: $limit, offset: $offset) {
			nodes {
				id domain shopName chargeType grossAmountCents netAmountCents currency transactionDate
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

func errorResult(toolCallID, msg string) chat.ToolResult {
	return chat.ToolResult{
		ToolCallID: toolCallID,
		Content:    fmt.Sprintf(`{"error": %q}`, msg),
		IsError:    true,
	}
}
