package sync

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sachin-sivadasan/ledgerguard/internal/chat"
)

// SyncTrigger abstracts the sync service so the module doesn't import application layer.
type SyncTrigger interface {
	TriggerSync(ctx context.Context, appID string) error
	TriggerSyncAll(ctx context.Context) error
}

type executor struct {
	gql     *chat.GraphQLExecutor
	trigger SyncTrigger
}

func (e *executor) execute(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	switch call.Name {
	case "trigger_sync":
		return e.triggerSync(ctx, call)
	case "get_sync_status":
		return e.getSyncStatus(ctx, call)
	default:
		return errorResult(call.ID, fmt.Sprintf("unknown sync tool: %s", call.Name))
	}
}

type triggerSyncArgs struct {
	AppID *string `json:"app_id,omitempty"`
}

func (e *executor) triggerSync(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	var args triggerSyncArgs
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		return errorResult(call.ID, "invalid arguments: "+err.Error())
	}

	if e.trigger == nil {
		return errorResult(call.ID, "sync service not available")
	}

	// Fire and forget — sync runs in background
	if args.AppID != nil {
		go func() {
			_ = e.trigger.TriggerSync(context.Background(), *args.AppID)
		}()
		return chat.ToolResult{
			ToolCallID: call.ID,
			Content:    fmt.Sprintf(`{"status": "sync_started", "message": "Sync triggered for app %s. Use get_sync_status to check progress."}`, *args.AppID),
		}
	}

	go func() {
		_ = e.trigger.TriggerSyncAll(context.Background())
	}()
	return chat.ToolResult{
		ToolCallID: call.ID,
		Content:    `{"status": "sync_started", "message": "Sync triggered for all apps. Use get_sync_status to check progress."}`,
	}
}

type statusArgs struct {
	AppID string `json:"app_id"`
}

func (e *executor) getSyncStatus(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	var args statusArgs
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		return errorResult(call.ID, "invalid arguments: "+err.Error())
	}

	// Get app details which include tracking status
	query := `query($id: ID!) {
		app(id: $id) {
			id name trackingEnabled installCount
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
