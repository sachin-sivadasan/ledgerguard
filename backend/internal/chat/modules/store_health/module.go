package store_health

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

func (m *Module) Name() string { return "store_health" }
func (m *Module) Description() string {
	return "Store health — check payment status for individual stores"
}

func (m *Module) PromptFragment() string {
	return `## Store Health Module
You can check the payment health of individual stores:
- isPaid: whether the store's subscription is current
- riskState: the store's current risk classification
- lastPaymentDate: when the store last paid
- daysPastDue: how many days since expected charge date
- subscription: full subscription details

Use check_store for a single store, compare_stores to compare multiple stores side by side.`
}

func (m *Module) Tools() []chat.ToolDefinition { return toolDefinitions() }
func (m *Module) ExecuteTool(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	return m.exec.execute(ctx, call)
}
