# 27. Flutter AI Insights & Chat

## What It Does
Displays AI-generated daily revenue insights and provides a chat interface for natural language queries about revenue data. The insights screen shows daily briefs with trend indicators and actionable recommendations. In production, this connects to the backend AI Chat via WebSocket.

## Architecture
Presentation layer. `InsightsProvider` manages the list of AI-generated insights from mock data. The `InsightsScreen` renders insight cards with severity indicators and action suggestions. The production Flutter app (`frontend/app/`) has a full `ChatBloc` with SSE streaming.

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `ledgerguard-flutter/lib/providers/insights_provider.dart` | ~40 | AI insights data provider |
| `ledgerguard-flutter/lib/screens/insights/insights_screen.dart` | ~200 | Insights list with cards |
| `ledgerguard-flutter/lib/models/insight_model.dart` | ~30 | Insight data model |
| `ledgerguard-flutter/lib/mock_data/mock_insights.dart` | ~60 | Mock AI-generated insights |

## Data Flow
```
MockInsights → InsightsProvider → InsightsScreen
                                      │
                                 insight cards
                                      │
                                      ▼
                                 InsightDetailView
                                   ├── Summary text
                                   ├── Severity badge
                                   ├── Trend indicators
                                   └── Suggested actions
```

## Configuration
None — mock data. Production requires OpenAI API key configured on backend.

## Widget Tree
```
InsightsScreen
├── LgPage (title: "AI Insights")
│   ├── Date selector
│   └── ListView.builder
│       └── LgCard per insight
│           ├── Icon (severity-based color)
│           ├── Insight title
│           ├── Summary text (80-120 words)
│           ├── Severity badge (info/warning/critical)
│           └── Action button (view details)
```

## State Machine
```
InsightsProvider (ChangeNotifier)
  State:
    _selectedDate: DateTime?  → null (latest)
    _selectedAppId: String?   → null (all apps)

  Events:
    setDate(date)         → filter insights by date
    setSelectedApp(appId) → filter by app

  Computed:
    insights → filtered list of AI insights
    latestInsight → most recent insight
```

## Gotchas
- Prototype only shows insights, not the full chat interface
- The production app (`frontend/app/`) has `ChatBloc` with SSE streaming for real AI chat
- Insights are Pro tier only — backend gate checks `plan_tier`
- Mock insights have static text; production insights are generated daily by GPT-4o-mini
- Insight severity: `info` (positive), `warning` (needs attention), `critical` (action required)
