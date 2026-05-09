package metrics

import (
	"encoding/json"

	"github.com/sachin-sivadasan/ledgerguard/internal/chat"
)

func toolDefinitions() []chat.ToolDefinition {
	return []chat.ToolDefinition{
		{
			Name:        "get_latest_metrics",
			Description: "Get the latest metrics snapshot for an app: active MRR, revenue at risk, renewal rate, usage revenue, and subscription counts by risk state.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"app_id": {"type": "string", "description": "The app UUID"}
				},
				"required": ["app_id"]
			}`),
		},
		{
			Name:        "get_metrics_trend",
			Description: "Get historical metrics trend from daily snapshots over a period of months. Supports granularity downsampling (daily/weekly/monthly).",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"app_id": {"type": "string", "description": "The app UUID"},
					"months": {"type": "integer", "description": "Number of months of history (default 6)"},
					"granularity": {"type": "string", "enum": ["DAILY", "WEEKLY", "MONTHLY"], "description": "Time granularity (default DAILY)"}
				},
				"required": ["app_id"]
			}`),
		},
		{
			Name:        "get_aggregate_metrics",
			Description: "Get aggregate metrics across all apps for the current user.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
		},
	}
}
