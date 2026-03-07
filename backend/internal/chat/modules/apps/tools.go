package apps

import (
	"encoding/json"

	"github.com/sachin-sivadasan/ledgerguard/internal/chat"
)

func toolDefinitions() []chat.ToolDefinition {
	return []chat.ToolDefinition{
		{
			Name:        "list_apps",
			Description: "List all tracked Shopify apps for the current user. Returns app IDs and names. Call this when you need to find the app ID.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {}
			}`),
		},
		{
			Name:        "get_app_detail",
			Description: "Get detailed information about a specific app including install count and tracking status.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"app_id": {
						"type": "string",
						"description": "The app ID to look up"
					}
				},
				"required": ["app_id"]
			}`),
		},
	}
}
