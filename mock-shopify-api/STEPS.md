# Mock Shopify Partner API — Step-by-Step Guide

## 1. Start the Mock Server

```bash
cd mock-shopify-api
/opt/homebrew/opt/ruby/bin/bundle install   # first time only
/opt/homebrew/opt/ruby/bin/ruby app.rb
```

Server runs at **http://localhost:4000**

> System Ruby (2.6) is too old. Always use Homebrew Ruby: `/opt/homebrew/opt/ruby/bin/ruby`

---

## 2. Start the Backend (with mock config)

In a second terminal:

```bash
cd backend
go run ./cmd/server -config config.mock.yaml
```

Backend runs at **http://localhost:8080** and calls the mock server instead of real Shopify.

---

## 3. Open the Admin Dashboard

Open in browser: **http://localhost:4000/admin**

You'll see 4 persona cards:

| Org ID | Persona | Apps | Shops |
|--------|---------|------|-------|
| 1001 | Solo Developer | 1 | 15 |
| 1002 | Growing Developer | 2 | 80 |
| 1003 | Power Developer | 4 | 210 |
| 1004 | Churning Developer | 1 | 40 |

Click any persona card to see its detail page.

---

## 4. View Persona Detail

URL: **http://localhost:4000/admin/personas/1002** (example: Growing Developer)

The page shows:
- Summary cards (apps, shops, active/frozen/cancelled counts)
- Apps table
- Shops table (first 50 shown)
- Subscriptions table with status badges
- Events table (most recent 30)

---

## 5. Add Data

### Add a Shop

On the persona detail page, scroll to the **Add Shop** form:

1. Enter **Domain**: `my-new-store.myshopify.com`
2. Enter **Name**: `My New Store`
3. Click **Add**

The shop gets a random GID and is saved to the YAML file.

### Add a Subscription

Scroll to the **Add Subscription** form:

1. Select **App Index** (0 = first app, 1 = second app, etc.)
2. Enter **Shop Index** (the # column from the shops table)
3. Enter **Plan Name**: `Pro`
4. Enter **Price**: `29.99`
5. Select **Status**: `active`, `frozen`, or `cancelled`
6. Click **Add**

### Change Subscription Status

In the subscriptions table, each row has a status dropdown:

1. Select new status (`active` / `frozen` / `cancelled`)
2. Optionally enter **days** (last charge days ago — used for risk engine)
3. Click **Set**

Example: Set a subscription to `frozen` with `45` days to trigger ONE_CYCLE_MISSED risk.

### Add an Event

Scroll to the **Add Event** form:

1. Select **Type**:
   - `RELATIONSHIP_INSTALLED` — app installed
   - `RELATIONSHIP_UNINSTALLED` — app uninstalled
   - `SUBSCRIPTION_CHARGE_ACCEPTED` — successful charge
   - `SUBSCRIPTION_CHARGE_CANCELED` — subscription cancelled
2. Select **App Index**
3. Enter **Shop Index**
4. Enter **Days Ago** (how many days in the past this event occurred)
5. Click **Add**

---

## 6. Trigger a Sync

After adding/editing data, trigger the backend to sync from the mock API:

```bash
curl -X POST -H "Authorization: Bearer YOUR_FIREBASE_TOKEN" \
  http://localhost:8080/api/v1/sync
```

The backend fetches transactions and events from the mock server, runs them through the real pipeline (risk engine, metrics, ledger rebuild), and stores results in the database.

---

## 7. Send Webhooks

Open: **http://localhost:4000/admin/webhooks**

### Steps:

1. Set **Backend URL** (default: `http://localhost:8080`)
2. Select a **Persona** from the dropdown
3. Select a **Shop** (loads automatically based on persona)
4. Select **Webhook Type**:

| Type | What it does |
|------|-------------|
| App Installed | Simulates a shop installing the app |
| App Uninstalled | Simulates a shop uninstalling the app |
| Subscription Updated | Simulates a subscription status change |
| Billing Failure | Simulates a failed billing attempt |

5. For **Subscription Updated**: also select a **Status** (ACTIVE / FROZEN / CANCELLED / EXPIRED)
6. Click **Send Webhook**

The result appears in the **Webhook History** table below with status code (200 = success, 0 = connection error).

### Example: Simulate a billing failure

1. Persona: Growing Developer (1002)
2. Shop: `acme-store.myshopify.com`
3. Type: Billing Failure
4. Click Send

---

## 8. Switch Personas

To test a different persona in the Flutter app:

1. Update the `organization_id` in the `partner_accounts` DB row to the desired org ID (1001–1004)
2. Trigger a sync (step 6)
3. The backend fetches different data from the mock server
4. Flutter app now shows that persona's data

```sql
UPDATE partner_accounts SET organization_id = '1003' WHERE user_id = 'YOUR_USER_ID';
```

---

## 9. Test with curl (no browser needed)

```bash
# Health check
curl http://localhost:4000/health

# Fetch transactions for Power Developer
curl -s http://localhost:4000/1003/api/2025-07/graphql.json \
  -X POST -H "Content-Type: application/json" \
  -d '{"query":"{ transactions(first: 5) { edges { node { id __typename grossAmount { amount } app { name } shop { myshopifyDomain } } } pageInfo { hasNextPage } } }"}' | python3 -m json.tool

# Fetch uninstall events for Churning Developer
curl -s http://localhost:4000/1004/api/2025-07/graphql.json \
  -X POST -H "Content-Type: application/json" \
  -d '{"query":"{ app(id: \"gid://partners/App/4001\") { events(first: 5, types: [RELATIONSHIP_UNINSTALLED]) { edges { node { type occurredAt shop { myshopifyDomain } } } } } }"}' | python3 -m json.tool
```

---

## 10. Reload Data

If you edit the YAML files directly (in `mock-shopify-api/data/`), click **Reload Data** in the nav bar or:

```bash
curl -X POST http://localhost:4000/admin/reload
```

This reloads all persona YAML files into memory without restarting the server.
