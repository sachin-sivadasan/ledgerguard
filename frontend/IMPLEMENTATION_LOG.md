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

## Test Summary

| Layer | Tests |
|-------|-------|
| presentation/blocs/auth | 11 |
| presentation/pages/login | 9 |
| presentation/pages/signup | 8 |
| widget | 1 |
| **Total** | **29** |

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
│   └── repositories/   → Repository implementations (FirebaseAuthRepository)
├── domain/
│   ├── entities/       → Business entities (UserEntity)
│   ├── repositories/   → Repository interfaces (AuthRepository)
│   └── usecases/       → Business logic
└── presentation/
    ├── blocs/          → Bloc state management (AuthBloc)
    ├── pages/          → Screen widgets (LoginPage, SignupPage)
    ├── widgets/        → Reusable components
    └── router/         → GoRouter with auth redirects
```
