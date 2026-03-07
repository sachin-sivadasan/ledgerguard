package risk

import (
	"encoding/json"

	"github.com/sachin-sivadasan/ledgerguard/internal/chat"
)

func toolDefinitions() []chat.ToolDefinition {
	return []chat.ToolDefinition{
		{
			Name:        "get_risk_summary",
			Description: "Get risk distribution for an app: counts of safe, one-cycle-missed, two-cycles-missed, and churned subscriptions, plus total revenue at risk in cents.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"app_id": {
						"type": "string",
						"description": "The app UUID to get risk summary for"
					}
				},
				"required": ["app_id"]
			}`),
		},
		{
			Name:        "list_at_risk",
			Description: "List subscriptions that are at risk (ONE_CYCLE_MISSED or TWO_CYCLES_MISSED). Optionally filter by specific risk state.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"app_id": {
						"type": "string",
						"description": "The app UUID"
					},
					"risk_state": {
						"type": "string",
						"enum": ["ONE_CYCLE_MISSED", "TWO_CYCLES_MISSED"],
						"description": "Filter by specific risk state. Omit to get all at-risk."
					}
				},
				"required": ["app_id"]
			}`),
		},
		{
			Name:        "get_risk_timeline",
			Description: "Get historical risk distribution over time from daily snapshots. Shows how safe, missed, and churned counts changed over months.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"app_id": {
						"type": "string",
						"description": "The app UUID"
					},
					"months": {
						"type": "integer",
						"description": "Number of months of history (default 6)"
					}
				},
				"required": ["app_id"]
			}`),
		},
	}
}
