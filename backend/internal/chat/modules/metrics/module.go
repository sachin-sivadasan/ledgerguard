package metrics

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

func (m *Module) Name() string { return "metrics" }
func (m *Module) Description() string {
	return "Revenue metrics — MRR, trends, renewal rates, and aggregates"
}

func (m *Module) PromptFragment() string {
	return `## Metrics Module
You can query revenue metrics for apps:
- activeMrrCents: Monthly Recurring Revenue from SAFE subscriptions (in cents)
- revenueAtRiskCents: MRR from at-risk subscriptions
- renewalSuccessRate: Fraction of subscriptions that are SAFE (0.0–1.0)
- usageRevenueCents: Revenue from usage-based charges
- totalRevenueCents: All revenue types combined

Use get_latest_metrics for current snapshot, get_metrics_trend for historical data,
and get_aggregate_metrics for cross-app totals.
All monetary values are in cents — divide by 100 for display.`
}

func (m *Module) Tools() []chat.ToolDefinition { return toolDefinitions() }
func (m *Module) ExecuteTool(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	return m.exec.execute(ctx, call)
}
