# Hetzner Deployment Setup - Interactive Visualization

## Context
You are a senior DevOps + visualization engineer building an interactive animated guide showing how LedgerGuard deploys to Hetzner Cloud infrastructure.

Build an educational visualization that helps developers understand:
1. Infrastructure components and their relationships
2. CI/CD pipeline from GitHub to production
3. SSL/TLS termination and domain routing
4. Database and backup strategy
5. Monitoring and logging setup

---

## Design Philosophy

### Target Audience
Developers and DevOps engineers who:
- Want to understand the production architecture
- Need to deploy or maintain the infrastructure
- Want to replicate the setup for their own projects
- Are evaluating Hetzner as a cloud provider

### Key Principles
1. **Show the flow** - Animated deployment pipeline
2. **Component relationships** - How services connect
3. **Security layers** - SSL, firewall, secrets
4. **Cost transparency** - Hetzner pricing for each component

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      LEDGERGUARD PRODUCTION ARCHITECTURE                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────┐     ┌─────────────────────────────────────────────────┐   │
│  │   Hetzner   │     │              Hetzner Cloud VPS                   │   │
│  │   DNS/CDN   │     │  ┌─────────────────────────────────────────┐    │   │
│  │             │────▶│  │          Caddy (Reverse Proxy)          │    │   │
│  │ ledgerguard │     │  │    - SSL/TLS (Let's Encrypt)            │    │   │
│  │   .com      │     │  │    - Auto HTTPS                          │    │   │
│  └─────────────┘     │  └─────────────────────────────────────────┘    │   │
│                      │              │                │                  │   │
│                      │              ▼                ▼                  │   │
│                      │  ┌───────────────┐  ┌───────────────────┐       │   │
│                      │  │  Go Backend   │  │  Marketing Site   │       │   │
│                      │  │  (API :8080)  │  │  (Next.js :3000)  │       │   │
│                      │  └───────┬───────┘  └───────────────────┘       │   │
│                      │          │                                       │   │
│                      │          ▼                                       │   │
│                      │  ┌───────────────┐                               │   │
│                      │  │  PostgreSQL   │                               │   │
│                      │  │   (Managed)   │                               │   │
│                      │  └───────────────┘                               │   │
│                      └─────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Flow 1: CI/CD Deployment Pipeline

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         CI/CD DEPLOYMENT FLOW                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  👨‍💻 Developer    →    📦 GitHub    →    ⚙️ Actions    →    🚀 Hetzner     │
│  git push             Repository        CI/CD Pipeline       Production     │
│                                                                              │
│  Steps:                                                                      │
│  1. Push to main branch                                                      │
│  2. GitHub Actions triggered                                                 │
│  3. Run tests (go test, flutter test)                                       │
│  4. Build binaries (go build, flutter build)                                │
│  5. SSH to Hetzner VPS                                                       │
│  6. Deploy new version                                                       │
│  7. Restart services (systemd)                                              │
│  8. Health check                                                             │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### GitHub Actions Workflow

```yaml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Go tests
        run: go test ./...
      - name: Run Flutter tests
        run: flutter test

  deploy:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - name: SSH Deploy
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.HETZNER_HOST }}
          username: ${{ secrets.HETZNER_USER }}
          key: ${{ secrets.HETZNER_SSH_KEY }}
          script: |
            cd /opt/ledgerguard
            git pull origin main
            make build
            sudo systemctl restart ledgerguard
```

---

## Flow 2: Request Routing

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         REQUEST ROUTING FLOW                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  🌐 User         →    🔒 Caddy      →    🎯 Route      →    🖥️ Service      │
│  HTTPS Request        SSL Termination    Decision          Handler          │
│                                                                              │
│  Routes:                                                                     │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  api.ledgerguard.com/*    → localhost:8080  (Go Backend)           │   │
│  │  ledgerguard.com/*        → localhost:3000  (Next.js Marketing)    │   │
│  │  app.ledgerguard.com/*    → Static files    (Flutter Web)          │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Caddyfile Configuration

```
api.ledgerguard.com {
    reverse_proxy localhost:8080
    encode gzip
    log {
        output file /var/log/caddy/api.log
    }
}

ledgerguard.com {
    reverse_proxy localhost:3000
    encode gzip
}

app.ledgerguard.com {
    root * /var/www/ledgerguard-app
    file_server
    try_files {path} /index.html
}
```

---

## Flow 3: Database & Backup Strategy

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       DATABASE & BACKUP STRATEGY                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Option A: Hetzner Managed PostgreSQL                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  + Automatic backups (daily)                                         │   │
│  │  + Point-in-time recovery                                            │   │
│  │  + High availability option                                          │   │
│  │  + Automatic updates                                                 │   │
│  │  - Higher cost (~€15/month)                                         │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  Option B: Self-Managed PostgreSQL on VPS                                    │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  + Lower cost (included in VPS)                                      │   │
│  │  + Full control                                                       │   │
│  │  - Manual backup scripts                                             │   │
│  │  - Manual updates                                                    │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  Backup Script (Option B):                                                  │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  0 2 * * * pg_dump ledgerguard | gzip > /backups/$(date +%Y%m%d).gz │   │
│  │  0 3 * * * rclone sync /backups hetzner-storage:ledgerguard-backups │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Flow 4: Monitoring & Logging

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       MONITORING & LOGGING STACK                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐             │
│  │   Go Backend    │  │     Caddy       │  │   PostgreSQL    │             │
│  │   (stdout)      │  │   (access.log)  │  │   (pg_stat)     │             │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘             │
│           │                    │                    │                       │
│           └────────────────────┼────────────────────┘                       │
│                                ▼                                            │
│                    ┌───────────────────────┐                               │
│                    │    journalctl / logs  │                               │
│                    └───────────┬───────────┘                               │
│                                │                                            │
│              ┌─────────────────┼─────────────────┐                         │
│              ▼                 ▼                 ▼                         │
│     ┌─────────────┐   ┌─────────────────┐  ┌─────────────┐                │
│     │  UptimeRobot │   │  Hetzner Alerts │  │   Grafana   │                │
│     │  (External)  │   │  (CPU, Disk)    │  │  (Optional) │                │
│     └─────────────┘   └─────────────────┘  └─────────────┘                │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Infrastructure Components

| Component | Hetzner Product | Specs | Monthly Cost |
|-----------|-----------------|-------|--------------|
| VPS | CX21 | 2 vCPU, 4GB RAM, 40GB SSD | €5.18 |
| Managed DB | PostgreSQL | 2 vCPU, 4GB RAM | €14.50 |
| Storage | Storage Box | 100GB for backups | €3.81 |
| DNS | Hetzner DNS | Included | Free |
| **Total** | | | **~€23/month** |

### Alternative: Self-Hosted DB

| Component | Hetzner Product | Specs | Monthly Cost |
|-----------|-----------------|-------|--------------|
| VPS | CX31 | 4 vCPU, 8GB RAM, 80GB SSD | €9.74 |
| Storage | Storage Box | 100GB for backups | €3.81 |
| DNS | Hetzner DNS | Included | Free |
| **Total** | | | **~€14/month** |

---

## Security Layers

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          SECURITY LAYERS                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Layer 1: Hetzner Firewall                                                   │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  - Allow: 22 (SSH), 80 (HTTP), 443 (HTTPS)                          │   │
│  │  - Block: All other inbound                                          │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  Layer 2: UFW (Host Firewall)                                                │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  - SSH: Only from known IPs (optional)                               │   │
│  │  - PostgreSQL: localhost only                                        │   │
│  │  - API: localhost only (behind Caddy)                               │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  Layer 3: SSL/TLS (Caddy)                                                    │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  - Automatic Let's Encrypt certificates                              │   │
│  │  - Auto-renewal                                                       │   │
│  │  - HSTS enabled                                                       │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  Layer 4: Application Security                                               │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  - Firebase Auth (JWT validation)                                    │   │
│  │  - Rate limiting (per IP)                                            │   │
│  │  - Input validation                                                   │   │
│  │  - SQL injection prevention (parameterized queries)                  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Deployment Checklist

### Initial Setup
- [ ] Create Hetzner Cloud account
- [ ] Provision VPS (CX21 or CX31)
- [ ] Add SSH key
- [ ] Configure firewall rules
- [ ] Point DNS to server IP

### Server Configuration
- [ ] Update system packages
- [ ] Install Go, PostgreSQL, Node.js
- [ ] Install Caddy
- [ ] Create ledgerguard user
- [ ] Clone repository

### Application Setup
- [ ] Copy config.yaml (secrets from GitHub Secrets or Vault)
- [ ] Run database migrations
- [ ] Build Go binary
- [ ] Build Next.js marketing site
- [ ] Build Flutter web app

### Service Configuration
- [ ] Create systemd service for Go backend
- [ ] Create systemd service for Next.js
- [ ] Configure Caddy with domains
- [ ] Enable and start services

### Final Steps
- [ ] Test all endpoints
- [ ] Set up monitoring (UptimeRobot)
- [ ] Configure backup script
- [ ] Document runbook

---

## Technical Requirements

### Framework
- Next.js 14+ with App Router
- TailwindCSS for styling
- React hooks for state and animation

### Animation Approach
- SVG-based architecture diagrams
- requestAnimationFrame for smooth animation
- Step-by-step deployment flow
- Play/pause controls

### Visual Style
- Dark theme (slate-950 background)
- Gradient accents (blue to cyan for infrastructure)
- Glowing effects for data flow
- Hetzner brand colors

### Interactions
- Flow type selector (tabs)
- Play/pause deployment animation
- Cost calculator toggle
- Expand/collapse configuration sections

---

## Component Structure

```
marketing/site/
├── app/deployment/page.tsx             # Page wrapper
└── components/
    └── DeploymentFlowVisualization.tsx  # Main visualization
        ├── ArchitectureDiagram          # Infrastructure overview
        ├── FlowSelector                  # Tab buttons
        ├── DeploymentPipeline           # CI/CD animation
        ├── CostCalculator               # Pricing breakdown
        └── SecurityLayers               # Security documentation
```

---

## Implementation Notes

1. **Animation Loop:** Use requestAnimationFrame with 2s step duration
2. **SVG Architecture:** Server boxes with service icons inside
3. **Flow Arrows:** Animated dashed lines showing data/deploy flow
4. **Cost Display:** Toggle between managed vs self-hosted pricing
5. **Copy Buttons:** Allow copying configuration snippets
