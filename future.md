# Future Features – LedgerGuard

Postponed ideas and features for later implementation.

---

## Backlog

| Feature | Priority | Notes |
|---------|----------|-------|
| Welcome & Onboarding Flow (Hybrid) | P1 | n8n + Postmark email drip + in-app checklist; custom webhook support for third-party flow builders; see `docs/prompts/welcome-onboarding-flow.md` |
| Billing System — Live mode + GST | P2 | Razorpay test mode done; need live keys, GST tax handling, invoice generation before launch |
| AI Chat + Internal GraphQL | P2 | OpenAI-powered chat widget; GraphQL as query layer; AIClient interface for provider swap |
| Public GraphQL Developer API | P3 | Promote internal GraphQL to public API for external devs (after AI chat ships) |
| Claude API as Parallel Provider | P3 | Add Claude alongside OpenAI — user picks preferred AI provider in settings |
| ~~Revenue forecasting~~ | ~~P2~~ | ~~Done — Linear regression + exponential smoothing in ForecastingEngine (2026-05-09)~~ |
| Anomaly detection | P2 | Alert on unusual patterns |
| Stripe integration | P3 | Non-Shopify revenue |
| Native mobile app | P3 | iOS/Android standalone |
| Custom report builder | P3 | User-defined reports |
| Quick Stats dashboard row | P3 | Always-visible compact stats row at top of dashboard: Active MRR, At-Risk Stores, Renewal Rate, Churned (30d) with deltas; uses existing metrics API + KpiCardCompact |
| Dark mode support | P3 | System/manual theme toggle with dark color palette |
| Home screen widgets | P3 | iOS/Android widgets for MRR, at-risk count |
| Smart search | P3 | Fuzzy matching for store names |
| Voice AI assistant | P4 | Voice commands for navigation and queries |
| Affiliate program | P4 | Referral system |
| GCP staging custom domain | P4 | Map staging.ledgerspear.com to Cloud Run URL |
| Marketing site on GCP staging | P4 | Deploy Next.js marketing to Cloud Run for staging |
| Smart shop logo fetch (active-first) | P2 | During sync, only fetch brand data for stores with ≥2 months of recurring transactions (confirmed active). Deprioritize one-time/trial stores. Avoids wasting Storefront API calls on churned/inactive domains. Current implementation fetches all domains equally. |
| Dashboard: Recent Events card (re-introduce) | P3 | The "Recent Events" card was replaced by "This Week Activity" summary table. Consider re-introducing it as a Row 4 or an expandable section below the activity summary for drill-down into individual events. |
| Dedicated Setup Wizard (fallback) | P3 | Separate `/setup` route with step-by-step flow (Connect Partner → Select App → First Sync). Redirect from dashboard until complete. Fallback if inline checklist isn't sufficient for user onboarding. |
| Flutter frontend consuming real reviews API | P2 | Replace mock reviews in Flutter with real `/api/v1/apps/{appID}/reviews` API data |
| AI-based sentiment analysis for reviews | P3 | Beyond simple rating threshold — use LLM to classify sentiment, extract themes |
| Review reply/response functionality | P3 | Allow responding to reviews from within LedgerGuard |
| Multi-source review scraping | P4 | Google Play, App Store (iOS), beyond Shopify |
| Webhook notification on new negative reviews | P2 | Detect new 1-2 star reviews and push notification to user |
| Sync scheduler fallback when Redis unavailable | P3 | When `queue.enabled: true` but Redis fails to connect, fall back to direct SyncScheduler instead of running no sync at all |
| Migration failure notification | P3 | Send notification (Slack/push) when database migrations fail on startup; currently only logged as WARNING |
| Faster periodic recovery (heartbeatTTL threshold) | P2 | Change periodic recovery threshold from `lockTTL` (2h) to `heartbeatTTL` (20 min). Currently if server hard-crashes and restarts within 20 min, job is stuck for up to 2 hours. With this change, worst case becomes ~30 min (20 min heartbeat expire + 10 min periodic tick). Safe because: grace period protects new jobs, and heartbeatTTL guarantees live workers have a heartbeat key. |
| Move event_sync to Wave 1 (remove subscription dependency) | P2 | EventProcessor currently iterates subscriptions to get shop GIDs, requiring Wave 2. The Shopify API supports fetching all events without a shop filter (`shopId` is optional). Refactor to call `FetchAppEvents(appGID, "")` once → no subscription dependency → can move to Wave 1 for faster full_sync. |
| Per-shop event sync API | P3 | Add dedicated endpoint `POST /api/v1/sync/enqueue/{appID}?type=event_sync&shop={shopGID}` to trigger event sync for a single shop. Useful for on-demand refresh after a specific store event without running full event_sync across all subscriptions. Pass shopGID via payload to EventProcessor; if set, skip iteration and fetch events for that shop only. |
| Normalize GID casing at ingestion (Select App) | P2 | Currently `PrepareProcessorContext` has a defensive fix to normalize `gid://partners/app/` → `gid://partners/App/`. The proper fix: validate and normalize the GID in the `POST /apps/select` handler at ingestion time, and run a one-time DB migration to fix existing lowercase GIDs. Remove the defensive fix from `processor_context.go` after migration. |
| Per-user rate limiting on Firebase-auth routes | P1 | Backend rate limiter only covers Revenue API (API key routes). Firebase-authenticated routes (`/api/v1/apps`, `/sync`, `/subscriptions`, etc.) have no per-user rate limiting. A leaked Firebase token or malicious script can hammer the backend. Add per-UID token-bucket middleware (e.g., 60 req/min per user) on the main authenticated router group in `router.go`. Reuse existing `InMemoryRateLimitStore` with Firebase UID as key instead of API key ID. |
| ~~Revert sync window to 12 months~~ | ~~P1~~ | ✅ Done (2026-05-09). Snapshot processor fixed to `AddDate(-1, 0, 0)`. Transaction processor still needs revert — see `transaction_processor.go:73`. |
| Revert event/status sync subscription limit | P1 | EventProcessor and StatusProcessor are capped to first 10 subscriptions for dev testing. Remove the `if len(subscriptions) > 10` blocks: `event_processor.go:69-72`, `status_processor.go:65-68`. |
| Flutter Provider → Bloc migration | P3 | The Provider prototype (`frontend-flutter/`) now uses DataLoadingMixin + DemoModeCoordinator pattern. The Bloc frontend (`frontend/app/`) already implements the same patterns properly with events/states, BlocListener for cascading loads, and get_it DI. If Provider prototype becomes primary app, migrate phased: (1) replace ChangeNotifier with Bloc per screen, (2) replace DataLoadingMixin with BlocListener on AppsBloc, (3) replace DemoModeCoordinator with DemoModeBloc events. |
| Remove `DELETE FROM shops` from ResetAppData | P1 | `shops` is a global cache shared across apps. Currently `ResetAppData` deletes all rows for dev convenience (single app). In production with multiple apps, remove the blanket delete — only clear app-specific tables. File: `admin_repository.go`. |
| Migrate from stdlib `log` to `zap` logger | P2 | Replace `log.Printf` with `zap.Logger` across the codebase. Benefits: structured fields, log levels (info/warn/error), JSON output for production, better performance. Touches many files — do as a dedicated refactor. |
| FCM push notifications in Flutter Provider (`frontend-flutter/`) | P2 | Add `firebase_messaging` dependency, request permission, get FCM token, register with `POST /api/v1/devices`, handle foreground/background messages. Backend device registration API already exists. |
| Churn-return analysis via StableDomainKey | P2 | Use `stable_domain_key` (deterministic `lg_sub_` + SHA1 of domain) to detect when a previously churned store reinstalls. Compare old subscription (soft-deleted, matched by StableDomainKey) with new subscription (different ShopifyGID) to track reinstall patterns, win-back rates, and time-to-return metrics. |
| Store lifetime metrics via StableDomainKey | P3 | Use `stable_domain_key` to compute total store lifetime value across multiple subscription cycles. Track: first-ever install date, total months active, cumulative revenue, number of reinstalls. Enables cohort analysis and lifetime value (LTV) reporting per store. |
| Revenue API: Webhooks for risk state changes | P2 | When a subscription's risk state changes (SAFE → ONE_CYCLE_MISSED → CHURNED), POST to a developer-configured webhook URL. Enables real-time reactions: show in-app payment banners, restrict premium features, trigger dunning emails. Config: `POST /api/v1/api-keys/{id}/webhooks` with URL + events filter. Delivers signed payloads (HMAC) with retry logic. |
| Revenue API: Subscription list with filters | P2 | `GET /subscriptions?risk_state=ONE_CYCLE_MISSED&status=ACTIVE&limit=50&cursor=...` — list all subscriptions with cursor pagination and filter by risk_state, status, plan_name. Currently external API only supports single/batch lookup by ID or domain. Internal API has list but it's Firebase-auth only. |
| Revenue API: Revenue summary endpoint | P3 | `GET /revenue/summary` — returns aggregate stats: active_subscriptions, mrr_cents, at_risk_count, total_usage_cents, churned_30d. Single call gives developers a dashboard-ready overview without fetching individual subscriptions. Reads from existing `daily_metrics_snapshot` table. |
| Real-time sync status via push (MQTT/WebSocket/SSE) | P3 | Replace sync polling (5s interval) with real-time push. Backend already has Redis progress tracking — publish state changes to MQTT topic, WebSocket channel, or SSE stream. Eliminates individual API calls for sync status. Options: MQTT (mqtt_client package, needs broker), WebSocket (backend already has /api/v1/chat WS endpoint — extend or add /sync/ws), SSE (simplest, one-way server-to-client over HTTP, no new infra). |
| Settings: Move all customization to sub-pages | P3 | Each settings card (Notifications, Sync, Workspace) gets its own sub-page like Dashboard. Cleaner main settings page with only navigation tiles. |
| Settings: Collapsible sections | P3 | Replace cards with `ExpansionTile` widgets. All settings stay on one page but sections collapse/expand. Saves vertical space without adding navigation. |
| Desktop sidebar: expand to 220px with grouped sections | P3 | Current desktop NavigationRail is 100px with 13 icon+label items stacked vertically (requires scrolling, tiny 11px labels). Upgrade to 220px expanded sidebar with horizontal icon+label, grouped sections (Core, Analytics, Admin), and collapse toggle to icon-only (56px). See `00-navigation.svg` wireframe. File: `frontend-flutter/lib/shell/app_shell.dart` `_buildRail()`. |

### Missing UI Features (from User Personas Analysis)

Cross-referenced 15 user personas (`docs/USER_PERSONAS.md`) with existing Flutter screens. These are features that personas need but have **no UI screen or menu** in either `frontend/app` (Bloc) or `frontend-flutter` (Provider).

**Conversion & Growth (P10: Freemium Dev, P15: Growth/Marketing)**
| Free-to-paid conversion funnel screen | P2 | New screen showing upgrade funnel: total installs → free users → trial starts → paid conversions → churned. Funnel chart + conversion rate metrics. No backend endpoint yet either — needs new query on subscription status transitions. |
| Trial expiry tracking screen | P2 | List of subscriptions approaching trial end with days remaining, auto-renewal status. Filter by "expiring in 7 days". Helps Freemium devs intervene before trials lapse. Backend: needs query on subscriptions with trial end dates. |
| Install velocity chart | P3 | Time-series chart showing installs per day/week overlaid with uninstalls. Currently only a total install count exists (`GET /apps/{appID}/install-count`). Needs historical install data aggregation. Persona P15 uses this to correlate marketing campaigns with install spikes. |

**Reporting & Export (P7: Agency, P8: Finance, P11: Investor, P15: Growth)**
| PDF/CSV export buttons on dashboard & metrics | P2 | Add export action to: dashboard KPI cards, subscription list, earnings timeline, fee breakdown, metrics by period. Use `pdf` and `csv` Flutter packages. No backend changes — frontend formats existing API data into downloadable files. Unlocks Investor (due diligence) and Finance (reconciliation) personas. |
| ~~Revenue concentration analysis screen~~ | ~~P3~~ | ✅ Done (2026-05-08). Top stores by revenue on Revenue tab (live mode). |
| ~~Historical data browser (12-month snapshots)~~ | ~~P3~~ | ✅ Done (2026-05-08). Date picker + snapshot comparison on Analytics page. |

**Notifications & Digests (P4: Notification-Only, P14: Side-Project Dev)**
| Weekly email digest toggle in notification settings | P2 | Add "Weekly Digest" option (day + time picker) alongside existing daily summary. Backend: needs new `weekly_digest_enabled` + `weekly_digest_day` fields in notification_preferences table + scheduler logic. Frontend: toggle + day picker in notification settings page. |
| Notification history / event log screen | P3 | Chronological feed of all notifications sent to the user: churn alerts, billing failures, daily summaries, install events. Backend: needs `notification_log` table. Frontend: new screen accessible from settings or bell icon. |

**Risk & Customer Success (P12: CS Manager, P13: Support Lead)**
| At-risk stores outreach list | P2 | Filtered view of stores by risk state with action buttons: "Contact Store", "View Details", "Dismiss". Groups by risk level (1-cycle missed, 2-cycles missed, churned). CS Managers use this as their daily work queue. Backend: existing subscription list API with risk_state filter. Frontend: new screen or tab in Risk section. |
| Global domain search bar | P2 | Search bar in app header to instantly look up any store by domain name. Returns subscription status, risk state, last payment, plan. Support Team Leads use this for ticket triage. Backend: existing `GET /subscriptions/status?domain=` endpoint. Frontend: search icon in AppBar → overlay search with results dropdown. |

**Dashboard Enhancements (P1: Embedded-Only, P6: AI Power, P7: Agency)**
| AI Daily Brief card on dashboard | P2 | Compact card on dashboard showing today's AI insight summary (1-2 sentences + key metric). Tap to expand or navigate to full insights page. Backend: existing `GET /apps/{appID}/insights/daily` endpoint. Frontend: new dashboard widget. Currently insights are only on a separate page. |
| Revenue by charge type breakdown | P3 | Pie/donut chart showing revenue split: Recurring vs Usage vs One-Time vs Refund. Backend: existing earnings API has charge type data. Frontend: new chart widget on dashboard or earnings page. Finance persona (P8) needs this for reconciliation. |
| Customizable dashboard widget layout | P3 | Allow users to show/hide and reorder dashboard cards (drag-and-drop or toggle list). Persist layout via existing `PUT /user/preferences/dashboard` endpoint. Agency persona (P7) managing multiple apps wants different layouts per context. |

**Reviews & Reputation (P9: Marketplace Veteran, P15: Growth)**
| Review sentiment dashboard | P3 | Beyond star ratings — show sentiment trends over time, common themes (positive/negative word clouds), comparison with competitor apps. Backend: needs AI-based sentiment analysis (future.md already has this). Frontend: new tab or enhancement to existing reviews tab in `frontend-flutter`. |

**Mobile-Specific (P5: Mobile-First)**
| FCM push notification setup flow | P2 | In-app prompt to enable push notifications, request permission, register FCM token via `POST /api/v1/devices`. Show notification preview. Already in future.md as backend item — this is the frontend counterpart. |
| Mobile home screen widgets | P3 | iOS/Android widgets showing MRR, at-risk count, last event. Already in future.md — adding persona context: Mobile-First Dev (P5) primary use case. |

**Sync & Data (P7: Agency, P8: Finance)**
| Sync jobs history page | P2 | List of recent sync jobs with status, duration, items processed, errors. Filter by job type and status. Cancel button for pending jobs. Backend: existing `GET /sync/jobs` API. Frontend: mentioned as "Future" in REQUIREMENTS.md — still not built. |
| Fee verification / reconciliation screen | P3 | Side-by-side view: LedgerGuard calculated fees vs expected Shopify payout. Highlights discrepancies. Backend: existing `GET /apps/{appID}/fees/summary` and `/fees/breakdown`. Frontend: new screen under earnings or settings. Finance persona (P8) primary use case. |

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
| Billing System (Razorpay test mode) | 2026-03-18 | Razorpay Subscriptions integration — backend (8 commits) + Flutter UI; Starter $249/mo, Pro $499/mo; hosted checkout via short_url |

---

## Technical Debt / Code Quality

All items resolved. See "Completed" section above.

---

## Ideas (Unvalidated)

- **Snapshot gap detection** — Detect missing daily snapshots and alert/backfill automatically
- **Materialized views at scale** — When query-time downsampling bottlenecks at 1000+ apps with 3+ years data, add `frequency` column to `daily_metrics_snapshot` table (DAILY/WEEKLY/MONTHLY) and pre-compute rollups during sync. See ADR-041 "Future Extensibility" section in DECISIONS.md
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

---

## Webhooks Backend API + Persistence

Add a real backend API for webhooks instead of relying on mock data. This includes:

- `webhook_events` table to store incoming webhook payloads
- `POST /api/v1/webhooks/shopify` and `POST /api/v1/webhooks/razorpay` endpoints to receive webhooks
- `GET /api/v1/apps/:appId/webhooks` list endpoint with pagination, filtering by source/status/date range
- HMAC signature verification for Shopify webhooks
- Razorpay webhook signature verification
- Webhook replay/retry mechanism for failed deliveries
- Connect WebhookProvider to live API (replace mock data path)
- **Dynamic KPI Registry via Admin Panel** — Admin page to manage available KPIs/widgets with status (active/coming_soon/requested). Users see "Coming Soon" badge on not-yet-available metrics. If a user requests a feature, admin enables it for them. Requires backend KPI definition table + admin API + dynamic frontend registry replacing the hardcoded `kAllKpis`/`kAllWidgets`.
