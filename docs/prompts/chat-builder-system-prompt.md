# LedgerGuard AI Chat Builder — Implementation Prompt

> **Purpose:** Paste this prompt into an AI coding assistant to implement the AI Chat Assistant for LedgerGuard. Adapted from OpenAI Explorer's chat builder — same architecture patterns, different domain. Covers module system, tool call loop, GraphQL integration, Flutter frontend, and all gotchas.

---

## SYSTEM CONTEXT

You are building an **AI-powered revenue data assistant** for LedgerGuard — a conversational interface where Shopify app developers query their revenue data (subscriptions, MRR, risk states, store health, earnings) entirely through natural language. The AI calls backend tools via **OpenAI function calling** (gpt-4o), and the frontend shows live data previews (tables, risk cards, trend charts).

This is NOT a simple chatbot. It is a **tool-orchestrated system** with:
- Plugin architecture (add query modules without touching the core)
- Multi-step tool call loops (AI reasons, calls tools, sees results, reasons again)
- Live state extraction (Flutter updates in real-time from tool results)
- GraphQL as the internal query execution layer (modules resolve through gqlgen)
- Firebase Auth for access control (tenant isolation — users query only their own data)
- **Provider-agnostic AI interface** — ships with OpenAI, designed for easy Claude API swap later

---

## ARCHITECTURE OVERVIEW

```
┌─────────── FLUTTER FRONTEND ──────────┐    ┌─────────── GO BACKEND ──────────────┐
│                                        │    │                                      │
│  ChatPane (messages + input)           │───>│  WebSocket /api/v1/chat              │
│  ├── @module autocomplete              │    │  ├── Firebase Auth middleware         │
│  ├── Quote selection                   │    │  ├── Chat Handler                    │
│  └── Context pick chips               │    │  │   ├── Build system prompt          │
│                                        │    │  │   ├── Call Claude API              │
│  DataPanel (live preview)              │<───│  │   ├── Tool call loop (max 5)      │
│  ├── SubscriptionTable                 │    │  │   ├── State extraction             │
│  ├── MetricsCard                       │    │  │   └── Suggestive replies           │
│  ├── RiskBreakdown                     │    │  │                                    │
│  ├── StoreHealthCard                   │    │  └── Module Registry                  │
│  └── EarningsChart                     │    │      ├── Subscriptions Module (4 tools)│
│                                        │    │      ├── Metrics Module (3 tools)      │
│  ChatBloc (state management)           │    │      ├── Risk Module (3 tools)         │
│                                        │    │      ├── Store Health Module (2 tools)  │
│                                        │    │      ├── Earnings Module (2 tools)      │
│                                        │    │      └── Sync Module (2 tools)          │
│                                        │    │                                      │
│                                        │    │  GraphQL Layer (gqlgen) ← resolvers  │
│                                        │    │  AIClient interface ─┐               │
│                                        │    │    ├─ OpenAIClient   │ ← DEFAULT     │
│                                        │    │    └─ ClaudeClient   │ (future)      │
│                                        │    │  PostgreSQL (domain data)             │
└────────────────────────────────────────┘    └──────────────────────────────────────┘
```

### How GraphQL Fits In

The GraphQL layer is **internal infrastructure**, not a separate endpoint (Phase 1). Modules execute their tools by running GraphQL queries against the gqlgen resolvers, which call existing domain services (RiskEngine, MetricsEngine, LedgerService). This means:

- Modules don't talk to the database directly
- Modules don't import domain services directly
- Modules call `graphqlExecutor.Execute(query, variables)` → get structured data
- The GraphQL schema is also injected into the AI system prompt so it understands what's queryable

### Provider-Agnostic AI Interface

The chat handler talks to an `AIClient` interface, not a specific provider. Ships with OpenAI, designed for Claude swap later.

```go
// AIClient abstracts the LLM provider — swap OpenAI for Claude without touching chat handler
type AIClient interface {
    // ChatCompletion sends messages + tools, returns response (may include tool calls)
    ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error)
}

type ChatCompletionRequest struct {
    Model        string
    SystemPrompt string
    Messages     []ChatMessage
    Tools        []ToolDefinition  // provider-agnostic tool schema
}

type ChatCompletionResponse struct {
    Content   string            // text response (empty if tool calls)
    ToolCalls []ToolCallRequest  // tool calls to execute (empty if text response)
    Usage     TokenUsage
}

type ToolCallRequest struct {
    ID        string  // provider assigns this (needed for result pairing)
    Name      string  // e.g., "risk__get_risk_summary"
    Arguments string  // JSON string of arguments
}
```

**OpenAI Implementation (default — ships first):**
```go
type OpenAIClient struct {
    client *openai.Client
    model  string  // "gpt-4o"
}

func (c *OpenAIClient) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
    // Convert ToolDefinition → openai.FunctionDefinition
    // Convert ChatMessage → openai.ChatCompletionMessage
    // Call client.CreateChatCompletion()
    // Convert openai.ToolCall → ToolCallRequest
    // Return unified response
}
```

**Claude Implementation (future):**
```go
type ClaudeClient struct {
    client *anthropic.Client
    model  string  // "claude-sonnet-4-6"
}

func (c *ClaudeClient) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
    // Convert ToolDefinition → anthropic.ToolDefinition (input_schema)
    // Convert ChatMessage → anthropic.MessageParam
    // Call client.Messages.Create() with tool_use
    // Convert anthropic tool_use blocks → ToolCallRequest
    // Return unified response
}
```

**AIProviderRegistry (per-user provider selection):**
```go
type AIProviderRegistry struct {
    mu        sync.RWMutex
    providers map[string]AIClient  // "openai" → OpenAIClient, "claude" → ClaudeClient
    fallback  string               // default provider if user has no preference
}

func (r *AIProviderRegistry) Register(name string, client AIClient)
func (r *AIProviderRegistry) Get(providerName string) (AIClient, error)

// Chat handler resolves per request:
func (h *Handler) handleMessage(ctx context.Context, user *User, msg ChatMessage) {
    providerName := user.Preferences.AIProvider  // "openai" or "claude"
    if providerName == "" {
        providerName = "openai"  // default
    }
    aiClient, _ := h.providers.Get(providerName)
    // ... use aiClient for this conversation
}
```

**Why this works:** OpenAI's function calling and Claude's tool use are structurally identical:
- Both: system prompt + messages + tool definitions → response with optional tool calls
- Both: tool results feed back as messages for the next iteration
- Only the wire format differs (JSON schema shape, message role names, tool result format)
- The `AIClient` interface hides these differences
- Both providers run simultaneously — each user picks their preferred one

```
Module.ExecuteTool()
    → Build GraphQL query from tool arguments
    → graphqlExecutor.Execute(query, vars)
    → gqlgen resolver
    → Domain Service (RiskEngine, MetricsEngine, etc.)
    → PostgreSQL
    → Return structured result
```

---

## PATTERN 1: MODULE INTERFACE (Plugin Architecture)

### The Interface

Every feature is a **Module** that registers itself. The chat handler and registry are generic — they never change when you add features.

```go
type Module interface {
    Name() string              // unique identifier: "subscriptions", "metrics", "risk"
    Description() string       // human-readable, shown in @module autocomplete
    PromptFragment() string    // AI instructions appended to system prompt
    Tools() []ToolDefinition   // OpenAI function calling schemas (provider-agnostic via AIClient)
    ExecuteTool(ctx context.Context, call ToolCall) ToolResult
}
```

### Why This Matters
- **Zero changes to core** when adding a new query module
- Each module is self-contained: tools.go, executor.go, module.go
- New modules auto-appear in @module autocomplete, system prompt, and tool list
- Modules don't import each other — they all go through GraphQL

### Module Folder Structure
```
internal/chat/
├── module.go           — Module interface definition
├── registry.go         — Registry: register, route, list, filter
├── handler.go          — WebSocket chat handler with tool loop
├── types.go            — ToolDefinition, ToolCall, ToolResult, ChatMessage
└── modules/
    ├── subscriptions/
    │   ├── tools.go    — OpenAI function schemas (list, detail, summary, search)
    │   ├── executor.go — Switch on tool name → build GraphQL query → execute
    │   └── module.go   — Implements Module interface
    ├── metrics/
    │   ├── tools.go    — (latest, trend, aggregate)
    │   ├── executor.go
    │   └── module.go
    ├── risk/
    │   ├── tools.go    — (summary, at_risk_list, risk_timeline)
    │   ├── executor.go
    │   └── module.go
    ├── store_health/
    │   ├── tools.go    — (health_check, store_compare)
    │   ├── executor.go
    │   └── module.go
    ├── earnings/
    │   ├── tools.go    — (breakdown, status)
    │   ├── executor.go
    │   └── module.go
    └── sync/
        ├── tools.go    — (trigger_sync, sync_status)
        ├── executor.go
        └── module.go
```

---

## PATTERN 2: REGISTRY (Tool Routing)

```go
type Registry struct {
    mu      sync.RWMutex
    modules []Module
}

func (r *Registry) Register(m Module)
func (r *Registry) ListAllTools() []ToolDefinition
func (r *Registry) ListModuleTools(name string) []ToolDefinition
func (r *Registry) RouteToolCall(ctx context.Context, call ToolCall) ToolResult
func (r *Registry) BuildSystemPrompt(scoped string) string
```

### CRITICAL: Tool Naming Convention

**Use double underscore (`__`) as separator: `module__tool_name`**

```
subscriptions__list_subscriptions
subscriptions__get_subscription_detail
metrics__get_latest_metrics
metrics__get_metrics_trend
risk__get_risk_summary
risk__list_at_risk
store_health__check_store
earnings__get_breakdown
sync__trigger_sync
```

**WHY `__`?** OpenAI function names must match `^[a-zA-Z0-9_-]+$`. Dots cause HTTP 400. Module names and tool names use single underscores, so `__` is unambiguous for routing. (Same constraint applies to Claude tool names.)

**Routing logic:**
```go
for _, m := range r.modules {
    prefix := m.Name() + "__"
    if strings.HasPrefix(call.Name, prefix) {
        localCall := ToolCall{
            Name:      strings.TrimPrefix(call.Name, prefix),
            Arguments: call.Arguments,
        }
        return m.ExecuteTool(ctx, localCall)
    }
}
```

---

## PATTERN 3: CHAT HANDLER (Tool Call Loop)

The core orchestration. A WebSocket handler that:
1. Builds a dynamic system prompt from module fragments + GraphQL schema
2. Calls AIClient (OpenAI) with the user's messages + available tools
3. If response contains tool calls, executes them and loops back
4. Extracts structured state from tool results for Flutter preview
5. Returns AI response + state + tool records + suggestions

### The Loop (pseudocode)

```
WebSocket /api/v1/chat
  │
  ├── 1. Extract user from Firebase Auth context
  ├── 2. Get user's active app ID (from preferences or message context)
  ├── 3. Build system prompt (module fragments + GraphQL schema summary)
  ├── 4. Determine available tools (all modules — no role filtering in Phase 1)
  ├── 5. Assemble context (prepend quotes + picks to user message)
  │
  └── 6. TOOL LOOP (max 5 iterations):
       │
       ├── Call AIClient.ChatCompletion (system prompt + messages + tools)
       │
       ├── If response has NO tool calls → DONE, return text response
       │
       └── If response HAS tool calls:
            │
            ├── For each tool call:
            │   ├── Inject user context (appID, userID) into tool arguments
            │   ├── registry.RouteToolCall(ctx, call)
            │   │   └── Module.ExecuteTool()
            │   │       └── Build GraphQL query → graphqlExecutor.Execute()
            │   │           └── gqlgen resolver → domain service → PostgreSQL
            │   └── Record in allToolRecords
            │
            ├── Append assistant message (with tool_calls) to messages
            ├── Append tool results as "tool" role messages (with tool_call_id)
            └── Loop back to AIClient.ChatCompletion
  │
  ├── 7. Extract state from allToolRecords (reverse scan, latest-wins)
  ├── 8. Generate suggestive replies from tool records + state
  └── 9. Stream response to Flutter via WebSocket
```

### CRITICAL: Max 5 Iterations
Prevents runaway loops. Safety net for token budget. In practice, most revenue queries need 1-2 iterations.

### CRITICAL: Tool Results Must Include Full Objects
When `list_subscriptions` returns results, include the FULL subscription list — not just `{count: 5}`. The Flutter frontend needs the data to render tables, risk cards, and charts without a separate API call.

```go
// BAD - Flutter can't render anything
result := map[string]any{"count": 5}

// GOOD - Flutter renders a subscription table immediately
result := map[string]any{
    "subscriptions": subscriptions,
    "total_count":   len(subscriptions),
    "risk_summary":  riskCounts,
}
```

### State Extraction (latest-wins)
After the loop, scan `allToolRecords` in **reverse order** to find the latest state:

```go
for i := len(records) - 1; i >= 0; i-- {
    rec := records[i]
    if rec.IsError { continue }

    switch rec.Module {
    case "subscriptions":
        if subscriptionState == nil {
            subscriptionState = json.RawMessage(rec.Result)
        }
    case "metrics":
        if metricsState == nil {
            metricsState = json.RawMessage(rec.Result)
        }
    case "risk":
        if riskState == nil {
            riskState = json.RawMessage(rec.Result)
        }
    case "store_health":
        if storeHealthState == nil {
            storeHealthState = json.RawMessage(rec.Result)
        }
    case "earnings":
        if earningsState == nil {
            earningsState = json.RawMessage(rec.Result)
        }
    }
}
```

**Why reverse?** If the AI queries risk, then drills into a specific store — you want the store health result (most recent), not the risk summary.

---

## PATTERN 4: DYNAMIC SYSTEM PROMPT

The system prompt is assembled from:

```
Base instructions (conversation style, revenue domain rules)
    +
GraphQL schema summary (so AI understands available data)
    +
Module fragments (each module adds its own instructions)
    +
Scope notes (if @module filter active)
    +
User context (active app name, plan tier)
```

### Base Prompt Rules (critical for good AI behavior)
```
You are LedgerGuard's Revenue Intelligence Assistant. You help Shopify app developers
understand their subscription revenue, identify at-risk stores, and track earnings.

RULES:
- Never ask a numbered checklist of questions
- Infer as much as possible from context (e.g., if user says "my MRR" → use their active app)
- Ask only ONE clarifying question at a time
- When you have enough info, query immediately — don't ask for permission
- Always include specific numbers and store names in your responses
- Format data as tables when showing multiple items
- Suggest 2-3 follow-up actions after each answer
- When showing monetary values, convert cents to dollars (divide by 100)
- Risk states: SAFE (green), ONE_CYCLE_MISSED (yellow), TWO_CYCLES_MISSED (red), CHURNED (dark)
```

### Module Fragment Example (risk)
```
You can analyze subscription risk for the user's app.

When the user asks about risk, at-risk stores, or revenue safety:
1. Call get_risk_summary first for an overview
2. If they want details, call list_at_risk for specific stores
3. Always mention the revenue impact (sum of at-risk MRR)
4. Suggest concrete actions: "You might want to check on store-x directly"

RISK DEFINITIONS:
- SAFE: Active and paying, or ≤30 days past due (grace period)
- ONE_CYCLE_MISSED: 31-60 days past due or FROZEN status
- TWO_CYCLES_MISSED: 61-90 days past due
- CHURNED: >90 days past due or CANCELLED/EXPIRED status
```

### GraphQL Schema in System Prompt
Include a condensed version of the schema so the AI understands the data model:

```
AVAILABLE DATA (via GraphQL):
- Subscriptions: domain, shopName, status, riskState, mrrCents, daysPastDue, plan, billingInterval
- Metrics: activeMrrCents, revenueAtRiskCents, renewalSuccessRate, safe/missed/churned counts
- MetricsTrend: historical snapshots by month
- StoreHealth: isPaid, riskState, lastPaymentDate, subscription details
- Earnings: recurringCents, usageCents, oneTimeCents, refundCents
- RiskSummary: counts per state + at-risk subscription list with revenue impact
```

### Scoped Prompts (@module filtering)
When user types `@risk`, only include the risk module's fragment and add:
```
NOTE: This message is scoped to the 'risk' module. Focus only on risk analysis capabilities.
```

---

## PATTERN 5: CONTEXT-AWARE INPUT (3 mechanisms)

### A. @Module Tags (ScopedModule)
User types `@` → dropdown shows modules → selects one → only that module's tools sent to AI.

```
Input: @risk show me at-risk stores
Backend receives: { scoped_module: "risk", messages: [...] }
Effect: Only risk tools + prompt fragment sent to Claude
```

**Implementation:** Flutter detects `@` keystroke, calls module list, shows filtered dropdown with keyboard navigation.

Available modules for autocomplete:
```
@subscriptions — Query subscription status, details, and lists
@metrics       — MRR, renewal rates, revenue trends
@risk          — Risk analysis, at-risk stores, revenue impact
@store_health  — Individual store payment status
@earnings      — Revenue breakdown by type
@sync          — Trigger data sync
```

### B. Quote Selection (QuotedContext)
User highlights text in an AI message → floating "Quote" button appears → click adds quoted text to next message.

```
User quotes: "store-x.myshopify.com has been 45 days overdue"
Backend receives: { quoted_context: "store-x.myshopify.com has been 45 days overdue" }
Effect: Prepended as: [Quoted from previous message]\n> quoted text
```

### C. Context Picks (ContextPicks)
User clicks elements in the DataPanel (store name, risk state, MRR value) → chips appear in input area.

```
User clicks: "store-x.myshopify.com" and "Risk: TWO_CYCLES_MISSED"
Backend receives: { context_picks: [{label: "Store", value: "store-x.myshopify.com"}, ...] }
Effect: Prepended as: [Context]\n- Store: store-x.myshopify.com\n- Risk: TWO_CYCLES_MISSED
```

### Assembly Order
```go
func assembleContext(userMessage, quotedContext string, picks []ContextPick) string {
    var sb strings.Builder
    if quotedContext != "" {
        sb.WriteString("[Quoted from previous message]\n> ")
        sb.WriteString(quotedContext)
        sb.WriteString("\n\n")
    }
    if len(picks) > 0 {
        sb.WriteString("[Context]\n")
        for _, p := range picks {
            sb.WriteString("- " + p.Label + ": " + p.Value + "\n")
        }
        sb.WriteString("\n")
    }
    sb.WriteString(userMessage)
    return sb.String()
}
```

---

## PATTERN 6: TENANT ISOLATION (replaces complex RBAC)

LedgerGuard uses Firebase Auth — every request has a user. Tenant isolation is enforced at the GraphQL resolver level:

### Layer 1: User Context Injection
Before any tool executes, the chat handler injects the user's context into every tool call:

```go
// Chat handler injects before routing
enrichedCall := ToolCall{
    Name:      call.Name,
    Arguments: call.Arguments,
    UserID:    user.ID,            // from Firebase Auth
    AppID:     user.ActiveAppID,   // from preferences
}
```

### Layer 2: Resolver-Level Filtering
Every GraphQL resolver filters by app_id, which is linked to the user's partner_account:

```go
func (r *queryResolver) Subscriptions(ctx context.Context, appID string, ...) ([]*Subscription, error) {
    // Verify user owns this app (via partner_account chain)
    if !r.authz.UserOwnsApp(ctx, appID) {
        return nil, errors.New("access denied")
    }
    return r.subscriptionRepo.FindByAppID(ctx, appID)
}
```

### Layer 3: Read-Only Schema
The GraphQL schema has **no mutations** (Phase 1). The only write action is `sync__trigger_sync`, which calls the existing sync endpoint with the same auth checks.

---

## PATTERN 7: FLUTTER FRONTEND (Bloc Pattern)

```
┌──────────────────────────────────────────────┐
│ ChatPage (STATE OWNER via ChatBloc)          │
├─────────────────┬────────────────────────────┤
│ ChatPane        │ DataPanel                  │
│ ├── Messages    │ ├── SubscriptionTable (if  │
│ ├── Input       │ │   activeModule=subs)     │
│ ├── @module     │ ├── MetricsCard (if        │
│ │   chips       │ │   activeModule=metrics)  │
│ ├── Quote chips │ ├── RiskBreakdown (if      │
│ └── Pick chips  │ │   activeModule=risk)     │
│                 │ ├── StoreHealthCard (if     │
│ Suggestions     │ │   activeModule=health)   │
│ └── Quick-reply │ ├── EarningsChart (if       │
│     chips       │ │   activeModule=earnings) │
│                 │ └── Loading skeleton        │
└─────────────────┴────────────────────────────┘
```

### ChatBloc (State Management)

```dart
// Events
abstract class ChatEvent {}
class SendMessage extends ChatEvent {
  final String message;
  final String? scopedModule;
  final String? quotedContext;
  final List<ContextPick>? contextPicks;
}
class ReceiveResponse extends ChatEvent {
  final ChatResponse response;
}
class ClearChat extends ChatEvent {}

// States
abstract class ChatState {}
class ChatInitial extends ChatState {}
class ChatLoaded extends ChatState {
  final List<ChatMessage> messages;
  final String? activeModule;
  final Map<String, dynamic>? subscriptionState;
  final Map<String, dynamic>? metricsState;
  final Map<String, dynamic>? riskState;
  final Map<String, dynamic>? storeHealthState;
  final Map<String, dynamic>? earningsState;
  final List<String> suggestions;
  final bool isLoading;
}
class ChatError extends ChatState {
  final String message;
}
```

### State Lifting
ChatBloc owns all shared state. When a response arrives:
- `activeModule` = last tool's module name
- Module-specific state extracted from `response.{module}_state`
- These flow to DataPanel widgets for rendering

### Suggestive Replies
After each AI response, generate 2-3 context-aware quick-reply suggestions:

| After this tool result | Suggest these |
|------------------------|---------------|
| risk_summary | "Show at-risk stores", "What's the MRR impact?", "Any churned this month?" |
| list_subscriptions | "Filter by at-risk only", "Show subscription details for [first store]" |
| get_latest_metrics | "Show trend over 6 months", "Compare to last month" |
| store_health | "Show subscription history", "Check other stores" |
| earnings_breakdown | "Which type grew most?", "Show monthly trend" |
| trigger_sync | "Check sync status", "Show updated metrics" |

Render as clickable chips below the last assistant message. Click → auto-send as user message.

---

## PATTERN 8: THREAD SAFETY

### GraphQL Executor
The GraphQL executor is shared across modules and must be thread-safe:

```go
type GraphQLExecutor struct {
    schema *graphql.Schema  // gqlgen schema (immutable after init)
}

func (e *GraphQLExecutor) Execute(ctx context.Context, query string, vars map[string]any) (json.RawMessage, error) {
    // gqlgen handles concurrency internally
    resp := e.schema.Exec(ctx, query, "", vars)
    if len(resp.Errors) > 0 {
        return nil, fmt.Errorf("graphql: %s", resp.Errors[0].Message)
    }
    return resp.Data, nil
}
```

### WebSocket Connection Pool
Each WebSocket connection runs in its own goroutine. Chat state (messages) is per-connection — no shared mutable state between connections.

### Sync Module Gotcha
`trigger_sync` starts an async operation. Don't block the tool call waiting for sync to finish. Return immediately with a status, let the user poll with `sync_status`:

```go
func (e *Executor) executeTriggerSync(ctx context.Context, args TriggerSyncArgs) ToolResult {
    // Fire and forget — sync runs in background
    go e.syncService.SyncApp(context.Background(), args.AppID)

    return ToolResult{
        Content: `{"status": "sync_started", "message": "Sync triggered. Use sync_status to check progress."}`,
    }
}
```

---

## MODULE DEFINITIONS

### Subscriptions Module (4 tools)

```
subscriptions__list_subscriptions
  Input:  { app_id, risk_state?, status?, domain?, limit?, offset? }
  Output: { subscriptions: [...], total_count, risk_summary }
  GraphQL: query { subscriptions(appId, riskState, status, domain, limit, offset) { nodes { ... } totalCount } }

subscriptions__get_subscription_detail
  Input:  { subscription_id }
  Output: { subscription: { ...full details + events history } }
  GraphQL: query { subscription(id) { domain shopName status riskState mrrCents events { ... } } }

subscriptions__get_subscription_summary
  Input:  { app_id }
  Output: { total, by_status: {active, cancelled, ...}, by_risk: {safe, missed, churned}, avg_mrr }
  GraphQL: query { subscriptions(appId) { nodes { status riskState mrrCents } } } → aggregate in executor

subscriptions__search_subscriptions
  Input:  { app_id, query }  // fuzzy search by domain or shop name
  Output: { subscriptions: [...], total_count }
  GraphQL: query { subscriptions(appId, domain: $query) { ... } }
```

### Metrics Module (3 tools)

```
metrics__get_latest_metrics
  Input:  { app_id }
  Output: { active_mrr, revenue_at_risk, renewal_rate, usage_revenue, safe/missed/churned counts }
  GraphQL: query { metrics(appId) { activeMrrCents revenueAtRiskCents renewalSuccessRate ... } }

metrics__get_metrics_trend
  Input:  { app_id, months? }  // default 6
  Output: { snapshots: [{ date, active_mrr, renewal_rate, ... }] }
  GraphQL: query { metricsTrend(appId, months) { date activeMrrCents ... } }

metrics__get_aggregate_metrics
  Input:  { }  // all apps for this user
  Output: { total_mrr, total_at_risk, app_count, per_app: [...] }
  GraphQL: Uses existing REST /metrics/aggregate endpoint (bridge)
```

### Risk Module (3 tools)

```
risk__get_risk_summary
  Input:  { app_id }
  Output: { total, safe, one_cycle_missed, two_cycles_missed, churned, revenue_at_risk_cents }
  GraphQL: query { riskSummary(appId) { safe oneCycleMissed twoCyclesMissed churned revenueAtRiskCents } }

risk__list_at_risk
  Input:  { app_id, risk_state? }  // ONE_CYCLE_MISSED, TWO_CYCLES_MISSED, or both
  Output: { subscriptions: [{ domain, risk_state, days_past_due, mrr }] }
  GraphQL: query { riskSummary(appId) { atRiskSubscriptions { domain riskState daysPastDue mrrCents } } }

risk__get_risk_timeline
  Input:  { app_id, months? }
  Output: { timeline: [{ date, safe, missed, churned }] }  // from daily snapshots
  GraphQL: query { metricsTrend(appId, months) { date safeCount oneCycleMissedCount churnedCount } }
```

### Store Health Module (2 tools)

```
store_health__check_store
  Input:  { app_id, domain }
  Output: { domain, shop_name, is_paid, risk_state, last_payment, days_past_due, subscription }
  GraphQL: query { storeHealth(appId, domain) { domain isPaid riskState lastPaymentDate subscription { ... } } }

store_health__compare_stores
  Input:  { app_id, domains: [string] }  // max 10
  Output: { stores: [{ domain, risk_state, mrr, days_past_due }] }
  GraphQL: Multiple storeHealth queries batched
```

### Earnings Module (2 tools)

```
earnings__get_breakdown
  Input:  { app_id }
  Output: { recurring, usage, one_time, refund, total }  // all in cents
  GraphQL: query { earnings(appId) { recurringCents usageCents oneTimeCents refundCents totalCents } }

earnings__get_status
  Input:  { app_id }
  Output: { current_month_earnings, last_month, growth_pct, by_type }
  Uses existing REST /earnings/status endpoint (bridge)
```

### Sync Module (2 tools)

```
sync__trigger_sync
  Input:  { app_id? }  // omit for all apps
  Output: { status: "sync_started", message }
  Calls existing SyncService.SyncApp() in background goroutine

sync__get_sync_status
  Input:  { app_id }
  Output: { last_sync_at, transaction_count, status }
  Reads from app record
```

---

## API CONTRACTS

### WebSocket /api/v1/chat

**Client → Server (send message):**
```json
{
  "type": "message",
  "messages": [
    {"role": "user", "content": "Which stores are at risk?"},
    {"role": "assistant", "content": "Let me check..."}
  ],
  "scoped_module": "risk",
  "quoted_context": null,
  "context_picks": []
}
```

**Server → Client (response):**
```json
{
  "type": "response",
  "message": "You have 3 stores at risk with $99.97/mo in jeopardy...",
  "subscription_state": null,
  "metrics_state": null,
  "risk_state": {
    "total": 45,
    "safe": 38,
    "one_cycle_missed": 2,
    "two_cycles_missed": 1,
    "churned": 4,
    "revenue_at_risk_cents": 9997,
    "at_risk_subscriptions": [
      {"domain": "acme.myshopify.com", "risk_state": "TWO_CYCLES_MISSED", "days_past_due": 72, "mrr_cents": 4999}
    ]
  },
  "store_health_state": null,
  "earnings_state": null,
  "tools_used": [
    {
      "module": "risk",
      "tool": "get_risk_summary",
      "arguments": "{\"app_id\": \"...\"}",
      "result": "{...}",
      "is_error": false
    }
  ],
  "suggestions": ["Show at-risk store details", "What's the MRR impact?", "Trigger a sync"],
  "active_module": "risk"
}
```

**Server → Client (streaming — optional):**
```json
{
  "type": "stream",
  "delta": "You have ",
  "done": false
}
```

### GET /api/v1/chat/modules
```json
[
  {"name": "subscriptions", "description": "Query subscription status and details", "tool_count": 4},
  {"name": "metrics", "description": "MRR, renewal rates, revenue trends", "tool_count": 3},
  {"name": "risk", "description": "Risk analysis and at-risk store identification", "tool_count": 3},
  {"name": "store_health", "description": "Individual store payment health", "tool_count": 2},
  {"name": "earnings", "description": "Revenue breakdown by charge type", "tool_count": 2},
  {"name": "sync", "description": "Trigger and monitor data sync", "tool_count": 2}
]
```

---

## WIRING (main.go pattern)

```go
// 1. Create GraphQL executor (wraps gqlgen schema)
graphqlExec := chat.NewGraphQLExecutor(gqlgenSchema)

// 2. Create modules (each gets GraphQL executor injected)
subscriptionsModule := subscriptions.NewModule(graphqlExec)
metricsModule := metrics.NewModule(graphqlExec)
riskModule := risk.NewModule(graphqlExec)
storeHealthModule := storehealth.NewModule(graphqlExec)
earningsModule := earnings.NewModule(graphqlExec)
syncModule := sync.NewModule(graphqlExec, syncService) // sync needs SyncService for trigger

// 3. Register with registry
registry := chat.NewRegistry()
registry.Register(subscriptionsModule)
registry.Register(metricsModule)
registry.Register(riskModule)
registry.Register(storeHealthModule)
registry.Register(earningsModule)
registry.Register(syncModule)

// 4. Create AI clients — both available, user picks at runtime
openaiClient := external.NewOpenAIClient(cfg.OpenAI.APIKey, "gpt-4o")
// claudeClient := external.NewClaudeClient(cfg.Claude.APIKey, "claude-sonnet-4-6")  // Phase 9

// 5. Create provider registry — resolves AIClient per user preference
aiProviders := chat.NewAIProviderRegistry()
aiProviders.Register("openai", openaiClient)
// aiProviders.Register("claude", claudeClient)  // Phase 9

// 6. Create chat handler — resolves provider per request from user preference
chatHandler := chat.NewHandler(aiProviders, registry)

// 5. Wire WebSocket route
r.Route("/api/v1/chat", func(r chi.Router) {
    r.Use(authMiddleware)
    r.Get("/", chatHandler.HandleWebSocket)        // WebSocket upgrade
    r.Get("/modules", chatHandler.HandleListModules) // Module list for @autocomplete
})
```

**Key insights:**
- All modules receive the same `graphqlExec` — they build different queries but share the execution layer
- Only `syncModule` receives an extra dependency (`syncService`) because it triggers writes
- `aiClient` satisfies the `AIClient` interface — swap OpenAI for Claude by changing one line
- AIClient is injected, making it testable with mocks (no real API calls in tests)

---

## IMPLEMENTATION ORDER (recommended)

Build in this order — each step is independently testable:

### Phase 1: GraphQL Layer
1. Define `.graphql` schema files (types, queries)
2. Generate gqlgen boilerplate (`gqlgen generate`)
3. Implement resolvers (delegate to existing domain services)
4. Write resolver tests (query → expected response)
5. **Expose `/graphql` endpoint with Firebase Auth** (for testing)
6. Test queries directly in GraphQL playground

### Phase 2: Module Framework
7. Module interface + types (`ToolDefinition`, `ToolCall`, `ToolResult`)
8. Registry (register, list tools, route calls)
9. GraphQL executor wrapper

### Phase 3: First Module (Risk — highest value)
10. Risk module tools (OpenAI function schemas)
11. Risk executor (tool → GraphQL query → execute → return)
12. Risk module registration
13. Test: call tool directly, verify GraphQL query and response

### Phase 4: Chat Handler + OpenAI Integration
14. AIClient interface + OpenAI implementation (function calling)
15. Chat handler with tool loop (max 5 iterations)
16. Dynamic system prompt assembly
17. State extraction from tool results
18. WebSocket endpoint
19. Test: send message → verify OpenAI called → tools executed → response returned

### Phase 5: Remaining Modules
20. Subscriptions module (4 tools)
21. Metrics module (3 tools)
22. Store Health module (2 tools)
23. Earnings module (2 tools)
24. Sync module (2 tools)

### Phase 6: Flutter Frontend
25. `ChatBloc` (events, states, WebSocket connection)
26. `ChatPane` (messages, input, send)
27. `MessageBubble` (markdown rendering, expandable tool calls)
28. `DataPanel` (conditional preview based on activeModule)
29. Split-pane `ChatPage` (state lifting from bloc)
30. Suggestive reply chips

### Phase 7: Context-Aware Input
31. @module autocomplete (ModuleChip widget)
32. Quote selection (floating button on text highlight)
33. Context picks (clickable DataPanel elements → chips in input)
34. Backend context assembly

### Phase 8: Polish
35. Loading states + typing indicator
36. Error handling + reconnection
37. Message history persistence (local storage)
38. Streaming responses (optional)

### Phase 9: Add Claude API as Parallel Provider (future)
39. Implement `ClaudeClient` satisfying `AIClient` interface
40. Map ToolDefinition → Claude `input_schema` format
41. Map tool results → Claude `tool_result` content blocks
42. Add `ai_provider` user preference (openai | claude) — per-user choice
43. Frontend: provider selector in chat settings (dropdown or toggle)
44. Both providers run in production simultaneously — user picks which one
45. Track per-provider metrics (latency, token cost, response quality ratings)

---

## GOTCHAS & LESSONS LEARNED

| Issue | Symptom | Fix |
|-------|---------|-----|
| Tool name format | OpenAI 400: "does not match pattern" | Use `__` separator, not `.` or `/` |
| State extraction order | Frontend shows old risk summary instead of store detail | Scan tool records in **reverse** (latest-wins) |
| Thin tool results | Flutter can't render data table | Return full objects, not just `{count: 5}` |
| Numbered questions | AI asks 5 questions before querying | System prompt: "infer, query immediately, don't ask permission" |
| AI hallucinates tools | Calls `billing__get_invoices` (doesn't exist) | Return error ToolResult — AI sees it and self-corrects next iteration |
| Cents vs dollars | AI shows "$4999" instead of "$49.99" | System prompt: "divide cents by 100 for display" |
| Missing app context | Tool fails: "app_id required" | Chat handler auto-injects active app_id before routing |
| WebSocket disconnect | Flutter loses connection on mobile backgrounding | Implement reconnection with exponential backoff in ChatBloc |
| GraphQL N+1 | Store health for 50 stores = 50 DB queries | Use dataloaders in gqlgen resolvers |
| Sync blocking | `trigger_sync` blocks for 30s waiting for completion | Return immediately, let user poll with `sync_status` |
| Token budget | User has 100 subscriptions → huge tool result | Paginate: default limit 20, max 50 per tool call |
| Conversation length | 20+ messages → context window fills up | Trim conversation to last 10 messages, summarize older ones |

---

## ADAPTING FROM OPENAI EXPLORER

| OpenAI Explorer | LedgerGuard | Same Pattern |
|-----------------|-------------|-------------|
| WhatsApp Templates module | Subscriptions module | Module with tools + executor |
| Broadcast module | Sync module | Module with async operations |
| Template validation | Risk classification | Domain logic in tool result |
| Template preview | Subscription table | State extraction → DataPanel |
| Broadcast progress | Sync status | Async polling via tool |
| OpenAI function calling | OpenAI function calling (same!) | Same API, same loop, same patterns |
| React split-pane | Flutter split-pane | Same layout, Bloc instead of useState |
| In-memory stores | GraphQL → PostgreSQL | Modules query via GraphQL, not direct stores |
| Direct OpenAI client | AIClient interface | Provider-agnostic — Claude swap later |
| CSV file processing | (not needed) | Removed |
| Phone validation | (not needed) | Removed |
| Role-based RBAC | Tenant isolation | Simpler — Firebase Auth + app ownership |
| REST POST /chat | WebSocket /chat | Streaming support |

**The module interface, registry, chat handler, tool loop, state extraction, context assembly, and suggestive replies are all reused. OpenAI function calling is the same proven pattern from Explorer — just wrapped behind AIClient interface for future Claude migration. Only the modules and data layer change.**
