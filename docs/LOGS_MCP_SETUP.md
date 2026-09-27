# Logs-over-MCP setup — LedgerGuard

**Decision (3 lines):** LedgerGuard has no centralized log store (Go stdlib `log`
→ Docker `json-file` → `docker logs` on the shared Hetzner box). We stand up a
**single-node Elasticsearch 8.x** (loopback-only) + **Vector** shipper + the
**`@elastic/mcp-server-elasticsearch` @0.1.1** MCP server — parity with the
existing `zoko-logs` setup. Access is a read-only ES API key over an SSH tunnel;
no ports exposed, no secrets in Claude config.

> Governing rule honored: nothing centralized existed, so this is the minimal store
> that gives ES parity. **Resource note (shared box):** ES heap is capped at 512 MB
> (`ES_JAVA_OPTS`); the box co-hosts checkoutmate and already needs disk hygiene
> ([[hetzner-migration]], DECISIONS ADR-049) — watch RAM/disk after Phase 1, and keep
> ILM retention short (15d) so log data can't fill the volume.

Phase order is fixed: **1 → 2 → 3 → 4 → 5**, verify each before the next. Phase 2
(retention + schema) MUST land before any data or dynamic mappings win and you'll
reindex. Run these **on the Hetzner box** (you run them — I don't SSH in;
see [[no-ssh-to-servers]]).

---

## Phase 0 — inventory (done, from the codebase)
- **Deploy:** `deploy/cohost/docker-compose.cohost.yml` — `ledgerguard-api` (Go), `ledgerguard-db` (pg16), `ledgerguard-redis`, nginx front; networks `lg-internal` + `checkoutmate_default`. **FACT.**
- **Logging (at inventory time, pre-migration):** Go **stdlib `log`** — ~376 `log.Printf`/`Println` (approx., grep-counted), **no structured logger**. *(Now migrated to zap/ecszap — Phase 4 below; only ~2 deliberate stdlib `log` calls remain.)* **FACT.**
- **Correlation:** chi `RequestID` middleware + `lgmw.ResponseLogger` are wired in `router.go`, but the request ID is **not** in the `log.Printf` lines. **FACT** — this is the highest-leverage gap (Phase 4).
- **Logs today:** default `json-file` driver → `docker logs ledgerguard-api`. No ES/Loki/Datadog. **FACT.**
- **Tenant fields available:** every request has an org (`X-Org-Id`, `OrgContextMW`) and most an `appID` — these become the keyword fields that make "one org's / one app's story" queryable.

---

## Phase 1 — Store (single-node ES, loopback only) — ✅ SHIPPED IN COMPOSE
The `elasticsearch` service + `lg_esdata` volume are committed in
`deploy/cohost/docker-compose.cohost.yml` (on `lg-internal`, bound to **127.0.0.1:9200**,
heap capped at 512 MB + `mem_limit: 1200m` for the shared box, auth ON, TLS off behind the
loopback + SSH-tunnel boundary). It **bootstraps the `elastic` password from `.env`**
(`ELASTIC_PASSWORD=${ELASTIC_PASSWORD:?…}`) — no manual `elasticsearch-reset-password` step —
and has an auth'd `_cluster/health` healthcheck.

**Run on the Hetzner box (you run these — I don't SSH in; [[no-ssh-to-servers]]):**
```bash
# 1. add a strong password to the compose env file (once)
echo "ELASTIC_PASSWORD=$(openssl rand -hex 24)" >> /path/to/deploy/cohost/.env
# 2. pull the new compose + start ONLY elasticsearch first
git pull
docker compose --env-file .env -f deploy/cohost/docker-compose.cohost.yml up -d elasticsearch
# 3. disk hygiene on the shared box (per DECISIONS ADR-049)
docker builder prune -f && docker image prune -f
```

**Verify:**
```bash
source deploy/cohost/.env
docker ps --filter name=ledgerguard-es          # State = healthy after ~30-60s
curl -su elastic:$ELASTIC_PASSWORD http://127.0.0.1:9200/_cluster/health   # status green/yellow (yellow is fine, single node)
ss -ltnp | grep 9200                             # bound to 127.0.0.1 ONLY — never 0.0.0.0
```
Keep the same `ELASTIC_PASSWORD` — it's the elastic superuser; the read-only MCP key (Phase 5)
is derived from it and never leaves the box.

---

## Phase 2 — Retention + schema (BEFORE any data)
**ILM — delete at 15 days** (retention cliff so logs can't fill the shared disk):
```bash
curl -su elastic:PASS -XPUT http://127.0.0.1:9200/_ilm/policy/logs-ledgerguard \
 -H 'content-type: application/json' -d '{"policy":{"phases":{
   "hot":{"actions":{"rollover":{"max_age":"1d","max_primary_shard_size":"5gb"}}},
   "delete":{"min_age":"15d","actions":{"delete":{}}}}}}'
```
**Index template** — type the core fields (single node → `replicas: 0`):
```bash
curl -su elastic:PASS -XPUT http://127.0.0.1:9200/_index_template/logs-ledgerguard \
 -H 'content-type: application/json' -d '{
  "index_patterns":["logs-ledgerguard-*"],
  "template":{
    "settings":{"number_of_replicas":0,"index.lifecycle.name":"logs-ledgerguard"},
    "mappings":{"properties":{
      "@timestamp":{"type":"date"},
      "log.level":{"type":"keyword"},
      "service.name":{"type":"keyword"},
      "request_id":{"type":"keyword"},        // ← correlation id — the highest-leverage field
      "org_id":{"type":"keyword"},
      "user_id":{"type":"keyword"},
      "app_id":{"type":"keyword"},
      "job_id":{"type":"keyword"},
      "message":{"type":"text"}
    }}}}'
```
**Verify:** `curl -su elastic:PASS http://127.0.0.1:9200/_index_template/logs-ledgerguard` returns it.

---

## Phase 3 — Shipper (Vector reads Docker logs → ES)
Add a `vector` service (pin `timberio/vector:0.41.X-debian`) to the compose, mount
the Docker socket read-only + a config:
```toml
# vector.toml
[sources.docker]
type = "docker_logs"
include_containers = ["ledgerguard-api"]     # label-filtered — not the whole box

[transforms.parse]
type = "remap"
inputs = ["docker"]
source = '''
  . = parse_json(.message) ?? { "message": .message }   # zap/ecszap JSON (Phase 4); plain text tolerated during migration
  ."@timestamp" = to_timestamp(.time) ?? now()
  .service.name = "ledgerguard-api"
'''

[sinks.es]
type = "elasticsearch"
inputs = ["parse"]
endpoints = ["http://elasticsearch:9200"]
auth = { strategy = "basic", user = "vector_writer", password = "${ES_WRITER_PASS}" }
mode = "data_stream"
bulk.index = "logs-ledgerguard-%Y.%m.%d"
```
Use a **separate write-only** user for Vector (not the read-only MCP key).
**Verify:** hit the API, then `curl -su elastic:PASS 'http://127.0.0.1:9200/logs-ledgerguard-*/_count'` → count > 0.

---

## Phase 4 — Structured logging (zap + ecszap) — ✅ IMPLEMENTED (foundation)
Chosen logger: **zap** with the **`go.elastic.co/ecszap`** encoder, so field names
(`@timestamp`, `log.level`, `message`, `service.name`) match the Phase-2 index template.
Pins: `zap v1.27.0`, `ecszap v1.0.2` (ecszap maintenance lags zap — smoke-test its output
on upgrades). Implemented:

1. **`internal/infrastructure/logging`** — `New`/`Init` build the ECS logger; `Init`
   installs it globally **and calls `zap.RedirectStdLog`** so the **remaining stdlib
   `log.Printf` sites emit structured JSON immediately** (no big-bang rewrite). Wired in
   `cmd/server/main.go` (level from `LOG_LEVEL` / `cfg.Log.Level`).
2. **`request_id` context logger** — `lgmw.RequestLogger` (added after `chimw.RequestID`)
   attaches a request-scoped logger carrying `request_id`; handlers/services log via
   `logging.FromContext(ctx)` and get the correlation id for free. Demo conversion:
   `middleware/auth.go` token-verification failure.
   > **Emitted fields (current):** `request_id` on every request line (via `RequestLogger`),
   > `org_id` + `user_id` on org-scoped routes (via `OrgContextMW`), and `app_id`/`job_id`
   > (+ inherited `request_id`) on sync-job lines (via `jobLogger`). All are typed in the
   > Phase-2 index template (add `user_id` there too — see below).
3. **Opportunistic migration:** convert hot paths off `log.Printf` to
   `logging.FromContext(ctx).Info/Error(...)`.
   - ✅ **Sync pipeline processors** (`internal/infrastructure/queue/processors/*`) — DONE.
     A `jobLogger(ctx, name, payload)` helper attaches a job-scoped logger carrying
     `processor` / `app_id` / `job_id` structured fields (the "one sync's story"
     correlation keys) and puts it on ctx; all 17 processor `log.Printf` sites now log
     via it. Hand-interpolated `"for app %s (job %s)"` suffixes became queryable fields.
   - ✅ **HTTP handlers** (117 sites across 33 files) — DONE. Request handlers log via
     `logging.FromContext(r.Context())` (so `request_id` flows onto every line);
     interpolated ids/errors became structured fields (`app_id`, `user_id`, `partner_id`,
     `topic`, `zap.Error(err)`, …). The shared helpers (repo-error / CSV writers, some
     `buildX`) were also ctx-threaded (follow-up, now DONE), so the handler package has
     **zero** `zap.L()` calls left — every handler line carries `request_id`.
   - ✅ **Everything else** — DONE. `cmd/server/main.go` (startup → `zap.L()`),
     `application/service`, `application/scheduler`, `infrastructure/queue` core (worker /
     recovery / client), `infrastructure/external` clients, `revenue_api`,
     `domain/service`, `chat`, and the remaining middleware. Background code with a
     `ctx` uses `logging.FromContext(ctx)` (inherits `request_id`/`job_id` when set);
     ctx-less spots use `zap.L()`.
   - **Net result:** the whole non-test backend now emits ECS JSON directly. Only **two**
     deliberate stdlib `log` calls remain — a pre-`Init` warning in `logging.go` (logger not
     built yet) and the top-level `log.Fatalf` in `main()` (runs after `logCleanup` tears the
     logger down). The `RedirectStdLog` bridge is now a safety net for third-party libs, not
     a crutch for our own code.

   Enrichment follow-ups — all ✅ DONE:
   - ✅ **`request_id` into the job payload:** `SyncJobPayload.RequestID` is populated at
     enqueue time (`chimw.GetReqID(ctx)` in `QueueSyncService`), inherited by child/snapshot
     jobs in `FullSyncProcessor`, and emitted by `jobLogger` — so a sync ties back to the
     HTTP request that triggered it. Empty for background/recovery-enqueued jobs.
   - ✅ **`org_id`/`user_id` on HTTP lines:** `OrgContextMW.RequireOrg` wraps the request
     logger with `org_id` + `user_id` after resolving the tenant, so every downstream
     handler line carries them ("one org's story").
   - ✅ **ctx into the report-handler shared helpers:** every `writeXRepoError` / `writeXCSV`
     / logging `buildX` helper now takes `ctx` and logs via `logging.FromContext(ctx)`; the
     handler package has **zero** `zap.L()` calls left — all handler-layer logs carry
     `request_id` (+ `org_id`/`user_id` for org-scoped routes).

**Verify:** `logging` unit tests assert ECS field names + `request_id`; end-to-end, one
API request produces N log lines sharing one `request_id` in ES.

---

## Phase 5 — Access, MCP wiring, end-to-end
**Read-only API key** scoped to the log indices only:
```bash
curl -su elastic:PASS -XPOST http://127.0.0.1:9200/_security/api_key \
 -H 'content-type: application/json' -d '{
  "name":"ledgerguard-logs-mcp-ro",
  "role_descriptors":{"logs_ro":{
    "indices":[{"names":["logs-ledgerguard-*"],"privileges":["read","view_index_metadata"]}]}}}'
```
Save the returned `encoded` value to `~/.config/ledgerguard-logs/es.key` and **`chmod 600`** it. Secrets never go in Claude config.

**Private network path — SSH tunnel** (ES is 127.0.0.1-only on the box), reused if already up:
```bash
# ~/.local/bin/ledgerguard-logs-mcp   (chmod +x)
#!/usr/bin/env bash
set -euo pipefail
nc -z 127.0.0.1 19200 2>/dev/null || \
  ssh -f -N -L 19200:127.0.0.1:9200 <you>@<hetzner-host>
export ES_URL="http://127.0.0.1:19200"
export ES_API_KEY="$(cat ~/.config/ledgerguard-logs/es.key)"   # chmod 600
exec npx -y @elastic/mcp-server-elasticsearch@0.1.1            # PIN: 0.2.0+ needs the ES v9 client; ES 8.x rejects it
```
Register (user scope):
`claude mcp add ledgerguard-logs ~/.local/bin/ledgerguard-logs-mcp`

**Verify (fresh session):** `list_indices` → `logs-ledgerguard-*` present → mappings show the **typed** fields (not just a `message` blob — if blob-only, Phase 4 wasn't done) → one real query filtering `request_id` returns the full story of one request.

---

## After the first real debugging session
Capture tribal knowledge as a `/debug-ledgerguard` skill: which query answers which
question, the `service.name` vocabulary, top incident signatures (sync `partial_failure`,
503 repo errors, risk-convergence drift), and the 15-day retention cliff. Add the index
name + field list to the **Observability** sections of the as-built tech-designs
(`docs/designs/06-sync-pipeline.md`, `08-revenue-api.md`, `14-org-and-team.md` all have
"no log index" divergence notes to update).

## Invariants checklist (all enforced above)
- [ ] Wrapper launches the MCP server; creds from a `chmod 600` file — never in Claude config
- [ ] **Read-only** API key scoped to `logs-ledgerguard-*` (separate write user for Vector)
- [ ] Private path: ES bound `127.0.0.1` + SSH tunnel — no open port
- [ ] `request_id` is a typed keyword and flows through every hop (Phase 4) — the leverage field
- [ ] MCP server pinned `@0.1.1` (ES 8.x); ES image + Vector pinned
```
