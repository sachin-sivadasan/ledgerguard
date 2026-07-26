# PLAN-15: Deployment (Firebase Hosting + GCP Staging)

**Date:** 2026-03-04
**Status:** Completed

## Scope
- Deploy Flutter web frontend to Firebase Hosting
- Production entry point (`main_prod.dart`) with Firebase initialization
- Staging environment (`main_staging.dart`) pointing to GCP Cloud Run backend
- GCP Cloud Run deployment for staging backend
- Cloud SQL (PostgreSQL 14, db-f1-micro) for staging database
- CI/CD pipeline (GitHub Actions) for build and deploy
- CORS configuration for Firebase Hosting domains

## Key Decisions
- ADR-009: GCP Cloud Run for Staging Environment
- ADR-010: Staging Frontend via Firebase Hosting + main_staging.dart

## Environment URLs
| Environment | Frontend | Backend |
|-------------|----------|---------|
| Dev | localhost | http://localhost:8080 |
| Staging | ledgerguard-c7557.web.app | ledgerspear-api-ineifpjrdq-uc.a.run.app |
| Production | TBD | api.ledgerspear.com (Hetzner) |

## Docker
- `--platform linux/amd64` required for Apple Silicon → Cloud Run
- `gcp-deploy.sh` script handles this automatically
