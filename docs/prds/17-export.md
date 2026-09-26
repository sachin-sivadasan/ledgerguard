# PRD: Data Export (as-built)

**Date:** 2026-09-26 · **Author:** sachin.s · **Status:** Draft (backfilled from implementation)

> **As-built PRD**, reverse-derived from code. Markers: **FACT** / **INFERRED** / **UNKNOWN**.
> Success metrics are **PROPOSED** (future), not historical. Sources:
> `handler/export_handler.go`, `application/service/export_service.go`, `audit_service.go`,
> `router.go` (notable for what's absent). (Per-report CSV export is a different mechanism — #1/#7/#12.)

## ⚠️ Headline finding (read first)

**FACT — Export is fully-written but DEAD CODE.** `export_handler.go` + `export_service.go` implement app-scoped transaction/subscription/metrics export to CSV/JSON — but the routes are **not registered in `router.go`**, the handler is **never instantiated in `main.go`**, there is **no frontend**, and there are **no tests**. It is unreachable by any client. The real decision this doc exists to force: **wire it (with fixes) or delete it.**

## Problem & evidence

Partners may want a bulk, app-level dump of their raw data (transactions, subscriptions, daily metrics) for their own spreadsheets/BI/accountant — beyond the per-report CSV that report screens already offer.

- **FACT:** the code to do this exists and is sensibly designed (CSV/JSON formatting, audit hooks).
- **UNKNOWN — was this ever a real requirement, or speculative scaffolding?** No evidence recorded; its unrouted state suggests it never shipped. Open question — owner: Product.

## Target users

**INFERRED:** a partner/analyst wanting raw app data offline. Per-app, per-org.
- **Not the target:** merchants; GDPR data-subject requests (this is *not* a compliance/portability tool — no user-data export, no deletion, no PII masking).

## Proposed solution (as-built behavior — currently unreachable)

**FACT — three defined endpoints** (not routed): `GET /api/v1/apps/{appID}/export/{transactions|subscriptions|metrics}?format=csv|json` (transactions/metrics take `start`/`end`; subscriptions unfiltered). Each returns a **file download** (`Content-Disposition: attachment`, `X-Record-Count`), scoped via `resolveAppFromRequest`, and calls `LogExportRequest` for audit.

**FACT — synchronous, in-memory.** Rows are collected into a `bytes.Buffer` and returned in the response body — **no streaming, pagination, size cap, or async job** → OOM risk on large apps (100k txns ≈ 15MB; millions → GB).

**FACT — distinct from per-report CSV** (#1/#7/#12): those are per-report, custom-columned, live, and reachable; this is coarse raw-entity dumps. Genuinely separate, but overlapping in user value.

### Key screens

**No UI exists** — this is backend-only *and unrouted*. Its absence from both `router.go` and the frontend is the finding.

## Platform & policy constraints

**FACT.** Would depend on nothing external (pure DB reads). Audit-logs each request (`ExportRequest` action). No compliance/GDPR framing despite the audit hook.

## Pricing-tier impact

**UNKNOWN — not wired, so no gating exists.** If shipped, bulk export is a plausible paid-tier feature. Owner: Product.

## Migration for existing users

**INFERRED — N/A.** Read-only; nothing to migrate.

## Success metrics (PROPOSED — future-facing, only if wired)

1. **Reachability:** if kept, it's routed + has a UI + tests (today: none of the three).
2. **Safety at scale:** a whale-app export completes without OOM — requires streaming (today it buffers in memory).
3. **Value vs per-report CSV:** ≥ X% of exports use the bulk endpoint over per-report CSV — else delete it as redundant.

## Non-goals (deliberately absent in code)

1. **Not a GDPR/compliance export** — no user-data portability, deletion, or PII masking. **FACT.**
2. **No async/streaming/size-cap** — synchronous in-memory only. **FACT.**
3. **No UI and no routing** — unreachable. **FACT.**
4. **No date filter on subscriptions** — full dump only. **FACT.**

## Open questions

- **Wire-or-delete decision** — the code is unrouted, untested, UI-less. Ship it (with the fixes below) or remove it as dead code? Owner: Product/Eng. **(the point of this doc)**
- **If wired: fix the OOM risk** — stream rows / cap size / make it an async job before exposing it. Owner: Eng.
- **Overlap with per-report CSV** — does bulk export add enough value to justify a second export mechanism? Owner: Product.
- **Compliance:** if the real need is GDPR data-export/deletion, this doesn't meet it — separate scope. Owner: Product/Sec.
