# 10. AI Insight Service

## What It Does
Generates a daily AI-powered revenue brief for each app, summarizing key metrics in 80-120 words of natural language. The service takes a `DailyMetricsSnapshot` as input, constructs a structured prompt including MRR, revenue at risk, usage revenue, total revenue, and subscription health counts, then sends it to an LLM for completion. The result is stored as a `DailyInsight` entity, one per app per day, using an upsert strategy for idempotency. This is a Pro tier-only feature; FREE tier users receive `ErrProTierRequired`.

## Architecture
Application layer service (`internal/application/service/`). The `AIInsightService` depends on three interfaces:
- `AIProvider` — abstracts the LLM API call behind `GenerateCompletion(ctx, prompt)`. Currently implemented with OpenAI GPT-4o-mini, but the interface is provider-agnostic and can target Claude or any other LLM.
- `DailyInsightRepository` — persistence interface for upsert and retrieval of insights.
- `UserRepository` — used solely to check the user's plan tier before generating an insight.

The service follows the dependency inversion principle: the application layer defines the `AIProvider` interface, and the infrastructure layer provides the concrete implementation. The domain layer has zero knowledge of AI providers.

## Key Files
| File | Purpose |
|------|---------|
| `backend/internal/application/service/ai_insight_service.go` | AIInsightService: GenerateInsight, BuildPrompt, tier check |
| `backend/internal/domain/entity/daily_insight.go` | DailyInsight entity: ID, AppID, Date, InsightText, CreatedAt |
| `backend/internal/domain/repository/daily_insight_repository.go` | DailyInsightRepository interface: Upsert, FindByAppIDAndDate, FindByAppIDRange, FindLatestByAppID |

## Data Flow
```
GenerateInsight(ctx, userID, appID, snapshot, now)
│
├── Look up user by userID
│     └── If user.PlanTier != PRO → return ErrProTierRequired
│
├── BuildPrompt(snapshot)
│     ├── Convert all cent values to dollars (divide by 100)
│     ├── Multiply RenewalSuccessRate by 100 for percentage display
│     └── Format into structured prompt with:
│           - Revenue section: ActiveMRR, RevenueAtRisk, UsageRevenue, TotalRevenue
│           - Subscription health: RenewalRate, TotalSubs, Safe/OneCycle/TwoCycle/Churned counts
│           - Instructions: 80-120 words, one sentence summary, one insight, one recommendation
│
├── aiProvider.GenerateCompletion(ctx, prompt)
│     └── Returns insightText string from LLM
│
├── entity.NewDailyInsight(appID, now, insightText)
│     ├── Generates new UUID
│     ├── Truncates date to start of day (midnight UTC)
│     └── Sets CreatedAt to current UTC time
│
└── insightRepo.Upsert(insight)
      └── ON CONFLICT (app_id, date) DO UPDATE
```

## Configuration
| Setting | Value | Notes |
|---------|-------|-------|
| Target word count | 80-120 words | Enforced in prompt instructions, not validated programmatically |
| Date truncation | Midnight UTC | `time.Date(year, month, day, 0, 0, 0, 0, time.UTC)` |
| Plan tier gate | PRO only | Checked via `valueobject.PlanTierPro` |
| LLM model | GPT-4o-mini (via AIProvider) | Provider-agnostic; swap by changing the infrastructure implementation |

The prompt instructs the LLM to:
1. Summarize overall health in one sentence
2. Highlight the most critical insight or trend
3. Provide one actionable recommendation
4. Use professional but conversational tone
5. No bullet points in the response

## API Surface
The AI insight is generated as part of the sync/snapshot pipeline, not exposed as a standalone HTTP endpoint for triggering generation. Insights are surfaced through:

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| — | Generated during sync pipeline | Firebase | One insight per app per day, after snapshot is computed |

The `DailyInsightRepository` interface supports retrieval patterns:
- `FindByAppIDAndDate` — single insight for a specific day
- `FindByAppIDRange` — insights within a date range (for history views)
- `FindLatestByAppID` — most recent insight (for dashboard display)

## Extension Points
- **Swap LLM provider** — implement the `AIProvider` interface for a different model (Claude, Gemini, local models). The service is provider-agnostic by design.
- **Prompt tuning** — `BuildPrompt()` is a separate public method. Override or extend the prompt template without changing generation logic.
- **Additional metrics in prompt** — add more fields from `DailyMetricsSnapshot` to the prompt string (e.g., top churning stores, growth rate) for richer insights.
- **Insight scoring** — add a quality/relevance score to `DailyInsight` by running a second LLM pass or heuristic check.
- **Free tier preview** — relax the tier gate to allow, say, one insight per week for FREE users.

## Gotchas
- **No word count validation.** The 80-120 word target is a prompt instruction only. The LLM may generate shorter or longer text, and the service stores whatever it returns without truncation or rejection.
- **Date truncation matters.** `NewDailyInsight` truncates the date to midnight UTC. If you pass a `time.Time` with a non-UTC timezone, it will still truncate to midnight of that date in UTC, which may be a different calendar day than intended.
- **Upsert overwrites previous insight.** If the sync runs multiple times in a day, the last insight wins. There is no versioning or history of overwritten insights.
- **Tier check uses UserRepository, not the snapshot.** The tier is checked by loading the full user from the database, not inferred from the snapshot. If the user downgrades mid-day, re-running the sync will fail with `ErrProTierRequired`.
- **AIProvider errors are not retried.** If the LLM call fails (rate limit, timeout, network), the error propagates up and the insight is not stored. The caller is responsible for retry logic.
- **Prompt formatting uses `%.2f` for all dollar amounts.** Very large revenue numbers (e.g., $1,234,567.89) will display without comma separators, which may confuse the LLM's interpretation.
