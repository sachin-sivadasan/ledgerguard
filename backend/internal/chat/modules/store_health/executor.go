package store_health

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
	case "check_store":
		return e.checkStore(ctx, call)
	case "compare_stores":
		return e.compareStores(ctx, call)
	default:
		return errorResult(call.ID, fmt.Sprintf("unknown store_health tool: %s", call.Name))
	}
}

type checkStoreArgs struct {
	AppID  string `json:"app_id"`
	Domain string `json:"domain"`
}

func (e *executor) checkStore(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	var args checkStoreArgs
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		return errorResult(call.ID, "invalid arguments: "+err.Error())
	}

	query := `query($appId: ID!, $domain: String!) {
		storeHealth(appId: $appId, domain: $domain) {
			domain shopName isPaid riskState lastPaymentDate daysPastDue
			subscription {
				id plan status mrrCents currency billingInterval
				lastPaymentDate expectedNextCharge
			}
		}
	}`

	data, err := e.gql.Execute(ctx, query, map[string]any{"appId": args.AppID, "domain": args.Domain})
	if err != nil {
		return errorResult(call.ID, err.Error())
	}
	return chat.ToolResult{ToolCallID: call.ID, Content: string(data)}
}

type compareStoresArgs struct {
	AppID   string   `json:"app_id"`
	Domains []string `json:"domains"`
}

func (e *executor) compareStores(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	var args compareStoresArgs
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		return errorResult(call.ID, "invalid arguments: "+err.Error())
	}

	if len(args.Domains) > 10 {
		args.Domains = args.Domains[:10]
	}

	type storeResult struct {
		Domain string `json:"domain"`
		Data   any    `json:"data,omitempty"`
		Error  string `json:"error,omitempty"`
	}

	results := make([]storeResult, 0, len(args.Domains))
	for _, domain := range args.Domains {
		query := `query($appId: ID!, $domain: String!) {
			storeHealth(appId: $appId, domain: $domain) {
				domain shopName isPaid riskState lastPaymentDate daysPastDue
				subscription { mrrCents billingInterval }
			}
		}`

		data, err := e.gql.Execute(ctx, query, map[string]any{"appId": args.AppID, "domain": domain})
		if err != nil {
			results = append(results, storeResult{Domain: domain, Error: err.Error()})
			continue
		}
		var parsed struct {
			StoreHealth any `json:"storeHealth"`
		}
		json.Unmarshal(data, &parsed)
		results = append(results, storeResult{Domain: domain, Data: parsed.StoreHealth})
	}

	out, _ := json.Marshal(map[string]any{"stores": results})
	return chat.ToolResult{ToolCallID: call.ID, Content: string(out)}
}

func errorResult(toolCallID, msg string) chat.ToolResult {
	return chat.ToolResult{
		ToolCallID: toolCallID,
		Content:    fmt.Sprintf(`{"error": %q}`, msg),
		IsError:    true,
	}
}
