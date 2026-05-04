# LedgerGuard Developer Documentation

> Revenue Intelligence Platform for Shopify App Developers

---

## Foundation

| Doc | Description |
|-----|-------------|
| [Architecture Overview](architecture-overview.md) | System context, DDD layers, data pipeline, deployment topology |
| [Tech Stack](tech-stack.md) | All technologies with versions and rationale |
| [Glossary](glossary.md) | Domain terms and definitions |
| [Contributing to Docs](CONTRIBUTING-DOCS.md) | Doc template, conventions, checklist |

---

## Backend Features (01–20)

### Auth & Integration

| # | Feature | Diagram | Key ADRs |
|---|---------|---------|----------|
| [01](01-project-init-ddd-structure.md) | Project Init & DDD Structure | [Component](../diagrams/puml/01-ddd-layers-component.puml) | ADR-001, ADR-005 |
| [02](02-firebase-auth-user-management.md) | Firebase Auth & User Management | [Sequence](../diagrams/puml/02-firebase-auth-sequence.puml) | ADR-003 |
| [03](03-shopify-partner-integration.md) | Shopify Partner Integration | [Sequence](../diagrams/puml/03-shopify-partner-oauth-sequence.puml) | ADR-006 |
| [04](04-app-management-selection.md) | App Management & Selection | [Component](../diagrams/puml/04-app-management-component.puml) | — |

### Core Domain

| # | Feature | Diagram | Key ADRs |
|---|---------|---------|----------|
| [05](05-transaction-sync-engine.md) | Transaction Sync Engine | [Sequence](../diagrams/puml/05-sync-pipeline-sequence.puml) | ADR-002 |
| [06](06-ledger-rebuild-engine.md) | Ledger Rebuild Engine | [Activity](../diagrams/puml/06-ledger-rebuild-activity.puml) | ADR-002 |
| [07](07-risk-engine.md) | Risk Engine | [Activity](../diagrams/puml/07-risk-engine-activity.puml) | — |
| [08](08-metrics-engine-kpi-computation.md) | Metrics Engine & KPI Computation | [Component](../diagrams/puml/08-metrics-computation-component.puml) | — |
| [09](09-daily-snapshots.md) | Daily Snapshots | [Sequence](../diagrams/puml/09-snapshot-upsert-sequence.puml) | — |

### Extended Backend

| # | Feature | Diagram | Key ADRs |
|---|---------|---------|----------|
| [10](10-ai-insight-service.md) | AI Insight Service | [Sequence](../diagrams/puml/10-ai-insight-sequence.puml) | — |
| [11](11-notification-service.md) | Notification Service | [Sequence](../diagrams/puml/11-notification-dispatch-sequence.puml) | — |
| [12](12-subscription-management.md) | Subscription Management | [Component](../diagrams/puml/12-subscription-management-component.puml) | — |
| [13](13-revenue-api.md) | Revenue API | [Sequence](../diagrams/puml/13-revenue-api-sequence.puml) | — |
| [14](14-webhook-integration.md) | Webhook Integration | [Sequence](../diagrams/puml/14-webhook-processing-sequence.puml) | — |
| [15](15-revenue-share-fee-tracking.md) | Revenue Share & Fee Tracking | [Activity](../diagrams/puml/15-fee-calculation-activity.puml) | ADR-008 |
| [16](16-earnings-timeline.md) | Earnings Timeline | [Sequence](../diagrams/puml/16-earnings-timeline-sequence.puml) | — |

### Advanced Backend

| # | Feature | Diagram | Key ADRs |
|---|---------|---------|----------|
| [17](17-store-health-brand-data.md) | Store Health & Brand Data | [Sequence](../diagrams/puml/17-store-health-sequence.puml) | — |
| [18](18-ai-chat-graphql.md) | AI Chat & GraphQL | [Sequence](../diagrams/puml/18-ai-chat-tool-loop-sequence.puml), [Component](../diagrams/puml/18-ai-chat-modules-component.puml) | ADR-011–014 |
| [19](19-razorpay-billing.md) | Razorpay Billing | [Sequence](../diagrams/puml/19-razorpay-billing-sequence.puml) | ADR-020 |
| [20](20-app-store-review-scraper.md) | App Store Review Scraper | [Activity](../diagrams/puml/20-review-scraper-activity.puml) | — |

### Platform Operations

| # | Feature | Diagram | Key ADRs |
|---|---------|---------|----------|
| [32](32-admin-dashboard-mixpanel.md) | Admin Dashboard API + Mixpanel Tracking | [Sequence](../diagrams/puml/32-admin-dashboard-sequence.puml) | ADR-025 |

---

## Frontend Features (21–28)

Source: `ledgerguard-flutter/` (Provider + mock data prototype). ASCII diagrams inline, no PlantUML.

| # | Feature | Key Provider | Key Screen |
|---|---------|-------------|------------|
| [21](21-flutter-dashboard.md) | Flutter Dashboard | DashboardProvider | DashboardScreen |
| [22](22-flutter-subscription-ui.md) | Flutter Subscription UI | SubscriptionProvider | SubscriptionListScreen, SubscriptionDetailScreen |
| [23](23-flutter-store-health-crm.md) | Flutter Store Health & CRM | StoreProvider | StoreListScreen, StoreDetailScreen |
| [24](24-flutter-transactions-events.md) | Flutter Transactions & Events | TransactionProvider, EventsProvider | TransactionListScreen |
| [25](25-flutter-risk-analysis.md) | Flutter Risk Analysis | RiskProvider | RiskFunnelScreen |
| [26](26-flutter-analytics-suite.md) | Flutter Analytics Suite | AnalyticsProvider | AnalyticsScreen (5 tabs) |
| [27](27-flutter-ai-insights-chat.md) | Flutter AI Insights & Chat | InsightsProvider | ChatScreen |
| [28](28-flutter-settings-workspace.md) | Flutter Settings & Workspace | SettingsProvider, ApiKeyProvider | SettingsScreen |

---

## Infrastructure (29–30)

ASCII deployment topology inline, no PlantUML.

| # | Feature | Environment |
|---|---------|-------------|
| [29](29-gcp-cloud-run-staging.md) | GCP Cloud Run Staging | Staging |
| [30](30-firebase-hosting-deployment.md) | Firebase Hosting Deployment | All |

---

## Future Specs (F01–F05)

Planned features — not yet implemented.

| # | Feature | Priority |
|---|---------|----------|
| [F01](future/F01-onboarding-welcome-flow.md) | Onboarding & Welcome Flow | P1 |
| [F02](future/F02-public-graphql-api.md) | Public GraphQL API | P3 |
| [F03](future/F03-claude-ai-provider.md) | Claude AI Provider | P3 |
| [F04](future/F04-dark-mode.md) | Dark Mode | P3 |
| [F05](future/F05-revenue-forecasting-ml.md) | Revenue Forecasting (ML) | P2 |

---

## Diagrams

All PlantUML diagrams are in [`docs/diagrams/puml/`](../diagrams/puml/). Validate with:

```bash
plantuml -checkonly docs/diagrams/puml/*.puml
```

System architecture (Excalidraw): [`docs/diagrams/excalidraw/system-architecture.excalidraw`](../diagrams/excalidraw/system-architecture.excalidraw)

---

## Related Project Docs

| Doc | Purpose |
|-----|---------|
| [PRD.md](../../PRD.md) | Product requirements |
| [TAD.md](../../TAD.md) | Technical architecture |
| [DATABASE_SCHEMA.md](../../DATABASE_SCHEMA.md) | Full database schema |
| [DECISIONS.md](../../DECISIONS.md) | Architecture Decision Records |
| [IMPLEMENTATION_LOG.md](../../IMPLEMENTATION_LOG.md) | Backend feature log |
| [future.md](../../future.md) | Backlog and deferred features |
