# Tech Design: AI Chat (as-built)

**PRD:** docs/prds/02-ai-chat.md · **Author:** sachin.s · **Date:** 2026-09-25
**Status:** Draft (backfilled from implementation)
**Reversibility:** Public streaming API (`POST /api/v1/chat`, SSE event shapes) + the internal GraphQL schema (`internal/chat/graphql/schema.graphql`) are the hard-to-undo surfaces. No new persistent data model. The **tenant-isolation** behavior (below) is the highest-stakes correctness concern.

> **As-built.** "Current state IS the design." Every claim cites a real path. Markers: **FACT / INFERRED / UNKNOWN**. **Nothing is refactored or fixed** — gaps are LISTED as DIVERGENCE.

## Current state

**FACT.** Chat is a single SSE endpoint: `POST /chat` under auth middleware (`router.go:488-496`), served by `chat.Handler.HandleChat` (`handler.go:69`). It rejects non-POST (405), decodes a `ChatRequest` (405/400 on bad input), resolves the default AI provider (503 if unconfigured, `handler.go:87-91`), then sets SSE headers (`text/event-stream`, `no-cache`, `X-Accel-Buffering: no`, `handler.go:100-103`).

**FACT.** It builds a system prompt from a base + per-module `PromptFragment()` (`buildSystemPrompt`, `handler.go:107-108`) and runs a **tool-call loop, max 5 iterations** (`maxToolIterations = 5`, `handler.go:11`): send thread + tool defs to the LLM → if the model returns `module__tool` calls, route them through the `Registry` (`registry.go`, splits on `__`), execute, append results, iterate; else emit the final `response`. Tool failures return `{error, isError:true}` and the loop continues (`handler.go`). After tools, it extracts a per-module **state bundle** and generates **follow-up suggestions** (`handler.go:291-344`), all streamed as typed SSE events.

**FACT.** Provider abstraction: `AIClient` interface (`ai_client.go:6-10`) + `AIProviderRegistry` (`ai_provider_registry.go`); only **OpenAI** is implemented (`openai_client.go`, default `gpt-4o`), converting `ToolDefinition` ↔ OpenAI function schemas. Tools resolve to **internal GraphQL** via `graphql_executor.go` (wraps the gqlgen HTTP handler through an in-process `httptest.ResponseRecorder`) against a **query-only** schema (`graphql/schema.graphql`). 7 modules register in `main.go:812-851`: risk, subscriptions, metrics, store_health, earnings, apps, sync.

**FACT (frontend gap).** The Insights UI (`screens/insights/insights_screen.dart`) shows a chat box, but `insights_provider.dart:73-98` implements `sendMessage` as a **mock** (800ms-delayed canned strings via `_generateResponse`); it never calls `POST /api/v1/chat`. Only daily insights are live.

## Proposed design

**No redesign — documents the shipped server design.** Pattern: **SSE handler → LLM tool-call loop → module registry → in-process GraphQL executor → repos.** Happy path:

See `docs/designs/02-ai-chat-sequence.puml` (validated `plantuml -checkonly`). Component interaction spans ≥3 units (handler, LLM provider, GraphQL/resolvers, repos), captured in that diagram.

### Data model

**FACT — no new tables.** Chat is stateless per request over existing read models: `subscriptions`, `daily_metrics_snapshot`, `transactions`, `apps` (via GraphQL resolvers `schema.resolvers.go`). Conversation history is client-supplied in each `ChatRequest.Messages` — **not persisted server-side** (INFERRED: no chat-history table). Daily insights text lives in `daily_insights` (separate feature). No migration.

### API & events

**FACT.** `POST /api/v1/chat` — body `{messages[], scopedModule?, appId?}`; response is an SSE stream of newline-delimited JSON events: `tool_call`, `tool_result`, `response` (`{text, state, suggestions}`), `error`. `GET /api/v1/chat/modules` lists modules+tools. Auth: Firebase (`AuthMW`, `router.go:491`).
**Breaking-change check:** SSE event field names + the internal GraphQL schema are the contract; renames break the (future) real client and any external GraphQL consumer. **DOC DIVERGENCE:** CLAUDE.md §12 documents this as a **WebSocket** (`/api/v1/chat` WS) — the code is **SSE over POST**. Doc must be corrected (listed, not done here).

## Alternatives considered

- **WebSocket vs SSE.** CLAUDE.md §12 describes WebSocket; the code ships SSE (`handler.go:93-103`). This looks like a **genuine past decision that changed** (WS → SSE) but the rationale is **UNKNOWN — not recorded**. **Choose WS instead if** bidirectional/interactive cancellation becomes required (SSE is server→client only).
- **In-process GraphQL executor (httptest recorder) vs direct resolver/repo calls.** As-built wraps the gqlgen HTTP handler in-process (`graphql_executor.go`). Rationale not recorded. **Choose direct calls instead if** the HTTP-recorder round-trip becomes a measurable latency/serialization cost.
- **Multi-provider LLM registry vs single client.** The registry exists but only OpenAI is wired. No alternative rationale recorded. **Choose collapsing the abstraction if** Claude/others never land; **keep it if** multi-provider is on the roadmap (repo guidance: new AI work should default to latest Claude).

## Failure modes

| Failure | Detection | Behavior (as-built) | Recovery |
|---------|-----------|---------------------|----------|
| Non-POST / bad body / empty messages | method + JSON decode | **FACT** 405 / 400 (`handler.go:70-84`) | client fixes request |
| AI provider not configured | `aiProviders.Default()` err | **FACT** 503 (`handler.go:87-91`) | ops sets OpenAI key |
| A tool call errors | ToolResult `isError` | **FACT** error appended to thread, loop continues, model can recover/explain (`handler.go`) | model retries / apologizes |
| Model keeps calling tools | iteration counter | **FACT** hard stop at 5 iterations (`handler.go:11`) | returns best-effort text |
| Streaming unsupported by writer | `w.(http.Flusher)` assert | **FACT** 500 (`handler.go:94-97`) | n/a (infra) |

### DIVERGENCE — failure paths the code does NOT handle (LISTED, not fixed)

- **D1 (highest stakes). Cross-org data access is not verified in resolvers.** `schema.resolvers.go` parses `appID` from the tool **arguments** and calls `Repo.FindByAppID(ctx, appUUID)` with **no check that the app belongs to the caller's org** (only `Apps()` uses `UserFromContext`, line 335). If the LLM is induced (prompt injection, guessed UUID) to call a tool with another org's `appID`, isolation depends entirely on whether the repos scope by ctx-injected org — **UNVERIFIED (UNKNOWN)**. → **top open question + proposed adversarial test.**
- **D2. OpenAI rate-limit / timeout mid-stream** — no explicit retry/backoff or request timeout in the handler; a hung LLM call hangs the SSE stream (no server-side deadline observed). → proposed test.
- **D3. Prompt-injection via tenant data** — tool results (store names, plan names) are fed back to the model; no sanitization/guardrail observed. → open question (Product/Sec).
- **D4. Token/context overflow** — long `messages[]` + large tool results are sent to the model with no truncation/window management observed. → proposed test.
- **D5. No cost metering** — every message spends OpenAI tokens; no per-org accounting/limit. → open question (mirrors PRD pricing).
- **D6. Frontend not wired** — the shipped client can never exercise any of the above (mock). → the enabling work item.

## Test plan

**Acceptance (from PRD PROPOSED metrics, as observable checks):**
1. Groundedness: given a fixed dataset, a "what's my MRR?" turn must include a `tool_result` whose value matches the number in the final `response` (no hallucinated figure). Extend `handler_test.go`.
2. Wire-up: an integration test that the Flutter client hits `POST /api/v1/chat` (currently impossible — mock). **TODO after D6.**

**Adversarial (one per Failure-modes row):**
- 405/400/503/500 mapping, tool-error-continues, 5-iteration cap — partly covered by `handler_test.go`, `registry_test.go`; confirm the iteration cap has an explicit test.

**Adversarial — DIVERGENCE TODO (none exist today):**
- **D1 cross-org:** call `subscriptions(appID=<other org's app>)` as org A → MUST return empty/forbidden. **Highest priority.**
- D2 LLM timeout/rate-limit injection · D4 oversized-thread truncation · D3 injection payload in a store name. **TODO.**

**Manual verification (copy-paste):**
1. `cd backend && go test ./internal/chat/... -v`
2. Run server; `curl -N -H "Authorization: Bearer <idToken>" -H "Content-Type: application/json" -d '{"messages":[{"role":"user","content":"which stores are at risk?"}],"appId":"<appID>"}' http://localhost:8080/api/v1/chat` → observe streamed `tool_call`/`tool_result`/`response` events.
3. `curl -H "Authorization: Bearer <idToken>" http://localhost:8080/api/v1/chat/modules | jq .`

**External-platform reality:** the LLM (OpenAI) is the external dependency; unit tests fake `AIClient` below the provider interface (`ai_client.go`). Real-model behavior (groundedness, injection resistance) can only be smoke-tested against a real OpenAI key — never fully deterministic.

## Pricing & policy touchpoints

**UNKNOWN — no metering/gating in code.** Per-message OpenAI cost is real and unbounded per org (D5). No Shopify-policy interaction (read-only over already-synced data). → Open question (Product).

## Rollout

**N/A for the server (already deployed).** The real enabling rollout is **wiring the Flutter client** (D6): replace `insights_provider._generateResponse` with an SSE client to `POST /api/v1/chat`. Suggested order: fix D1 (tenant isolation) → add cost guard (D5) → wire client behind a flag → enable per org. Rollback = revert to mock; no data stranded (chat is stateless).

## Observability

**FACT (partial):** handler/resolvers `log.Printf` on errors. **DIVERGENCE:** no per-request token/cost metric, no tool-call success-rate or latency dashboard, no groundedness signal. On-call greps container logs on the Hetzner box (no chat-specific index). → proposed observability TODO tied to PRD metrics.

## Open questions

- **D1 tenant isolation** — verify/enforce org ownership of `appID` in resolvers or repos. Owner: Eng/Sec. **(highest priority)**
- WS→SSE decision rationale (undocumented) + fix CLAUDE.md §12. Owner: Eng.
- Cost metering/gating (D5). Owner: Product.
- Wire the real client (D6) — intended beta gate or gap? Owner: Eng/Product.
- Multi-provider: implement Claude or collapse the registry. Owner: Eng.
