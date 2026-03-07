package risk

import (
	"context"

	"github.com/sachin-sivadasan/ledgerguard/internal/chat"
)

// Module implements the chat.Module interface for risk analysis.
type Module struct {
	exec *executor
}

// New creates a risk module that executes tools via the GraphQL executor.
func New(gql *chat.GraphQLExecutor) *Module {
	return &Module{exec: &executor{gql: gql}}
}

func (m *Module) Name() string        { return "risk" }
func (m *Module) Description() string { return "Risk analysis — payment risk states, at-risk stores, and risk trends" }

func (m *Module) PromptFragment() string {
	return `## Risk Module
You can analyze subscription payment risk. Risk states:
- SAFE: Payment on schedule
- ONE_CYCLE_MISSED: 31–60 days past expected charge date
- TWO_CYCLES_MISSED: 61–90 days past expected charge date
- CHURNED: 90+ days past expected charge date

Use get_risk_summary for an overview, list_at_risk to see specific stores, and get_risk_timeline for historical trends.
Revenue at risk is the MRR from ONE_CYCLE_MISSED + TWO_CYCLES_MISSED subscriptions (not CHURNED — those are already lost).`
}

func (m *Module) Tools() []chat.ToolDefinition {
	return toolDefinitions()
}

func (m *Module) ExecuteTool(ctx context.Context, call chat.ToolCall) chat.ToolResult {
	return m.exec.execute(ctx, call)
}
