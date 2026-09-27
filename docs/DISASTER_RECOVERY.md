# Disaster Recovery — rebuild after host loss

**Context:** 2026-09 the Hetzner box (account K0387809126) was cancelled for non-payment and
its data (Postgres `lg_pgdata`, Redis) was lost. This runbook rebuilds LedgerGuard on a new
host. It doubles as the standing DR procedure — **and the backup routine (§5) so this can
never cost data again.**

## 0. What's safe vs lost vs rebuildable

| Asset | Status | Source of truth |
|---|---|---|
| All code (backend, frontend, docs) | ✅ safe | GitHub |
| Firebase Auth users, Hosting, FCM | ✅ safe | Firebase (separate from the host) |
| Transactions / subscriptions / ledger / earnings | ♻️ **rebuildable** | re-sync from Shopify Partner API (deterministic rebuild — TAD §8) |
| `daily_metrics_snapshot` history | ❌ lost | point-in-time; not re-derivable — accepted loss |
| Partner-account Shopify tokens (encrypted) | ❌ lost | users **re-connect** their Partner account |
| Org/team/members, API keys, plan labels, notification prefs, device tokens | ❌ lost | re-provisioned (auto on login / manual) |

## 1. Provision a new host
- Ubuntu LTS VPS (Hetzner **new account**, or another provider). Install Docker + Compose plugin.
- Open only 80/443 publicly; keep Postgres/Redis/ES on the internal docker network (never public — see the logs-MCP loopback rule).
- **Set up auto-pay / a card on file** so a missed invoice can't terminate the box again.

## 2. Deploy from GitHub
```bash
git clone <repo> ledgerguard && cd ledgerguard
cp deploy/.env.example deploy/cohost/.env   # then fill every value (see below)
docker compose --env-file deploy/cohost/.env -f deploy/cohost/docker-compose.cohost.yml up -d --build
docker builder prune -f && docker image prune -f   # disk hygiene (ADR-049)
```
**`.env` values to set** (grep the compose for the full list): `DB_NAME/DB_USER/DB_PASSWORD`,
`REDIS_PASSWORD`, `ENCRYPTION_MASTER_KEY`, `INTERNAL_KEY`, `OPENAI_API_KEY/MODEL`,
`QUEUE_ENABLED/NUM_WORKERS/FULL_SYNC_WORKERS`, `RAZORPAY_*`, `SHOPIFY_*`, `ELASTIC_PASSWORD`.
- **`ENCRYPTION_MASTER_KEY`:** generate a **fresh** key — the old ciphertext (Shopify tokens)
  is gone anyway; users re-enter tokens, which get encrypted with the new key.
- **Migrations run automatically** on API startup (`cmd/server/main.go` → `migrator.Up()` when
  `MigrationsPath` is set), so the schema is created on first boot. Confirm in the logs.

## 3. DNS + TLS
- Point **`api.ledgerspear.com`** A record at the new host IP.
- Reverse proxy / TLS: reuse `deploy/cohost/nginx-ledgerguard.conf` (or `deploy/Caddyfile`) and
  re-issue the certificate (Let's Encrypt) for `api.ledgerspear.com`.
- Frontend (`app.ledgerspear.com`, Firebase Hosting) needs **no change** — once DNS resolves
  and the cert is live, the existing web/APK builds hit the new backend.

## 4. Re-provision + rebuild data
1. A user signs in (Firebase Auth works) → a default org is auto-provisioned on first login
   (`middleware/auth.go`).
2. The user **re-connects their Shopify Partner account** (re-enter the Partner API token via the
   integration screen) — this recreates the partner account + encrypted token.
3. Selecting an app **auto-triggers a full sync** (`handler/app.go` → `TriggerSync`), or trigger
   manually: `POST /api/v1/sync/enqueue/{appID}?type=full`. The sync pipeline **rebuilds
   transactions → ledger → subscriptions → risk → forward snapshots** from Shopify (idempotent,
   deterministic). Historical snapshots start accumulating again from now.
4. Re-create any API keys, plan labels, and notification preferences (small, manual).

**Verify:** `curl -sf https://api.ledgerspear.com/health` → ok; an app dashboard shows synced
MRR/subscriptions after the full sync completes.

## 5. Backups going forward (so this never recurs)
**Nightly off-box `pg_dump`** via host cron:
```bash
# /etc/cron.d/ledgerguard-pgbackup  — 02:30 UTC daily
30 2 * * * root docker exec ledgerguard-db pg_dump -U "$DB_USER" "$DB_NAME" \
  | gzip > /var/backups/lg/lg-$(date +\%F).sql.gz && \
  find /var/backups/lg -name 'lg-*.sql.gz' -mtime +14 -delete
```
- **Copy off the box** (the whole point): sync `/var/backups/lg` to object storage / another
  provider (rclone/S3) — a backup on the same box dies with the box.
- Restore test quarterly: `gunzip -c lg-YYYY-MM-DD.sql.gz | docker exec -i ledgerguard-db psql -U $DB_USER -d $DB_NAME`.
- Also: **billing auto-pay + a monitored email**, and a low-balance/invoice alert, so a
  payment lapse is caught long before cancellation.

## 6. Accepted losses (document, don't chase)
- Historical `daily_metrics_snapshot` rows before the incident (trend/audit history).
- Org/team structure, API keys, prefs created before the incident (re-provisioned in §4).
Everything financial re-derives from Shopify; only the above are genuinely gone.
