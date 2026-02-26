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

## Test Summary

| Layer | Tests |
|-------|-------|
| presentation/blocs/auth | 11 |
| presentation/blocs/role | 11 |
| presentation/blocs/partner_integration | 10 |
| presentation/blocs/app_selection | 12 |
| presentation/blocs/dashboard | 6 |
| presentation/pages/login | 9 |
| presentation/pages/signup | 8 |
| presentation/pages/manual_integration | 4 |
| presentation/pages/partner_integration | 16 |
| presentation/pages/app_selection | 15 |
| presentation/pages/dashboard | 19 |
| presentation/widgets/role_guard | 10 |
| widget | 1 |
| **Total** | **132** |

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
│   └── repositories/   → FirebaseAuthRepository, ApiUserProfileRepository, MockPartnerIntegrationRepository, MockAppRepository, MockDashboardRepository
├── domain/
│   ├── entities/       → UserEntity, UserProfile, PartnerIntegration, ShopifyApp, DashboardMetrics
│   ├── repositories/   → AuthRepository, UserProfileRepository, PartnerIntegrationRepository, AppRepository, DashboardRepository
│   └── usecases/       → Business logic
└── presentation/
    ├── blocs/          → AuthBloc, RoleBloc, PartnerIntegrationBloc, AppSelectionBloc, DashboardBloc
    ├── pages/          → LoginPage, SignupPage, ManualIntegrationPage, PartnerIntegrationPage, AppSelectionPage, DashboardPage
    ├── widgets/        → RoleGuard, ProGuard, KpiCard, RevenueMixChart, RiskDistributionChart
    └── router/         → GoRouter with auth/role redirects
```
