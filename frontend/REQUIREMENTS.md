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
- API: `https://api.ledgerguard.com`
- Firebase: Production project

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
- [ ] Subscription List
- [ ] Subscription Detail

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

---

## Notes

- Web-first, but structure supports mobile later
- No UI implementation in initial setup
- Firebase core only (auth screens come later)
