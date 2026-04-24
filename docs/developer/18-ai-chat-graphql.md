# 18. AI Chat & Internal GraphQL

## What It Does
Provides an AI-powered conversational interface for querying Shopify app revenue data. Users ask natural-language questions ("What's my MRR?", "Show at-risk stores") and the system uses LLM function calling to execute tools against an internal GraphQL schema. The architecture is fully modular: new data domains are added as plugins without touching the chat core.

The system has three main capabilities:
1. **AI Chat** -- SSE-streamed conversation with tool call loop (max 5 iterations)
2. **Internal GraphQL** -- Schema covering subscriptions, metrics, risk, store health, earnings, transactions, and apps
3. **Module Registry** -- Plugin system where each module declares its tools, prompt fragments, and execution logic

## Architecture
```
┌─────────────────────────────────────────────────────────────────┐
│                        Chat Handler                             │
│  POST /api/v1/chat (SSE streaming)                              │
│  GET  /api/v1/chat/modules (list available modules)             │
│                                                                 │
│  ┌─────────────┐    ┌──────────────────┐    ┌───────────────┐  │
│  │ AIProvider  │    │    Registry       │    │  GraphQL      │  │
│  │ Registry    │    │                  │    │  Executor     │  │
│  │             │    │  ┌────────────┐  │    │               │  │
│  │ ┌─────────┐│    │  │risk       │  │    │  httptest     │  │
│  │ │ OpenAI  ││    │  │subscriptions│ │    │  recorder     │  │
│  │ └─────────┘│    │  │metrics    │  │    │  wrapper      │  │
│  │ ┌─────────┐│    │  │store_health│ │    └───────────────┘  │
│  │ │ Claude  ││    │  │earnings   │  │                       │
│  │ │(planned)││    │  │sync       │  │                       │
│  │ └─────────┘│    │  │apps       │  │                       │
│  └─────────────┘    │  └────────────┘  │                       │
│                     └──────────────────┘                       │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
              ┌───────────────────────────┐
              │ Internal GraphQL Schema   │
              │ (gqlgen-generated)        │
              │                           │
              │ Resolvers delegate to:    │
              │ - SubscriptionRepository  │
              │ - TransactionRepository   │
              │ - MetricsEngine           │
              │ - RiskEngine              │
              │ - AppRepository           │
              └───────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility |
|-----------|---------------|
| `Handler` | Receives chat requests, orchestrates SSE streaming, runs tool call loop |
| `Registry` | Manages module plugins, routes tool calls, builds system prompts |
| `Module` (interface) | Plugin contract: Name, Description, PromptFragment, Tools, ExecuteTool |
| `AIClient` (interface) | Provider-agnostic LLM abstraction |
| `AIProviderRegistry` | Manages multiple LLM providers, selects by name or default |
| `GraphQLExecutor` | Thread-safe gqlgen wrapper using httptest for in-process execution |

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `backend/internal/chat/handler.go` | ~355 | SSE chat handler with tool call loop and module list endpoint |
| `backend/internal/chat/module.go` | ~31 | Module plugin interface |
| `backend/internal/chat/registry.go` | ~126 | Module registry: register, route, list tools, build system prompt |
| `backend/internal/chat/types.go` | ~34 | Core types: ToolDefinition, ToolCall, ToolResult, ChatMessage |
| `backend/internal/chat/ai_client.go` | ~40 | AIClient interface and provider-agnostic request/response types |
| `backend/internal/chat/ai_provider_registry.go` | ~50 | Multi-provider registry with fallback selection |
| `backend/internal/chat/graphql_executor.go` | ~75 | Thread-safe GraphQL execution via httptest recorder |
| `backend/internal/chat/graphql/schema.graphql` | ~184 | Internal GraphQL schema (types, enums, queries) |
| `backend/internal/chat/graphql/resolver.go` | -- | Root resolver with injected repos/services |
| `backend/internal/chat/graphql/schema.resolvers.go` | -- | Generated resolver implementations |
| `backend/internal/chat/modules/risk/module.go` | -- | Risk module (3 tools) |
| `backend/internal/chat/modules/subscriptions/module.go` | -- | Subscriptions module (4 tools) |
| `backend/internal/chat/modules/metrics/module.go` | -- | Metrics module (3 tools) |
| `backend/internal/chat/modules/store_health/module.go` | -- | Store health module (2 tools) |
| `backend/internal/chat/modules/earnings/module.go` | -- | Earnings module (2 tools) |
| `backend/internal/chat/modules/sync/module.go` | -- | Sync module (2 tools) |
| `backend/internal/chat/modules/apps/module.go` | -- | Apps module (3 tools) |

## Data Flow

### Chat Request Lifecycle
```
Client (Flutter)                     Server
     │                                  │
     │  POST /api/v1/chat              │
     │  { messages, scoped_module,     │
     │    app_id }                     │
     │────────────────────────────────▶│
     │                                  │
     │  SSE: Content-Type: text/event-stream
     │◀────────────────────────────────│
     │                                  │
     │   ┌─── Tool Call Loop (max 5 iterations) ───┐
     │   │                                          │
     │   │  1. Build system prompt (base + modules) │
     │   │  2. Send messages + tools to AI provider │
     │   │  3. AI returns tool_calls or text        │
     │   │                                          │
     │   │  If tool_calls:                          │
     │   │    SSE: { type: "tool_call", data: ... } │
     │◀──│────────────────────────────────────────  │
     │   │    Registry.RouteToolCall()              │
     │   │    Module.ExecuteTool()                  │
     │   │    → GraphQL query via Executor          │
     │   │    SSE: { type: "tool_result", data: ...}│
     │◀──│──────────────────────────────────────── │
     │   │    Append result to messages             │
     │   │    Loop back to step 2                   │
     │   │                                          │
     │   │  If text (no tool_calls):                │
     │   │    SSE: { type: "response", data: ... }  │
     │◀──│──────────────────────────────────────── │
     │   │    Done                                  │
     │   └──────────────────────────────────────────┘
     │                                  │
```

### Tool Routing
```
Fully qualified tool name: "risk__get_risk_summary"
                            ^^^^   ^^^^^^^^^^^^^^^^
                            module   local tool name
                              │
Registry.RouteToolCall()      │
  │                           │
  ├─ Split on "__" double underscore
  ├─ Find module by prefix
  ├─ Strip prefix → local name "get_risk_summary"
  └─ Module.ExecuteTool(ctx, localCall)
       │
       └─ GraphQLExecutor.Execute(query, vars)
            │
            └─ httptest.NewRecorder() + gqlgen handler.ServeHTTP()
```

### SSE Event Types
| Type | Payload | When |
|------|---------|------|
| `tool_call` | `{ iteration, tool_name, arguments }` | AI requests a tool call |
| `tool_result` | `{ tool_name, content, is_error }` | Tool execution completed |
| `response` | `{ message, *_state, suggestions, usage }` | Final AI text response |
| `error` | `{ message }` | Unrecoverable error |

### System Prompt Assembly
1. Base prompt with rules (only answer Shopify revenue questions, format tables, convert cents to dollars)
2. Module prompt fragments appended (filtered by `scoped_module` if set)
3. Active app ID appended if provided in request

## Configuration
| Setting | Source | Description |
|---------|--------|-------------|
| OpenAI API key | Environment / config | Required for chat to function |
| Default AI provider | `AIProviderRegistry` fallback param | `"openai"` in production |
| Max tool iterations | Constant `maxToolIterations = 5` | Prevents infinite tool loops |
| Scoped module | Request field `scoped_module` | Limits tools and prompt to one module |

### GraphQL Code Generation
```bash
cd backend && go generate ./internal/chat/graphql/
```
This runs gqlgen to regenerate `generated.go` and `models_gen.go` from `schema.graphql` and `gqlgen.yml`.

## API Surface
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | /api/v1/chat | Firebase | AI chat with SSE streaming |
| GET | /api/v1/chat/modules | Firebase | List registered modules and tool counts |
| GET | /graphql | Firebase | GraphQL Playground (browser) |
| POST | /graphql | Firebase | Execute GraphQL query |

### Chat Request Body
```json
{
  "messages": [
    { "role": "user", "content": "What's my MRR?" }
  ],
  "scoped_module": "metrics",
  "app_id": "uuid-of-active-app"
}
```

### Module List Response
```json
[
  { "name": "risk", "description": "...", "tool_count": 3 },
  { "name": "subscriptions", "description": "...", "tool_count": 4 },
  { "name": "metrics", "description": "...", "tool_count": 3 },
  { "name": "store_health", "description": "...", "tool_count": 2 },
  { "name": "earnings", "description": "...", "tool_count": 2 },
  { "name": "sync", "description": "...", "tool_count": 2 },
  { "name": "apps", "description": "...", "tool_count": 3 }
]
```

### Internal GraphQL Schema (Key Queries)
```graphql
subscriptions(appId: ID!, riskState: RiskState, status: AppSubscriptionStatus, domain: String, limit: Int, offset: Int): AppSubscriptionConnection!
subscription(id: ID!): AppSubscription
metrics(appId: ID!): Metrics!
metricsTrend(appId: ID!, months: Int): [DailySnapshot!]!
storeHealth(appId: ID!, domain: String!): StoreHealth!
earnings(appId: ID!): Earnings!
transactions(appId: ID!, chargeType: ChargeType, domain: String, startDate: DateTime, endDate: DateTime, limit: Int, offset: Int): TransactionConnection!
riskSummary(appId: ID!): RiskSummary!
apps: [App!]!
app(id: ID!): App
```

## Extension Points
- **New module**: Create a directory under `internal/chat/modules/`, implement the `Module` interface, and call `registry.Register()` in `main.go`. Zero changes to the chat core.
- **New AI provider**: Implement the `AIClient` interface and register with `AIProviderRegistry`. The handler uses `aiProviders.Default()` to select.
- **New GraphQL types**: Add to `schema.graphql`, run `go generate`, implement resolver. Tools in modules can immediately query the new types.
- **Custom suggestions**: Extend `generateSuggestions()` in `handler.go` to map new tool names to follow-up prompts.
- **State extraction**: The `responseEvent` carries per-module state (`SubscriptionState`, `MetricsState`, etc.). Add new `*State` fields for new modules to pass structured data to the frontend alongside the AI text.

## Gotchas
- **Tool naming convention**: Must use double underscore (`module__tool_name`). Single underscore within tool names is fine. The registry splits on the first `__` occurrence.
- **Max 5 iterations**: If the AI keeps requesting tools after 5 rounds, the handler returns a generic "I've gathered the data above" message with whatever state was collected.
- **SSE not WebSocket**: Despite earlier design docs mentioning WebSocket, the current implementation uses Server-Sent Events (SSE) via `http.Flusher`. The `X-Accel-Buffering: no` header prevents nginx from buffering the stream.
- **GraphQLExecutor uses httptest**: Tool execution goes through `httptest.NewRecorder()` to invoke the gqlgen handler in-process. This means no network overhead but also no HTTP middleware (auth context must be forwarded via `context.Context`).
- **Module prompt fragments**: Each module's `PromptFragment()` is included in every request's system prompt (unless `scoped_module` filters it). Keep fragments concise to avoid exceeding token limits.
- **Token usage is cumulative**: The `Usage` field in the response sums tokens across all iterations of the tool call loop, not just the final completion.
- **Error handling in tools**: Tool results with `IsError: true` are still sent to the AI as context. The AI interprets errors and may retry with different parameters or report the failure to the user.
- **No auth in GraphQL executor**: The internal GraphQL is accessed in-process. Firebase auth is checked at the HTTP handler level before the chat handler runs. The GraphQL resolvers do not re-check auth.
