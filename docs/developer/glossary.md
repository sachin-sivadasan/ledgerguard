# Glossary

Domain terms used throughout the LedgerGuard codebase and documentation.

---

## Revenue & Billing

| Term | Definition |
|------|-----------|
| **MRR** | Monthly Recurring Revenue. Sum of all ACTIVE subscription amounts (RECURRING charges only). |
| **Revenue at Risk** | MRR from subscriptions in ONE_CYCLE_MISSED or TWO_CYCLES_MISSED states. |
| **Renewal Success Rate** | (Renewed subscriptions / Expected renewals) × 100. Core KPI. |
| **Charge Type** | Classification of a transaction: `RECURRING`, `USAGE`, `ONE_TIME`, or `REFUND`. |
| **Revenue Share Tier** | Shopify's fee percentage on app earnings. Tiers: `SMALL_DEV_0` (0% on first $1M), `SMALL_DEV_15` (15%), `DEFAULT_20` (20%), `LARGE_DEV_15` (15% for large devs). |
| **Gross Amount** | Transaction amount before Shopify fees. |
| **Net Amount** | Transaction amount after Shopify fees (what the developer receives). |
| **Earnings Status** | Availability of funds: `PENDING` (7-day hold), `AVAILABLE`, `PAID_OUT`. |
| **Billing Interval** | Subscription billing frequency: `MONTHLY` or `ANNUAL`. |

## Risk States

| Term | Definition |
|------|-----------|
| **SAFE** | Subscription is ACTIVE or ≤30 days past expected charge date. Healthy. |
| **ONE_CYCLE_MISSED** | 31–60 days past expected charge. At risk — one billing cycle missed. |
| **TWO_CYCLES_MISSED** | 61–90 days past expected charge. High risk — two billing cycles missed. |
| **CHURNED** | >90 days past expected charge. Revenue is considered lost. |
| **Days Past Due** | Number of days since the expected next charge date. Drives risk classification. |

## Subscription States

| Term | Definition |
|------|-----------|
| **ACTIVE** | Subscription is currently billing. Always classified as SAFE risk. |
| **CANCELLED** | Merchant cancelled the subscription. May still be in grace period. |
| **FROZEN** | Billing paused (usually by Shopify due to store issues). |
| **EXPIRED** | Subscription term ended without renewal. |
| **PENDING** | Subscription created but not yet confirmed/charged. |

## Sync & Ledger

| Term | Definition |
|------|-----------|
| **Sync** | Process of fetching transactions from Shopify Partner API and rebuilding the ledger. |
| **Ledger Rebuild** | Deterministic recalculation of all subscriptions, risk states, and metrics from raw transactions. Same input always produces the same output. |
| **Daily Snapshot** | Immutable record of all KPIs for one app on one date. Upserted (never deleted). |
| **Rolling Window** | 12-month period of transaction history used for ledger rebuild. |
| **Backfill** | Initial sync that fetches all historical data (up to 12 months) for a newly connected app. |
| **Idempotent** | Safe to re-run — produces the same result regardless of how many times executed. |

## Architecture

| Term | Definition |
|------|-----------|
| **DDD** | Domain-Driven Design. Architecture pattern where business logic lives in the domain layer with zero external dependencies. |
| **Port** | Repository interface defined in the domain layer. Describes what data access is needed. |
| **Adapter** | Repository implementation in the infrastructure layer. Provides the actual database access. |
| **Modular Monolith** | Single deployable unit with clear module boundaries. Can be split into microservices later. |
| **ADR** | Architecture Decision Record. Documented in `DECISIONS.md`. |

## Entities

| Term | Definition |
|------|-----------|
| **App** | A Shopify app being tracked. Belongs to a Partner Account. Has transactions, subscriptions, snapshots. |
| **Partner Account** | A Shopify Partner API connection. Stores encrypted access token. Belongs to a User. |
| **User** | A LedgerGuard user account. Linked to Firebase Auth via `firebase_uid`. |
| **Shop** | A Shopify store (merchant). Identified by `myshopify_domain`. Has brand data (logo, name, country). |
| **Subscription** | A merchant's subscription to a Shopify app. Tracked for risk analysis and MRR. |
| **Transaction** | A financial event from the Partner API. Immutable source of truth. |
| **Subscription Event** | A lifecycle event for a subscription (activated, cancelled, renewed, etc.). |
| **Daily Insight** | AI-generated summary of an app's revenue health for one day. Pro tier only. |
| **Billing Subscription** | The LedgerGuard user's own SaaS subscription (via Razorpay). Separate from Shopify subscriptions. |
| **App Review** | A review from the Shopify App Store. Scraped periodically. |

## Infrastructure

| Term | Definition |
|------|-----------|
| **Firebase Auth** | Authentication provider. Users sign in with Google or email/password. Backend verifies ID tokens. |
| **Partner API** | Shopify's GraphQL API for app developers. Provides transaction data, app info, install counts. |
| **Storefront API** | Shopify's public API for store metadata (brand, logo, description). |
| **Cloud Run** | GCP serverless container platform. Used for staging backend. |
| **Cloud SQL** | GCP managed PostgreSQL. Used for staging database. |
| **Razorpay** | Indian payment gateway. Used for LedgerGuard's own B2B billing (Subscriptions API). |

## AI Chat

| Term | Definition |
|------|-----------|
| **Module** | A self-contained plugin for the AI chat system. Provides tools and prompt fragments. |
| **Tool** | A function the AI can call to query data. Named `module__tool_name` (double underscore). |
| **Tool Loop** | AI sends message → receives tool calls → executes tools → sends results back → AI responds. Max 5 iterations. |
| **AIClient** | Provider-agnostic interface for LLM communication. Currently OpenAI, Claude planned. |
| **GraphQL Executor** | Thread-safe wrapper around gqlgen that modules use to query data. |

## Plans & Tiers

| Term | Definition |
|------|-----------|
| **FREE** | Legacy tier (deprecated). |
| **STARTER** | Base paid plan. $249/mo. Dashboard, sync, notifications, 1 app. |
| **PRO** | Advanced plan. $499/mo. AI Chat, API keys, Slack, data export, unlimited apps. |
| **ENTERPRISE** | Custom plan. Custom pricing, custom risk rules, priority support. |
| **Plan Tier** | User's current billing plan (`plan_tier` column on users table). |

## Frontend (Flutter)

| Term | Definition |
|------|-----------|
| **Provider** | State management pattern used in the Flutter prototype (`ledgerguard-flutter/`). |
| **Bloc** | State management pattern used in the production Flutter app (`frontend/app/`). Events in, states out. |
| **GoRouter** | Declarative routing library for Flutter. Handles deep links and navigation. |
| **AppShell** | Main navigation wrapper. BottomNavigationBar on mobile, NavigationRail on tablet/web. |
