# F02. Public GraphQL API

## What It Will Do
Promote the internal GraphQL schema (currently used by AI Chat) into a public developer API. External developers can query their revenue data, subscription metrics, and risk states programmatically. API keys provide authentication, rate limiting enforces fair usage.

## Why It Matters
LedgerGuard's internal GraphQL already has resolvers for subscriptions, metrics, risk, earnings, and store health. Exposing this as a public API creates a developer platform and opens integration possibilities (custom dashboards, CI/CD checks on revenue health, Slack bots).

## Dependencies
- Internal GraphQL schema and resolvers (implemented — `backend/internal/chat/graphql/`)
- API key system (implemented — `api_keys` table, `ApiKeyHandler`)
- Rate limiting middleware (partially implemented)

## Integration Points
- Reuse existing gqlgen resolvers (no new business logic needed)
- Add API key authentication middleware (separate from Firebase)
- Rate limiting per API key tier
- Usage tracking in `api_audit_log` table

## Estimated Scope
- API key auth middleware: 1 day
- Rate limiting: 1-2 days
- Schema cleanup (remove internal-only fields): 1 day
- Documentation (GraphQL Playground customization): 1 day
- Total: ~4-5 days

## Open Questions
- Should the public schema be a subset of the internal schema, or a separate schema?
- Rate limits per plan: Free (100 req/day), Pro (10,000 req/day)?
- Do we need field-level permissions (e.g., hide revenue data from VIEWER role)?
- Versioning strategy: URL-based (`/api/v1/graphql`) or schema-based?

## Suggested Approach
1. Create a separate GraphQL handler for the public API (`/api/v1/graphql`)
2. Add API key middleware that validates `X-API-Key` header
3. Create a curated public schema (subset of internal schema)
4. Add rate limiting per API key with bucket-based enforcement
5. Log all queries to `api_audit_log` for usage tracking
6. Deploy with GraphQL Playground for developer exploration
