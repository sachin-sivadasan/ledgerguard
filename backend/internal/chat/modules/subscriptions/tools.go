package subscriptions

import (
	"encoding/json"

	"github.com/sachin-sivadasan/ledgerguard/internal/chat"
)

func toolDefinitions() []chat.ToolDefinition {
	return []chat.ToolDefinition{
		{
			Name:        "list_subscriptions",
			Description: "List subscriptions for an app with optional filters for risk state, status, and domain search. Returns paginated results with risk summary.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"app_id": {"type": "string", "description": "The app UUID"},
					"risk_state": {"type": "string", "enum": ["SAFE", "ONE_CYCLE_MISSED", "TWO_CYCLES_MISSED", "CHURNED"], "description": "Filter by risk state"},
					"status": {"type": "string", "enum": ["ACTIVE", "CANCELLED", "FROZEN", "PENDING", "UNINSTALLED"], "description": "Filter by subscription status"},
					"domain": {"type": "string", "description": "Search by store domain"},
					"limit": {"type": "integer", "description": "Page size (default 50)"},
					"offset": {"type": "integer", "description": "Offset for pagination (default 0)"}
				},
				"required": ["app_id"]
			}`),
		},
		{
			Name:        "get_subscription_detail",
			Description: "Get full details of a single subscription including event history.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"subscription_id": {"type": "string", "description": "The subscription UUID"}
				},
				"required": ["subscription_id"]
			}`),
		},
		{
			Name:        "get_subscription_summary",
			Description: "Get aggregate summary of all subscriptions for an app: counts by status and risk state, average MRR.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"app_id": {"type": "string", "description": "The app UUID"}
				},
				"required": ["app_id"]
			}`),
		},
		{
			Name:        "search_subscriptions",
			Description: "Search subscriptions by store domain or shop name.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"app_id": {"type": "string", "description": "The app UUID"},
					"query": {"type": "string", "description": "Search term (domain or shop name)"}
				},
				"required": ["app_id", "query"]
			}`),
		},
	}
}
