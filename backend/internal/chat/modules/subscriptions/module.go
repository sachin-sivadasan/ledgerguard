package subscriptions

import (
	"context"

	"github.com/sachin-sivadasan/ledgerguard/internal/chat"
)

// Module implements chat.Module for subscription queries.
type Module struct {
	exec *executor
}

func New(gql *chat.GraphQLExecutor) *Module {
	return &Module{exec: &executor{gql: gql}}
}

func (m *Module) Name() string        { return "subscriptions" }
func (m *Module) Description() string { return "Subscription management — list, search, detail, and summary" }

func (m *Module) PromptFragment() string {
	return `## Subscriptions Module
You can query app subscriptions. Each subscription has:
- domain: the myshopify.com domain
- shopName: human-readable store name
- status: ACTIVE, CANCELLED, FROZEN, PENDING, UNINSTALLED
- riskState: SAFE, ONE_CYCLE_MISSED, TWO_CYCLES_MISSED, CHURNED
- mrrCents: monthly recurring revenue in cents
- billingInterval: MONTHLY or ANNUAL
- daysPastDue: days since expected charge date (null if not past due)

Use list_subscriptions to browse with filters, search_subscriptions for domain/name search,
get_subscription_detail for full info including event history, and get_subscription_summary for aggregate stats.`
}

func (m *Module) Tools() []chat.ToolDefinition           { return toolDefinitions() }
func (m *Module) ExecuteTool(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	return m.exec.execute(ctx, call)
}
