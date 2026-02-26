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

### 2. Architecture Diagrams
Update when architecture changes:
- `docs/C4.puml` – System context & containers
- `docs/ER.puml` – Entity relationships
- `docs/SEQUENCE.puml` – Key flows (sync, auth, etc.)

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
1. Fetch transactions from Partner API (12-month window)
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

### 10. Risk Engine (Authoritative)
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
8. Commit with message & push
    ↓
9. Log prompt to prompts.md (original + improved)
```

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
└── frontend/
    └── lib/
        ├── domain/
        ├── data/
        ├── application/
        └── presentation/
```

---

## QUICK REFERENCE

| Action | Command |
|--------|---------|
| Run tests | `go test ./... -v` |
| Run server | `go run ./cmd/server` |
| Format code | `go fmt ./...` |
| Lint | `golangci-lint run` |

---

## REMEMBER

- **Ask before assuming** – Clarify ambiguous requirements
- **Small commits** – Easier to review and revert
- **Docs are code** – Keep them in sync
- **Future.md is your friend** – Don't scope creep
