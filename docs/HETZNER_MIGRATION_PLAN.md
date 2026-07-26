# LedgerGuard → Hetzner Migration Plan

**Status:** ✅ Deployed — co-host POC live at `https://api.ledgerspear.com` (2026-07-26). GCP decommissioned.
**Created:** 2026-07-25
**Goal:** Move the LedgerGuard backend off GCP (Cloud Run + Cloud SQL + VPC connector) onto a single Hetzner VPS running Docker Compose — cutting hosting from ~₹2000/mo to ~₹360/mo.
**Reference implementation:** `/Users/sachins/personal/checkoutmate` (already in production on Hetzner — proven pattern).

> **Chosen rollout (2026-07-25):** POC first by **co-hosting on the existing checkoutmate VPS** (`46.224.203.174`) to prove it out at ₹0 extra, then graduate to a dedicated CX22 once verified. The box is near-idle (load 0.06, ~2.8 GB free RAM, 64 GB free disk), so it fits. Co-host artifacts live in `deploy/cohost/` (see `deploy/cohost/README.md`). Because config is env-based and Dockerized, moving to a dedicated box later is a copy-`.env` + `up -d` + DNS repoint — no lock-in. The standalone `docker-compose.prod.yml` (with its own nginx+certbot) is the dedicated-box target.

---

## Why

Current GCP staging (`ledgerspear` project) bills ~₹2000/mo (~$24), dominated by:

| Resource | Config | Est. /mo |
|---|---|---|
| VPC connector `cloudrun-sql-connector` | min 2 × e2-micro, always-on | ~₹1,100 |
| Cloud SQL `ledgerspear-staging` | `db-f1-micro`, activation ALWAYS, 10 GB | ~₹800 |
| Cloud Run `ledgerspear-api` | min-scale 0 | ~₹0 idle |
| Artifact Registry / misc | 65 MB | ~₹5 |

The VPC connector exists only because Cloud SQL is private-IP-only — a cost with no analog on a single-box Hetzner setup.

---

## Target architecture

**Moves to Hetzner (one VPS, Docker Compose):**

```
Hetzner CX22 (2 vCPU / 4 GB / 40 GB SSD, Ubuntu 24.04, ~€3.79/mo)
└─ docker-compose.prod.yml
   ├─ postgres:16-alpine   self-hosted, volume-backed        ← replaces Cloud SQL
   ├─ redis:7-alpine       password-protected, volume-backed ← queue/sync pipeline
   ├─ api                  Go backend (existing Dockerfile), 127.0.0.1:8080
   ├─ nginx:alpine         reverse proxy :80/:443            ← replaces Cloud Run ingress
   └─ certbot              Let's Encrypt auto-renew          ← replaces Google-managed TLS
```

**Stays put:**
- **Frontend** → Firebase Hosting (free tier — not the cost problem). Only change: repoint API URL to `https://api.ledgerspear.com`, rebuild, redeploy.
- **Firebase Auth** → unchanged. The backend needs `firebase-credentials.json` **mounted** into the `api` container (the one LedgerGuard-specific addition vs checkoutmate, which uses JWT).

### Why this maps cleanly from checkoutmate
The Go backend already reads **all** config from environment variables (`backend/internal/infrastructure/config/config.go`): `DB_HOST/PORT/USER/PASSWORD/NAME/SSLMODE`, `REDIS_*`, `FIREBASE_CREDENTIALS_FILE`, `SHOPIFY_*`, `OPENAI_*`, `RAZORPAY_*`, `ENCRYPTION_MASTER_KEY`, `SERVER_PORT`, `DB_MIGRATIONS_PATH`. These are the same vars Cloud Run passes today, so they drop straight into a compose `.env`. Migrations already auto-run on startup.

---

## Cost swap

| | GCP now | Hetzner |
|---|---|---|
| Monthly | ~₹2000 ($24) | **~₹360 (€3.79)** + optional ~₹70 (€0.76) backups |
| Savings | — | **~₹1,600/mo (~80%)** |

---

## Migration phases

### Phase 1 — Repo scaffolding (local, in a PR)
Generate, modeled on checkoutmate's files, under the LedgerGuard repo:
- `docker-compose.prod.yml` — postgres + redis + api + nginx + certbot, wired to the backend's env-var names, **+ `firebase-credentials.json` volume mount**.
- `deploy/setup.sh` — fresh-Ubuntu bootstrap (install Docker + compose plugin, create `/opt/ledgerguard`, first-time certbot).
- `deploy/nginx.conf` — `api.ledgerspear.com` → `127.0.0.1:8080`, gzip, access logs.
- `deploy/.env.example` — all keys (Postgres, Redis, Shopify, Firebase, OpenAI, Razorpay, encryption, server).

### Phase 2 — Provision Hetzner (user, ~10 min)
1. cloud.hetzner.com → Create Server → Ubuntu 24.04 → **CX22**.
2. Location: Nuremberg/Helsinki. Enable **Backups** (worth it for the DB).
3. Add SSH key (`ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519_ledgerguard`).
4. Note the IPv4. *(Requires your Hetzner account — cannot be automated from here.)*

### Phase 3 — DNS (user)
Registrar DNS for `ledgerspear.com` (GoDaddy):

| Type | Name | Value | TTL |
|---|---|---|---|
| A | api | `<VPS_IP>` | 600 |

Verify: `dig +short api.ledgerspear.com` → VPS IP.

### Phase 4 — Deploy (guided; run on the box)
```bash
apt-get update && apt-get install -y git
git clone https://github.com/sachin-sivadasan/ledgerguard.git /opt/ledgerguard
cd /opt/ledgerguard

# Secrets
echo "POSTGRES_PASSWORD=$(openssl rand -base64 24)"
echo "REDIS_PASSWORD=$(openssl rand -base64 24)"
echo "ENCRYPTION_MASTER_KEY=$(openssl rand -hex 16)"   # match existing 32-hex-char key length
cp deploy/.env.example .env && nano .env               # fill all secrets

# Firebase service account (copy from local, do NOT commit)
scp firebase-credentials.json ledgerguard-prod:/opt/ledgerguard/

# Bring up the stack (migrations auto-run on api startup)
docker compose -f docker-compose.prod.yml up -d --build
curl -s https://api.ledgerspear.com/health   # {"status":"ok","database":"connected"}
```

### Phase 5 — Point frontend at Hetzner
- Update staging API URL in `frontend/app/lib/main_staging.dart` → `https://api.ledgerspear.com`.
- `flutter build web --release -t lib/main_staging.dart && firebase deploy --only hosting`.
- Verify login (test account `test@zoko.io`) and the events `storeDomain` fix end-to-end.

### Phase 6 — Tear down GCP
```bash
gcloud sql instances delete ledgerspear-staging --project=ledgerspear
gcloud compute networks vpc-access connectors delete cloudrun-sql-connector --region=us-central1 --project=ledgerspear
gcloud run services delete ledgerspear-api --region=us-central1 --project=ledgerspear
gcloud artifacts repositories delete ledgerspear --location=us-central1 --project=ledgerspear
# then the VPC network / reserved private range
```
Confirm billing → ₹0 in the GCP Billing report.

### Phase 7 — CI/CD + docs
- Replace `scripts/gcp-deploy.sh` in GitHub Actions with SSH-based deploy (`appleboy/ssh-action`: `git pull && docker compose -f docker-compose.prod.yml up -d --build`).
- Update `docs/GCP_SETUP_LOG.md` (mark decommissioned) and add a Hetzner setup log.
- Log to `prompts.md`, `IMPLEMENTATION_LOG.md`, `verification.md`.

---

## Decisions (recommendation in **bold**)

1. **Data migration** — ledger rebuilds deterministically from Shopify on each sync → **start fresh + re-sync** (no `pg_dump`/restore). Only loses staging daily-snapshot history. *(Alt: export from Cloud SQL via a temporary public IP and restore.)*
2. **GCP cutover timing** — **run both in parallel**, verify Hetzner, then delete GCP (avoids downtime; one overlap month of GCP cost). *(Alt: stop Cloud SQL + delete the VPC connector now for immediate savings, accept a staging gap.)*
3. **Database hosting** — **self-host Postgres in Docker** (checkoutmate-proven, cheapest) + Hetzner VPS backups. *(Alt: Hetzner Managed Postgres ~€16/mo for auto-backups + PITR.)*

---

## Secrets checklist (move to `.env` on the box, never commit)

`backend/config.local.yaml` (gitignored) currently holds live-looking keys that must move to the server `.env`:
- **OpenAI API key** — ⚠️ was surfaced in plaintext during planning; **rotate it** as part of this migration.
- **Razorpay** key_id / key_secret / webhook_secret (currently test keys).
- **Encryption master key** — reuse the existing value so encrypted data stays readable (do NOT regenerate if preserving data).
- Shopify client_id / client_secret.
- `firebase-credentials.json` — copy the file to the box, mount into `api`.

---

## Rollback

Until Phase 6 (GCP teardown), GCP stays fully intact — rollback = repoint DNS / frontend API URL back to the Cloud Run URL. After teardown, rollback means re-provisioning GCP (avoid until Hetzner is verified stable for a few days).

---

## Backups (self-hosted Postgres)

Cron on the VPS (from checkoutmate's pattern):
```
0 2 * * *  docker compose -f /opt/ledgerguard/docker-compose.prod.yml exec -T postgres \
             pg_dump -U ledgerguard ledgerguard | gzip > /opt/backups/$(date +\%Y\%m\%d).sql.gz
```
Optionally `rclone` the dumps to a Hetzner Storage Box. VPS-level backups (enabled in Phase 2) cover the whole disk.

---

## Reference files (checkoutmate)

| Purpose | checkoutmate path |
|---|---|
| Compose stack | `docker-compose.prod.yml` |
| Bootstrap script | `deploy/setup.sh` |
| nginx reverse proxy | `deploy/nginx.conf` |
| Env template | `deploy/.env.example` |
| Full runbook | `dev_docs/production-deployment.md` |
