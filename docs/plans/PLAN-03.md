# PLAN-03: Shopify Partner Integration (OAuth + Manual)

**Date:** 2026-02-26
**Status:** Completed

## Scope
- Implement Shopify Partner OAuth flow (startOAuth, callback, token exchange)
- Manual partner token integration (admin-only fallback)
- Fetch apps from Shopify Partner API
- PartnerSyncService orchestrating full sync cycle
- Store partner accounts and apps in database

## Key Decisions
- ADR-006: OAuth State Validation for CSRF Protection
- ADR-007: Tenant Isolation in Sync Handler

## API Endpoints
- `GET /api/v1/integrations/shopify/oauth` — Start OAuth flow
- `GET /api/v1/integrations/shopify/callback` — OAuth callback (public)
- `GET /api/v1/integrations/shopify/status` — Check integration status
- `POST /api/v1/admin/manual-token` — Manual token entry (admin)
- `GET /api/v1/apps` — List apps from Partner API
