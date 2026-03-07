package apps

import (
	"context"

	"github.com/sachin-sivadasan/ledgerguard/internal/chat"
)

type Module struct {
	exec *executor
}

func New(gql *chat.GraphQLExecutor) *Module {
	return &Module{exec: &executor{gql: gql}}
}

func (m *Module) Name() string        { return "apps" }
func (m *Module) Description() string { return "App discovery — list and inspect tracked Shopify apps" }

func (m *Module) PromptFragment() string {
	return `## Apps Module
You can discover the user's tracked Shopify apps:
- list_apps: Returns all tracked apps with IDs and names. Use this when the user doesn't specify an app or you need to find the right app ID.
- get_app_detail: Returns detailed info for a specific app.

IMPORTANT: If no app_id is provided in the context, call list_apps first to find it.
If the user has only one app, use it automatically without asking.`
}

func (m *Module) Tools() []chat.ToolDefinition { return toolDefinitions() }
func (m *Module) ExecuteTool(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	return m.exec.execute(ctx, call)
}
