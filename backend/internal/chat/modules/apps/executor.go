package apps

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
	case "list_apps":
		return e.listApps(ctx, call)
	case "get_app_detail":
		return e.getAppDetail(ctx, call)
	default:
		return errorResult(call.ID, fmt.Sprintf("unknown apps tool: %s", call.Name))
	}
}

func (e *executor) listApps(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	query := `query {
		apps {
			id name partnerAppId trackingEnabled installCount
		}
	}`

	data, err := e.gql.Execute(ctx, query, nil)
	if err != nil {
		return errorResult(call.ID, err.Error())
	}
	return chat.ToolResult{ToolCallID: call.ID, Content: string(data)}
}

type appIDArgs struct {
	AppID string `json:"app_id"`
}

func (e *executor) getAppDetail(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	var args appIDArgs
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		return errorResult(call.ID, "invalid arguments: "+err.Error())
	}

	query := `query($id: ID!) {
		app(id: $id) {
			id name partnerAppId trackingEnabled installCount
		}
	}`

	data, err := e.gql.Execute(ctx, query, map[string]any{"id": args.AppID})
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
