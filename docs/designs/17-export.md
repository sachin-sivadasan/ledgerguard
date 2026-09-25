# Tech Design: Data Export (as-built)

**PRD:** docs/prds/17-export.md · **Author:** sachin.s · **Date:** 2026-09-26
**Status:** Draft (backfilled from implementation)
**Reversibility:** Currently zero — the code is **unreachable** (unrouted), so it affects nothing in production. It becomes irreversible only if wired (a public export API + file contracts). Right-sized doc: this is dead code; the design's job is the wire-or-delete decision + the fixes required *before* any wiring.

> **As-built.** "Current state IS the design." Every claim cites a real path:line and was verified. Markers: **FACT / INFERRED / UNKNOWN**. **Nothing is refactored or fixed** — gaps are LISTED.

## Current state

**FACT (verified) — complete code, fully disconnected.** `NewExportHandler` + `ExportTransactions/ExportSubscriptions/ExportMetrics` exist (`export_handler.go:22,38,109,158`) over `export_service.go` (CSV/JSON for the three entity types). But: **no `export` routes in `router.go`** (grep: none), **no `ExportHandler` instantiation in `main.go`** (grep: none), **no frontend**, **no tests**. The endpoints return 404 — the feature is unreachable.

**FACT — synchronous in-memory** (`export_service.go:196,297,421` → `buf.Bytes()`): each export loads all rows and writes to a `bytes.Buffer` returned in the response — no streaming/pagination/size-cap → OOM risk at scale.

**FACT — design (if it ran):** app-scoped via `resolveAppFromRequest` (auth + app), `start`/`end` on transactions/metrics (subscriptions unfiltered), `LogExportRequest` audit, file download (`Content-Disposition`, `X-Record-Count`). Not a GDPR/compliance export (no portability/deletion/PII masking).

## Proposed design

**No redesign — documents the shipped (dead) code.** Pattern: **app-scoped synchronous entity dump to CSV/JSON**, currently unrouted. Would-be flow + the disconnection: see `docs/designs/17-export-sequence.puml` (validated `plantuml -checkonly`).

**The decision this doc forces:** either (A) **delete** the handler+service as dead code, or (B) **wire** it — which requires: register routes, instantiate in `main.go`, add a frontend entry point, add tests, and **replace in-memory buffering with streaming + a size cap / async job** before exposing it.

### Data model

**FACT — reads existing tables only** (`transactions`, `subscriptions`, `daily_metrics_snapshot`); writes an `audit_log` `ExportRequest` entry when run. No new tables, no migration.

### API & events

**FACT (defined, unrouted).** `GET /api/v1/apps/{appID}/export/{transactions|subscriptions|metrics}?format=csv|json[&start&end]`; file-download responses. **Not a live contract** (unreachable). No events.

## Alternatives considered

- **Synchronous in-memory dump vs streaming/async job.** As-built buffers everything in RAM. **Choose streaming (or a queued job producing a downloadable artifact) if** ever wired — mandatory for whale apps. The current approach is only safe for tiny datasets.
- **Dedicated bulk export vs the existing per-report CSV** (#1/#7/#12). Per-report CSV is live, tested, reachable; this bulk export is none of those. **Choose per-report CSV** unless a genuine raw-multi-entity dump need exists — then wire this with the fixes above.
- **Delete vs keep.** Keeping unrouted, untested code is a maintenance/confusion liability. **Delete** unless there's a committed near-term plan to ship it.

## Failure modes

*(Hypothetical — the code doesn't run today. Listed for the wire path.)*

| Failure | Detection | Behavior (if wired, as-built) | Recovery |
|---------|-----------|-------------------------------|----------|
| Large app export | none | **DIVERGENCE** in-memory buffer → OOM (`export_service.go:196…`) | must add streaming/cap |
| Bad date param | parse | **FACT** JSON error (`export_handler.go:239-248`) | fix params |
| App not owned | `resolveAppFromRequest` | **FACT** 404/403 (but see systemic appID gap, #14 D4) | — |
| Any run | audit hook | **FACT** `LogExportRequest` recorded | — |

### DIVERGENCE — the whole capability is a gap

- **D1. Unrouted / uninstantiated / no UI / no tests** — dead code, unreachable. → wire-or-delete.
- **D2. In-memory buffering, no streaming/size-cap/async** — OOM risk; must be fixed before any exposure.
- **D3. Not a compliance export** — if GDPR data-portability/deletion is the real need, this doesn't meet it (separate scope).
- **D4. Subscriptions unfiltered** — full dump only; no date/paging.
- **D5. Systemic appID scoping** — shares the `resolveAppFromRequest` org-cross-check gap (#14 D4) if wired.

## Test plan

**There are no tests today.** If wired:
1. **Reachability:** routed + instantiated + a smoke test hitting each export endpoint returns the right `Content-Type` + `X-Record-Count`.
2. **Scale/safety:** an export of N=100k+ rows completes within memory/time budget (forces the streaming fix, D2).
3. **Scoping:** an app not owned by the caller's org → denied (D5).
4. **Format parity:** CSV and JSON contain identical row sets.

*If deleting:* the test plan is "confirm nothing references the handler/service, then remove."

**Manual verification (today):**
1. `cd backend && grep -rn "export" internal/interfaces/http/router/router.go` → **expect no matches** (confirms unrouted).
2. `go build ./...` after any decision (delete or wire) to confirm the tree still compiles.

**External-platform reality:** none — pure DB reads. Fully unit-testable if wired.

## Pricing & policy touchpoints

**UNKNOWN — not wired.** If shipped, bulk export is a plausible paid-tier feature; GDPR export is a separate, unmet need. → Open question (Product).

## Rollout

**Nothing to roll out (dead code).** If wiring: routes + main + UI + tests + streaming, behind a flag, ideally paid-tier-gated. If deleting: remove handler + service + audit action reference; confirm build. Rollback trivial either way.

## Observability

**FACT:** `LogExportRequest` audit hook exists. **DIVERGENCE:** irrelevant until wired; if wired, add export-size/duration metrics (tied to the OOM-safety metric).

## Open questions

- **Wire-or-delete (D1)** — the core decision. Owner: Product/Eng.
- **If wired: streaming/async + size cap (D2)** before exposure. Owner: Eng.
- **Is the real need GDPR export/deletion (D3)?** — separate scope. Owner: Product/Sec.
- **Value vs per-report CSV** — justify a second export mechanism. Owner: Product.
