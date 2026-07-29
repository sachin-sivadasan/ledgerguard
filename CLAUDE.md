# LedgerGuard – Development Directive

You are building **LedgerGuard**, a Revenue Intelligence Platform for Shopify App Developers.

---

## MANDATORY RULES

### 1. Documentation First
Update these files when relevant changes occur:

| File | Update When |
|------|-------------|
| `PRD.md` | Requirements change |
| `TAD.md` | Architecture decisions |
| `DATABASE_SCHEMA.md` | Schema changes |
| `DECISIONS.md` | Any non-trivial technical choice |
| `TEST_PLAN.md` | New test scenarios |
| `prompts.md` | Log every prompt executed |
| `future.md` | Postponed features/ideas |
| `verification.md` | Append verification checklist from every plan |
| `IMPLEMENTATION_LOG.md` | Backend feature implementations |
| `marketing/REQUIREMENTS.md` | Marketing site changes |
| `frontend/REQUIREMENTS.md` | Frontend app changes |
| `frontend/prompts.md` | Frontend-specific prompts |
| `frontend/IMPLEMENTATION_LOG.md` | Frontend feature implementations |
| `docs/GCP_SETUP_LOG.md` | GCP/infrastructure commands run (for learning) |
| `docs/FIREBASE_HOSTING_SETUP_LOG.md` | Firebase Hosting commands run (for learning) |

### 2. Architecture Diagrams
Update when architecture changes:
- `docs/C4.puml` – System context & containers
- `docs/ER.puml` – Entity relationships
- `docs/SEQUENCE.puml` – Key flows (sync, auth, etc.)

**Diagram Audit:** When the user says **"run diagram audit"**, execute the full prompt at `docs/prompts/diagram-audit-prompt.md` — audit all diagrams (PlantUML, Excalidraw, sequence), update outdated ones, create missing ones (backend + frontend-flutter), and log to `prompts.md`.

### 3. Domain-Driven Design (Go)
```
cmd/server/main.go                    → Entry point only
internal/domain/entity/               → Domain entities (User, Subscription, Transaction)
internal/domain/valueobject/          → Value objects (Money, RiskState, ChargeType)
internal/domain/service/              → Domain services (RiskEngine, MetricsEngine)
internal/domain/repository/           → Repository interfaces (ports)
internal/application/service/         → Application services (use cases)
internal/application/dto/             → Data transfer objects
internal/infrastructure/config/       → Configuration
internal/infrastructure/persistence/  → Database implementations (adapters)
internal/infrastructure/external/     → External service clients (Firebase, Shopify)
internal/interfaces/http/handler/     → HTTP handlers
internal/interfaces/http/middleware/  → HTTP middleware
internal/interfaces/http/router/      → Route definitions
pkg/                                  → Shared utilities
```

**Dependency Rule:** Outer layers depend on inner. Domain has ZERO external dependencies.

### 4. TDD – Test-Driven Development
```
1. Write failing test
2. Write minimal code to pass
3. Refactor
4. Commit
```
Never skip tests. Run `go test ./...` before every commit.

### 5. Incremental Development
- **One feature per commit**
- Commit messages: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`
- Do NOT implement multiple major modules in one step
- If scope creeps, log to `future.md` and continue

### 6. Architecture Changes
**Always confirm before:**
- Adding new external dependencies
- Changing database schema
- Modifying core domain entities
- Altering sync/rebuild logic

### 7. Revenue Classification (Strict Separation)
```go
type ChargeType string
const (
    ChargeTypeRecurring ChargeType = "RECURRING"  // Monthly/annual subscriptions
    ChargeTypeUsage     ChargeType = "USAGE"      // Usage-based billing
    ChargeTypeOneTime   ChargeType = "ONE_TIME"   // Setup fees, add-ons
    ChargeTypeRefund    ChargeType = "REFUND"     // Negative adjustments
)
```
- Never mix RECURRING and USAGE calculations
- MRR = RECURRING only
- Usage Revenue = USAGE only

### 8. Ledger Rebuild Strategy
```
Every sync:
1. Fetch transactions from Partner API (full history from `SyncHistoryStart`; catch-up syncs fetch a small `LookbackDays` delta)
2. Store raw transactions (immutable)
3. Rebuild entire ledger from scratch
4. Recalculate all risk states
5. Compute KPIs
6. Store daily snapshot
```
- **Deterministic:** Same input → Same output
- **Idempotent:** Safe to re-run
- No incremental updates in MVP

### 9. Daily Snapshots
```sql
INSERT INTO daily_metrics_snapshot (app_id, date, ...)
ON CONFLICT (app_id, date) DO UPDATE SET ...
```
- One snapshot per app per day
- **Never delete** – permanent audit trail
- Used for trends, AI insights, reconciliation

### 10. Marketing Site (Next.js)
```
marketing/
├── REQUIREMENTS.md           → Site requirements and copy
└── site/                     → Next.js 14+ App Router
    ├── app/                  → Pages and layouts
    ├── components/           → Reusable UI components
    └── public/               → Static assets
```
- **TailwindCSS** for styling
- **Minimal professional design** – clean, focused
- **No authentication** – public landing page only
- **Responsive** – mobile-first approach

**Visualization Pages:**
- Use `docs/prompts/PROMPT-interactive-visualization-chat.md` as template for new visualizations
- Available pages: `/kpi-guide`, `/money-flow`, `/architecture`, `/api-guide`, `/affiliate-program`, `/notifications`, `/pitch`, `/voice-assistant`, `/hetzner-infrastructure`, `/gcp-staging`, `/billing-flow`
- Each visualization has: animated flow diagrams, step-by-step progression, reference cards

**Visualization Prompt Requirements:**
- **Always create a separate prompt file** in `docs/prompts/` for each visualization page
- Prompt file naming: `<page-name>-flow.md` or `<page-name>-visualization.md`
- Include: context, audience, flow diagrams (ASCII), data definitions, technical requirements
- **Log to `prompts.md`** with reference to the prompt file

| Page | Prompt File |
|------|-------------|
| `/kpi-guide` | `docs/prompts/kpi-metrics-visualization.md` |
| `/money-flow` | `docs/prompts/shopify-money-flow-diagram.md` |
| `/architecture` | `docs/prompts/internal-architecture-flow.md` |
| `/api-guide` | `docs/prompts/ledgerguard-api-integration.md` |
| `/affiliate-program` | `docs/prompts/affiliate-program-flow.md` |
| `/notifications` | `docs/prompts/notification-engine-flow.md` |
| `/pitch` | `docs/prompts/customer-pitch-ui.md` |
| `/voice-assistant` | `docs/prompts/voice-assistant-flow.md` |
| `/hetzner-infrastructure` | `docs/prompts/hetzner-infrastructure-visualization.md` |
| `/gcp-staging` | `docs/prompts/gcp-staging-visualization.md` |
| `/billing-flow` | `docs/prompts/billing-system-flow.md` |

### 11. Frontend App

> **⚠️ LIVE FRONTEND = `frontend-flutter/` (Provider), NOT `frontend/app/` (Bloc).**
> As of 2026-07-26, the actively-developed **`frontend-flutter/` (Provider, ChangeNotifier)** app is the one deployed to Firebase Hosting (`ledgerguard-c7557`) at **app.ledgerspear.com** / ledgerguard-c7557.web.app, pointed at the Hetzner backend (`api.ledgerspear.com`).
> - **Do all Flutter UI work in `frontend-flutter/`.**
> - Config: API URL via `--dart-define=API_BASE_URL=...` (see `frontend-flutter/lib/core/config/app_config.dart`); Firebase via `lib/firebase_options.dart`.
> - Deploy: `cd frontend-flutter && flutter build web --release --dart-define=API_BASE_URL=https://api.ledgerspear.com && firebase deploy --only hosting`.
> - `frontend/app/` (Bloc) below is **stale/not deployed** — a possible future migration target, don't touch unless asked.

#### 11a. `frontend/app/` — Bloc version (stale, migration target — reference only)
```
frontend/
├── REQUIREMENTS.md           → App requirements, screens, flows
├── prompts.md                → Frontend-specific prompts log
├── docs/
│   └── SCREENS.puml          → Screen flow diagram
└── app/                      → Flutter application
    └── lib/
        ├── core/
        │   ├── config/       → Environment configs (dev/prod)
        │   ├── constants/    → App constants
        │   ├── theme/        → App theme
        │   └── utils/        → Utilities
        ├── data/
        │   ├── datasources/  → API clients, local storage
        │   ├── models/       → JSON serializable models
        │   └── repositories/ → Repository implementations
        ├── domain/
        │   ├── entities/     → Business entities
        │   ├── repositories/ → Repository interfaces
        │   └── usecases/     → Business logic
        └── presentation/
            ├── blocs/        → Bloc state management
            ├── pages/        → Screen widgets
            ├── widgets/      → Reusable components
            └── router/       → GoRouter configuration
```

**Architecture:**
- **Clean Architecture** with domain, data, presentation layers
- **Bloc** for state management (events → states)
- **GoRouter** for navigation
- **get_it** for dependency injection
- **Firebase Auth** for authentication

**Bloc Pattern:**
```dart
// Events
abstract class AppsEvent {}
class LoadApps extends AppsEvent {}
class RefreshApps extends AppsEvent {}

// States
abstract class AppsState {}
class AppsInitial extends AppsState {}
class AppsLoading extends AppsState {}
class AppsLoaded extends AppsState {
  final List<App> apps;
  AppsLoaded(this.apps);
}
class AppsError extends AppsState {
  final String message;
  AppsError(this.message);
}

// Bloc
class AppsBloc extends Bloc<AppsEvent, AppsState> {
  final GetAppsUseCase getAppsUseCase;

  AppsBloc(this.getAppsUseCase) : super(AppsInitial()) {
    on<LoadApps>(_onLoadApps);
  }

  Future<void> _onLoadApps(LoadApps event, Emitter<AppsState> emit) async {
    emit(AppsLoading());
    try {
      final apps = await getAppsUseCase.execute();
      emit(AppsLoaded(apps));
    } catch (e) {
      emit(AppsError(e.toString()));
    }
  }
}
```

**Testing:**
- Unit tests for Blocs, UseCases, Repositories
- Widget tests for UI components
- Run: `flutter test`

### 12. AI Chat + Internal GraphQL
```
backend/internal/chat/
├── module.go                         → Module interface
├── registry.go                       → Module registry (register, route, build prompt)
├── types.go                          → ToolDefinition, ToolCall, ToolResult, ChatMessage
├── handler.go                        → Chat WebSocket handler + /chat/modules endpoint
├── ai_client.go                      → AIClient interface (provider-agnostic)
├── ai_provider_registry.go           → AIProviderRegistry (OpenAI, Claude, etc.)
├── graphql_executor.go               → Thread-safe gqlgen executor wrapper
├── graphql/
│   ├── schema.graphql                → Internal GraphQL schema
│   ├── gqlgen.yml                    → gqlgen config
│   ├── generated.go                  → Generated runtime (gqlgen)
│   ├── models_gen.go                 → Generated models (gqlgen)
│   ├── resolver.go                   → Root resolver with injected repos/services
│   ├── schema.resolvers.go           → Resolver implementations
│   └── handler.go                    → HTTP handler (GET=playground, POST=execute)
└── modules/
    ├── risk/                         → Risk module (3 tools)
    ├── subscriptions/                → Subscriptions module (4 tools)
    ├── metrics/                      → Metrics module (3 tools)
    ├── store_health/                 → Store health module (2 tools)
    ├── earnings/                     → Earnings module (2 tools)
    └── sync/                         → Sync module (2 tools)
```

**Commands:**
| Action | Command |
|--------|---------|
| Generate GraphQL | `cd backend && go generate ./internal/chat/graphql/` |
| Run chat tests | `cd backend && go test ./internal/chat/... -v` |

**Endpoints:**
| Endpoint | Auth | Description |
|----------|------|-------------|
| `/graphql` (GET) | Firebase | GraphQL Playground |
| `/graphql` (POST) | Firebase | Execute GraphQL query |
| `/api/v1/chat` | Firebase (WebSocket) | AI chat via WebSocket |
| `/api/v1/chat/modules` | Firebase | List available modules + tools |

**Module Tool Naming:** `module__tool_name` (double underscore, OpenAI-compliant)

### 13. Risk Engine (Authoritative)
```go
func ClassifyRisk(status string, expectedNextCharge time.Time, now time.Time) RiskState {
    if status == "ACTIVE" {
        return RiskSafe
    }
    daysLate := int(now.Sub(expectedNextCharge).Hours() / 24)
    switch {
    case daysLate <= 30:
        return RiskSafe
    case daysLate <= 60:
        return RiskOneCycleMissed
    case daysLate <= 90:
        return RiskTwoCycleMissed
    default:
        return RiskChurned
    }
}
```

---

## WORKFLOW

```
User Prompt
    ↓
1. IMPROVE PROMPT – Correct/enhance the prompt for clarity and completeness
    ↓
2. COMMIT & PUSH FIRST – Commit and push any pending changes before starting new work
    ↓
3. Clarify scope (if still ambiguous)
    ↓
4. Update relevant docs
    ↓
5. Write tests (TDD)
    ↓
6. Implement minimal code
    ↓
7. Run tests
    ↓
8. PRE-COMMIT CHECKLIST (see below)
    ↓
9. Commit with message & push
    ↓
10. Log prompt to prompts.md (original + improved)
```

### Pre-Commit Checklist
**Before every commit, verify these are updated if relevant:**

- [ ] `TEST_PLAN.md` – New test scenarios added?
- [ ] `DATABASE_SCHEMA.md` – New migrations documented?
- [ ] `IMPLEMENTATION_LOG.md` – Feature summary added?
- [ ] `docs/ER_current.puml` – New entities/relationships?
- [ ] `docs/SEQUENCE_current.puml` – New flows?
- [ ] `docs/C4_current.puml` – New containers/components?
- [ ] `prompts.md` – Prompt logged?
- [ ] `frontend/REQUIREMENTS.md` – Frontend changes documented?
- [ ] `frontend/prompts.md` – Frontend prompt logged?
- [ ] `frontend/docs/SCREENS.puml` – Screen flow updated?
- [ ] `verification.md` – Plan verification checklist appended?

### Prompt Improvement Rule
Before executing any user prompt:
1. Show the **original prompt**
2. Show the **improved prompt** (clearer, more specific, fills in obvious gaps)
3. Execute the improved version

### Commit Before New Work Rule
- Always commit and push pending changes BEFORE starting a new prompt implementation
- This ensures clean git history with one feature per commit
- Never mix work from different prompts in the same commit

---

## FILE STRUCTURE

```
ledgerguard/
├── PRD.md
├── TAD.md
├── DATABASE_SCHEMA.md
├── DECISIONS.md
├── TEST_PLAN.md
├── prompts.md
├── future.md
├── docs/
│   ├── C4.puml
│   ├── ER.puml
│   └── SEQUENCE.puml
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── domain/
│   │   │   ├── entity/
│   │   │   ├── valueobject/
│   │   │   ├── service/
│   │   │   └── repository/
│   │   ├── application/
│   │   │   ├── service/
│   │   │   └── dto/
│   │   ├── infrastructure/
│   │   │   ├── config/
│   │   │   ├── persistence/
│   │   │   └── external/
│   │   └── interfaces/
│   │       └── http/
│   │           ├── handler/
│   │           ├── middleware/
│   │           └── router/
│   ├── pkg/
│   └── migrations/
├── marketing/
│   ├── REQUIREMENTS.md
│   └── site/                 → Next.js marketing site
│       ├── app/
│       ├── components/
│       └── public/
└── frontend/
    ├── REQUIREMENTS.md       → App requirements, screens, flows
    ├── prompts.md            → Frontend-specific prompts
    ├── docs/
    │   └── SCREENS.puml      → Screen flow diagram
    └── app/                  → Flutter application
        └── lib/
            ├── core/         → Constants, themes, utils
            ├── data/
            │   ├── datasources/
            │   ├── models/
            │   └── repositories/
            ├── domain/
            │   ├── entities/
            │   ├── repositories/
            │   └── usecases/
            └── presentation/
                ├── blocs/
                ├── pages/
                └── widgets/
```

---

## QUICK REFERENCE

### Backend (Go)
| Action | Command |
|--------|---------|
| Run tests | `go test ./... -v` |
| Run server | `go run ./cmd/server -config config.local.yaml` |
| Format code | `go fmt ./...` |
| Lint | `golangci-lint run` |

### Marketing Site (Next.js)
| Action | Command |
|--------|---------|
| Dev server | `cd marketing/site && npm run dev` |
| Build | `cd marketing/site && npm run build` |
| Lint | `cd marketing/site && npm run lint` |

### Frontend App (Flutter) — LIVE app is `frontend-flutter/` (Provider)
| Action | Command |
|--------|---------|
| Run tests | `cd frontend-flutter && flutter test` |
| Run app | `cd frontend-flutter && flutter run --dart-define=API_BASE_URL=http://localhost:8080` |
| Build web (live → Hetzner) | `cd frontend-flutter && flutter build web --release --dart-define=API_BASE_URL=https://api.ledgerspear.com` |
| Deploy Firebase | `cd frontend-flutter && firebase deploy --only hosting` |
| Build APK | `cd frontend-flutter && flutter build apk --dart-define=API_BASE_URL=https://api.ledgerspear.com` |
| Analyze | `cd frontend-flutter && flutter analyze` |

_(Legacy Bloc app `frontend/app/` — not deployed; use only if explicitly asked.)_

### GCP Staging Deployment
| Action | Command |
|--------|---------|
| Deploy backend | `./scripts/gcp-deploy.sh ledgerspear` |
| Check logs | `gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=ledgerspear-api AND textPayload:*" --project=ledgerspear --limit=30 --format="table(timestamp,textPayload)"` |
| Health check | `curl -sf https://ledgerspear-api-ineifpjrdq-uc.a.run.app/health` |

**IMPORTANT: Docker builds for Cloud Run MUST use `--platform linux/amd64`** (dev machine is Apple Silicon ARM). Without this flag, the binary gets `exec format error` at runtime. The `gcp-deploy.sh` script handles this automatically.

### Environment URLs
| Environment | Frontend | Backend API |
|-------------|----------|-------------|
| Dev | `localhost` | `http://localhost:8080` |
| Staging | `https://app.ledgerspear.com` (custom domain) / `ledgerguard-c7557.web.app` | `https://api.ledgerspear.com` (Hetzner, co-hosted) |
| Production | TBD | `https://api.ledgerspear.com` (Hetzner) |

> **Backend moved off GCP → Hetzner (2026-07-26).** GCP (`ledgerspear`: Cloud Run + Cloud SQL + VPC connector, ~₹2000/mo) was decommissioned. Backend now runs in Docker Compose on the shared Hetzner VPS via `deploy/cohost/` (see ADR-043 / `docs/HETZNER_MIGRATION_PLAN.md`). Deploy = SSH to box, `git pull`, `docker compose --env-file .env -f deploy/cohost/docker-compose.cohost.yml up -d --build`. `scripts/gcp-deploy.sh` is retired.

---

## Reusable Blueprints

| Blueprint | File | Description |
|-----------|------|-------------|
| Sync Pipeline | `docs/blueprints/SYNC_PIPELINE_BLUEPRINT.md` | Language-agnostic async job pipeline with distributed locking, wave orchestration, crash recovery, and 14 production pitfalls. Drop into any project's CLAUDE.md. |

---

## REMEMBER

- **Ask before assuming** – Clarify ambiguous requirements
- **Small commits** – Easier to review and revert
- **Docs are code** – Keep them in sync
- **Future.md is your friend** – Don't scope creep
- **Log all infra commands** – Every GCP/Firebase/deployment command goes in `docs/GCP_SETUP_LOG.md` or `docs/FIREBASE_HOSTING_SETUP_LOG.md`
- **Check dirty migrations after deploy** – If Cloud Run logs show `Dirty database version N`, follow §12 in `docs/GCP_SETUP_LOG.md` to fix
- **Always disable Cloud SQL public IP** after direct DB access — it's private-only by default
