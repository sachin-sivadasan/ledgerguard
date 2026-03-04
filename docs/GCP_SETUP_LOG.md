# GCP Staging Setup – Command Log

A step-by-step record of every command run to provision the GCP staging environment.

---

## 1. Authentication

```bash
# Logout all existing gcloud accounts
gcloud auth revoke --all

# Login with personal Gmail (sachinkarunagappally@gmail.com)
gcloud auth login

# Verify active account
gcloud auth list
# ACTIVE  ACCOUNT
# *       sachinkarunagappally@gmail.com

# List available projects
gcloud projects list

# Check billing is enabled on the project
gcloud billing projects describe ledgerspear
# billingEnabled: true
```

---

## 2. Enable APIs & Create Service Account

```bash
# Run the setup script (equivalent to make gcp-setup GCP_PROJECT=ledgerspear)
# This script does the following:

# Set active project
gcloud config set project ledgerspear

# Enable required GCP APIs
gcloud services enable \
  run.googleapis.com \
  sqladmin.googleapis.com \
  artifactregistry.googleapis.com \
  secretmanager.googleapis.com \
  vpcaccess.googleapis.com \
  servicenetworking.googleapis.com \
  compute.googleapis.com

# Create service account for GitHub Actions CI/CD
gcloud iam service-accounts create github-deployer \
  --display-name="GitHub Actions Deployer"

# Grant IAM roles to the service account
for ROLE in roles/run.admin roles/artifactregistry.writer \
  roles/secretmanager.secretAccessor roles/iam.serviceAccountUser \
  roles/cloudsql.client; do
  gcloud projects add-iam-policy-binding ledgerspear \
    --member="serviceAccount:github-deployer@ledgerspear.iam.gserviceaccount.com" \
    --role="$ROLE" --quiet
done

# Create service account key for GitHub Actions
gcloud iam service-accounts keys create gcp-sa-key.json \
  --iam-account="github-deployer@ledgerspear.iam.gserviceaccount.com"
# Output: gcp-sa-key.json (add as GitHub secret GCP_SA_KEY)

# Configure Docker to authenticate with Artifact Registry
gcloud auth configure-docker us-central1-docker.pkg.dev --quiet
```

---

## 3. Install Terraform

```bash
brew install terraform
# Installed: terraform 1.5.7
```

---

## 4. Terraform Init & Plan

```bash
cd deploy/gcp

# Initialize Terraform (downloads google provider ~5.45.2)
terraform init

# Plan infrastructure (uses gcloud access token for auth)
GOOGLE_OAUTH_ACCESS_TOKEN=$(gcloud auth print-access-token) \
  terraform plan -var="project_id=ledgerspear" -var="db_password=$(openssl rand -hex 16)"
# Plan: 21 to add, 0 to change, 0 to destroy
```

**Why `GOOGLE_OAUTH_ACCESS_TOKEN`?**
Terraform uses Application Default Credentials (ADC) by default. Instead of running `gcloud auth application-default login` (which opens a browser), we pass the existing gcloud CLI token directly. This works because the active gcloud account (sachinkarunagappally@gmail.com) is the project Owner.

---

## 5. Terraform Apply (First Attempt)

```bash
DB_PASS=$(openssl rand -hex 16) && \
GOOGLE_OAUTH_ACCESS_TOKEN=$(gcloud auth print-access-token) \
  terraform apply -auto-approve \
    -var="project_id=ledgerspear" \
    -var="db_password=$DB_PASS"
```

**Result:** 19/21 resources created. Cloud Run failed because no Docker image existed yet.

**Resources created:**
- `google_compute_network.vpc` — VPC (55s)
- `google_compute_global_address.private_ip_range` — Private IP range (12s)
- `google_service_networking_connection.private_vpc` — VPC peering (1m40s)
- `google_vpc_access_connector.connector` — VPC connector (2m56s)
- `google_sql_database_instance.staging` — Cloud SQL PostgreSQL 14 (12m46s)
- `google_sql_database.ledgerspear` — Database (5s)
- `google_sql_user.ledgerspear` — DB user (9s)
- `google_artifact_registry_repository.backend` — Docker repo (16s)
- `google_secret_manager_secret.*` — 5 secrets (3-5s each)
- `google_secret_manager_secret_iam_member.*` — 5 IAM bindings (7-9s each)
- `google_secret_manager_secret_version.db_password` — DB password value (7s)

**Failed:**
- `google_cloud_run_v2_service.backend` — Image not found

---

## 6. Build & Push Docker Image

```bash
# First attempt failed: Dockerfile used golang:1.23 but go.mod requires 1.24
# Fix: Updated backend/Dockerfile FROM golang:1.23-alpine → golang:1.24-alpine

# Build for linux/amd64 (Cloud Run runs amd64, Mac is ARM)
docker build --platform linux/amd64 \
  -t us-central1-docker.pkg.dev/ledgerspear/ledgerspear/backend:latest \
  -f backend/Dockerfile backend/

# Push to Artifact Registry
docker push us-central1-docker.pkg.dev/ledgerspear/ledgerspear/backend:latest
```

**Lesson:** Always use `--platform linux/amd64` when building on Apple Silicon for Cloud Run. Without it, you get `exec format error` at runtime.

---

## 7. Populate Secret Manager Values

```bash
# Encryption key (random 32-byte hex)
echo -n "$(openssl rand -hex 16)" | \
  gcloud secrets versions add ledgerspear-encryption-key --data-file=- --project=ledgerspear

# Shopify OAuth placeholders (replace with real values later)
echo -n "staging-placeholder" | \
  gcloud secrets versions add ledgerspear-shopify-client-id --data-file=- --project=ledgerspear

echo -n "staging-placeholder" | \
  gcloud secrets versions add ledgerspear-shopify-client-secret --data-file=- --project=ledgerspear

# Firebase credentials (reuse existing file from backend/)
gcloud secrets versions add ledgerspear-firebase-credentials \
  --data-file=backend/firebase-credentials.json \
  --project=ledgerspear
```

**Why needed?** Cloud Run mounts secrets as env vars and volumes. If a secret has no versions, Cloud Run refuses to start.

---

## 8. Terraform Apply (Final — Cloud Run)

```bash
GOOGLE_OAUTH_ACCESS_TOKEN=$(gcloud auth print-access-token) \
  terraform apply -auto-approve \
    -var="project_id=ledgerspear" \
    -var="db_password=$(gcloud secrets versions access latest --secret=ledgerspear-db-password --project=ledgerspear)"
```

**Result:**
```
Apply complete! Resources: 2 added, 0 changed, 1 destroyed.

Outputs:
artifact_registry_url = "us-central1-docker.pkg.dev/ledgerspear/ledgerspear"
backend_url = "https://ledgerspear-api-ineifpjrdq-uc.a.run.app"
db_private_ip = <sensitive>
```

---

## 9. Health Check

```bash
curl -sf https://ledgerspear-api-ineifpjrdq-uc.a.run.app/health
# {"status":"ok","database":"connected"}
```

---

## Summary of Resources Created

| Resource | Type | Details |
|----------|------|---------|
| `ledgerspear-staging-vpc` | VPC | Auto-create subnets |
| `ledgerspear-staging-private-ip` | Global Address | /16 for VPC peering |
| Private VPC Connection | Service Networking | servicenetworking.googleapis.com |
| `cloudrun-sql-connector` | VPC Connector | e2-micro, 2-3 instances |
| `ledgerspear-staging` | Cloud SQL | PostgreSQL 14, db-f1-micro, private IP |
| `ledgerspear` | SQL Database | Main app database |
| `ledgerspear` | SQL User | With generated password |
| `ledgerspear` | Artifact Registry | Docker repository |
| `ledgerspear-db-password` | Secret | Auto-generated |
| `ledgerspear-encryption-key` | Secret | Auto-generated |
| `ledgerspear-firebase-credentials` | Secret | From existing file |
| `ledgerspear-shopify-client-id` | Secret | Placeholder |
| `ledgerspear-shopify-client-secret` | Secret | Placeholder |
| 5x Secret IAM | IAM Binding | Compute SA → secretAccessor |
| `ledgerspear-api` | Cloud Run | 0-2 instances, 512MB |
| Public IAM | IAM Binding | allUsers → run.invoker |

**Staging URL:** https://ledgerspear-api-ineifpjrdq-uc.a.run.app

---

## Troubleshooting Encountered

| Issue | Cause | Fix |
|-------|-------|-----|
| Permission denied on APIs | Logged in as `sachin@zoko.io` (wrong account) | `gcloud auth login` with personal Gmail |
| Billing not enabled | New project, no billing linked | Link billing at console.cloud.google.com/billing |
| Billing domain mismatch | Org restriction on `zoko.io` workspace | Created project under personal Gmail |
| `terraform: command not found` | Not installed | `brew install terraform` |
| `data.google_project` permission error | Terraform ADC not set | Used `GOOGLE_OAUTH_ACCESS_TOKEN=$(gcloud auth print-access-token)` |
| Image not found | No Docker image pushed yet | Built and pushed image before Cloud Run |
| `exec format error` | Built ARM image on Mac, Cloud Run needs amd64 | `docker build --platform linux/amd64` |
| Secret versions not found | Secrets created but empty | Populated with `gcloud secrets versions add` |
| Startup probe failed | exec format error (wrong arch) | Rebuilt with correct platform |

---

## 10. Redeploy Backend (Update Migrations & Code)

When backend code or migrations change, rebuild and redeploy:

```bash
# From repo root — builds Docker image, pushes to Artifact Registry, deploys to Cloud Run
./scripts/gcp-deploy.sh ledgerspear

# What it does:
# 1. docker build -t us-central1-docker.pkg.dev/ledgerspear/ledgerspear/backend:<git-sha> backend/
# 2. docker push (both :sha and :latest tags)
# 3. gcloud run deploy ledgerspear-api --image <image> --region us-central1 --project ledgerspear
# 4. Health check: curl <service-url>/health

# Migrations auto-run on startup (baked into Docker image at /app/migrations)
```

**When to redeploy backend:**
- New database migrations added
- Backend code changes (handlers, middleware, domain logic)
- Config changes (env vars in cloudrun.tf need `terraform apply` instead)

---

## 11. Check Cloud Run Logs

```bash
# Recent application logs
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=ledgerspear-api AND textPayload:*" \
  --project=ledgerspear --limit=30 --format="table(timestamp,textPayload)"

# Error logs only
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=ledgerspear-api AND severity>=WARNING" \
  --project=ledgerspear --limit=10 --format=json

# Full request logs (includes HTTP status codes)
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=ledgerspear-api" \
  --project=ledgerspear --limit=20 --format="table(timestamp,httpRequest.status,httpRequest.requestUrl,textPayload)"
```

---

## Troubleshooting Encountered (continued)

| Issue | Cause | Fix |
|-------|-------|-----|
| `failed to lookup user` (500) | Cloud SQL schema missing columns — stale Docker image without latest migrations | Redeploy backend: `./scripts/gcp-deploy.sh ledgerspear` |
| `column "daily_summary" does not exist` | Migration 009+ not applied | Same fix — redeploy triggers auto-migration on startup |
