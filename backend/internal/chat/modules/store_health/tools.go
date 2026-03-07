package store_health

import (
	"encoding/json"

	"github.com/sachin-sivadasan/ledgerguard/internal/chat"
)

func toolDefinitions() []chat.ToolDefinition {
	return []chat.ToolDefinition{
		{
			Name:        "check_store",
			Description: "Check the payment health of a specific store by its myshopify.com domain. Returns risk state, last payment, days past due, and subscription details.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"app_id": {"type": "string", "description": "The app UUID"},
					"domain": {"type": "string", "description": "The myshopify.com domain (e.g. acme.myshopify.com)"}
				},
				"required": ["app_id", "domain"]
			}`),
		},
		{
			Name:        "compare_stores",
			Description: "Compare payment health across multiple stores side by side. Useful for spotting patterns. Max 10 domains.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"app_id": {"type": "string", "description": "The app UUID"},
					"domains": {
						"type": "array",
						"items": {"type": "string"},
						"description": "List of myshopify.com domains to compare (max 10)"
					}
				},
				"required": ["app_id", "domains"]
			}`),
		},
	}
}
