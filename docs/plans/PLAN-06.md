# PLAN-06: Marketing Site (Next.js)

**Date:** 2026-02-27
**Status:** Completed

## Scope
- Initialize Next.js 14+ project with App Router
- TailwindCSS styling, responsive mobile-first design
- Public landing page (no authentication)
- Interactive visualization pages for product education

## Visualization Pages
| Page | Prompt File |
|------|-------------|
| `/kpi-guide` | `docs/prompts/kpi-metrics-visualization.md` |
| `/money-flow` | `docs/prompts/shopify-money-flow-diagram.md` |
| `/architecture` | `docs/prompts/internal-architecture-flow.md` |
| `/api-guide` | `docs/prompts/ledgerguard-api-integration.md` |
| `/affiliate-program` | `docs/prompts/affiliate-program-flow.md` |
| `/notifications` | `docs/prompts/notification-engine-flow.md` |
| `/pitch` | `docs/prompts/customer-pitch-ui.md` |
| `/voice-assistant` | `docs/prompts/voice-assistant-flow.md` |
| `/hetzner-infrastructure` | `docs/prompts/hetzner-infrastructure-visualization.md` |
| `/gcp-staging` | `docs/prompts/gcp-staging-visualization.md` |

## Structure
```
marketing/site/
├── app/          → Pages and layouts
├── components/   → Reusable UI components
└── public/       → Static assets
```
