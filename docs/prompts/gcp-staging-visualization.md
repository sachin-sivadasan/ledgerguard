# GCP Staging Architecture – Prompt Specification

## Context
LedgerGuard uses a dual-environment setup: Hetzner Cloud for production ($15/mo) and GCP Cloud Run for staging (free credits). This visualization shows both architectures side by side, the CI/CD branching strategy, and the GCP service topology.

## Audience
- Developers understanding the staging environment
- DevOps engineers deploying to staging
- Stakeholders reviewing infrastructure costs

## Page Route
`/gcp-staging`

## Flow Diagrams

### Flow 1: Dual Architecture
```
Hetzner (Production):  VPS → Caddy → Go API → PostgreSQL
GCP (Staging):         Cloud Run → Go API → VPC → Cloud SQL
```

### Flow 2: GCP Service Topology
```
Artifact Registry → Cloud Run → Secret Manager
                         │
                    VPC Connector → Cloud SQL
```

### Flow 3: CI/CD Branching
```
Developer → GitHub → Actions → [main] → Hetzner
                             → [staging] → Cloud Run
```

### Flow 4: Request Flow (GCP)
```
User → HTTPS → Go API → VPC Connector → Cloud SQL
```

## Static Sections

### GCP Services (4 cards)
- Cloud Run: serverless, scale-to-zero, auto-HTTPS
- Cloud SQL: managed PostgreSQL 14, private IP, db-f1-micro
- Artifact Registry: Docker storage, SHA-tagged, vulnerability scanning
- Secret Manager: DB password, Firebase creds, encryption key, Shopify OAuth

### Environment Comparison
- Hetzner Production: VPS CX31, self-hosted PostgreSQL, Caddy, $15/mo
- GCP Staging: Cloud Run, Cloud SQL, auto-HTTPS, $0 (free credits)

### CI/CD Branching Strategy
- main → Hetzner (SSH + go build)
- staging → GCP (Docker build + Cloud Run deploy)

### Required GitHub Secrets
- Existing: HETZNER_HOST, HETZNER_USER, HETZNER_SSH_KEY
- New: GCP_PROJECT_ID, GCP_REGION, GCP_SA_KEY

### Cost Comparison
- Hetzner: ~$15/mo (recommended for production)
- GCP: $0/mo (free credits, ~$20-30/mo after expiry)

## Color Theme
- Primary: blue-400 to purple-400 gradient
- Background: from-slate-950 via-blue-950 to-slate-950
- Accent colors: blue, purple, indigo
- Distinct from orange (Hetzner infra) and cyan (deployment)

## Technical Requirements
- Next.js 14+ App Router
- TailwindCSS
- SVG animated flows with play/pause
- `'use client'` directive
- Responsive design
