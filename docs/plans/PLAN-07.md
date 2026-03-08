# PLAN-07: Flutter App Foundation (Auth, Navigation, DI)

**Date:** 2024-01-XX
**Status:** Completed

## Scope
- Initialize Flutter Web project with Clean Architecture folder structure
- Bloc for state management, GoRouter for navigation
- get_it + injectable for dependency injection
- Firebase Auth integration (email/password + Google login)
- AuthBloc (events: AuthCheckRequested, SignInWithEmail, SignInWithGoogle, SignOut)
- Login and Signup screens with form validation
- Role fetching from backend (RoleBloc)
- Environment configs (dev/prod)

## Architecture
```
lib/
├── core/          → Config, constants, theme, utils, DI
├── data/          → Datasources, models, repositories
├── domain/        → Entities, repository interfaces, use cases
└── presentation/  → Blocs, pages, widgets, router
```

## Key Patterns
- Bloc pattern: Events → States with Emitter
- GoRouter with auth-aware redirect (unauthenticated → /login, authenticated → /dashboard)
- Firebase ID token passed as Bearer token to backend
