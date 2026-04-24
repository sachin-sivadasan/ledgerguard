# F03. Claude AI Provider

## What It Will Do
Add Anthropic's Claude as a parallel AI provider alongside OpenAI for the AI Chat feature. Users can select their preferred provider in settings. Claude would handle the same function calling workflow (tool definitions, tool loop, response generation).

## Why It Matters
Provider diversity reduces single-vendor risk and gives users choice. Claude's tool use capabilities are well-suited for structured data queries. Some users may prefer Claude's response style for financial analysis.

## Dependencies
- AIClient interface (implemented — `backend/internal/chat/ai_client.go`)
- AIProviderRegistry (implemented — `backend/internal/chat/ai_provider_registry.go`)
- User preference storage (implemented — `ai_provider` column in `notification_preferences`)
- See [ADR-012](../../../DECISIONS.md)

## Integration Points
- Implement `AIClient` interface with Anthropic's API
- Register in `AIProviderRegistry` alongside OpenAI
- Convert tool definitions to Claude's tool_use format
- Handle Claude's tool_use response format (content blocks vs function calls)
- User selects provider in Settings → saved to preferences

## Estimated Scope
- Claude API client implementation: 2-3 days
- Tool format conversion (OpenAI ↔ Claude): 1-2 days
- Provider selection UI in Flutter: 0.5 day
- Testing with all 6 chat modules: 1-2 days
- Total: ~5-7 days

## Open Questions
- Which Claude model? (claude-sonnet-4-20250514 for cost-efficiency, claude-opus-4-20250514 for quality)
- Should we support streaming responses from Claude?
- How to handle tool format differences (OpenAI uses `function` calling, Claude uses `tool_use`)?
- Pricing comparison: should users see cost estimates per query?

## Suggested Approach
1. Create `ClaudeClient` implementing `AIClient` interface
2. Use Anthropic Go SDK or HTTP client for API calls
3. Convert `ToolDefinition` to Claude's `tool` format in the client
4. Parse Claude's `tool_use` content blocks back to `ToolCall` structs
5. Register in `AIProviderRegistry` with key `"claude"`
6. Add provider selector to Flutter settings screen
7. Test all 6 modules with Claude to verify tool calling accuracy
