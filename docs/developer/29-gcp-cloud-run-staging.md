# 29. GCP Cloud Run Staging

## What It Does
Provides a staging environment for the LedgerGuard backend on Google Cloud Platform. The backend runs as a serverless container on Cloud Run, connected to Cloud SQL PostgreSQL via private networking. Scales to zero when idle to preserve free credits. See [ADR-009](../../DECISIONS.md).

## Architecture
Infrastructure layer. Cloud Run serves the same Go binary as production but with staging-specific configuration. Docker images stored in Artifact Registry. Secrets managed via GCP Secret Manager. Private networking via VPC Connector between Cloud Run and Cloud SQL.

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `scripts/gcp-deploy.sh` | ~50 | Deployment script (Docker build + push + deploy) |
| `backend/Dockerfile` | ~30 | Multi-stage Go build |
| `docs/GCP_SETUP_LOG.md` | ~400 | Complete command history for GCP setup |
| `.github/workflows/deploy-staging.yml` | ~60 | GitHub Actions CI/CD (staging branch) |

## Data Flow
```
┌─────────────────────────────────────────────────────┐
│                  GCP Project: ledgerspear            │
│                                                      │
│  ┌──────────────┐     ┌─────────────────────────┐   │
│  │ Artifact     │     │ Secret Manager           │   │
│  │ Registry     │     │ (DB creds, Firebase key, │   │
│  │ (Docker imgs)│     │  encryption key, etc.)   │   │
│  └──────┬───────┘     └───────────┬─────────────┘   │
│         │                         │                   │
│         ▼                         ▼                   │
│  ┌──────────────────────────────────────────────┐   │
│  │          Cloud Run Service                    │   │
│  │          ledgerspear-api                      │   │
│  │          (scale 0-2 instances)                │   │
│  │          --platform linux/amd64               │   │
│  └──────────────────┬───────────────────────────┘   │
│                     │ VPC Connector                   │
│                     │ (private IP)                    │
│                     ▼                                 │
│  ┌──────────────────────────────────────────────┐   │
│  │          Cloud SQL                            │   │
│  │          PostgreSQL 14 (db-f1-micro)          │   │
│  │          Private IP only                      │   │
│  └──────────────────────────────────────────────┘   │
│                                                      │
└─────────────────────────────────────────────────────┘

External access:
  https://ledgerspear-api-ineifpjrdq-uc.a.run.app
```

## Configuration
| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| GCP_PROJECT | ledgerspear | Yes | GCP project ID |
| GCP_REGION | us-central1 | Yes | Cloud Run region |
| CLOUD_SQL_INSTANCE | ledgerspear:us-central1:ledgerspear-db | Yes | Cloud SQL connection name |
| DATABASE_URL | — | Yes | PostgreSQL connection string (via Secret Manager) |
| FIREBASE_SERVICE_ACCOUNT | — | Yes | Firebase Admin SDK credentials (Secret Manager) |
| ENCRYPTION_KEY | — | Yes | AES-256 encryption key (Secret Manager) |

## Command Reference
| Action | Command |
|--------|---------|
| Deploy backend | `./scripts/gcp-deploy.sh ledgerspear` |
| Check service status | `gcloud run services describe ledgerspear-api --region us-central1` |
| View logs | `gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=ledgerspear-api" --project=ledgerspear --limit=30` |
| Health check | `curl -sf https://ledgerspear-api-ineifpjrdq-uc.a.run.app/health` |
| List revisions | `gcloud run revisions list --service ledgerspear-api --region us-central1` |
| Force new deployment | `gcloud run deploy ledgerspear-api --image {image} --region us-central1` |
| Enable Cloud SQL public IP | `gcloud sql instances patch ledgerspear-db --assign-ip` |
| Disable Cloud SQL public IP | `gcloud sql instances patch ledgerspear-db --no-assign-ip` |

## Extension Points
- Add custom domain mapping via `gcloud run domain-mappings create`
- Scale configuration: `--min-instances`, `--max-instances`, `--concurrency`
- Add Cloud Armor for WAF/DDoS protection
- Enable Cloud CDN for static response caching

## Gotchas
- **Docker builds MUST use `--platform linux/amd64`** — dev machine is Apple Silicon ARM. Without this, Cloud Run gets `exec format error`
- **Cloud SQL public IP must be disabled after direct DB access** — it's private-only by default
- **Cold starts**: ~2 seconds for Go binary when scaling from zero (acceptable for staging)
- **Dirty database migrations**: If Cloud Run logs show `Dirty database version N`, follow §12 in `docs/GCP_SETUP_LOG.md` to fix
- **Free credits**: $300 for 90 days. After expiry, estimated ~$20-30/month
- **No custom domain** — uses auto-generated Cloud Run URL
