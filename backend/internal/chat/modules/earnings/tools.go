package earnings

import (
	"encoding/json"

	"github.com/sachin-sivadasan/ledgerguard/internal/chat"
)

func toolDefinitions() []chat.ToolDefinition {
	return []chat.ToolDefinition{
		{
			Name:        "get_breakdown",
			Description: "Get earnings breakdown by charge type: recurring, usage, one-time, and refunds. All values in cents.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"app_id": {"type": "string", "description": "The app UUID"}
				},
				"required": ["app_id"]
			}`),
		},
		{
			Name:        "get_transactions",
			Description: "List individual transactions with optional filters for charge type, domain, and date range.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"app_id": {"type": "string", "description": "The app UUID"},
					"charge_type": {"type": "string", "enum": ["RECURRING", "USAGE", "ONE_TIME", "REFUND"], "description": "Filter by charge type"},
					"domain": {"type": "string", "description": "Filter by store domain"},
					"start_date": {"type": "string", "description": "Start date (ISO 8601)"},
					"end_date": {"type": "string", "description": "End date (ISO 8601)"},
					"limit": {"type": "integer", "description": "Page size (default 50)"},
					"offset": {"type": "integer", "description": "Offset for pagination (default 0)"}
				},
				"required": ["app_id"]
			}`),
		},
	}
}
