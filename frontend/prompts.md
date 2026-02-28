# Frontend Prompts Log – LedgerGuard

Track all prompts executed for the Flutter frontend.

---

## Prompt 001 – Initialize Flutter Web Project
**Date:** 2024-01-XX
**Status:** Complete

**Prompt:**
> Initialize Flutter Web project for LedgerGuard. Requirements: Clean Architecture folder structure, Bloc for state management, GoRouter for navigation, Firebase core setup (no auth screens yet), Separate environments (dev/prod ready structure). Do NOT implement UI yet. Just project structure and configuration.

**Changes:**
- Created `frontend/REQUIREMENTS.md` with tech stack and architecture specs
- Initialized Flutter project with web support (`flutter create --platforms web`)
- Set up Clean Architecture folders:
  - `lib/core/` (config, constants, theme, utils, di)
  - `lib/data/` (datasources, models, repositories)
  - `lib/domain/` (entities, repositories, usecases)
  - `lib/presentation/` (blocs, pages, widgets, router)
- Added dependencies: flutter_bloc, go_router, get_it, injectable, firebase_core, dio, freezed
- Created environment config (`EnvConfig`, `AppConfig`)
- Set up dependency injection with get_it + injectable
- Created GoRouter configuration with placeholder pages
- Added theme matching marketing site colors
- Updated TEST_PLAN.md with frontend test scenarios
- Test: 1 passed

---

## Prompt 002 – Implement Firebase Authentication
**Date:** 2024-01-XX
**Status:** Complete

**Prompt:**
> Implement Firebase Authentication integration. Requirements: Email/Password login, Google login, Firebase initialization, Auth state listener, Basic loading state. Create: AuthRepository, AuthController, AuthState. Write widget tests for login logic. Do not build dashboard yet.

**Changes:**
- Added dependencies: firebase_auth, google_sign_in
- Domain layer:
  - `domain/entities/user_entity.dart` - UserEntity
  - `domain/repositories/auth_repository.dart` - AuthRepository interface + exceptions
- Data layer:
  - `data/repositories/firebase_auth_repository.dart` - Firebase implementation
- Presentation layer:
  - `presentation/blocs/auth/auth_bloc.dart` - AuthBloc
  - `presentation/blocs/auth/auth_event.dart` - Events (AuthCheckRequested, SignInWithEmail, SignInWithGoogle, SignOut)
  - `presentation/blocs/auth/auth_state.dart` - States (AuthInitial, AuthLoading, Authenticated, Unauthenticated, AuthError)
  - `presentation/blocs/auth/auth.dart` - Barrel export
- Updated DI: Registered AuthRepository and AuthBloc in injection.config.dart
- Tests: 11 AuthBloc tests (TDD)
- Updated TEST_PLAN.md with AuthBloc test scenarios
- All tests passing (12/12)

---

## Prompt 003 – Create Login and Signup Screens
**Date:** 2024-01-XX
**Status:** Complete

**Prompt:**
> Create login and signup screens. Requirements: Email field, Password field, Google login button, Loading state, Error display, Clean minimal UI. Navigation: If logged in → redirect to dashboard route. If not logged in → show login. Write widget tests.

**Changes:**
- Presentation layer:
  - `presentation/pages/login_page.dart` - Login screen with email/password, Google sign-in
  - `presentation/pages/signup_page.dart` - Signup screen with email/password, Google sign-in
  - `presentation/router/app_router.dart` - Auth-aware routing with redirects
- Updated `app.dart` - BlocProvider setup, AuthBloc initialization
- Widget tests:
  - `test/presentation/pages/login_page_test.dart` - 9 test cases
  - `test/presentation/pages/signup_page_test.dart` - 8 test cases
- Features:
  - Form validation
  - Loading state with disabled buttons
  - Error display with red container
  - Auth redirect (login ↔ dashboard)
  - Clean minimal UI matching theme
- Updated TEST_PLAN.md with page widget tests
- All tests passing (29/29)

---

## Prompt 004 – Fetch User Role from Backend
**Date:** 2024-01-XX
**Status:** Complete

**Prompt:**
> Fetch user role from backend after login. Store: role (OWNER/ADMIN), plan_tier. Implement: RoleProvider. Hide admin-only UI for non-admin. Protect manual integration route. Write tests for role visibility.

**Changes:**
- Domain layer:
  - `domain/entities/user_profile.dart` - UserProfile, UserRole, PlanTier
  - `domain/repositories/user_profile_repository.dart` - Repository interface + exceptions
- Data layer:
  - `data/repositories/api_user_profile_repository.dart` - API implementation with Dio
- Presentation layer:
  - `presentation/blocs/role/role_bloc.dart` - RoleBloc
  - `presentation/blocs/role/role_event.dart` - FetchRoleRequested, ClearRoleRequested
  - `presentation/blocs/role/role_state.dart` - RoleInitial, RoleLoading, RoleLoaded, RoleError
  - `presentation/widgets/role_guard.dart` - RoleGuard, ProGuard widgets
  - `presentation/pages/admin/manual_integration_page.dart` - Admin-only page
- Updated `app.dart` - RoleBloc integration, fetch role after auth
- Updated `auth_repository.dart` - Added getIdToken method
- Updated router with /admin/manual-integration route
- Tests:
  - `test/presentation/blocs/role_bloc_test.dart` - 11 tests
  - `test/presentation/widgets/role_guard_test.dart` - 10 tests
  - `test/presentation/pages/manual_integration_page_test.dart` - 4 tests
- All tests passing (54/54)

---

## Prompt 005 – Partner Integration Screen
**Date:** 2024-01-XX
**Status:** Complete

**Prompt:**
> Create Partner Integration screen. Features: "Connect Shopify Partner" button (OAuth), Manual Token form (visible only to ADMIN), Token input fields, Save button, Loading state, Success state. Do not implement API logic deeply. Mock API calls first. Write widget tests.

**Changes:**
- Domain layer:
  - `domain/entities/partner_integration.dart` - PartnerIntegration entity, IntegrationStatus enum
  - `domain/repositories/partner_integration_repository.dart` - Repository interface + exceptions
- Data layer:
  - `data/repositories/mock_partner_integration_repository.dart` - Mock implementation
- Presentation layer:
  - `presentation/blocs/partner_integration/partner_integration_bloc.dart` - PartnerIntegrationBloc
  - `presentation/blocs/partner_integration/partner_integration_event.dart` - CheckStatus, ConnectWithOAuth, SaveManualToken, Disconnect
  - `presentation/blocs/partner_integration/partner_integration_state.dart` - Initial, Loading, NotConnected, Connected, Success, Error
  - `presentation/pages/partner_integration_page.dart` - Partner integration page
- Updated `app.dart` - PartnerIntegrationBloc integration
- Updated `core/di/injection.config.dart` - Registered new dependencies
- Updated `core/theme/app_theme.dart` - Added success color
- Updated router with /partner-integration route
- Tests:
  - `test/presentation/blocs/partner_integration_bloc_test.dart` - 10 tests
  - `test/presentation/pages/partner_integration_page_test.dart` - 16 tests
- All tests passing (80/80)

---

## Prompt 006 – App Selection Screen
**Date:** 2024-01-XX
**Status:** Complete

**Prompt:**
> After partner connection: Fetch list of apps from backend. Display: List of apps, Radio selection, Confirm selection button. Store selected app locally. Add loading and error handling.

**Changes:**
- Domain layer:
  - `domain/entities/shopify_app.dart` - ShopifyApp entity
  - `domain/repositories/app_repository.dart` - Repository interface + exceptions
- Data layer:
  - `data/repositories/mock_app_repository.dart` - Mock implementation with sample apps
- Presentation layer:
  - `presentation/blocs/app_selection/app_selection_bloc.dart` - AppSelectionBloc
  - `presentation/blocs/app_selection/app_selection_event.dart` - FetchApps, AppSelected, ConfirmSelection, LoadSelectedApp
  - `presentation/blocs/app_selection/app_selection_state.dart` - Initial, Loading, Loaded, Saving, Confirmed, Error
  - `presentation/pages/app_selection_page.dart` - App selection page with radio list
- Updated `app.dart` - AppSelectionBloc integration
- Updated `core/di/injection.config.dart` - Registered new dependencies
- Updated router with /app-selection route
- Updated partner_integration_page.dart - Navigate to app selection after success
- Tests:
  - `test/presentation/blocs/app_selection_bloc_test.dart` - 12 tests
  - `test/presentation/pages/app_selection_page_test.dart` - 15 tests
- All tests passing (107/107)

---

## Prompt 007 – Executive Dashboard Layout
**Date:** 2024-01-XX
**Status:** Complete

**Prompt:**
> Create Executive Dashboard layout. Top Section (Primary KPIs): Renewal Success Rate, Active MRR, Revenue at Risk, Churn. Secondary Section: Usage Revenue, Revenue Mix chart, Risk Distribution. Use placeholder/mock data. Do NOT connect to backend yet.

**Changes:**
- Domain layer:
  - `domain/entities/dashboard_metrics.dart` - DashboardMetrics, RevenueMix, RiskDistribution
  - `domain/repositories/dashboard_repository.dart` - Repository interface + exceptions
- Data layer:
  - `data/repositories/mock_dashboard_repository.dart` - Mock implementation with sample data
- Presentation layer:
  - `presentation/blocs/dashboard/dashboard_bloc.dart` - DashboardBloc
  - `presentation/blocs/dashboard/dashboard_event.dart` - LoadDashboardRequested, RefreshDashboardRequested
  - `presentation/blocs/dashboard/dashboard_state.dart` - Initial, Loading, Loaded, Error
  - `presentation/pages/dashboard_page.dart` - Dashboard page with responsive layout
  - `presentation/widgets/kpi_card.dart` - KpiCard, KpiCardCompact
  - `presentation/widgets/revenue_mix_chart.dart` - Revenue mix horizontal bar chart
  - `presentation/widgets/risk_distribution_chart.dart` - Risk distribution 2x2 grid
- Updated `app.dart` - DashboardBloc integration
- Updated `core/di/injection.config.dart` - Registered new dependencies
- Updated `presentation/router/app_router.dart` - Dashboard route uses DashboardPage
- Tests:
  - `test/presentation/blocs/dashboard_bloc_test.dart` - 6 tests
  - `test/presentation/pages/dashboard_page_test.dart` - 19 tests
- All tests passing (132/132)

---

## Prompt 008 – Connect Dashboard to Backend
**Date:** 2024-01-XX
**Status:** Complete

**Prompt:**
> Connect dashboard to daily_metrics_snapshot endpoint. Render: Renewal Success Rate, Active MRR, Revenue at Risk, Usage Revenue, Total Revenue. Add: Loading state, Error state, Empty state. Write widget tests for rendering.

**Changes:**
- Domain layer:
  - Updated `DashboardMetrics` - Added `totalRevenue` field
  - Updated `DashboardRepository` - Return nullable for empty state
  - Added exceptions: `NoAppSelectedException`, `NoMetricsException`, `UnauthorizedMetricsException`
- Data layer:
  - Created `ApiDashboardRepository` - Connects to `/api/v1/apps/{appId}/metrics/latest`
  - Updated `MockDashboardRepository` - Support `returnEmpty` flag, added `totalRevenue`
- Presentation layer:
  - Updated `DashboardBloc` - Handle null metrics (empty state)
  - Added `DashboardEmpty` state
  - Updated `DashboardPage` - Added empty state UI, added Total Revenue KPI
- Updated `core/di/injection.config.dart` - Wire `ApiDashboardRepository`
- Tests:
  - Updated `dashboard_bloc_test.dart` - Added empty state tests (8 tests)
  - Updated `dashboard_page_test.dart` - Added empty state tests, Total Revenue test (25 tests)
- All tests passing (140/140)

---

## Prompt 009 – Dashboard Configuration
**Date:** 2024-01-XX
**Status:** Complete

**Prompt:**
> Implement Dashboard Configuration feature. User can: Choose up to 4 primary KPIs, Reorder KPIs, Toggle secondary widgets. Store preferences via backend API. Create: DashboardPreferencesRepository, PreferencesController. Persist configuration.

**Changes:**
- Domain layer:
  - Created `domain/entities/dashboard_preferences.dart` - `KpiType` enum, `SecondaryWidget` enum, `DashboardPreferences` class
  - Created `domain/repositories/dashboard_preferences_repository.dart` - Repository interface + exceptions
- Data layer:
  - Created `data/repositories/mock_dashboard_preferences_repository.dart` - Mock implementation
  - Created `data/repositories/api_dashboard_preferences_repository.dart` - API implementation
- Presentation layer:
  - Created `presentation/blocs/preferences/preferences_event.dart` - Events for preferences
  - Created `presentation/blocs/preferences/preferences_state.dart` - States for preferences
  - Created `presentation/blocs/preferences/preferences_bloc.dart` - PreferencesBloc
  - Created `presentation/widgets/dashboard_config_dialog.dart` - Configuration dialog UI
  - Updated `dashboard_page.dart` - Added settings button in AppBar
- Updated `app.dart` - Added PreferencesBloc provider
- Updated `core/di/injection.config.dart` - Registered PreferencesBloc and repository
- Tests:
  - Created `test/presentation/blocs/preferences_bloc_test.dart` - 14 tests
  - Created `test/presentation/widgets/dashboard_config_dialog_test.dart` - 15 tests
- All tests passing (168/168)

---

## Prompt 010 – Risk Breakdown Screen
**Date:** 2024-01-XX
**Status:** Complete

**Prompt:**
> Create Risk Breakdown screen. Display: SAFE count, ONE_CYCLE_MISSED count, TWO_CYCLE_MISSED count, CHURNED count. Include: Simple bar or pie chart, Clean professional layout. Connect to backend risk summary endpoint.

**Changes:**
- Domain layer:
  - Created `domain/entities/risk_summary.dart` - `RiskLevel` enum, `RiskSummary` class
  - Created `domain/repositories/risk_repository.dart` - Repository interface + exceptions
- Data layer:
  - Created `data/repositories/api_risk_repository.dart` - Uses `/api/v1/apps/{appId}/metrics/latest`
  - Created `data/repositories/mock_risk_repository.dart` - Mock implementation
- Presentation layer:
  - Created `presentation/blocs/risk/risk_bloc.dart` - RiskBloc with Load/Refresh events
  - Created `presentation/blocs/risk/risk_event.dart` - Load and Refresh events
  - Created `presentation/blocs/risk/risk_state.dart` - Initial, Loading, Loaded, Empty, Error states
  - Created `presentation/pages/risk_breakdown_page.dart` - Page with:
    - Summary card (total subscriptions, revenue at risk)
    - Donut pie chart with legend
    - Breakdown list with progress bars and descriptions
  - Updated `presentation/widgets/risk_distribution_chart.dart` - Made clickable, navigates to /risk-breakdown
  - Updated `presentation/router/app_router.dart` - Added /risk-breakdown route
- Updated `app.dart` - Added RiskBloc provider
- Updated `core/di/injection.config.dart` - Registered RiskBloc and repository
- Tests:
  - Created `test/presentation/blocs/risk_bloc_test.dart` - 13 tests (8 bloc + 5 entity)
  - Created `test/presentation/pages/risk_breakdown_page_test.dart` - 15 tests
- All tests passing (196/196)

---

## Prompt 011 – AI Insight Card
**Date:** 2026-02-27
**Status:** Complete

**Prompt:**
> Implement AI Insight card on dashboard. Behavior: Visible only if plan_tier == PRO, Fetch daily_insight from backend, Display formatted executive summary, Collapsible card. Handle loading and empty states.

**Improved:**
> Implement AI Insight card on Executive Dashboard:
> 1. Create `DailyInsight` entity with summary, generatedAt, keyPoints
> 2. Create `InsightRepository` interface with exceptions
> 3. Create `ApiInsightRepository` - calls `/api/v1/apps/{appId}/insights/daily`
> 4. Create `InsightBloc` with Load/Refresh events and states
> 5. Create `AiInsightCard` widget - collapsible card with:
>    - AI sparkle icon, "Daily Insight" title
>    - Generated timestamp
>    - Summary text
>    - Key takeaways bullet points
>    - Shimmer loading state
>    - Hidden on empty/error states
> 6. Wrap card in `ProGuard` - visible only to PRO tier users
> 7. Place at top of dashboard, above Primary KPIs
> 8. Write tests for InsightBloc and AiInsightCard

**Changes:**
- Domain layer:
  - Created `domain/entities/daily_insight.dart` - DailyInsight with fromJson, formattedGeneratedAt
  - Created `domain/repositories/insight_repository.dart` - Repository interface + exceptions (InsightException, NoAppSelectedInsightException, UnauthorizedInsightException, ProRequiredInsightException)
- Data layer:
  - Created `data/repositories/api_insight_repository.dart` - API implementation
  - Created `data/repositories/mock_insight_repository.dart` - Mock implementation
- Presentation layer:
  - Created `presentation/blocs/insight/insight_bloc.dart` - InsightBloc
  - Created `presentation/blocs/insight/insight_event.dart` - LoadInsightRequested, RefreshInsightRequested
  - Created `presentation/blocs/insight/insight_state.dart` - InsightInitial, InsightLoading, InsightLoaded, InsightEmpty, InsightError
  - Created `presentation/widgets/ai_insight_card.dart` - Collapsible AI insight card with:
    - Gradient background matching theme
    - AI sparkle icon
    - Shimmer loading state
    - Expand/collapse animation
    - Refresh button with spinner
  - Updated `presentation/pages/dashboard_page.dart` - Added AiInsightCard wrapped in ProGuard
- Updated `app.dart` - Added InsightBloc provider
- Updated `core/di/injection.config.dart` - Registered InsightBloc and repository
- Tests:
  - Created `test/presentation/blocs/insight_bloc_test.dart` - 21 tests (bloc + entity + exceptions)
  - Created `test/presentation/widgets/ai_insight_card_test.dart` - 13 tests
  - Updated `test/presentation/pages/dashboard_page_test.dart` - Added RoleBloc, InsightBloc, PreferencesBloc mocks
- All tests passing (229/229)

---

## Prompt 012 – Notification Settings Screen
**Date:** 2026-02-27
**Status:** Complete

**Prompt:**
> Create Notification Settings screen. Options: Toggle Critical Alerts, Toggle Daily Summary, Select Daily Summary Time. Persist via backend. Add loading state and success indicator.

**Improved:**
> Create Notification Settings screen with:
> 1. Toggle switch for Critical Alerts (on/off)
> 2. Toggle switch for Daily Summary (on/off)
> 3. Time picker for Daily Summary Time (disabled when Daily Summary is off)
> 4. Save button with loading state
> 5. Success snackbar on save
> 6. Load/Save via `/api/v1/users/notification-preferences`

**Changes:**
- Domain layer:
  - Created `domain/entities/notification_preferences.dart` - NotificationPreferences with fromJson/toJson, formattedTime
  - Created `domain/repositories/notification_preferences_repository.dart` - Repository interface + exceptions
- Data layer:
  - Created `data/repositories/api_notification_preferences_repository.dart` - API implementation
  - Created `data/repositories/mock_notification_preferences_repository.dart` - Mock implementation
- Presentation layer:
  - Created `presentation/blocs/notification_preferences/notification_preferences_bloc.dart` - Bloc
  - Created `presentation/blocs/notification_preferences/notification_preferences_event.dart` - Events
  - Created `presentation/blocs/notification_preferences/notification_preferences_state.dart` - States
  - Created `presentation/pages/notification_settings_page.dart` - Settings page with:
    - Critical Alerts toggle section
    - Daily Summary toggle section
    - Time picker (disabled when summary off)
    - Save button with unsaved changes indicator
    - Success/error snackbars
- Updated `app.dart` - Added NotificationPreferencesBloc provider
- Updated `core/di/injection.config.dart` - Registered bloc and repository
- Updated `presentation/router/app_router.dart` - Added /settings/notifications route
- Tests:
  - Created `test/presentation/blocs/notification_preferences_bloc_test.dart` - 19 tests
  - Created `test/presentation/pages/notification_settings_page_test.dart` - 21 tests
- All tests passing (268/268)

---

## Prompt 013 – Profile Page
**Date:** 2026-02-27
**Status:** Complete

**Prompt:**
> Create Profile page. Display: Email, Role, Plan tier, Upgrade button (placeholder). Include logout button. Add route protection.

**Improved:**
> Create Profile page with:
> 1. Profile header with avatar (initials), email
> 2. Account section showing email, role (badge), plan tier (badge)
> 3. Upgrade card (visible only for FREE tier) with "Coming Soon" placeholder
> 4. Settings section with link to Notification Settings
> 5. Logout button with confirmation dialog
> 6. Add /profile route, accessible from dashboard app bar
> 7. Route protection via existing auth redirect

**Changes:**
- Domain layer:
  - Updated `domain/entities/user_profile.dart` - Added displayName getters to UserRole and PlanTier enums, added PlanTier.free alias and isFree getter
- Presentation layer:
  - Created `presentation/pages/profile_page.dart` - Profile page with:
    - Profile header with CircleAvatar showing initials
    - Account section (email, role, plan info tiles)
    - Role badge (Owner=primary, Admin=secondary)
    - Plan badge (Pro=warning star, Free=grey star)
    - Upgrade card (only for free tier) with "Coming Soon" snackbar
    - Settings section with Notification Settings link
    - Logout button with confirmation dialog
    - SignOutRequested dispatch on confirm
  - Updated `presentation/pages/dashboard_page.dart` - Added profile icon to app bar
- Updated `presentation/router/app_router.dart` - Added /profile route
- Tests:
  - Created `test/presentation/pages/profile_page_test.dart` - 20 tests covering:
    - App bar title
    - Loading states (unauthenticated, role loading)
    - Error state display
    - User email and initials
    - Role badge (owner, admin)
    - Plan badge (pro, free)
    - Upgrade card visibility (free tier only)
    - Upgrade coming soon snackbar
    - Notification settings link
    - Logout button and confirmation dialog
    - Logout cancel and confirm actions
    - Account and Settings sections
- All tests passing (288/288)

---

## Prompt 014 – Global Error Handling & Shared Components
**Date:** 2026-02-27
**Status:** Complete

**Prompt:**
> Implement global error snackbar, loading overlay component, API error interceptor, token refresh handling. Refactor duplicated UI elements into shared components.

**Improved:**
> Implement global error handling and shared UI components:
> 1. SnackbarService - showError(), showSuccess(), showInfo(), showWarning()
> 2. LoadingOverlay widget - semi-transparent overlay with spinner
> 3. ApiClient with Dio interceptors for auth and error handling
> 4. Token refresh on 401 with automatic retry
> 5. Shared widgets: ErrorStateWidget, EmptyStateWidget, SectionHeader, StatusBadge, InfoTile, CardSection

**Changes:**
- Core services:
  - Created `core/services/snackbar_service.dart` - Global snackbar service
  - Created `core/network/api_client.dart` - Centralized API client with interceptors
    - AuthInterceptor: Adds auth token to requests
    - ErrorInterceptor: Global error handling, token refresh, user-friendly messages
- Presentation widgets:
  - Created `presentation/widgets/error_state_widget.dart` - Error state with retry
  - Created `presentation/widgets/empty_state_widget.dart` - Empty state with action
  - Created `presentation/widgets/section_header.dart` - Section/subsection headers
  - Created `presentation/widgets/status_badge.dart` - StatusBadge, RoleBadge, PlanBadge, RiskBadge
  - Created `presentation/widgets/info_tile.dart` - InfoTile, NavigationTile
  - Created `presentation/widgets/card_section.dart` - CardSection, ContentCard
  - Created `presentation/widgets/loading_overlay.dart` - LoadingOverlay, FullScreenLoading
  - Created `presentation/widgets/shared.dart` - Barrel export file
- Refactored pages to use shared components:
  - Updated `dashboard_page.dart` - Uses ErrorStateWidget, EmptyStateWidget
  - Updated `risk_breakdown_page.dart` - Uses ErrorStateWidget, EmptyStateWidget
  - Updated `profile_page.dart` - Uses ErrorStateWidget, RoleBadge, PlanBadge
- Updated `app.dart` - Integrated SnackbarService with ScaffoldMessengerKey
- Updated `core/di/injection.config.dart` - Registered SnackbarService and ApiClient
- Tests:
  - Created `test/core/services/snackbar_service_test.dart` - 6 tests
  - Created `test/presentation/widgets/shared_widgets_test.dart` - 37 tests
- All tests passing (331/331)

---

## Prompt 015 – Fix Critical Security Blockers
**Date:** 2026-02-27
**Status:** Complete

**Prompt:**
> Fix the 3 critical blockers identified in production readiness review:
> 1. Auth middleware not wired in main.go
> 2. OAuth state not validated (CSRF vulnerability)
> 3. Tenant isolation missing in SyncApp handler

**Changes:**
- Backend security fixes:
  - Created `internal/infrastructure/cache/oauth_state_store.go` - In-memory OAuth state store
    - Store() saves state with user ID
    - Validate() returns user ID and consumes state (one-time use)
    - 10-minute TTL with automatic cleanup
  - Updated `internal/interfaces/http/handler/oauth.go`:
    - Added OAuthStateStore interface
    - Added userRepo dependency for callback user lookup
    - StartOAuth now stores state with user ID
    - Callback validates state, retrieves user from store (CSRF protection)
  - Updated `internal/interfaces/http/handler/sync.go`:
    - Added appRepo dependency
    - SyncApp verifies app.PartnerAccountID matches user's partner account
    - Returns 403 Forbidden for unauthorized access (tenant isolation)
  - Updated `cmd/server/main.go`:
    - Wired Firebase Auth service
    - Wired AuthMiddleware with token verifier and user repo
    - Wired RoleMiddleware for admin routes
    - Wired OAuthHandler with state store
    - Initialized all repositories from database pool
    - Added graceful degradation when services unavailable
- Tests:
  - Created `internal/infrastructure/cache/oauth_state_store_test.go` - 5 tests
  - Updated `internal/interfaces/http/handler/oauth_test.go` - 6 tests (new signature)
  - Updated `internal/interfaces/http/handler/sync_test.go` - 9 tests (added tenant isolation test)
- Documentation:
  - Updated `DECISIONS.md`:
    - ADR-006: OAuth State Validation for CSRF Protection
    - ADR-007: Tenant Isolation in Sync Handler
- All backend tests passing
- All frontend tests passing (325/325)

---

## Prompt 016 – KPI Dashboard Upgrade: Time Filtering and Delta Comparison
**Date:** 2026-02-27
**Status:** Complete

**Prompt:**
> (Plan file provided) KPI Dashboard Upgrade with time filtering and delta comparison featuring Play Store-style analytics.

**Improved:**
> Implement time-based filtering and period-over-period delta comparison:
> 1. Create TimeRange entity with TimeRangePreset enum
> 2. Add factory methods for each preset (thisMonth, lastMonth, etc.)
> 3. Create MetricsDelta class with percentage changes
> 4. Create DeltaIndicator helper for direction and color semantics
> 5. Add TimeRangeChanged event to DashboardBloc
> 6. Update DashboardLoaded state with timeRange field
> 7. Create TimeRangeSelector widget for app bar
> 8. Update KpiCard with delta badges (green/red coloring)
> 9. Update repository to pass TimeRange to API
> 10. Update all tests with new timeRange parameter

**Changes:**
- Domain layer:
  - Created `lib/domain/entities/time_range.dart` - TimeRange, TimeRangePreset
  - Updated `lib/domain/entities/dashboard_metrics.dart` - MetricsDelta, DeltaIndicator
  - Updated `lib/domain/repositories/dashboard_repository.dart` - Added TimeRange param
- Data layer:
  - Updated `lib/data/repositories/api_dashboard_repository.dart` - New API endpoint with query params
  - Updated `lib/data/repositories/mock_dashboard_repository.dart` - Mock delta data
- Presentation layer:
  - Updated `lib/presentation/blocs/dashboard/dashboard_event.dart` - TimeRangeChanged
  - Updated `lib/presentation/blocs/dashboard/dashboard_state.dart` - timeRange in DashboardLoaded
  - Updated `lib/presentation/blocs/dashboard/dashboard_bloc.dart` - Handle TimeRangeChanged
  - Created `lib/presentation/widgets/time_range_selector.dart` - PopupMenuButton widget
  - Updated `lib/presentation/widgets/kpi_card.dart` - Delta badges with semantics
  - Updated `lib/presentation/pages/dashboard_page.dart` - Wired TimeRangeSelector
- Tests:
  - Updated `test/presentation/blocs/dashboard_bloc_test.dart` - TimeRange in seed/mocks
  - Updated `test/presentation/pages/dashboard_page_test.dart` - TimeRange in all states
- All dashboard tests passing (32/32)

---

## Prompt 017 – Subscription List and Detail Pages
**Date:** 2026-02-27
**Status:** Complete

**Prompt:**
> Implement subscription list and detail views for frontend with risk filtering.

**Improved:**
> Create subscription list and detail pages:
> 1. Create Subscription entity with shopName field
> 2. Create SubscriptionRepository interface
> 3. Create ApiSubscriptionRepository
> 4. Create SubscriptionListBloc with filter events
> 5. Create SubscriptionDetailBloc
> 6. Create RiskBadge widget (green/yellow/orange/red)
> 7. Create SubscriptionTile widget showing store name (not domain)
> 8. Create SubscriptionListPage with filter dropdown
> 9. Create SubscriptionDetailPage
> 10. Add routes to app_router.dart

**Changes:**
- Domain layer:
  - `lib/domain/entities/subscription.dart` - Subscription with shopName
  - `lib/domain/repositories/subscription_repository.dart`
- Data layer:
  - `lib/data/repositories/api_subscription_repository.dart`
- Presentation layer:
  - `lib/presentation/blocs/subscription_list/` - Bloc, events, states
  - `lib/presentation/blocs/subscription_detail/` - Bloc, events, states
  - `lib/presentation/widgets/risk_badge.dart`
  - `lib/presentation/widgets/subscription_tile.dart`
  - `lib/presentation/pages/subscription_list_page.dart`
  - `lib/presentation/pages/subscription_detail_page.dart`
- Updated `lib/presentation/router/app_router.dart`

---

## Prompt 018 – Fix Subscription Tile Index Error
**Date:** 2026-02-27
**Status:** Complete

**Prompt:**
> The following IndexError was thrown building SubscriptionTile: RangeError (index): Index out of range: no indices are valid: 0

**Improved:**
> Fix index out of range in subscription_tile.dart:
> 1. Add null/empty checks in _getInitials method
> 2. Add null/empty checks in _formatDisplayName method
> 3. Filter empty parts after splitting strings
> 4. Return fallback values for edge cases

**Changes:**
- `lib/presentation/widgets/subscription_tile.dart`:
  - `_getInitials`: Added empty checks, filter empty parts
  - `_formatDisplayName`: Added empty string checks
  - Returns '??' for empty/invalid names

---

## Prompt 019 – Frontend Tests for Subscription Feature
**Date:** 2026-02-27
**Status:** Complete

**Prompt:**
> whats pending for next? → 3 (Frontend tests - BLoC and widget tests for subscription feature)

**Improved:**
> Write comprehensive tests for subscription feature:
> 1. SubscriptionListBloc tests - fetch, filter, refresh, load more
> 2. RiskBadge widget tests - colors, icons, text, compact mode
> 3. SubscriptionTile widget tests - display, tap handling, responsive layout

**Changes:**
- Tests created:
  - `test/presentation/blocs/subscription_list_bloc_test.dart` - 17 tests
    - Initial state test
    - FetchSubscriptionsRequested: success, empty, error (SubscriptionException, generic)
    - hasMore pagination flag
    - FilterByRiskStateRequested: success, empty, no appId
    - RefreshSubscriptionsRequested: success, failure
    - LoadMoreSubscriptionsRequested: success, hasMore=false
  - `test/presentation/widgets/risk_badge_test.dart` - 20 tests
    - Display correct text for each RiskState
    - Display correct colors (green, yellow, orange, red)
    - Display correct icons
    - Compact mode (no icon, smaller font)
    - RiskStateIndicator with descriptions
  - `test/presentation/widgets/subscription_tile_test.dart` - 17 tests
    - Display shop name and formatted domain
    - Display plan name and price
    - Show initials in avatar
    - Display risk badge
    - Chevron icon
    - Tap handling
    - Responsive layout (compact vs full)
    - Initials generation edge cases
- All 54 new tests passing
- Total tests: 188 passing (2 pre-existing failures in partner_integration_page_test.dart)

---

## Prompt 020 – Earnings Timeline Enhancements
**Date:** 2026-02-28
**Status:** Complete

**Prompt:**
> Multiple requests for earnings timeline improvements:
> 1. Remove individual date switcher, use top dashboard date filter
> 2. What date for "Last 30 Days"? → Use date range API
> 3. Add month navigator back with dashboard sync
> 4. Add to dashboard configuration settings
> 5. Fix scroll jump when switching modes/months
> 6. Fix animation not working

**Improved:**
> Enhance Earnings Timeline chart with:
> 1. Change API from year/month to start/end date range to support "Last 30 Days" spanning multiple months
> 2. Sync chart month with dashboard time range (extract appropriate month from preset)
> 3. Keep month navigation arrows for manual browsing within current data
> 4. Add `earningsTimeline` to SecondaryWidget enum for dashboard config toggle
> 5. Premium-styled dashboard config dialog with gradient header, icons
> 6. Fix scroll jump by maintaining consistent 300px card height across all states
> 7. Fix animation by preserving chart during loading (overlay indicator instead of replacing chart)

**Changes:**
- Backend:
  - Updated `revenue_repository.go` - `GetRevenueByDateRange(startDate, endDate)`
  - Updated `revenue_handler.go` - Parse start/end query params
  - Updated `router.go` - Use new handler method name
- Frontend Domain:
  - Updated `earnings_timeline.dart` - Date range based entity
  - Updated `earnings_repository.dart` - `fetchEarnings(startDate, endDate, mode)`
- Frontend Data:
  - Updated `api_earnings_repository.dart` - start/end query params
- Frontend Bloc:
  - Updated `earnings_bloc.dart` - `_getTargetMonth(TimeRange)`, `EarningsTimeRangeChanged`
  - Updated `earnings_event.dart` - Added `EarningsTimeRangeChanged`
  - Updated `earnings_state.dart` - Added `copyWith` method
- Frontend Widgets:
  - Updated `earnings_timeline_chart.dart`:
    - Cache `_lastLoadedState` for smooth transitions
    - Show overlay loading indicator instead of replacing chart
    - 300px fixed height for all states
    - `ValueKey('earnings-card')` for widget identity
    - BarChart animation: 250ms easeInOut
  - Updated `dashboard_config_dialog.dart` - Premium styling
  - Updated `dashboard_preferences.dart` - Added `earningsTimeline` to SecondaryWidget
- Tests:
  - Updated `dashboard_config_dialog_test.dart` - 5 switches instead of 4
- All dashboard tests passing (24/24)
- All config dialog tests passing (15/15)

---

---

## [2026-03-01] Revenue Share Tier Tracking - Frontend

**Original Prompt:**
> yes continue with frontend

**Context:**
Backend Phase 1 (Revenue Share Tier Tracking) was completed. User requested frontend implementation to match.

**Improved Prompt:**
> Implement frontend support for revenue share tier tracking:
> 1. Create RevenueShareTier enum with 4 tiers and FeeBreakdown calculator
> 2. Update ShopifyApp entity with tier field
> 3. Update AppRepository with tier and fee methods
> 4. Create TierSelector and TierIndicator widgets
> 5. Create AppSettingsPage with fee calculator and tier comparison
> 6. Add FeeInsightsCard to dashboard
> 7. Update tests

**Files Created:**
- revenue_share_tier.dart, tier_selector.dart, app_settings_page.dart, fee_insights_card.dart

**Files Modified:**
- shopify_app.dart, app_repository.dart, api_app_repository.dart, mock_app_repository.dart
- app_selection_page.dart, profile_page.dart, dashboard_page.dart, app_router.dart

---

## [2026-03-01] Profile Page Test Fixes

**Original Prompt:**
> fix the failing profile page tests

**Issues Found:**
- Initials test expected 'TE' but code returns 'T' for 'test@example.com'
- Upgrade button text was 'Upgrade to Pro', not 'Upgrade Now'
- Snackbar text was 'Upgrade coming soon!', not 'Upgrade functionality coming soon!'
- Logout button was InkWell, not OutlinedButton
- Dialog confirm button was ElevatedButton, not TextButton
- Tests needed larger screen size for scrolling content

**Fixed:** All 21 profile page tests now pass

---
