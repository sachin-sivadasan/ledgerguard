package chat

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const maxToolIterations = 5

// Handler orchestrates the AI chat: receives messages, runs the tool call loop,
// and streams results back via SSE.
type Handler struct {
	registry    *Registry
	aiProviders *AIProviderRegistry
}

// NewHandler creates a chat handler.
func NewHandler(registry *Registry, aiProviders *AIProviderRegistry) *Handler {
	return &Handler{
		registry:    registry,
		aiProviders: aiProviders,
	}
}

// --- SSE Event Types ---

// SSEEvent is sent to the client as a server-sent event.
type SSEEvent struct {
	Type string `json:"type"` // "tool_call", "tool_result", "response", "error"
	Data any    `json:"data"`
}

type toolCallEvent struct {
	Iteration int    `json:"iteration"`
	ToolName  string `json:"tool_name"`
	Arguments string `json:"arguments"`
}

type toolResultEvent struct {
	ToolName string `json:"tool_name"`
	Content  string `json:"content"`
	IsError  bool   `json:"is_error"`
}

type responseEvent struct {
	Message           string           `json:"message"`
	SubscriptionState json.RawMessage  `json:"subscription_state,omitempty"`
	MetricsState      json.RawMessage  `json:"metrics_state,omitempty"`
	RiskState         json.RawMessage  `json:"risk_state,omitempty"`
	StoreHealthState  json.RawMessage  `json:"store_health_state,omitempty"`
	EarningsState     json.RawMessage  `json:"earnings_state,omitempty"`
	Suggestions       []string         `json:"suggestions,omitempty"`
	Usage             TokenUsage       `json:"usage"`
}

// --- Request / Response ---

// ChatRequest is the JSON body for POST /api/v1/chat.
type ChatRequest struct {
	Messages     []ChatMessage `json:"messages"`
	ScopedModule string        `json:"scoped_module,omitempty"` // optional @module filter
	AppID        string        `json:"app_id,omitempty"`        // active app context
}

// HandleChat handles POST /api/v1/chat with SSE streaming.
func (h *Handler) HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Messages) == 0 {
		http.Error(w, "messages required", http.StatusBadRequest)
		return
	}

	// Get AI provider (default)
	aiClient, err := h.aiProviders.Default()
	if err != nil {
		http.Error(w, "AI provider not configured", http.StatusServiceUnavailable)
		return
	}

	// Set up SSE
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering

	ctx := r.Context()

	// Build system prompt
	systemPrompt := h.buildSystemPrompt(req.ScopedModule, req.AppID)

	// Get tools (all or scoped)
	var tools []ToolDefinition
	if req.ScopedModule != "" {
		tools = h.registry.ListModuleTools(req.ScopedModule)
	} else {
		tools = h.registry.ListAllTools()
	}

	// Run the tool call loop
	messages := make([]ChatMessage, len(req.Messages))
	copy(messages, req.Messages)

	var allToolRecords []toolRecord
	var totalUsage TokenUsage

	for iteration := 0; iteration < maxToolIterations; iteration++ {
		resp, err := aiClient.ChatCompletion(ctx, ChatCompletionRequest{
			SystemPrompt: systemPrompt,
			Messages:     messages,
			Tools:        tools,
		})
		if err != nil {
			writeSSE(w, flusher, SSEEvent{Type: "error", Data: map[string]string{"message": err.Error()}})
			return
		}

		totalUsage.PromptTokens += resp.Usage.PromptTokens
		totalUsage.CompletionTokens += resp.Usage.CompletionTokens
		totalUsage.TotalTokens += resp.Usage.TotalTokens

		// No tool calls → final text response
		if len(resp.ToolCalls) == 0 {
			state := extractState(allToolRecords)
			writeSSE(w, flusher, SSEEvent{
				Type: "response",
				Data: responseEvent{
					Message:           resp.Content,
					SubscriptionState: state["subscriptions"],
					MetricsState:      state["metrics"],
					RiskState:         state["risk"],
					StoreHealthState:  state["store_health"],
					EarningsState:     state["earnings"],
					Suggestions:       generateSuggestions(allToolRecords),
					Usage:             totalUsage,
				},
			})
			return
		}

		// Process tool calls
		assistantMsg := ChatMessage{
			Role:    "assistant",
			Content: resp.Content,
		}
		for _, tc := range resp.ToolCalls {
			assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, ToolCall{
				ID:        tc.ID,
				Name:      tc.Name,
				Arguments: tc.Arguments,
			})
		}
		messages = append(messages, assistantMsg)

		for _, tc := range resp.ToolCalls {
			// Stream tool_call event
			writeSSE(w, flusher, SSEEvent{
				Type: "tool_call",
				Data: toolCallEvent{
					Iteration: iteration + 1,
					ToolName:  tc.Name,
					Arguments: tc.Arguments,
				},
			})

			// Execute tool
			call := ToolCall{ID: tc.ID, Name: tc.Name, Arguments: tc.Arguments}
			result := h.registry.RouteToolCall(ctx, call)

			// Determine module name from tool name
			moduleName := ""
			if parts := strings.SplitN(tc.Name, "__", 2); len(parts) == 2 {
				moduleName = parts[0]
			}

			allToolRecords = append(allToolRecords, toolRecord{
				Module: moduleName,
				Name:   tc.Name,
				Result: result.Content,
				Error:  result.IsError,
			})

			// Stream tool_result event
			writeSSE(w, flusher, SSEEvent{
				Type: "tool_result",
				Data: toolResultEvent{
					ToolName: tc.Name,
					Content:  result.Content,
					IsError:  result.IsError,
				},
			})

			// Append tool result message for next iteration
			messages = append(messages, ChatMessage{
				Role: "tool",
				ToolResult: &ToolResult{
					ToolCallID: tc.ID,
					Content:    result.Content,
					IsError:    result.IsError,
				},
			})
		}
	}

	// Max iterations reached — return what we have
	state := extractState(allToolRecords)
	writeSSE(w, flusher, SSEEvent{
		Type: "response",
		Data: responseEvent{
			Message:           "I've gathered the data above. Let me know if you need anything else.",
			SubscriptionState: state["subscriptions"],
			MetricsState:      state["metrics"],
			RiskState:         state["risk"],
			StoreHealthState:  state["store_health"],
			EarningsState:     state["earnings"],
			Usage:             totalUsage,
		},
	})
}

// HandleListModules handles GET /api/v1/chat/modules.
func (h *Handler) HandleListModules(w http.ResponseWriter, r *http.Request) {
	type moduleInfo struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		ToolCount   int    `json:"tool_count"`
	}

	modules := h.registry.Modules()
	result := make([]moduleInfo, 0, len(modules))
	for _, m := range modules {
		result = append(result, moduleInfo{
			Name:        m.Name(),
			Description: m.Description(),
			ToolCount:   len(m.Tools()),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) buildSystemPrompt(scopedModule, appID string) string {
	base := `You are LedgerGuard's Revenue Intelligence Assistant. You help Shopify app developers
understand their subscription revenue, identify at-risk stores, and track earnings.

RULES:
- You ONLY answer questions related to Shopify app revenue, subscriptions, metrics, risk, store health, earnings, and app management
- If the user asks a general knowledge question, off-topic question, or anything unrelated to their Shopify app data, politely decline and suggest they ask about their app metrics instead
- Never ask a numbered checklist of questions
- Infer as much as possible from context (e.g., if user says "my MRR" → use their active app)
- Ask only ONE clarifying question at a time
- When you have enough info, query immediately — don't ask for permission
- Always include specific numbers and store names in your responses
- Format data as tables when showing multiple items
- Suggest 2-3 follow-up actions after each answer
- When showing monetary values, convert cents to dollars (divide by 100)
- Risk states: SAFE (green), ONE_CYCLE_MISSED (yellow), TWO_CYCLES_MISSED (red), CHURNED (dark)`

	moduleFragments := h.registry.BuildSystemPrompt(scopedModule)
	prompt := base
	if moduleFragments != "" {
		prompt += "\n\n" + moduleFragments
	}

	if appID != "" {
		prompt += fmt.Sprintf("\n\nThe user's active app ID is: %s. Use this for queries unless they specify otherwise.", appID)
	}

	return prompt
}

// extractState scans tool records in reverse to find the latest state per module.
func extractState(records []toolRecord) map[string]json.RawMessage {
	state := make(map[string]json.RawMessage)
	for i := len(records) - 1; i >= 0; i-- {
		rec := records[i]
		if rec.Error || rec.Module == "" {
			continue
		}
		if _, exists := state[rec.Module]; !exists {
			state[rec.Module] = json.RawMessage(rec.Result)
		}
	}
	return state
}

type toolRecord struct {
	Module string
	Name   string
	Result string
	Error  bool
}

// generateSuggestions returns context-aware follow-up suggestions based on tool records.
func generateSuggestions(records []toolRecord) []string {
	if len(records) == 0 {
		return nil
	}

	// Use the last successful tool to determine suggestions
	for i := len(records) - 1; i >= 0; i-- {
		if records[i].Error {
			continue
		}
		switch {
		case strings.Contains(records[i].Name, "risk_summary"):
			return []string{"Show at-risk stores", "What's the MRR impact?", "Any churned this month?"}
		case strings.Contains(records[i].Name, "list_at_risk"):
			return []string{"Check a specific store", "Show risk timeline", "Get earnings breakdown"}
		case strings.Contains(records[i].Name, "list_subscriptions"):
			return []string{"Filter by at-risk only", "Show subscription summary", "Check store health"}
		case strings.Contains(records[i].Name, "latest_metrics"):
			return []string{"Show trend over 6 months", "Compare to last month", "Show risk summary"}
		case strings.Contains(records[i].Name, "metrics_trend"):
			return []string{"Show current metrics", "Show risk breakdown", "Get earnings"}
		case strings.Contains(records[i].Name, "store_health") || strings.Contains(records[i].Name, "check_store"):
			return []string{"Show subscription history", "Check other stores", "Get risk summary"}
		case strings.Contains(records[i].Name, "earnings") || strings.Contains(records[i].Name, "breakdown"):
			return []string{"Which type grew most?", "Show monthly trend", "Show risk summary"}
		case strings.Contains(records[i].Name, "trigger_sync"):
			return []string{"Check sync status", "Show updated metrics"}
		}
	}
	return nil
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, event SSEEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("SSE marshal error: %v", err)
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}
