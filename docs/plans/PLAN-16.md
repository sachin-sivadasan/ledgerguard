# PLAN-16: AI Chat + Internal GraphQL (12 Commits)

**Date:** 2026-03-07
**Status:** Completed

## Scope
12-commit micro-phase implementation of AI-powered chat with internal GraphQL query layer.

### Backend
- Internal GraphQL schema + gqlgen code generation
- GraphQL resolvers delegating to existing domain services
- Module plugin architecture (6 modules, 16 tools)
- AIClient interface with OpenAI implementation (gpt-4o)
- SSE streaming chat handler
- Database migration for chat history

### Frontend
- ChatBloc with SSE streaming
- ChatPage with split-pane layout (ChatPane + DataPanel)
- MessageBubble (user/assistant styles)
- DataPanel for structured data display
- Suggestion chips for follow-up questions

## Key Decisions
- ADR-011: gqlgen for Internal GraphQL Layer
- ADR-012: OpenAI-First with AIClient Interface
- ADR-013: WebSocket for Chat Communication (later changed to SSE)
- ADR-014: Module Plugin Architecture for Chat Tools

## Modules
| Module | Tools |
|--------|-------|
| Risk | 3 tools (risk overview, breakdown, timeline) |
| Subscriptions | 4 tools (list, detail, search, filter) |
| Metrics | 3 tools (latest, snapshot, trends) |
| Store Health | 2 tools (health, history) |
| Earnings | 2 tools (summary, timeline) |
| Sync | 2 tools (trigger, status) |

## Tests
- 47 Go tests + 5 Flutter tests
- ChatBloc: simple response, tool call streaming, error handling, clear, data state

## Prompt File
- `docs/prompts/chat-builder-system-prompt.md`
