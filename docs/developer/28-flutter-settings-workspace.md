# 28. Flutter Settings & Workspace

## What It Does
Central settings hub providing navigation to: app settings (revenue share tier), API key management (create/revoke keys), authentication (logout, profile), notification preferences, and billing management. Manages workspace-level configuration for the logged-in user.

## Architecture
Presentation layer. Multiple providers collaborate: `SettingsProvider` manages general settings, `ApiKeyProvider` manages API key CRUD, `AuthProvider` handles authentication state. `SettingsScreen` is a navigation hub; detail screens handle each settings area.

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `ledgerguard-flutter/lib/providers/settings_provider.dart` | ~50 | General settings state |
| `ledgerguard-flutter/lib/providers/api_key_provider.dart` | ~60 | API key list, create, revoke |
| `ledgerguard-flutter/lib/providers/auth_provider.dart` | ~40 | Auth state (logged in/out) |
| `ledgerguard-flutter/lib/screens/settings/settings_screen.dart` | ~200 | Settings navigation hub |
| `ledgerguard-flutter/lib/screens/api_keys/api_keys_screen.dart` | ~200 | API key management |
| `ledgerguard-flutter/lib/screens/apps/apps_screen.dart` | ~150 | App settings (tier, slug) |
| `ledgerguard-flutter/lib/models/api_key_model.dart` | ~30 | API key data model |

## Data Flow
```
SettingsScreen (navigation hub)
    ├── App Settings → AppsScreen (revenue tier, tracking)
    ├── API Keys → ApiKeysScreen (create, list, revoke)
    ├── Notifications → NotificationPrefs
    ├── Billing → BillingScreen (plan tier)
    └── Profile → AuthProvider (logout)
```

## Configuration
None — mock data.

## Widget Tree
```
SettingsScreen
├── LgPage (title: "Settings")
│   └── Column
│       ├── ListTile: App Settings → AppsScreen
│       ├── ListTile: API Keys → ApiKeysScreen
│       ├── ListTile: Notifications → NotificationPrefs
│       ├── ListTile: Billing → BillingScreen
│       ├── ListTile: Profile → ProfileScreen
│       ├── Divider
│       └── ListTile: Logout (destructive)

ApiKeysScreen
├── LgPage (title: "API Keys")
│   ├── "Create API Key" button
│   └── ListView.builder
│       └── ListTile per API key
│           ├── Key name
│           ├── Masked key value (xxxx...xxxx)
│           ├── Created date
│           └── Delete button (with confirmation)
```

## State Machine
```
SettingsProvider (ChangeNotifier)
  State:
    _selectedTheme: ThemeMode
    _notificationSettings: Map

  Events:
    setTheme(), updateNotificationSetting()

ApiKeyProvider (ChangeNotifier)
  State:
    _apiKeys: List<ApiKey>
    _isCreating: bool

  Events:
    createKey(name)  → add new key
    revokeKey(id)    → remove key
    loadKeys()       → refresh list

  Computed:
    apiKeys → current key list
    isCreating → loading state for create

AuthProvider (ChangeNotifier)
  State:
    _isLoggedIn: bool
    _user: User?

  Events:
    login()  → set logged in
    logout() → clear state, navigate to login
```

## Gotchas
- API key is only shown in full once at creation (masked afterward) — production pattern
- Logout navigates to login screen and clears all provider state
- Revenue share tier defaults to SMALL_DEV_0 (0%) per ADR-008
- Delete operations show a confirmation dialog (`LgConfirmationDialog`)
- Billing screen links to external Razorpay checkout URL in production
