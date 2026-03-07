package sync

import (
	"encoding/json"

	"github.com/sachin-sivadasan/ledgerguard/internal/chat"
)

func toolDefinitions() []chat.ToolDefinition {
	return []chat.ToolDefinition{
		{
			Name:        "trigger_sync",
			Description: "Trigger a data sync for an app. Fetches latest transactions from Shopify, rebuilds the ledger, and recalculates risk states. Runs in background — use get_sync_status to check progress.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"app_id": {"type": "string", "description": "The app UUID to sync. Omit to sync all apps."}
				}
			}`),
		},
		{
			Name:        "get_sync_status",
			Description: "Check the sync status and last sync time for an app.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"app_id": {"type": "string", "description": "The app UUID"}
				},
				"required": ["app_id"]
			}`),
		},
	}
}
