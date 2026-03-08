# PLAN-18: Dashboard Load Fix & UI Polish

**Date:** 2026-03-09
**Status:** Completed

## Scope
Fix multiple UI issues discovered during Android APK testing:

### 1. Dashboard Initial Load Race Condition
- **Problem:** `LoadDashboardRequested` fired before `FetchAppsRequested` completed, causing null app ID
- **Fix:** Auto-select first app in `AppSelectionBloc` + defer dashboard load via `BlocListener`
- Added `didChangeDependencies` for returning users with app already in SharedPreferences

### 2. AppBar Visibility
- **Problem:** White text on white AppBar background (invisible on Android)
- **Fix:** Changed AppBar theme to primary blue (`#2563EB`) with white foreground

### 3. Scaffold Background Standardization
- **Problem:** 9 pages had hardcoded `backgroundColor: Colors.grey[50]`, others used theme default (white)
- **Fix:** Set `scaffoldBackgroundColor: Color(0xFFF9FAFB)` in theme, removed all 9 overrides

### 4. Preferences Save Button Visibility
- **Problem:** TextButton with default primary color (blue) on blue AppBar = invisible
- **Fix:** Added `foregroundColor: Colors.white` to Save TextButton

### 5. Popup/Dialog Tint
- **Problem:** M3 `ColorScheme.fromSeed` generated blue-tinted surfaces for popups/dialogs
- **Fix:** Added `popupMenuTheme` and `dialogTheme` with `surfaceTintColor: Colors.transparent`

## Files Modified
- `app_theme.dart` — AppBar, scaffold, popup, dialog themes
- `app_selection_bloc.dart` — Auto-select first app
- `dashboard_page.dart` — BlocListener + didChangeDependencies
- `preferences_page.dart` — Save button white foreground
- 9 page files — Removed hardcoded scaffold backgrounds
- 3 test files — Updated for new behavior

## Commits
- Auto-select app (b9aa36e), deferred load (e0123e9), AppBar (20aa855), scaffold bg (b311d44), Save button (c5d31b9), popup/dialog (3be9e1c)
