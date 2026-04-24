# Tech Stack

All technologies used in LedgerGuard with versions and rationale.

---

## Backend

| Technology | Version | Purpose | Rationale |
|-----------|---------|---------|-----------|
| Go | 1.22+ | Application language | Fast compilation, strong concurrency, single binary deploy |
| Chi | v5 | HTTP router | Lightweight, middleware-friendly, stdlib-compatible |
| pgx | v5 | PostgreSQL driver | Pure Go, connection pooling, prepared statements |
| golang-migrate | v4 | Database migrations | Up/down SQL migrations, dirty state recovery |
| Firebase Admin SDK | latest | Token verification | Stateless JWT verification, no session storage |
| gorilla/websocket | v1 | WebSocket (AI chat) | Mature, well-tested, supports ping/pong |
| gqlgen | latest | GraphQL code generation | Schema-first, type-safe resolvers for AI chat |
| openai-go | latest | OpenAI API client | Function calling for AI chat tools |
| razorpay-go | v1 | Payment processing | Razorpay Subscriptions API for billing |
| crypto/aes | stdlib | Token encryption | AES-256-GCM for Partner API token storage |

## Database

| Technology | Version | Purpose | Rationale |
|-----------|---------|---------|-----------|
| PostgreSQL | 16 (prod) / 14 (staging) | Primary database | ACID, JSON support, mature ecosystem |
| pgcrypto | extension | UUID generation | `gen_random_uuid()` for primary keys |

## Frontend (Flutter Prototype)

| Technology | Version | Purpose | Rationale |
|-----------|---------|---------|-----------|
| Flutter | 3.x | Cross-platform UI | Web + iOS + Android from single codebase |
| Dart | 3.x | Application language | Null safety, strong typing, hot reload |
| Provider | latest | State management | Simple, lightweight for prototype phase |
| GoRouter | latest | Navigation | Declarative routing, deep linking |
| http | latest | API client | Standard HTTP requests to backend |
| Firebase Auth (Flutter) | latest | Authentication | Google OAuth, email/password login |
| fl_chart | latest | Charts & graphs | Revenue charts, risk funnels, analytics |
| intl | latest | Formatting | Currency, dates, number formatting |

## Marketing Site

| Technology | Version | Purpose | Rationale |
|-----------|---------|---------|-----------|
| Next.js | 14+ | Static site framework | App Router, SSG, React Server Components |
| React | 18+ | UI library | Component model, ecosystem |
| TailwindCSS | 3.x | Styling | Utility-first, responsive, minimal CSS |
| TypeScript | 5.x | Type safety | Catch errors at build time |

## Infrastructure

| Technology | Purpose | Environment | Rationale |
|-----------|---------|-------------|-----------|
| Hetzner Cloud (CX31) | Production server | Production | 4 vCPU, 8GB RAM, ~$15/mo |
| Caddy | Reverse proxy + auto-SSL | Production | Automatic HTTPS, simple config |
| systemd | Process management | Production | Built into Linux, reliable |
| GCP Cloud Run | Serverless containers | Staging | Scale-to-zero, free credits |
| Cloud SQL | Managed PostgreSQL | Staging | Private IP, automatic backups |
| Artifact Registry | Docker image storage | Staging | Integrated with Cloud Run |
| Secret Manager | Credential storage | Staging | Encrypted at rest, versioned |
| Firebase Hosting | Static web hosting | All | CDN, SPA routing, free tier |
| Firebase Auth | Identity provider | All | Google OAuth, email/password |
| GitHub Actions | CI/CD | All | Automated tests, lint, deploy |

## External APIs

| Service | Purpose | Tier |
|---------|---------|------|
| Shopify Partner API (GraphQL) | Transaction data, subscriptions | All |
| Shopify Storefront API | Store brand data, logos | All |
| Shopify App Store | App review scraping | All |
| OpenAI GPT-4o | AI chat function calling | Pro |
| OpenAI GPT-4o-mini | Daily insight generation | Pro |
| Razorpay Subscriptions | B2B SaaS billing | All |
| Firebase Cloud Messaging | Push notifications | All |
| Postmark | Transactional email | All (planned) |
| n8n | Automation workflows | All (planned) |

---

## Build & Deploy Commands

| Action | Command |
|--------|---------|
| Backend tests | `cd backend && go test ./... -v` |
| Backend server | `cd backend && go run ./cmd/server -config config.local.yaml` |
| Backend lint | `cd backend && golangci-lint run` |
| Backend format | `cd backend && go fmt ./...` |
| Flutter tests | `cd frontend/app && flutter test` |
| Flutter web (staging) | `cd frontend/app && flutter build web --release -t lib/main_staging.dart` |
| Flutter web (prod) | `cd frontend/app && flutter build web --release -t lib/main_prod.dart` |
| Marketing dev | `cd marketing/site && npm run dev` |
| Marketing build | `cd marketing/site && npm run build` |
| GCP deploy | `./scripts/gcp-deploy.sh ledgerspear` |
| Firebase deploy | `cd frontend/app && firebase deploy --only hosting` |
| GraphQL generate | `cd backend && go generate ./internal/chat/graphql/` |
