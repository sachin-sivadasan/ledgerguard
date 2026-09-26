# PRD: AI Chat (as-built)

**Date:** 2026-09-25 · **Author:** sachin.s · **Status:** Draft (backfilled from implementation)

> **As-built PRD**, reverse-derived from code. Markers: **FACT** (code proves it),
> **INFERRED** (behavior implies it), **UNKNOWN** (rationale not recorded → Open question).
> Success metrics are **PROPOSED** (future), not historical. Primary sources:
> `backend/internal/chat/` (handler, registry, modules, graphql), `openai_client.go`,
> `frontend-flutter/lib/screens/insights/` + `providers/insights_provider.dart`.

## ⚠️ Headline finding (read first)

**FACT — the backend AI chat is fully built and tested, but the shipped Flutter UI is NOT connected to it.** `insights_provider.dart:73-98` implements `sendMessage` as a hardcoded mock: an 800ms-delayed `_generateResponse` that returns canned strings by keyword matching (`mrr`, `churn`, `forecast`), never calling `POST /api/v1/chat`. Only the *daily insights* feed (`/api/v1/apps/{appId}/insights/daily`) is live. So "chat with your revenue data" is real on the server and a demo on the client.

## Problem & evidence

A partner has revenue/risk/earnings data across many stores and wants **answers in natural language** ("Which stores are at risk?", "What's my MRR trend?") instead of navigating 20+ report screens.

- **INFERRED:** the backend design (7 tool modules over an internal GraphQL read-API) exists precisely to answer such questions grounded in the partner's own data.
- **UNKNOWN — original demand evidence not recorded.** No tickets/usage metrics stored. Open question: validate demand before wiring the real client — owner: Product.

## Target users

**INFERRED (from module scope):** the same **Shopify app partner** persona as Reports — founder/growth/success wanting fast answers about subscriptions, risk, earnings, store health, metrics, and sync status for a selected app/org.
- **Not the target:** merchants; users needing write/actions via chat (the assistant is read-only — see Non-goals).

## Proposed solution (as-built behavior)

**FACT (backend).** A user sends a message thread to `POST /api/v1/chat`; the server streams back **Server-Sent Events** (`text/event-stream`, `handler.go:100-103`) of typed events (`tool_call`, `tool_result`, `response`, `error`). The server builds a system prompt from a base + per-module fragments (`handler.go:107-108`, `buildSystemPrompt`), sends the thread + available tools to an LLM (OpenAI `gpt-4o` default), and runs a **tool-call loop up to 5 iterations** (`maxToolIterations = 5`, `handler.go:11`). Each tool call (`module__tool` naming) resolves to an **internal GraphQL query** (`graphql_executor.go`) over the partner's data, results stream back, and a final text `response` event carries the answer plus a per-module **state bundle** and auto-generated **follow-up suggestions** (`handler.go:291-344`).

Tool modules available (**FACT**, `internal/chat/modules/`): **risk** (summary / list-at-risk / timeline), **subscriptions** (list / detail / summary / search), **metrics** (latest / trend / aggregate), **store_health** (check / compare ≤10), **earnings** (breakdown / transactions), **apps** (list / detail), **sync** (trigger / status). `GET /api/v1/chat/modules` lists them.

**FACT (frontend).** The Insights screen (`screens/insights/insights_screen.dart`) shows a chat sidebar ("Ask about your revenue"), but its responses are the mock described above — **the real SSE contract is unused by the app today.**

### Key screens

**No new wireframe** (as-built). Real surface: `frontend-flutter/lib/screens/insights/insights_screen.dart` (chat sidebar + daily-insights feed). The intended-but-unwired data path is `insights_provider → insights_service → POST /api/v1/chat` (SSE).

## Platform & policy constraints

**FACT/INFERRED.** Answers are grounded only in data LedgerGuard already syncs (Shopify Partner API–derived): subscriptions, transactions, metrics snapshots, app reviews. The assistant is **read-only** — GraphQL schema is query-only (`schema.graphql`), so it cannot mutate data (the lone exception: the **sync** module can *trigger* a background refresh). LLM calls depend on an external OpenAI key/quota (upstream rate limits apply). Tenant isolation relies on org context injected by middleware and enforced in GraphQL resolvers (see design doc).

## Pricing-tier impact

**UNKNOWN — no plan-gating for chat found in code.** AI calls carry a real per-message OpenAI cost, so this is a live cost/pricing question. Open question: meter or gate AI chat by tier? — owner: Product.

## Migration for existing users

**INFERRED.** Wiring the real client would *replace* the mock responses users see today. Comms/expectation-setting needed: answers become real (and occasionally "I don't have that data") instead of always-confident canned text. No data migration.

## Success metrics (PROPOSED — future-facing)

None measured today. Proposed once the real client is wired:
1. **Wire-up done:** 0% of chat responses come from `_generateResponse`; 100% from `POST /api/v1/chat` (the mock is deleted).
2. **Groundedness:** ≥ 90% of answers that state a number cite a value returned by a tool call in the same turn (no hallucinated figures).
3. **Usefulness:** ≥ 30% of chat sessions use ≥ 1 follow-up suggestion within 60 days.

## Non-goals (deliberately absent in code)

1. **No write actions via chat** — read-only GraphQL; assistant can't change subscriptions/settings (only trigger a sync). **FACT.**
2. **No multi-provider LLM today** — the provider registry abstracts providers but only **OpenAI** is implemented; no Claude/other. **FACT** (`ai_provider_registry.go`).
3. **No cross-org/global assistant** — scoped to the caller's org + a selected app. **INFERRED.**
4. **No voice** (a separate marketing concept) — chat is text/SSE only. **FACT.**

## Open questions

- **Wire the real client** — the app still ships the mock. Is this intended (beta gate) or a gap? Owner: Eng/Product.
- Validate demand (no evidence recorded). Owner: Product.
- Pricing/metering for per-message LLM cost. Owner: Product.
- Multi-provider: is Claude support planned, or is the registry over-abstraction? Owner: Eng. (Repo default should be latest Claude for new AI work.)
- CLAUDE.md §12 documents chat as **WebSocket**; code is **SSE** — doc fix needed (tracked in design). Owner: Eng.
