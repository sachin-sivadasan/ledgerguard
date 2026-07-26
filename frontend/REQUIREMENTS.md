# Frontend Requirements – LedgerGuard

## Overview
Flutter Web application for LedgerGuard Revenue Intelligence Platform.

## Tech Stack
- **Framework:** Flutter 3.x (Web)
- **State Management:** Bloc
- **Navigation:** GoRouter
- **Dependency Injection:** get_it
- **Authentication:** Firebase Auth
- **Architecture:** Clean Architecture + TDD

> **Note:** A Provider-based prototype exists at `frontend-flutter/` (same repo root). It uses ChangeNotifier + Provider instead of Bloc and serves as a rapid-prototyping environment. The Bloc version at `frontend/app/` is the primary app. A Provider-to-Bloc migration is planned — see MEMORY.md for details.

---

## Project Structure

```
lib/
├── core/
│   ├── config/           → Environment configs (dev/prod)
│   ├── constants/        → App constants
│   ├── theme/            → App theme
│   └── utils/            → Utilities
├── data/
│   ├── datasources/      → API clients, local storage
│   ├── models/           → JSON serializable models
│   └── repositories/     → Repository implementations
├── domain/
│   ├── entities/         → Business entities
│   ├── repositories/     → Repository interfaces
│   └── usecases/         → Business logic
└── presentation/
    ├── blocs/            → Bloc state management
    ├── pages/            → Screen widgets
    ├── widgets/          → Reusable components
    └── router/           → GoRouter configuration
```

---

## Environments

### Development
- API: `http://localhost:8080`
- Firebase: Development project

### Production
- API: `https://api.ledgerspear.com`
- Firebase: Production project

---

## Theming & Material Design Standards

### Typography
- All text MUST use `Theme.of(context).textTheme.*` semantic styles
- Custom type scale defined in `core/theme/app_theme.dart`
- Never use hardcoded `TextStyle(fontSize:)` in the presentation layer
- Colors remain inline via `.copyWith(color: ...)`

| Semantic Style | Size | Weight | Usage |
|---|---|---|---|
| `headlineLarge` | 26 | bold | Page hero text |
| `headlineMedium` | 24 | bold | Section titles |
| `titleLarge` | 18 | bold | Card/section headers |
| `titleMedium` | 16 | w600 | Sub-headers, emphasis |
| `titleSmall` | 14 | w600 | Small headers |
| `bodyLarge` | 16 | normal | Primary body text |
| `bodyMedium` | 14 | normal | Default body text |
| `bodySmall` | 13 | normal | Secondary text |
| `labelMedium` | 12 | w500 | Labels, badges |
| `labelSmall` | 11 | w500 | Tiny labels |

### Scaffold Background
- Set via `scaffoldBackgroundColor` in theme (grey-50)
- Never set `backgroundColor` on individual `Scaffold` widgets

### AppBar
- Primary blue background (`AppTheme.primary`) with white text/icons
- `scrolledUnderElevation: 0` to prevent M3 tint on scroll
- Any `TextButton`/action in AppBar must use explicit white foreground

### Dialogs & Popups
- White background with `surfaceTintColor: Colors.transparent`
- Prevents Material 3 blue tint on overlays
- Defined in `popupMenuTheme` and `dialogTheme` in app_theme.dart

### Card Backgrounds
- Use `Colors.white` for cards on grey scaffold background
- Provides visual elevation contrast

---

## Screens (Implementation Status)

### Authentication
- [x] Login (Firebase Auth)
- [x] Signup
- [ ] Forgot Password

### Onboarding
- [x] Connect Shopify Partner (Partner Integration)
- [x] Select App (App Selection)

### Dashboard
- [x] Overview (MRR, Renewal Rate, At Risk, Revenue Mix, Risk Distribution)
- [x] Subscription List
- [x] Subscription Detail

### Admin
- [x] Manual Integration (Admin-only token entry)

### Settings
- [ ] Notification Preferences
- [ ] Account Settings

---

## Dependencies

```yaml
dependencies:
  flutter_bloc: ^8.x
  go_router: ^13.x
  get_it: ^7.x
  injectable: ^2.x
  firebase_core: ^2.x
  firebase_auth: ^4.x
  dio: ^5.x
  freezed_annotation: ^2.x
  json_annotation: ^4.x
  equatable: ^2.x

dev_dependencies:
  build_runner: ^2.x
  freezed: ^2.x
  json_serializable: ^6.x
  injectable_generator: ^2.x
  bloc_test: ^9.x
  mocktail: ^1.x
```

---

## Testing Strategy

- **Unit Tests:** Blocs, UseCases, Repositories (TDD)
- **Widget Tests:** UI components
- **Integration Tests:** Full flows (future)

### Test Coverage
- [x] AuthBloc tests
- [x] RoleBloc tests
- [x] PartnerIntegrationBloc tests
- [x] AppSelectionBloc tests
- [x] DashboardBloc tests
- [x] PreferencesBloc tests
- [x] RiskBloc tests
- [x] InsightBloc tests
- [x] NotificationPreferencesBloc tests
- [x] SubscriptionListBloc tests
- [x] RiskBadge widget tests
- [x] SubscriptionTile widget tests

---

## Notes

- Web-first with mobile-responsive layouts
- Responsive breakpoints: 400px (compact), 600px (mobile), 800px (tablet)
- Firebase Auth integrated
- AppBar overflow menu on small screens
- KPI cards scale down for narrow widths

---

## Chat Page (`/chat`)

AI-powered Revenue Intelligence Assistant accessible via dashboard menu.

### Features
- Natural language queries about subscriptions, revenue, risk, and metrics
- SSE streaming for real-time tool call progress and responses
- Responsive split-pane layout: ChatPane (left) + DataPanel (right) on wide screens
- Suggestion chips for follow-up questions
- Active tool indicator showing which query is running
- Welcome screen with starter suggestions

### Architecture
- `ChatBloc` — Events: SendMessage, SuggestionTapped, ClearChat
- `ApiChatRepository` — SSE client parsing `data:` events from POST `/api/v1/chat`
- `ChatMessage` entity with user/assistant/loading states and data panel fields
- `MessageBubble` — User (blue, right) / Assistant (surface, left) with suggestion chips
- `DataPanel` — Displays structured risk/metrics/subscription/earnings/store_health data

### Tests
- [x] ChatBloc: simple response, tool call streaming, error handling, clear, data state

---

## Sync Progress & Job Management

### Auto-Sync on App Selection
- After selecting an app during onboarding, a `full_sync` is auto-triggered
- Show progress indicator: "Syncing your app data..."
- Poll `GET /api/v1/sync/jobs/{jobID}/progress` every 3-5 seconds
- On completion, refresh dashboard data

### Sync Status Indicator (Dashboard)
- Show active sync badge when any job is `processing` for the selected app
- Display job type being processed (e.g., "Syncing transactions...")
- Progress bar: `completed_items / total_items`

### Sync Jobs History (Future)
- List recent sync jobs for selected app
- Columns: job_type, status, started_at, completed_at, error_message
- Filter by status
- Cancel button for pending/processing jobs

### API Endpoints Used
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/sync/enqueue/{appID}?type=full_sync` | Trigger sync |
| GET | `/api/v1/sync/jobs/{jobID}/progress` | Poll progress |
| GET | `/api/v1/sync/jobs?app_id={appID}` | Job history |
| POST | `/api/v1/sync/jobs/{jobID}/cancel` | Cancel job |
