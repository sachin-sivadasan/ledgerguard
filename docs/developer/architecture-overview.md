# Architecture Overview

LedgerGuard is a Revenue Intelligence Platform for Shopify App Developers. It monitors subscription health, detects churn risk, and provides AI-powered revenue insights.

---

## System Context

```
┌─────────────────────────────────────────────────────────┐
│                    LedgerGuard System                    │
│                                                         │
│  ┌──────────┐    ┌──────────┐    ┌──────────────────┐  │
│  │ Flutter   │    │ Next.js  │    │ Go Backend       │  │
│  │ Frontend  │───▶│ Marketing│    │ (Modular Monolith│  │
│  │ (Web/iOS/ │    │ Site     │    │  + DDD)          │  │
│  │  Android) │    └──────────┘    │                  │  │
│  └─────┬─────┘                    │  ┌────────────┐  │  │
│        │ REST + WebSocket         │  │ PostgreSQL │  │  │
│        └─────────────────────────▶│  └────────────┘  │  │
│                                   └──────────────────┘  │
└────────────────────────┬────────────────────────────────┘
                         │
            ┌────────────┼────────────┐
            ▼            ▼            ▼
     ┌────────────┐ ┌─────────┐ ┌─────────┐
     │ Shopify    │ │ Firebase│ │ OpenAI  │
     │ Partner API│ │ Auth    │ │ GPT-4o  │
     └────────────┘ └─────────┘ └─────────┘
```

---

## Architecture Style

**Backend:** Go modular monolith with Domain-Driven Design (DDD). See [ADR-001](../../DECISIONS.md) and [ADR-005](../../DECISIONS.md).

**Frontend:** Flutter prototype using Provider pattern (at `ledgerguard-flutter/`). A Bloc-based version exists at `frontend/app/` but the prototype is the primary reference.

**Marketing:** Next.js 14 with TailwindCSS, static public site.

---

## DDD Layer Structure

```
┌─────────────────────────────────────────────┐
│              Interfaces (HTTP)               │
│   handlers, middleware, router               │
├─────────────────────────────────────────────┤
│           Application (Use Cases)            │
│   sync_service, billing_service, DTOs        │
├─────────────────────────────────────────────┤
│              Domain (Core Logic)             │
│   entities, value objects, domain services,  │
│   repository interfaces (ports)              │
├─────────────────────────────────────────────┤
│          Infrastructure (Adapters)           │
│   PostgreSQL repos, Firebase, Shopify,       │
│   OpenAI, Razorpay clients                   │
└─────────────────────────────────────────────┘
```

**Dependency Rule:** Each layer depends only on layers below it. The Domain layer has zero external dependencies — pure Go.

| Layer | Package Path | Depends On |
|-------|-------------|------------|
| Domain | `internal/domain/` | Nothing |
| Application | `internal/application/` | Domain |
| Infrastructure | `internal/infrastructure/` | Domain, Application |
| Interfaces | `internal/interfaces/` | Application |

---

## Core Data Pipeline

The heart of LedgerGuard is a deterministic sync-rebuild pipeline:

```
1. FETCH     Shopify Partner API → raw transactions (12-month window)
2. STORE     Immutable transaction records in PostgreSQL
3. REBUILD   Full ledger recalculation from raw data
4. CLASSIFY  Risk engine assigns states (SAFE → CHURNED)
5. COMPUTE   Metrics engine calculates KPIs (MRR, renewal rate)
6. SNAPSHOT  Daily metrics snapshot (upsert, never delete)
7. INSIGHT   AI generates daily revenue brief (Pro tier)
8. NOTIFY    Alerts via email, Slack, in-app
```

This pipeline is **idempotent** — same input always produces the same output. There are no incremental updates in the current implementation; every sync rebuilds from scratch. See [ADR-002](../../DECISIONS.md).

---

## Authentication Flow

```
Flutter App                    Go Backend                   Firebase
    │                              │                            │
    │──── Login (Google/Email) ───▶│                            │
    │                              │                            │
    │◀─── Firebase ID Token ───────│                            │
    │                              │                            │
    │──── API Request + Bearer ───▶│                            │
    │                              │──── Verify Token ─────────▶│
    │                              │◀─── User Claims ───────────│
    │                              │                            │
    │◀─── JSON Response ──────────│                            │
```

Stateless JWT verification on every request. No server-side sessions. See [ADR-003](../../DECISIONS.md).

---

## Deployment Topology

| Environment | Backend | Frontend | Database |
|-------------|---------|----------|----------|
| Development | `localhost:8080` | `localhost` (Flutter) | Local PostgreSQL |
| Staging | GCP Cloud Run | Firebase Hosting | Cloud SQL PostgreSQL |
| Production | Hetzner VPS + Caddy | Firebase Hosting | Self-hosted PostgreSQL |

See [29. GCP Cloud Run Staging](29-gcp-cloud-run-staging.md) and [30. Firebase Hosting](30-firebase-hosting-deployment.md) for details.

---

## Key Design Decisions

| ADR | Decision | Rationale |
|-----|----------|-----------|
| 001 | Modular monolith | Small team, rapid iteration, extract later |
| 002 | Full ledger rebuild | Deterministic, debuggable, audit-friendly |
| 003 | Firebase Auth | Fast setup, Google OAuth included, stateless |
| 004 | PostgreSQL | ACID, JSON support, mature ecosystem |
| 005 | DDD over Clean Architecture | Better domain isolation, explicit modeling |
| 011 | gqlgen for internal GraphQL | Schema-first, type-safe, AI chat query layer |
| 020 | Razorpay for billing | Available in India (Stripe is invite-only) |

Full list: [DECISIONS.md](../../DECISIONS.md)

---

## Revenue Classification

LedgerGuard strictly separates revenue types:

| Type | Code | Description | Included in MRR? |
|------|------|-------------|-------------------|
| Recurring | `RECURRING` | Monthly/annual subscriptions | Yes |
| Usage | `USAGE` | Usage-based billing | No |
| One-Time | `ONE_TIME` | Setup fees, add-ons | No |
| Refund | `REFUND` | Negative adjustments | No |

MRR = RECURRING revenue only. Usage revenue is tracked separately.

---

## Risk Classification

Subscriptions are classified into risk states based on billing cycle analysis:

| State | Condition | Meaning |
|-------|-----------|---------|
| `SAFE` | Status ACTIVE or ≤30 days late | Healthy |
| `ONE_CYCLE_MISSED` | 31–60 days late | At risk |
| `TWO_CYCLE_MISSED` | 61–90 days late | High risk |
| `CHURNED` | >90 days late | Lost revenue |

See [07. Risk Engine](07-risk-engine.md) for implementation details.

---

## Related Docs

- [Tech Stack](tech-stack.md) — versions and rationale
- [Glossary](glossary.md) — domain terms
- [00. Index](00-index.md) — master feature index
- [TAD.md](../../TAD.md) — full technical architecture document
- [PRD.md](../../PRD.md) — product requirements
