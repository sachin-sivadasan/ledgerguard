# PLAN-13: Revenue API & API Key Management

**Date:** 2026-02-28
**Status:** Completed

## Scope

### Backend
- Revenue API endpoint `/api/v1/apps/{appId}/revenue` with date range and charge type filtering
- Live FetchTransactions from Shopify Partner API (12-month rolling window, pagination)
- Shop name extraction, gross amount calculation, period-based usage revenue
- API Key CRUD (generate, list, revoke) with rate limiting per key
- API audit log for tracking key usage

### Frontend
- API Key management UI in settings (generate, copy, revoke with confirmation)
- Revenue API documentation page on marketing site

## API Endpoints
- `GET /api/v1/apps/{appId}/revenue` — Revenue data with filters
- `POST /api/v1/api-keys` — Generate new API key
- `GET /api/v1/api-keys` — List user's API keys
- `DELETE /api/v1/api-keys/{id}` — Revoke API key

## Blocs
- `ApiKeyBloc` — Key management state
