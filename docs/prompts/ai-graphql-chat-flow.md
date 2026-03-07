# AI Chat + GraphQL - Interactive Visualization

## Context

You are a senior frontend + visualization engineer building an interactive animated diagram showing how **LedgerGuard's AI Chat Assistant** uses an **internal GraphQL layer** to let Shopify app developers query their revenue data via natural language.

This visualization helps developers understand:
1. **Natural language → structured query** — Ask in English, get GraphQL results
2. **GraphQL schema** — What data is queryable (subscriptions, metrics, risk, store health)
3. **The AI pipeline** — Intent classification → query generation → response formatting
4. **Phased rollout** — Internal GraphQL for AI first, public API later

---

## The Concept

### Two-Phase Approach

```
┌─────────────────────────────────────────────────────────────────────────┐
│                                                                         │
│   PHASE 1 (P2): AI Chat + Internal GraphQL                             │
│   ─────────────────────────────────────────                             │
│   • gqlgen-based GraphQL layer (internal only, not public)             │
│   • Used exclusively by AI Chat as its query engine                    │
│   • Chat widget in Flutter dashboard                                   │
│   • Claude API for NL → GraphQL translation                           │
│   • Conversational responses with inline data                          │
│   • Schema iterated freely without public API contract                 │
│                                                                         │
│   PHASE 2 (P3): Public GraphQL Developer API                           │
│   ──────────────────────────────────────────                            │
│   • Promote battle-tested schema to public /graphql endpoint           │
│   • Auth via existing API keys                                         │
│   • Playground with schema introspection                               │
│   • Supplements existing REST Revenue API                              │
│   • Schema is stable — validated by real AI queries in Phase 1         │
│                                                                         │
│   WHY THIS ORDER:                                                       │
│   • Ship user value (AI chat) without API governance overhead          │
│   • Iterate schema freely before committing to public contract         │
│   • Battle-test resolvers with real queries before exposing publicly   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## User Flow

```
┌──────────┐     ┌──────────────┐     ┌────────────────┐     ┌──────────────┐
│          │     │              │     │                │     │              │
│  User    │────▶│  Chat Widget │────▶│  Claude API    │────▶│  GraphQL     │
│  types   │     │  or GraphQL  │     │  (if chat)     │     │  Resolver    │
│  query   │     │  Playground  │     │                │     │  (gqlgen)    │
│          │     │              │     │  Intent →      │     │              │
└──────────┘     └──────────────┘     │  GraphQL query │     │  Query DB    │
                                      └────────────────┘     └──────┬───────┘
                                                                     │
┌──────────┐     ┌──────────────┐     ┌────────────────┐            │
│          │     │              │     │                │            │
│  User    │◀────│  Chat: text  │◀────│  Claude API    │◀───────────┘
│  sees    │     │  + tables    │     │  formats       │     PostgreSQL
│  answer  │     │  + charts    │     │  response      │     (domain data)
│          │     │  + follow-up │     │                │
└──────────┘     └──────────────┘     └────────────────┘
```

---

## Example Interactions

### Example 1: Risk Query
```
User:    "Which stores haven't paid in over 60 days?"

AI:      → Generates: query { subscriptions(riskState: TWO_CYCLES_MISSED) {
                         domain, shopName, daysPastDue, mrrCents
                       }}

         → Response: "3 stores are 60+ days overdue:

           | Store              | Days Late | MRR     |
           |--------------------|-----------|---------|
           | acme.myshopify.com | 72 days   | $49.99  |
           | beta.myshopify.com | 65 days   | $29.99  |
           | gamma.myshopify.com| 61 days   | $19.99  |

           Total revenue at risk: $99.97/mo

           Would you like to see the risk timeline for any of these?"
```

### Example 2: MRR Trend
```
User:    "What's my MRR trend over the last 6 months?"

AI:      → Generates: query { metricsTrend(appId: "...", months: 6) {
                         date, activeMrrCents
                       }}

         → Response: "Your MRR has grown 23% over 6 months:

           Jan: $2,450  →  Feb: $2,680  →  Mar: $2,890
           Apr: $2,750  →  May: $2,920  →  Jun: $3,010  📈

           Biggest growth: Feb (+$230, 9.4%)
           Only dip: Apr (-$140, likely from 2 churned stores)

           Want me to break down the churn in April?"
```

### Example 3: Store Health
```
User:    "Is acme-shop paying?"

AI:      → Generates: query { storeHealth(domain: "acme-shop.myshopify.com") {
                         isPaid, riskState, lastPaymentDate, daysPastDue,
                         subscription { plan, mrrCents }
                       }}

         → Response: "✅ acme-shop.myshopify.com is current!

           Plan: Pro ($49.99/mo)
           Last payment: Mar 1, 2026
           Risk state: SAFE
           Next charge: Apr 1, 2026

           No action needed."
```

---

## GraphQL Schema (Proposed)

```graphql
type Query {
  # Subscriptions
  subscriptions(
    appId: ID!
    riskState: RiskState
    status: SubscriptionStatus
    domain: String
    limit: Int = 50
    offset: Int = 0
  ): SubscriptionConnection!

  subscription(id: ID!): Subscription

  # Metrics
  metrics(appId: ID!): Metrics!
  metricsTrend(appId: ID!, months: Int = 6): [DailySnapshot!]!

  # Store Health
  storeHealth(appId: ID!, domain: String!): StoreHealth!

  # Earnings
  earnings(appId: ID!): Earnings!

  # Risk Summary
  riskSummary(appId: ID!): RiskSummary!
}

type Subscription {
  id: ID!
  domain: String!
  shopName: String!
  plan: String
  status: SubscriptionStatus!
  riskState: RiskState!
  mrrCents: Int!
  currency: String!
  daysPastDue: Int
  lastPaymentDate: DateTime
  expectedNextCharge: DateTime
  billingInterval: BillingInterval!
  events(limit: Int = 10): [SubscriptionEvent!]!
}

type Metrics {
  activeMrrCents: Int!
  revenueAtRiskCents: Int!
  usageRevenueCents: Int!
  totalRevenueCents: Int!
  renewalSuccessRate: Float!
  safeCount: Int!
  oneCycleMissedCount: Int!
  twoCyclesMissedCount: Int!
  churnedCount: Int!
}

type StoreHealth {
  domain: String!
  shopName: String!
  isPaid: Boolean!
  riskState: RiskState!
  subscription: Subscription
  lastPaymentDate: DateTime
  daysPastDue: Int
}

type RiskSummary {
  totalSubscriptions: Int!
  safe: Int!
  oneCycleMissed: Int!
  twoCyclesMissed: Int!
  churned: Int!
  revenueAtRiskCents: Int!
  atRiskSubscriptions: [Subscription!]!
}

type DailySnapshot {
  date: DateTime!
  activeMrrCents: Int!
  renewalSuccessRate: Float!
  revenueAtRiskCents: Int!
  safeCount: Int!
  churnedCount: Int!
}

type Earnings {
  recurringCents: Int!
  usageCents: Int!
  oneTimeCents: Int!
  refundCents: Int!
  totalCents: Int!
}

type SubscriptionEvent {
  eventType: String!
  fromStatus: SubscriptionStatus
  toStatus: SubscriptionStatus
  fromRiskState: RiskState
  toRiskState: RiskState
  occurredAt: DateTime!
}

type SubscriptionConnection {
  nodes: [Subscription!]!
  totalCount: Int!
  pageInfo: PageInfo!
}

type PageInfo {
  hasNextPage: Boolean!
  hasPreviousPage: Boolean!
}

enum RiskState {
  SAFE
  ONE_CYCLE_MISSED
  TWO_CYCLES_MISSED
  CHURNED
}

enum SubscriptionStatus {
  ACTIVE
  CANCELLED
  FROZEN
  PENDING
  UNINSTALLED
}

enum BillingInterval {
  MONTHLY
  ANNUAL
}

scalar DateTime
```

---

## AI Chat Pipeline (Technical)

```
┌─────────────────────────────────────────────────────────────────┐
│                     AI Chat Pipeline                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. USER INPUT                                                  │
│     "Which stores are at risk?"                                 │
│                                                                 │
│  2. CONTEXT ASSEMBLY                                            │
│     • User's app context (selected app ID)                      │
│     • Conversation history (last 10 messages)                   │
│     • GraphQL schema (injected as system prompt)                │
│     • Available query types and fields                          │
│                                                                 │
│  3. CLAUDE API CALL                                             │
│     System: "You are a revenue data assistant. Given the        │
│     GraphQL schema below, generate a query for the user's       │
│     question. Return JSON: {query, variables, explanation}"     │
│                                                                 │
│     Tools: [{name: "execute_graphql",                           │
│              input: {query: String, variables: JSON}}]          │
│                                                                 │
│  4. QUERY EXECUTION                                             │
│     • Validate generated query against schema                   │
│     • Execute via gqlgen resolver                               │
│     • Return structured data                                    │
│                                                                 │
│  5. RESPONSE FORMATTING                                         │
│     • Claude formats data as conversational text                │
│     • Adds inline tables/charts when appropriate                │
│     • Suggests 2-3 follow-up questions                          │
│     • Cites specific numbers from the data                      │
│                                                                 │
│  6. SAFETY & GUARDRAILS                                         │
│     • Query complexity limits (max depth, max fields)           │
│     • Rate limiting per user                                    │
│     • Schema-only queries (no mutations via chat)               │
│     • Audit logging of all AI-generated queries                 │
│     • Tenant isolation (user can only query own data)           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Architecture Integration

```
┌─────────────────────────────────────────────────────────────────────┐
│                        LedgerGuard Backend                          │
│                                                                     │
│  ┌─────────────────┐   ┌──────────────────┐   ┌────────────────┐  │
│  │  REST API        │   │  GraphQL API      │   │  AI Chat API   │  │
│  │  /api/v1/...     │   │  /graphql         │   │  /api/v1/chat  │  │
│  │  (existing)      │   │  (gqlgen)         │   │  (WebSocket)   │  │
│  │                  │   │                   │   │                │  │
│  │  Firebase Auth   │   │  API Key Auth     │   │  Firebase Auth │  │
│  └────────┬─────────┘   └────────┬──────────┘   └───────┬────────┘  │
│           │                      │                       │          │
│           └──────────────────────┼───────────────────────┘          │
│                                  │                                   │
│                    ┌─────────────▼──────────────┐                   │
│                    │     Domain Services         │                   │
│                    │  (RiskEngine, Metrics, etc.) │                   │
│                    └─────────────┬──────────────┘                   │
│                                  │                                   │
│                    ┌─────────────▼──────────────┐                   │
│                    │       PostgreSQL             │                   │
│                    └────────────────────────────┘                   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Visualization Requirements

### Page: `/ai-query-assistant`

**Animated Flow Steps:**
1. Chat bubble appears with user typing a natural language question
2. Question flows to "Intent Classification" node (Claude logo)
3. Claude generates a GraphQL query (syntax-highlighted code block appears)
4. Query flows to GraphQL resolver
5. Data flows back from PostgreSQL through resolver
6. Claude formats the response
7. Chat bubble appears with formatted answer + data table + follow-up suggestions
8. User clicks a follow-up → cycle repeats

**Interactive Elements:**
- Clickable example queries that animate the full pipeline
- Toggle between "Chat Mode" and "Direct GraphQL Mode"
- Schema explorer sidebar showing available types/fields
- Live query builder that shows the generated GraphQL

**Reference Cards:**
- "Supported Query Types" — what you can ask about
- "GraphQL Schema" — type definitions
- "Example Conversations" — 5+ real-world scenarios
- "Security & Guardrails" — how data isolation works

---

## Technical Requirements

### Frontend (Flutter)
- Chat widget using a custom `ChatBloc` (events: SendMessage, ReceiveResponse)
- WebSocket connection for streaming responses
- Markdown rendering for AI responses (tables, code blocks)
- Message history stored locally (Hive/SharedPreferences)
- Typing indicator while AI processes

### Backend (Go)
- `gqlgen` for GraphQL schema-first development
- `/graphql` endpoint with API key auth middleware
- `/api/v1/chat` WebSocket endpoint with Firebase auth
- Claude API integration for NL → GraphQL translation
- Query validation and complexity analysis
- Audit logging for all AI-generated queries

### AI (Claude API)
- System prompt with full GraphQL schema
- Tool use: `execute_graphql(query, variables)`
- Conversation memory (last 10 messages per session)
- Response formatting instructions (tables, charts, follow-ups)
- Guard against prompt injection (schema-only queries, no mutations)
