# Implementation Log – LedgerGuard Frontend

A chronological record of all frontend features implemented.

---

## [2026-02-27] Initialize Flutter Web Project

**Commit:** Initialize Flutter Web project with Clean Architecture

**Summary:**
Initialized Flutter Web project for LedgerGuard frontend with Clean Architecture folder structure, Bloc state management, and environment configuration.

**Implemented:**

1. **Project Structure (Clean Architecture):**
   - `lib/core/` - Config, constants, theme, utils, DI
   - `lib/data/` - Datasources, models, repository implementations
   - `lib/domain/` - Entities, repository interfaces, usecases
   - `lib/presentation/` - Blocs, pages, widgets, router

2. **Dependencies:**
   - flutter_bloc for state management
   - go_router for navigation
   - get_it + injectable for dependency injection
   - firebase_core for Firebase initialization
   - dio for HTTP networking
   - freezed for immutable data classes

3. **Environment Configuration:**
   - `EnvConfig` with dev/prod environments
   - `AppConfig` singleton for runtime configuration
   - Separate entry points: `main_dev.dart`, `main_prod.dart`

4. **Theme:**
   - Colors matching marketing site (primary blue, accent green, etc.)
   - Material 3 design system

5. **Router:**
   - GoRouter with placeholder pages
   - Routes: home, login, signup, dashboard, onboarding, settings

**Files Created:**
- `frontend/REQUIREMENTS.md`
- `frontend/prompts.md`
- `frontend/app/` - Full Flutter project
- 11 Dart source files

**Tests:** 1 passing

---

## [2026-02-27] Firebase Authentication

**Commit:** Implement Firebase Authentication with Bloc (TDD)

**Summary:**
Implemented Firebase Authentication integration using Clean Architecture and TDD approach.

**Implemented:**

1. **Domain Layer:**
   - `UserEntity` - User domain model with id, email, displayName, photoUrl
   - `AuthRepository` interface - Contract for auth operations
   - Exception classes: `InvalidCredentialsException`, `UserNotFoundException`, `WeakPasswordException`, `EmailAlreadyInUseException`, `GoogleSignInCancelledException`

2. **Data Layer:**
   - `FirebaseAuthRepository` - Firebase Auth + Google Sign-In implementation
   - Maps Firebase `User` to `UserEntity`
   - Maps `FirebaseAuthException` to domain exceptions

3. **Presentation Layer (Bloc):**
   - **AuthBloc** - Manages authentication state
   - **Events:**
     - `AuthCheckRequested` - Check current auth state
     - `SignInWithEmailRequested` - Email/password login
     - `SignInWithGoogleRequested` - Google OAuth login
     - `SignOutRequested` - Sign out
   - **States:**
     - `AuthInitial` - Before auth check
     - `AuthLoading` - During operations
     - `Authenticated(user)` - User logged in
     - `Unauthenticated` - No user
     - `AuthError(message)` - Error occurred

4. **Dependency Injection:**
   - Registered `AuthRepository` as lazy singleton
   - Registered `AuthBloc` as factory

**Tests (TDD):**
- Initial state test
- AuthCheckRequested (logged in / not logged in)
- SignInWithEmail (success / invalid credentials / user not found)
- SignInWithGoogle (success / cancelled / failure)
- SignOut (success / failure)

**Files Created:**
- `lib/domain/entities/user_entity.dart`
- `lib/domain/repositories/auth_repository.dart`
- `lib/data/repositories/firebase_auth_repository.dart`
- `lib/presentation/blocs/auth/` - Bloc, events, states, barrel
- `test/presentation/blocs/auth_bloc_test.dart`

**Tests:** 12 passing (11 AuthBloc + 1 widget)

---

## [2026-02-27] Login and Signup Screens

**Commit:** Create login and signup screens with auth navigation

**Summary:**
Created login and signup pages with clean minimal UI, form validation, loading states, and auth-aware routing.

**Implemented:**

1. **LoginPage:**
   - Email and password form fields with validation
   - Sign In button dispatches `SignInWithEmailRequested`
   - Google Sign In button dispatches `SignInWithGoogleRequested`
   - Loading state shows CircularProgressIndicator, disables buttons
   - Error state shows message in red container
   - Link to signup page

2. **SignupPage:**
   - Email and password form fields with validation
   - Create Account button (placeholder for email signup)
   - Google Sign In button dispatches `SignInWithGoogleRequested`
   - Loading state shows CircularProgressIndicator, disables buttons
   - Error state shows message in red container
   - Link to login page

3. **Auth-Aware Router:**
   - `GoRouterRefreshStream` listens to AuthBloc state changes
   - Redirect to `/login` if not authenticated
   - Redirect to `/dashboard` if authenticated on auth routes
   - `AppRouter` now requires `AuthBloc` instance

4. **App Integration:**
   - `LedgerGuardApp` creates and provides `AuthBloc`
   - Triggers `AuthCheckRequested` on startup
   - Provides BlocProvider to widget tree

**Widget Tests:**
- LoginPage: 9 test cases (fields, buttons, loading, error, events)
- SignupPage: 8 test cases (fields, buttons, loading, error, events)

**Files Created/Modified:**
- `lib/presentation/pages/login_page.dart`
- `lib/presentation/pages/signup_page.dart`
- `lib/presentation/router/app_router.dart` (updated)
- `lib/app.dart` (updated)
- `test/presentation/pages/login_page_test.dart`
- `test/presentation/pages/signup_page_test.dart`
- `test/widget_test.dart` (updated)

**Tests:** 29 passing (11 AuthBloc + 9 LoginPage + 8 SignupPage + 1 widget)

---

## [2026-02-27] Role-Based Access Control

**Commit:** Implement user role fetching and role-based UI guards

**Summary:**
Implemented user role fetching from backend after login, with role-based UI visibility guards and protected admin routes.

**Implemented:**

1. **Domain Layer:**
   - `UserProfile` - Domain entity with id, email, role, planTier
   - `UserRole` enum - `owner`, `admin` with permission hierarchy
   - `PlanTier` enum - `starter`, `pro` for plan-based features
   - `UserProfileRepository` interface - Contract for profile operations
   - Exception classes: `ProfileNotFoundException`, `ProfileFetchException`

2. **Data Layer:**
   - `ApiUserProfileRepository` - Fetches profile from `/api/v1/me`
   - Uses Dio with Bearer token authentication
   - Caches profile for quick access

3. **Presentation Layer (Bloc):**
   - **RoleBloc** - Manages user role state
   - **Events:**
     - `FetchRoleRequested(authToken)` - Fetch profile from backend
     - `ClearRoleRequested` - Clear cached role (on logout)
   - **States:**
     - `RoleInitial` - Before role fetch
     - `RoleLoading` - During fetch
     - `RoleLoaded(profile)` - Profile loaded with role checks
     - `RoleError(message)` - Error occurred

4. **Role Guard Widgets:**
   - `RoleGuard` - Shows child only for users with required role
     - `RoleGuard.ownerOnly()` - Owner-only content
     - `RoleGuard.adminOnly()` - Admin+ content (owner or admin)
   - `ProGuard` - Shows child only for Pro tier users

5. **Admin Pages:**
   - `ManualIntegrationPage` - Admin-only page with Partner API token form
   - Shows "Access Denied" for non-admin users
   - Route: `/admin/manual-integration`

6. **Auth Integration:**
   - `AuthRepository.getIdToken()` - Get Firebase ID token for API calls
   - `app.dart` - Fetches role after successful authentication
   - Clears role on logout

**Tests (TDD):**
- RoleBloc: 11 tests (initial state, fetch success/failure, clear)
- RoleGuard: 10 tests (owner, admin, fallback, loading states)
- ManualIntegrationPage: 4 tests (owner, admin, loading, access denied)

**Files Created/Modified:**
- `lib/domain/entities/user_profile.dart`
- `lib/domain/repositories/user_profile_repository.dart`
- `lib/domain/repositories/auth_repository.dart` (updated)
- `lib/data/repositories/api_user_profile_repository.dart`
- `lib/data/repositories/firebase_auth_repository.dart` (updated)
- `lib/presentation/blocs/role/` - Bloc, events, states, barrel
- `lib/presentation/widgets/role_guard.dart`
- `lib/presentation/pages/admin/manual_integration_page.dart`
- `lib/presentation/router/app_router.dart` (updated)
- `lib/app.dart` (updated)
- `lib/core/di/injection.config.dart` (updated)
- `test/presentation/blocs/role_bloc_test.dart`
- `test/presentation/widgets/role_guard_test.dart`
- `test/presentation/pages/manual_integration_page_test.dart`

**Tests:** 54 passing

---

## [2026-02-27] Partner Integration Screen

**Commit:** Create Partner Integration screen with OAuth and manual token form

**Summary:**
Created Partner Integration page for connecting Shopify Partner account via OAuth or manual token entry (admin-only).

**Implemented:**

1. **Domain Layer:**
   - `PartnerIntegration` entity with status, partnerId, connectedAt
   - `IntegrationStatus` enum: notConnected, connecting, connected, error
   - `PartnerIntegrationRepository` interface
   - `PartnerIntegrationException` for error handling

2. **Data Layer:**
   - `MockPartnerIntegrationRepository` - Mock implementation for development
   - Simulates OAuth connection and manual token save
   - Configurable delay for realistic loading states

3. **Presentation Layer (Bloc):**
   - **PartnerIntegrationBloc** - Manages integration state
   - **Events:**
     - `CheckIntegrationStatusRequested` - Check current status
     - `ConnectWithOAuthRequested` - Initiate OAuth flow
     - `SaveManualTokenRequested` - Save manual token (admin only)
     - `DisconnectRequested` - Disconnect integration
   - **States:**
     - `PartnerIntegrationInitial` - Before status check
     - `PartnerIntegrationLoading` - During operations
     - `PartnerIntegrationNotConnected` - No integration
     - `PartnerIntegrationConnected` - Integration active
     - `PartnerIntegrationSuccess` - Action completed
     - `PartnerIntegrationError` - Error occurred

4. **Partner Integration Page:**
   - "Connect Shopify Partner" OAuth button (all users)
   - Manual Token form (admin-only via RoleGuard)
   - Partner ID and API Token input fields
   - Loading state with progress indicator
   - Connected state with partner ID display
   - Disconnect button when connected
   - Form validation for required fields
   - Success snackbar on completion
   - Route: `/partner-integration`

**Tests (TDD):**
- PartnerIntegrationBloc: 10 tests (status check, OAuth, manual token, disconnect)
- PartnerIntegrationPage: 16 tests (rendering, events, states, validation)

**Files Created/Modified:**
- `lib/domain/entities/partner_integration.dart`
- `lib/domain/repositories/partner_integration_repository.dart`
- `lib/data/repositories/mock_partner_integration_repository.dart`
- `lib/presentation/blocs/partner_integration/` - Bloc, events, states, barrel
- `lib/presentation/pages/partner_integration_page.dart`
- `lib/presentation/router/app_router.dart` (updated)
- `lib/app.dart` (updated)
- `lib/core/di/injection.config.dart` (updated)
- `lib/core/theme/app_theme.dart` (updated - added success color)
- `test/presentation/blocs/partner_integration_bloc_test.dart`
- `test/presentation/pages/partner_integration_page_test.dart`

**Tests:** 80 passing

---

## [2026-02-27] App Selection Screen

**Commit:** Create App Selection screen with radio list and local storage

**Summary:**
Created App Selection page for choosing which Shopify app to track revenue for after partner connection.

**Implemented:**

1. **Domain Layer:**
   - `ShopifyApp` entity with id, name, iconUrl, description, installCount
   - `AppRepository` interface for app operations
   - `AppException`, `NoAppsFoundException`, `FetchAppsException`

2. **Data Layer:**
   - `MockAppRepository` - Mock implementation with sample apps
   - Simulates fetch with configurable delay
   - Local storage for selected app (in-memory for now)

3. **Presentation Layer (Bloc):**
   - **AppSelectionBloc** - Manages app selection state
   - **Events:**
     - `FetchAppsRequested` - Load apps from backend
     - `AppSelected(app)` - User selected an app
     - `ConfirmSelectionRequested` - Save selection
     - `LoadSelectedAppRequested` - Load previous selection
   - **States:**
     - `AppSelectionInitial` - Before fetch
     - `AppSelectionLoading` - Fetching apps
     - `AppSelectionLoaded(apps, selectedApp)` - Ready for selection
     - `AppSelectionSaving` - Saving selection
     - `AppSelectionConfirmed(app)` - Selection saved
     - `AppSelectionError(message)` - Error occurred

4. **App Selection Page:**
   - Radio selection list with app tiles
   - App name, description, install count display
   - Visual selection indicator (radio + check icon)
   - "Confirm Selection" button when app selected
   - Loading state with progress indicator
   - Error state with retry button
   - Navigates to dashboard after confirmation
   - Route: `/app-selection`

5. **Partner Integration Update:**
   - Navigates to `/app-selection` after successful connection

**Tests (TDD):**
- AppSelectionBloc: 12 tests (fetch, select, confirm, load)
- AppSelectionPage: 15 tests (rendering, events, states)

**Files Created/Modified:**
- `lib/domain/entities/shopify_app.dart`
- `lib/domain/repositories/app_repository.dart`
- `lib/data/repositories/mock_app_repository.dart`
- `lib/presentation/blocs/app_selection/` - Bloc, events, states, barrel
- `lib/presentation/pages/app_selection_page.dart`
- `lib/presentation/pages/partner_integration_page.dart` (updated)
- `lib/presentation/router/app_router.dart` (updated)
- `lib/app.dart` (updated)
- `lib/core/di/injection.config.dart` (updated)
- `test/presentation/blocs/app_selection_bloc_test.dart`
- `test/presentation/pages/app_selection_page_test.dart`

**Tests:** 107 passing

---

## [2026-02-27] Executive Dashboard Layout

**Commit:** Create Executive Dashboard with KPI cards and metrics display

**Summary:**
Created Executive Dashboard page with primary KPIs (Renewal Success Rate, Active MRR, Revenue at Risk, Churn) and secondary section (Usage Revenue, Revenue Mix chart, Risk Distribution chart) using mock data.

**Implemented:**

1. **Domain Layer:**
   - `DashboardMetrics` entity with renewalSuccessRate, activeMrr, revenueAtRisk, churnedRevenue, churnedCount, usageRevenue
   - `RevenueMix` entity with recurring, usage, oneTime breakdown
   - `RiskDistribution` entity with safe, atRisk, critical, churned counts
   - `DashboardRepository` interface
   - `DashboardException` for error handling
   - Formatting methods for currency display

2. **Data Layer:**
   - `MockDashboardRepository` - Mock implementation with sample data
   - $124,500 MRR, 94.2% renewal rate sample metrics

3. **Presentation Layer (Bloc):**
   - **DashboardBloc** - Manages dashboard state
   - **Events:**
     - `LoadDashboardRequested` - Load metrics
     - `RefreshDashboardRequested` - Refresh metrics
   - **States:**
     - `DashboardInitial` - Before load
     - `DashboardLoading` - Loading metrics
     - `DashboardLoaded` - Metrics loaded with isRefreshing flag
     - `DashboardError` - Error occurred

4. **Dashboard Page:**
   - Responsive layout (4-column, 2x2, single column)
   - Primary KPIs section with 4 KpiCard widgets
   - Secondary section with Usage Revenue, Revenue Mix chart, Risk Distribution chart
   - Pull-to-refresh functionality
   - Refresh button in app bar
   - Error state with retry
   - Route: `/dashboard`

5. **Widgets:**
   - `KpiCard` - Large card for primary KPIs with icon, title, value, subtitle
   - `KpiCardCompact` - Compact version for secondary metrics
   - `RevenueMixChart` - Horizontal bar chart with recurring/usage/one-time legend
   - `RiskDistributionChart` - 2x2 grid showing safe/at-risk/critical/churned counts

**Tests (TDD):**
- DashboardBloc: 6 tests (initial state, load success/failure, refresh success/failure)
- DashboardPage: 19 tests (rendering, events, states, KPIs, charts, refresh)

**Files Created/Modified:**
- `lib/domain/entities/dashboard_metrics.dart`
- `lib/domain/repositories/dashboard_repository.dart`
- `lib/data/repositories/mock_dashboard_repository.dart`
- `lib/presentation/blocs/dashboard/` - Bloc, events, states, barrel
- `lib/presentation/pages/dashboard_page.dart`
- `lib/presentation/widgets/kpi_card.dart`
- `lib/presentation/widgets/revenue_mix_chart.dart`
- `lib/presentation/widgets/risk_distribution_chart.dart`
- `lib/presentation/router/app_router.dart` (updated)
- `lib/app.dart` (updated)
- `lib/core/di/injection.config.dart` (updated)
- `test/presentation/blocs/dashboard_bloc_test.dart`
- `test/presentation/pages/dashboard_page_test.dart`

**Tests:** 132 passing

---

## [2026-02-27] Connect Dashboard to Backend

**Commit:** Connect dashboard to daily_metrics_snapshot endpoint

**Summary:**
Connected dashboard to backend API endpoint. Added Total Revenue KPI, empty state handling, and API repository implementation.

**Implemented:**

1. **Domain Layer:**
   - Updated `DashboardMetrics` - Added `totalRevenue` field and `formattedTotalRevenue`
   - Updated `DashboardRepository` - Return nullable `DashboardMetrics?` for empty state
   - Added `NoAppSelectedException`, `NoMetricsException`, `UnauthorizedMetricsException`

2. **Data Layer:**
   - Created `ApiDashboardRepository` - Calls `/api/v1/apps/{appId}/metrics/latest`
   - Parses backend response (cents, decimal rate)
   - Handles 404 as empty state
   - Updated `MockDashboardRepository` - Support `returnEmpty` flag

3. **Presentation Layer:**
   - Added `DashboardEmpty` state for no metrics available
   - Updated `DashboardBloc` - Handle null metrics → emit `DashboardEmpty`
   - Updated `DashboardPage` - Empty state UI with "Sync Data" button, Total Revenue KPI card

4. **DI:**
   - Switched from `MockDashboardRepository` to `ApiDashboardRepository`

**Tests (TDD):**
- DashboardBloc: 8 tests (+2 empty state tests)
- DashboardPage: 25 tests (+5 empty state, +1 Total Revenue)

**Files Created/Modified:**
- `lib/domain/entities/dashboard_metrics.dart` (modified)
- `lib/domain/repositories/dashboard_repository.dart` (modified)
- `lib/data/repositories/api_dashboard_repository.dart` (created)
- `lib/data/repositories/mock_dashboard_repository.dart` (modified)
- `lib/presentation/blocs/dashboard/dashboard_bloc.dart` (modified)
- `lib/presentation/blocs/dashboard/dashboard_state.dart` (modified)
- `lib/presentation/pages/dashboard_page.dart` (modified)
- `lib/core/di/injection.config.dart` (modified)
- Tests updated

**Tests:** 140 passing

---

## [2026-02-27] KPI Dashboard Upgrade: Time Filtering and Delta Comparison

**Commit:** feat: implement KPI dashboard time filtering and delta comparison

**Summary:**
Upgraded dashboard with Play Store-style analytics featuring time-based filtering and period-over-period delta comparisons on KPI cards.

**Implemented:**

1. **Domain Layer:**
   - `TimeRange` entity with `TimeRangePreset` enum (thisMonth, lastMonth, last30Days, last90Days, custom)
   - Factory methods: `TimeRange.thisMonth()`, `TimeRange.lastMonth()`, etc.
   - Date formatting helpers for API calls
   - `MetricsDelta` class with percentage changes for each KPI
   - `DeltaIndicator` helper for determining direction and color

2. **Data Layer:**
   - Updated `ApiDashboardRepository` to call `/api/v1/apps/{appId}/metrics` with start/end query params
   - Parse delta from API response
   - Updated `MockDashboardRepository` with mock delta data

3. **Presentation Layer - BLoC:**
   - Added `TimeRangeChanged` event to `DashboardEvent`
   - Added `timeRange` field to `DashboardLoaded` state
   - `DashboardBloc` now tracks current time range and passes to repository

4. **Presentation Layer - Widgets:**
   - Created `TimeRangeSelector` - PopupMenuButton with preset options in app bar
   - Updated `KpiCard` with delta display:
     - `_DeltaBadge` for large cards
     - `_DeltaBadgeSmall` for compact cards
     - Arrow icon (up/down) with percentage
     - Green for good changes, red for bad changes

5. **Presentation Layer - Pages:**
   - Updated `DashboardPage` with `TimeRangeSelector` in app bar
   - Pass delta indicators to all KPI cards
   - Period subtitle shows selected range (e.g., "This Month")

6. **Delta Semantics:**
   | Metric | Higher is Good? | Green When |
   |--------|-----------------|------------|
   | Renewal Success Rate | Yes | Positive delta |
   | Active MRR | Yes | Positive delta |
   | Revenue at Risk | No | Negative delta |
   | Churn Count | No | Negative delta |
   | Usage Revenue | Yes | Positive delta |

**Tests:**
- Updated `dashboard_bloc_test.dart` - Added TimeRange to DashboardLoaded
- Updated `dashboard_page_test.dart` - Added TimeRange to all state tests
- All dashboard tests passing (32/32)

**Files Created/Modified:**
- `lib/domain/entities/time_range.dart` (created)
- `lib/domain/entities/dashboard_metrics.dart` (modified)
- `lib/domain/repositories/dashboard_repository.dart` (modified)
- `lib/data/repositories/api_dashboard_repository.dart` (modified)
- `lib/data/repositories/mock_dashboard_repository.dart` (modified)
- `lib/presentation/blocs/dashboard/dashboard_event.dart` (modified)
- `lib/presentation/blocs/dashboard/dashboard_state.dart` (modified)
- `lib/presentation/blocs/dashboard/dashboard_bloc.dart` (modified)
- `lib/presentation/widgets/time_range_selector.dart` (created)
- `lib/presentation/widgets/kpi_card.dart` (modified)
- `lib/presentation/pages/dashboard_page.dart` (modified)
- Tests updated

**Tests:** All dashboard tests passing

---

## Test Summary

| Layer | Tests |
|-------|-------|
| presentation/blocs/auth | 11 |
| presentation/blocs/role | 11 |
| presentation/blocs/partner_integration | 10 |
| presentation/blocs/app_selection | 12 |
| presentation/blocs/dashboard | 8 |
| presentation/pages/login | 9 |
| presentation/pages/signup | 8 |
| presentation/pages/manual_integration | 4 |
| presentation/pages/partner_integration | 16 |
| presentation/pages/app_selection | 15 |
| presentation/pages/dashboard | 25 |
| presentation/widgets/role_guard | 10 |
| widget | 1 |
| **Total** | **140** |

---

## Architecture

```
frontend/app/lib/
├── core/
│   ├── config/         → EnvConfig, AppConfig
│   ├── constants/      → App constants
│   ├── di/             → get_it + injectable setup
│   ├── theme/          → AppTheme (Material 3)
│   └── utils/          → Utilities
├── data/
│   ├── datasources/    → API clients, local storage
│   ├── models/         → JSON serializable models
│   └── repositories/   → FirebaseAuthRepository, ApiUserProfileRepository, MockPartnerIntegrationRepository, MockAppRepository, ApiDashboardRepository
├── domain/
│   ├── entities/       → UserEntity, UserProfile, PartnerIntegration, ShopifyApp, DashboardMetrics
│   ├── repositories/   → AuthRepository, UserProfileRepository, PartnerIntegrationRepository, AppRepository, DashboardRepository
│   └── usecases/       → Business logic
└── presentation/
    ├── blocs/          → AuthBloc, RoleBloc, PartnerIntegrationBloc, AppSelectionBloc, DashboardBloc
    ├── pages/          → LoginPage, SignupPage, ManualIntegrationPage, PartnerIntegrationPage, AppSelectionPage, DashboardPage, SubscriptionListPage, SubscriptionDetailPage
    ├── widgets/        → RoleGuard, ProGuard, KpiCard, RevenueMixChart, RiskDistributionChart, RiskBadge, SubscriptionTile
    └── router/         → GoRouter with auth/role redirects
```

---

## [2026-02-27] Subscription List and Detail Pages

**Commit:** feat: implement subscription list and detail views

**Summary:**
Implemented subscription list and detail pages with risk filtering, showing store display name instead of domain.

**Implemented:**

1. **Domain Layer:**
   - `Subscription` entity with all fields including shopName
   - `SubscriptionRepository` interface with list, filter, and getById

2. **Data Layer:**
   - `ApiSubscriptionRepository` - API implementation

3. **Presentation Layer (Bloc):**
   - `SubscriptionListBloc` - List with filtering
   - `SubscriptionDetailBloc` - Single subscription detail

4. **Widgets:**
   - `SubscriptionTile` - List item with avatar, store name, plan, risk badge
   - `RiskBadge` - Colored badge for risk state

5. **Pages:**
   - `SubscriptionListPage` - List with risk filter dropdown
   - `SubscriptionDetailPage` - Full subscription details

6. **Fixes:**
   - Fixed index out of range in `_getInitials` and `_formatDisplayName`
   - Added defensive null/empty checks for string operations

**Files Created/Modified:**
- `lib/domain/entities/subscription.dart`
- `lib/domain/repositories/subscription_repository.dart`
- `lib/data/repositories/api_subscription_repository.dart`
- `lib/presentation/blocs/subscription_list/` - Bloc, events, states
- `lib/presentation/blocs/subscription_detail/` - Bloc, events, states
- `lib/presentation/pages/subscription_list_page.dart`
- `lib/presentation/pages/subscription_detail_page.dart`
- `lib/presentation/widgets/subscription_tile.dart`
- `lib/presentation/widgets/risk_badge.dart`
- `lib/presentation/router/app_router.dart` (updated)

---

## [2026-02-27] Mobile Responsiveness Improvements

**Commit:** feat(frontend): add mobile responsiveness improvements

**Summary:**
Added responsive layouts for small screens (phones) across dashboard and subscription pages.

**Implemented:**

1. **Dashboard Page:**
   - Replaced 5 AppBar action buttons with overflow menu (PopupMenuButton)
   - Responsive padding: 12px on mobile (<600px), 20px on desktop
   - Shortened title to "Dashboard"

2. **KPI Cards:**
   - LayoutBuilder for responsive sizing based on card width
   - FittedBox wraps value text to prevent overflow
   - Compact mode (<200px): smaller fonts, padding, icons
   - Delta badges scale with card size

3. **TimeRangeSelector:**
   - Added `shortName` property to TimeRangePreset ("Month", "30D", "90D")
   - Hide calendar icon on screens <400px
   - Compact padding on mobile

4. **Subscription List Page:**
   - Responsive list padding
   - Compact filter bar count on small screens

5. **Subscription Tile:**
   - Responsive avatar size (40px vs 48px)
   - Responsive padding and font sizes
   - Combined plan/price text on mobile

**Responsive Breakpoints:**
- `< 400px` - Compact (phone portrait)
- `< 600px` - Mobile
- `< 800px` - Tablet
- `≥ 800px` - Desktop

**Files Modified:**
- `lib/domain/entities/time_range.dart` - Added shortName
- `lib/presentation/pages/dashboard_page.dart` - Overflow menu, padding
- `lib/presentation/pages/subscription_list_page.dart` - Responsive padding
- `lib/presentation/widgets/kpi_card.dart` - FittedBox, compact mode
- `lib/presentation/widgets/subscription_tile.dart` - Responsive sizing
- `lib/presentation/widgets/time_range_selector.dart` - Compact mode

---

## [2026-02-28] API Key Management Frontend

**Commit:** feat(frontend): implement API key management for Revenue API

**Summary:**
Created API Key management screens allowing users to create, view, and revoke API keys for accessing the Revenue API (REST/GraphQL).

**Implemented:**

1. **Domain Layer:**
   - `ApiKey` entity with id, name, keyPrefix, createdAt, lastUsedAt
   - `ApiKeyCreationResult` for returning full key (shown only once after creation)
   - `ApiKeyRepository` interface with getApiKeys, createApiKey, revokeApiKey
   - Exception classes: `ApiKeyException`, `ApiKeyLimitException`, `ApiKeyNotFoundException`, `ApiKeyUnauthorizedException`

2. **Data Layer:**
   - `ApiApiKeyRepository` - API implementation calling `/v1/api-keys` endpoints
   - Uses Dio with Bearer token authentication
   - Handles error responses with proper exception mapping

3. **Presentation Layer (Bloc):**
   - **ApiKeyBloc** - Manages API key state
   - **Events:**
     - `LoadApiKeysRequested` - Load all API keys
     - `CreateApiKeyRequested(name)` - Create new key
     - `RevokeApiKeyRequested(keyId)` - Revoke/delete key
     - `DismissKeyCreatedRequested` - Dismiss key created dialog
   - **States:**
     - `ApiKeyInitial` - Before load
     - `ApiKeyLoading` - Loading keys
     - `ApiKeyLoaded(apiKeys, isCreating, isRevoking, revokingKeyId)` - Keys loaded
     - `ApiKeyCreated(apiKeys, fullKey, keyName)` - Key created with full secret
     - `ApiKeyEmpty` - No keys exist
     - `ApiKeyError(message, previousKeys)` - Error with recovery

4. **API Key List Page:**
   - List of existing API keys with name, prefix, dates
   - Create button in app bar opens dialog
   - Create dialog with name validation
   - Key created dialog shows full key once with copy button and warning
   - Revoke confirmation dialog with warning about breaking apps
   - Pull-to-refresh functionality
   - Empty state with create button
   - Error state with retry button
   - Route: `/settings/api-keys`

5. **Widgets:**
   - `ApiKeyTile` - Card showing key name, masked prefix, created date, last used
   - Copy prefix button, revoke button with loading indicator

6. **Profile Page Integration:**
   - Added "API Keys" navigation tile in Settings section
   - Links to `/settings/api-keys`

**Tests (TDD):**
- ApiKeyBloc: 14 tests
  - Initial state
  - Load success/empty/error
  - Create from loaded/empty state, error with limit
  - Revoke success/empty result/error
  - Dismiss key created state

**Files Created/Modified:**
- `lib/domain/entities/api_key.dart` (created)
- `lib/domain/repositories/api_key_repository.dart` (created)
- `lib/data/repositories/api_api_key_repository.dart` (created)
- `lib/presentation/blocs/api_key/` - Bloc, events, states, barrel (created)
- `lib/presentation/pages/api_key_list_page.dart` (created)
- `lib/presentation/widgets/api_key_tile.dart` (created)
- `lib/presentation/pages/profile_page.dart` (modified - added API Keys link)
- `lib/presentation/router/app_router.dart` (modified - added route)
- `lib/core/di/injection.config.dart` (modified - registered DI)
- `test/presentation/blocs/api_key_bloc_test.dart` (created)

**Tests:** 14 passing (ApiKeyBloc)

---

## [2026-02-28] Earnings Timeline Enhancements

**Commit:** feat: enhance earnings timeline with date range sync and animations

**Summary:**
Enhanced the Earnings Timeline chart with dashboard time range synchronization, smooth animations, and configurable visibility in dashboard settings.

**Implemented:**

1. **Backend Changes:**
   - Changed API from `year/month` to `start/end` date parameters
   - Updated `RevenueRepository` interface with `GetRevenueByDateRange(startDate, endDate)`
   - Updated handler to parse ISO date strings from query params

2. **Frontend - Domain Layer:**
   - Updated `EarningsTimeline` entity for date range-based queries
   - Updated `EarningsRepository` interface with `fetchEarnings(startDate, endDate, mode)`

3. **Frontend - Data Layer:**
   - Updated `ApiEarningsRepository` to use start/end query parameters

4. **Frontend - Presentation Layer (Bloc):**
   - Added `EarningsTimeRangeChanged` event for dashboard sync
   - Added `_getTargetMonth(TimeRange)` to extract appropriate month from presets
   - Month navigation syncs with dashboard filter but allows manual browsing

5. **Frontend - Widget Enhancements:**
   - Added `earningsTimeline` to `SecondaryWidget` enum for dashboard config
   - Premium-styled `DashboardConfigDialog` with gradient header, icons
   - Smooth chart animations (250ms easeInOut) when data changes
   - Loading overlay preserves chart during data fetch (prevents scroll jumps)
   - Consistent 300px card height across all states
   - `ValueKey` for widget identity preservation

6. **Animation Fix:**
   - Cache `_lastLoadedState` to preserve chart during loading
   - Show semi-transparent overlay with spinner during data fetch
   - BarChart widget persists for smooth fl_chart animations

**Files Modified:**
- Backend: `revenue_repository.go`, `revenue_metrics_service.go`, `revenue_handler.go`, `router.go`
- Frontend: `earnings_timeline.dart`, `earnings_repository.dart`, `api_earnings_repository.dart`
- Frontend: `earnings_bloc.dart`, `earnings_event.dart`, `earnings_state.dart`
- Frontend: `earnings_timeline_chart.dart`, `dashboard_page.dart`
- Frontend: `dashboard_preferences.dart`, `dashboard_config_dialog.dart`
- Tests: `dashboard_config_dialog_test.dart`

**Tests:** All dashboard tests passing (24/24), config dialog tests passing (15/15)

---

## [2026-02-28] Subscriptions Page Premium Analytics

**Commit:** feat: upgrade subscriptions page with advanced filtering, sorting, and pagination

**Summary:**
Transformed the basic subscriptions list into a premium SaaS-level reporting interface with server-side filtering, sorting, and pagination. Includes summary statistics, dynamic price ranges, and sortable table headers.

**Implemented:**

1. **Domain Layer:**
   - `SubscriptionFilters` class with riskStates, priceRange, billingInterval, searchQuery, sort, sortAscending
   - `PriceRange` entity with label, minCents, maxCents, count
   - `SubscriptionSummary` entity with activeCount, atRiskCount, churnedCount, avgPriceCents, totalCount
   - `SubscriptionSort` enum with apiValue and displayName (riskState, price, shopName)
   - `PaginatedSubscriptionResponse` with page, pageSize, totalPages, hasNextPage, hasPreviousPage, rangeText

2. **Data Layer:**
   - Updated `SubscriptionRepository` interface with:
     - `getSubscriptionsFiltered(appId, filters, page, pageSize)`
     - `getSummary(appId)` - Returns subscription statistics
     - `getPriceRanges(appId)` - Returns dynamic price tiers
   - Updated `ApiSubscriptionRepository` with new endpoint implementations

3. **Presentation Layer (Bloc):**
   - **New Events:**
     - `LoadSubscriptionsRequested` - Initial load with summary + price ranges + list
     - `ApplyFiltersRequested` - Apply filter changes (resets to page 1)
     - `ChangePageRequested` - Navigate to specific page
     - `ChangePageSizeRequested` - Change rows per page (resets to page 1)
     - `ChangeSortRequested` - Change sort column and direction
     - `SearchRequested` - Debounced search (resets to page 1)
     - `ClearFiltersRequested` - Reset all filters
   - **Enhanced State:**
     - `SubscriptionListLoaded` now includes summary, priceRanges, filters, page, pageSize, totalPages
     - `SubscriptionListEmpty` includes summary and filters for showing filter bar
   - Cached summary and priceRanges to avoid refetching on filter/page changes
   - Backward compatibility maintained with legacy events

4. **New Widgets:**
   - `SubscriptionSummaryBar` - 4 stat cards (Active, At Risk, Churned, Avg Price)
     - Horizontal scrollable on mobile
     - Color-coded icons matching risk semantics
     - Loading state with shimmer effect
   - `SubscriptionFilterBar` - Advanced filtering UI
     - Multi-select risk state chips
     - Price range dropdown (populated from API)
     - Billing interval dropdown (Monthly/Annual)
     - Search input with debounce (300ms)
     - Clear filters button with active filter count badge
   - `PaginationControls` - Server-side pagination
     - Page size selector (10, 25, 50)
     - Page navigation with ellipsis for large page counts
     - "Showing X-Y of Z" range text

5. **Subscription List Page:**
   - New structure: SummaryBar → FilterBar → TableHeader → ListView → PaginationControls
   - Sortable table header (Shop, Price, Risk) with sort indicators
   - Loading overlay preserves list during filter/page changes
   - Empty state varies based on filters (shows Clear Filters button if filters active)
   - Pull-to-refresh functionality

**Page Structure:**
```
Scaffold
├── AppBar (title, refresh button)
└── Column
    ├── SubscriptionSummaryBar
    ├── SubscriptionFilterBar
    ├── _SubscriptionTableHeader (sortable)
    ├── Expanded(ListView with SubscriptionTiles)
    └── PaginationControls
```

**Files Created:**
- `lib/domain/entities/subscription_filter.dart`
- `lib/presentation/widgets/subscription_summary_bar.dart`
- `lib/presentation/widgets/subscription_filter_bar.dart`
- `lib/presentation/widgets/pagination_controls.dart`

**Files Modified:**
- `lib/domain/entities/subscription.dart` - Added apiValue to BillingInterval
- `lib/domain/repositories/subscription_repository.dart` - Added new methods
- `lib/data/repositories/api_subscription_repository.dart` - Implemented new endpoints
- `lib/presentation/blocs/subscription_list/subscription_list_event.dart` - New events
- `lib/presentation/blocs/subscription_list/subscription_list_state.dart` - Enhanced states
- `lib/presentation/blocs/subscription_list/subscription_list_bloc.dart` - New handlers
- `lib/presentation/pages/subscription_list_page.dart` - Complete rewrite

**Backend Integration:**
- Uses `/api/v1/apps/{appId}/subscriptions` with query params: status, priceMin, priceMax, billingInterval, search, sortBy, sortOrder, page, pageSize
- Uses `/api/v1/apps/{appId}/subscriptions/summary` for summary stats
- Uses `/api/v1/apps/{appId}/subscriptions/price-stats` for distinct prices

---

## [2026-02-28] Price Stats with Distinct Prices

**Commit:** feat: change price filter to distinct prices dropdown

**Summary:**
Changed the price filter from tier-based ranges to a dropdown showing all distinct prices with counts, allowing developers to accurately filter by exact price points.

**Implemented:**

1. **Domain Layer:**
   - `PricePoint` entity with priceCents, count, and formatted getter
   - Updated `PriceStats` to include `prices: List<PricePoint>` (sorted ascending)
   - Removed `PriceRange` entity (replaced by PricePoint)
   - Updated `SubscriptionFilters` to use `priceMinCents/priceMaxCents` for exact price matching

2. **Data Layer:**
   - Updated `SubscriptionRepository.getPriceStats()` to parse prices array
   - `PriceStats.fromJson()` parses prices list from API response

3. **Presentation Layer (Bloc):**
   - Changed `priceRanges` to `priceStats` in `SubscriptionListLoaded` state
   - Updated bloc to call `getPriceStats()` instead of `getPriceRanges()`
   - Cached `priceStats` to avoid refetching on filter/page changes

4. **Widget Changes - SubscriptionFilterBar:**
   - Changed price filter from dialog to `_FilterDropdown<int?>` widget
   - Dropdown shows "All prices" + all distinct prices from API
   - Each option displays formatted price with count: `$4.99 (281)`
   - `_setPriceFilter(int? priceCents)` sets both priceMin and priceMax to same value for exact match
   - `_getCurrentPriceFilter()` returns current price filter if min == max, null otherwise
   - Prices displayed in ascending order (sorted by backend)

5. **Backend API Response:**
   ```json
   {
     "minCents": 499,
     "maxCents": 40499,
     "avgCents": 4389,
     "prices": [
       { "priceCents": 499, "count": 281 },
       { "priceCents": 1283, "count": 1 },
       { "priceCents": 4999, "count": 156 }
     ]
   }
   ```

**Files Modified:**
- `lib/domain/entities/subscription_filter.dart` - Added PricePoint, updated PriceStats
- `lib/data/repositories/api_subscription_repository.dart` - Updated parsing
- `lib/presentation/blocs/subscription_list/subscription_list_state.dart` - Changed to priceStats
- `lib/presentation/blocs/subscription_list/subscription_list_bloc.dart` - Updated caching
- `lib/presentation/widgets/subscription_filter_bar.dart` - Changed to dropdown
- `lib/presentation/pages/subscription_list_page.dart` - Pass priceStats instead of priceRanges
- `test/presentation/blocs/subscription_list_bloc_test.dart` - Updated mocks

---

---

## [2026-03-01] Revenue Share Tier Tracking - Frontend (Phase 1)

**Commit:** feat: add revenue share tier tracking with fee breakdown

**Summary:**
Implemented frontend support for Shopify revenue share tier tracking. Users can view and change their app's tier, see fee breakdowns, and compare savings across tiers.

**Files Created:**
- `lib/domain/entities/revenue_share_tier.dart` - RevenueShareTier enum with 4 tiers, FeeBreakdown class with calculate() factory
- `lib/presentation/widgets/tier_selector.dart` - TierSelector widget for display/selection, TierIndicator for compact display
- `lib/presentation/pages/app_settings_page.dart` - Full page for tier management with fee calculator and tier comparison
- `lib/presentation/widgets/fee_insights_card.dart` - FeeInsightsCard for dashboard, FeeKpiCard for compact view

**Files Modified:**
- `lib/domain/entities/shopify_app.dart` - Added revenueShareTier field, copyWith method
- `lib/domain/repositories/app_repository.dart` - Added updateAppTier, getFeeSummary, getFeeBreakdown methods, FeeSummary/TierSavings classes
- `lib/data/repositories/api_app_repository.dart` - Implemented new repository methods
- `lib/data/repositories/mock_app_repository.dart` - Mock implementations
- `lib/presentation/pages/app_selection_page.dart` - Shows tier indicator on each app
- `lib/presentation/pages/profile_page.dart` - Added "App Settings" navigation tile
- `lib/presentation/pages/dashboard_page.dart` - Added Fee Insights section
- `lib/presentation/router/app_router.dart` - Added /settings/app route

**Features:**
- View current tier with color-coded badges (green=0%, amber=15%, red=20%)
- Change tier with dropdown selector
- Interactive fee calculator with adjustable amounts ($10-$500 slider)
- Tier comparison showing fees/net for all tiers on $100 sample
- Dashboard fee insights showing total fees and savings vs default 20%

**Tests:**
- Updated dashboard_page_test.dart with AppRepository mock
- Fixed profile_page_test.dart (21 tests) - matched tests to actual implementation

---

## [2026-03-01] Earnings Timeline Tracking - Frontend (Phase 2)

**Commit:** feat: add earnings timeline with pending/available status tracking

**Summary:**
Implemented frontend support for Shopify earnings availability tracking. Shows pending vs available earnings with upcoming availability dates.

**Files Created:**
- `lib/domain/entities/earnings_status.dart` - EarningsStatus and EarningsDateEntry entities with:
  - `totalPendingCents`, `totalAvailableCents`, `totalPaidOutCents`
  - `pendingByDate`, `upcomingAvailability` lists
  - Formatting helpers: `formattedPending`, `formattedAvailable`, `formattedPaidOut`
  - `nextAvailable`, `daysUntilNextAvailable` computed properties
- `lib/presentation/widgets/earnings_status_card.dart` - Two display modes:
  - **Full card**: Shows pending/available tiles, upcoming timeline, paid out summary
  - **Compact card**: Shows available + pending in single row with "next available" badge
  - **EarningsKpiCard**: Compact KPI version for dashboard grid

**Files Modified:**
- `lib/domain/repositories/earnings_repository.dart` - Added `fetchEarningsStatus()` method
- `lib/data/repositories/api_earnings_repository.dart` - Implemented API call to `/api/v1/apps/{appId}/earnings/status`
- `lib/core/network/api_client.dart` - Added missing `patch()` method for tier updates

**Widget Features:**
- Loading state with progress indicator
- Error state with graceful fallback (shows empty)
- Color-coded status indicators (amber for pending, green for available)
- Upcoming availability timeline with days countdown
- Refresh button on full card

**Usage:**
```dart
// Full card
EarningsStatusCard()

// Compact card (for dashboard)
EarningsStatusCard(compact: true)

// KPI card (requires pre-loaded status)
EarningsKpiCard(status: earningsStatus)
```

---
