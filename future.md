# Future Features – LedgerGuard

Postponed ideas and features for later implementation.

---

## Backlog

| Feature | Priority | Notes |
|---------|----------|-------|
| Welcome & Onboarding Flow (Hybrid) | P1 | n8n + Postmark email drip + in-app checklist; custom webhook support for third-party flow builders; see `docs/prompts/welcome-onboarding-flow.md` |
| Billing System (Stripe, all-paid) | P1 | Stripe Checkout; Starter (14-day trial) / Pro / Enterprise; no free tier; read-only on expiry; plans + plan_features tables; see `docs/prompts/billing-system-flow.md` |
| AI Chat + Internal GraphQL | P2 | OpenAI-powered chat widget; GraphQL as query layer; AIClient interface for provider swap |
| Public GraphQL Developer API | P3 | Promote internal GraphQL to public API for external devs (after AI chat ships) |
| Claude API as Parallel Provider | P3 | Add Claude alongside OpenAI — user picks preferred AI provider in settings |
| Revenue forecasting | P2 | ML-based prediction |
| Anomaly detection | P2 | Alert on unusual patterns |
| Stripe integration | P3 | Non-Shopify revenue |
| Native mobile app | P3 | iOS/Android standalone |
| Custom report builder | P3 | User-defined reports |
| Dark mode support | P3 | System/manual theme toggle with dark color palette |
| Home screen widgets | P3 | iOS/Android widgets for MRR, at-risk count |
| Smart search | P3 | Fuzzy matching for store names |
| Voice AI assistant | P4 | Voice commands for navigation and queries |
| Affiliate program | P4 | Referral system |
| GCP staging custom domain | P4 | Map staging.ledgerspear.com to Cloud Run URL |
| Marketing site on GCP staging | P4 | Deploy Next.js marketing to Cloud Run for staging |

---

## Completed

| Feature | Completed | Notes |
|---------|-----------|-------|
| Subscription detail view | 2026-03-01 | GET /api/v1/subscriptions/{id}, /history, /risk-timeline |
| Subscription list page | 2026-02-28 | GET /api/v1/apps/{appID}/subscriptions with filters, pagination, sorting |
| Onboarding flow (backend) | 2026-03-01 | GET /api/v1/users/onboarding-status, POST /api/v1/users/onboarding-complete |
| Config validation | 2026-03-01 | Added Validate() and HasCriticalWarnings() to config.go |
| RegisterDevice error handling | 2026-03-01 | Fixed to only ignore duplicate key errors |
| Webhook integration | 2026-03-01 | Real-time subscription updates, billing failures, app uninstalls |
| GitHub Actions CI | 2026-03-01 | Backend tests, lint, frontend tests, marketing site build |
| io.ReadAll error handling | 2026-03-01 | Verified all usages handle errors correctly |
| Repository contract clarity | 2026-03-01 | Added documentation to AppRepository interface |
| Multi-app support | 2026-03-01 | Aggregate metrics, default app preference, app selector support |

---

## Technical Debt / Code Quality

All items resolved. See "Completed" section above.

---

## Ideas (Unvalidated)

-

---

## Feature Details

### Dark Mode Support (P3)
**Added:** 2026-02-27

**Description:**
Add dark theme support with system preference detection and manual toggle.

**Proposed Features:**
- Dark color palette matching brand identity
- System theme detection (follow device settings)
- Manual toggle in settings (Light/Dark/System)
- Persist preference locally
- Smooth transition animation between themes

**Implementation:**
- Create `AppTheme.darkTheme` in `core/theme/app_theme.dart`
- Add `ThemeBloc` or use `ValueNotifier` for theme state
- Update `MaterialApp` to use `themeMode` property
- Store preference in SharedPreferences
- Add theme toggle in Profile/Settings page

**Color Considerations:**
- Dark backgrounds: grey[900], grey[850]
- Card surfaces: grey[800]
- Primary colors remain consistent
- Ensure WCAG contrast compliance
- Charts and badges need dark-mode variants

---

### Voice AI Assistant (P4)
**Added:** 2026-03-02

**Description:**
Voice-enabled assistant for hands-free navigation and queries. Users can speak commands like "Show store Acme health" or "List subscriptions at risk".

**Why P4 (Low Priority):**
- Target users (developers) prefer tapping/typing over voice
- Privacy concerns with speaking financial data aloud
- High implementation complexity vs value delivered
- Better alternatives exist (widgets, smart search, better notifications)

**Specification:**
- Full visualization and spec at `/voice-assistant` marketing page
- Prompt file: `docs/prompts/voice-assistant-flow.md`

**Proposed Features:**
- Voice capture using `speech_to_text` Flutter package
- Intent classification via Claude API or local model
- Entity extraction (store names, filters, metrics)
- Navigation via GoRouter deep links
- Fallback: Show suggestions if intent unclear

**Supported Commands:**
- "Show details of store [name]" → Subscription details
- "Store [name] health" → Health score page
- "List subscriptions at risk" → Filtered list
- "What's my MRR?" → Dashboard metrics
- "Any billing failures?" → Alerts page

**Higher Priority Alternatives:**
1. Home screen widgets (P3) - Instant access without opening app
2. Smart search (P3) - Type "acme" to find store instantly
3. Better push notifications - Proactive alerts eliminate need to ask

**Implementation Effort:** High (speech recognition, AI integration, entity matching)

---

### Phase 1: AI Chat + Internal GraphQL (P2)
**Added:** 2026-03-07
**Updated:** 2026-03-07 — OpenAI function calling first, Claude API later via AIClient interface

**Description:**
AI-powered chat widget in the Flutter dashboard + a GraphQL endpoint (`/graphql`) for direct testing and development. The GraphQL layer serves two purposes: (1) the AI chat's internal query engine, and (2) a directly-queryable endpoint to test queries, debug resolvers, and validate the schema. Uses **OpenAI function calling (gpt-4o)** to translate natural language into GraphQL queries — the same proven pattern from OpenAI Explorer. Architecture uses an `AIClient` interface so Claude API can be swapped in later without changing the chat handler, modules, or frontend.

**Why Expose GraphQL in Phase 1:**
- Developers can test queries directly without going through AI chat
- Faster debugging — run a query in the playground, see exactly what comes back
- Validates resolvers independently from the AI layer
- GraphQL playground available for all authenticated users (Firebase auth)
- Schema still internal (not a public API contract) — can iterate freely

**Specification:**
- Prompt file: `docs/prompts/ai-graphql-chat-flow.md`
- Diagram: `docs/diagrams/ai-graphql-chat.excalidraw`
- Marketing page: `/ai-query-assistant`

**Internal GraphQL Schema (Key Types):**
- `Query.subscriptions(appId, riskState, status, domain)` → `[Subscription]`
- `Query.metrics(appId)` → `Metrics` (MRR, at-risk, renewal rate, counts)
- `Query.metricsTrend(appId, months)` → `[DailySnapshot]`
- `Query.storeHealth(appId, domain)` → `StoreHealth`
- `Query.earnings(appId)` → `Earnings` (recurring, usage, one-time, refund)
- `Query.riskSummary(appId)` → `RiskSummary` (counts + at-risk subscriptions)

**Example Interactions:**
- "Which stores haven't paid in 60+ days?" → subscriptions query filtered by risk
- "What's my MRR trend?" → metrics trend query → sparkline chart
- "Is acme-shop paying?" → store health query → status card
- "Show me churned stores this month" → filtered subscriptions
- "What's the revenue impact of at-risk stores?" → risk summary with dollar amounts

**Architecture:**
- **Frontend:** `ChatBloc` (Flutter), WebSocket for streaming, markdown rendering
- **Backend:**
  - `gqlgen` GraphQL layer with `/graphql` endpoint (Firebase auth)
  - GraphQL playground at `/graphql` (GET) for testing
  - `/api/v1/chat` WebSocket endpoint (Firebase auth)
  - AIClient interface + AIProviderRegistry (provider-agnostic)
  - OpenAI function calling (gpt-4o) — default provider
  - Claude API — added later as parallel provider (user picks in settings)
- **AI Pipeline:**
  1. User message + conversation history + GraphQL schema → AIClient (OpenAI)
  2. AI calls tools via function calling → module executes GraphQL query
  3. GraphQL resolvers call existing domain services (RiskEngine, MetricsEngine, LedgerService)
  4. AI formats results as conversational response
  5. Returns: text + data tables + 2-3 follow-up suggestions

**Safety & Guardrails:**
- Read-only: no mutations (schema enforces this)
- Tenant isolation: user can only query their own apps
- Query complexity limits (max depth: 5, max fields: 50)
- Rate limiting per user (10 queries/min)
- Audit logging of all AI-generated queries
- Prompt injection protection (schema-constrained queries only)

**Implementation:**
- Add `gqlgen` to Go backend (schema-first)
- Resolvers delegate to existing domain services — no new business logic
- `/graphql` endpoint with Firebase auth (same auth as dashboard)
- GraphQL playground enabled for all authenticated users (testing/debugging)
- Write integration tests that query `/graphql` directly
- `/api/v1/chat` WebSocket with Firebase auth
- Flutter `ChatBloc` with markdown rendering, typing indicator, follow-ups

**Implementation Effort:** High (OpenAI integration, gqlgen, WebSocket, chat UI) — but OpenAI function calling is proven pattern from Explorer project

---

### Phase 2: Public GraphQL Developer API (P3)
**Added:** 2026-03-07
**Depends on:** Phase 1 (AI Chat + Internal GraphQL)

**Description:**
Promote the battle-tested internal GraphQL schema to a public, developer-facing API. Shopify app developers can query their revenue data programmatically via GraphQL, authenticated with existing API keys. This supplements the existing REST Revenue API.

**Why P3 (After AI Chat Ships):**
- Schema has been validated by real AI-generated queries in Phase 1
- Confidence that types, fields, and resolvers are correct and complete
- Public API contract is stable — fewer breaking changes
- Shopify devs already think in GraphQL (Partner API is GraphQL)
- Schema introspection = self-documenting API

**What Changes from Phase 1:**
- Add API key auth as alternative to Firebase auth (external devs use API keys)
- Public-facing rate limiting per API key
- Audit logging via existing `api_audit_log` table
- Freeze schema versioning (v1) — breaking changes require v2
- Publish API docs and integration guides
- Publish API documentation and integration guides
- Version the schema (v1)

**Implementation Effort:** Low-Medium (schema + resolvers already exist from Phase 1, just add auth + public endpoint)

---

### Phase 3: Claude API as Parallel AI Provider (P3)
**Added:** 2026-03-07
**Depends on:** Phase 1 (AI Chat with OpenAI)

**Description:**
Add Claude API as a second AI provider running alongside OpenAI. Users choose their preferred provider in chat settings. Both run in production simultaneously — not a swap, a choice.

**Why Parallel (Not Replace):**
- Different models excel at different query types
- Users have provider preferences — let them choose
- Compare response quality, latency, and cost in real usage
- No migration risk — OpenAI keeps working if Claude has issues
- Future: could auto-route based on query type (e.g., Claude for complex reasoning, OpenAI for speed)

**Implementation:**
- Implement `ClaudeClient` satisfying `AIClient` interface
- Map ToolDefinition → Claude `input_schema` format
- Map tool results → Claude `tool_result` content blocks
- Register both providers in `AIProviderRegistry`
- Add `ai_provider` field to `user_preferences` table
- Flutter: provider selector in chat settings (OpenAI / Claude toggle)
- Track per-provider metrics: latency, token cost, user satisfaction ratings
- Both providers share the same modules, registry, and GraphQL layer

**Implementation Effort:** Medium (Claude client + preferences UI, everything else reused)
