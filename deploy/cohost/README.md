# Co-Host Deployment — LedgerGuard alongside checkoutmate

Run LedgerGuard on the **existing checkoutmate Hetzner VPS** (`46.224.203.174`) as a POC,
without disrupting checkoutmate production. Graduate to a dedicated box later using the
standalone `docker-compose.prod.yml` (nothing here creates lock-in).

**Isolation guarantees:** own containers (`ledgerguard-*`), own volumes (`lg_*`), own
internal network (`lg-internal`). Teardown removes only LedgerGuard. The *only* change to
checkoutmate is appending a vhost to its nginx config + one cert — both additive/reversible.

---

## Prerequisites
- DNS: `A  api.ledgerspear.com → 46.224.203.174`  (verify: `dig +short api.ledgerspear.com`)
- Local `firebase-credentials.json` to copy to the box
- A filled `.env` (see `deploy/.env.example`)

---

## Steps (run on the box unless noted)

### 1. Safety net — add 2 GB swap (box currently has none)
```bash
fallocate -l 2G /swapfile && chmod 600 /swapfile && mkswap /swapfile && swapon /swapfile
echo '/swapfile none swap sw 0 0' >> /etc/fstab
free -m   # confirm Swap: 2048
```

### 2. Clone LedgerGuard + configure
```bash
git clone https://github.com/sachin-sivadasan/ledgerguard.git /opt/ledgerguard
cd /opt/ledgerguard
cp deploy/.env.example .env && nano .env          # fill all secrets
# from local:  scp firebase-credentials.json checkoutmate-prod:/opt/ledgerguard/
```

### 3. Bring up LedgerGuard services (postgres + redis + api)
```bash
cd /opt/ledgerguard
# --env-file is REQUIRED: with -f pointing at deploy/cohost/, Compose looks for
# .env in that dir, not the repo root, so point it at the root .env explicitly.
docker compose --env-file /opt/ledgerguard/.env -f deploy/cohost/docker-compose.cohost.yml up -d --build
# migrations auto-run on api startup; confirm from inside the shared network:
docker exec checkoutmate-nginx-1 wget -qO- http://ledgerguard-api:8080/health
# expect {"status":"ok","database":"connected"}
```

### 4. Issue the TLS cert (webroot via checkoutmate's nginx) — two-phase
**4a.** Append ONLY the `listen 80` (ACME) block from `nginx-ledgerguard.conf` to
`/opt/checkoutmate/deploy/nginx.conf`, then reload:
```bash
docker exec checkoutmate-nginx-1 nginx -t && docker exec checkoutmate-nginx-1 nginx -s reload
```
**4b.** Obtain the certificate:
```bash
# --entrypoint certbot is REQUIRED: checkoutmate's certbot service has a custom
# auto-renew-loop entrypoint, so without this override the certonly args are
# ignored and it just runs `certbot renew`.
docker compose -f /opt/checkoutmate/docker-compose.prod.yml run --rm --entrypoint certbot certbot \
  certonly --webroot -w /var/www/certbot -d api.ledgerspear.com \
  --email sachinkarunagappally@gmail.com --agree-tos --non-interactive
```
**4c.** Append the `listen 443 ssl` block from `nginx-ledgerguard.conf`, then reload:
```bash
docker exec checkoutmate-nginx-1 nginx -t && docker exec checkoutmate-nginx-1 nginx -s reload
```

### 5. Verify end-to-end
```bash
curl -s https://api.ledgerspear.com/health     # {"status":"ok","database":"connected"}
```
Then repoint the frontend (`main_staging.dart` → `https://api.ledgerspear.com`),
rebuild, `firebase deploy`, and log in with the test account.

---

## Redeploy (update to latest) — with disk hygiene
```bash
cd /opt/ledgerguard
git pull && \
  docker compose --env-file /opt/ledgerguard/.env -f deploy/cohost/docker-compose.cohost.yml up -d --build && \
  docker builder prune -f && docker image prune -f
```
The two prunes keep the shared box from filling with orphaned **build cache** and
**dangling images** left by each `--build` (mirrors checkoutmate — which had 20+
deploys accumulate ~56 GB / 78% disk before adding `docker builder prune -f` to its
deploy). Scoped on purpose: `builder prune` = cache only; `image prune` (no `-a`) =
untagged images only, so no running/tagged image (LedgerGuard's or checkoutmate's) is
touched. **Never** `docker system prune -af` here — `-a` deletes images not tied to a
*running* container, which can wipe a stopped service's base image.

One-time reclaim if disk is already high: `docker builder prune -f && docker image prune -f`
(check first with `docker system df`).

---

## Rollback / teardown (LedgerGuard only — checkoutmate untouched)
```bash
cd /opt/ledgerguard
docker compose --env-file /opt/ledgerguard/.env -f deploy/cohost/docker-compose.cohost.yml down   # add -v to also drop data
# remove the api.ledgerspear.com blocks from /opt/checkoutmate/deploy/nginx.conf
docker exec checkoutmate-nginx-1 nginx -t && docker exec checkoutmate-nginx-1 nginx -s reload
```

## Notes
- Keep GCP running in parallel during the POC — instant rollback, zero downtime.
- Memory caps (api 768m / pg 512m / redis 256m) protect checkoutmate from a heavy "whale" sync.
- Graduation: on a new VPS use `docker-compose.prod.yml` (includes its own nginx+certbot),
  copy `.env` + creds, `up -d`, repoint DNS, then teardown here.
