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
