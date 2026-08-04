package sync

import (
	"context"

	"github.com/sachin-sivadasan/ledgerguard/internal/chat"
)

type Module struct {
	exec *executor
}

// New creates a sync module. trigger can be nil if sync service is unavailable.
func New(gql *chat.GraphQLExecutor, trigger SyncTrigger) *Module {
	return &Module{exec: &executor{gql: gql, trigger: trigger}}
}

func (m *Module) Name() string { return "sync" }
func (m *Module) Description() string {
	return "Data sync — trigger Shopify data refresh and check status"
}

func (m *Module) PromptFragment() string {
	return `## Sync Module
You can trigger data synchronization with Shopify:
- trigger_sync: Starts a background sync (fetches latest transactions, rebuilds ledger, recalculates risk)
- get_sync_status: Check app status and tracking info

Sync runs in the background — it does NOT block. After triggering, suggest the user check status
or look at updated metrics after a minute.`
}

func (m *Module) Tools() []chat.ToolDefinition { return toolDefinitions() }
func (m *Module) ExecuteTool(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	return m.exec.execute(ctx, call)
}
