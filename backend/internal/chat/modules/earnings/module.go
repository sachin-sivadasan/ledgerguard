package earnings

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

func (m *Module) Name() string        { return "earnings" }
func (m *Module) Description() string { return "Earnings data — revenue breakdown and transaction history" }

func (m *Module) PromptFragment() string {
	return `## Earnings Module
You can query earnings and transactions:
- Breakdown by charge type: RECURRING, USAGE, ONE_TIME, REFUND
- All amounts in cents (divide by 100 for dollars)
- Transactions include gross and net amounts (after Shopify fees)

Use get_breakdown for aggregate earnings by type, get_transactions for individual transaction records.
RECURRING = subscription fees, USAGE = usage-based charges, ONE_TIME = setup fees or add-ons, REFUND = negative adjustments.`
}

func (m *Module) Tools() []chat.ToolDefinition { return toolDefinitions() }
func (m *Module) ExecuteTool(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	return m.exec.execute(ctx, call)
}
