# Mock Shopify Partner API

A lightweight Ruby/Sinatra server that mimics the Shopify Partner API GraphQL endpoint. Used for local development and testing of the LedgerGuard backend with diverse data scenarios.

## Quick Start

```bash
cd mock-shopify-api
bundle install
ruby app.rb
# → Runs on http://localhost:4000
# → Admin UI at http://localhost:4000/admin
```

## Personas

| Org ID | Persona | Apps | Shops | Description |
|--------|---------|------|-------|-------------|
| 1001 | Solo Developer | 1 | 15 | All healthy, steady revenue |
| 1002 | Growing Developer | 2 | 80 | Moderate revenue, some risk |
| 1003 | Power Developer | 4 | 210 | High revenue, mixed risk states |
| 1004 | Churning Developer | 1 | 40 | Heavy churn, declining revenue |

## GraphQL Endpoint

```
POST /:org_id/api/2025-07/graphql.json
```

Supports `transactions` and `events` queries with pagination.

## Admin UI

- **Dashboard** — `/admin` — overview of all personas
- **Persona Detail** — `/admin/personas/:org_id` — view/edit shops, subscriptions, events
- **Webhook Tester** — `/admin/webhooks` — send test webhooks to backend

## Using with Backend

Start backend with mock config:

```bash
cd backend && go run ./cmd/server -config config.mock.yaml
```

Then trigger a sync to pull data from the mock API.
