# F04. Dark Mode

## What It Will Do
Add a dark color theme to the Flutter app with system-aware automatic switching and manual toggle. All screens, charts, and custom widgets adapt to the selected theme. User preference persists across sessions.

## Why It Matters
Dark mode reduces eye strain during long analysis sessions, saves battery on OLED devices, and is expected by modern app users. Many developers (the target audience) prefer dark themes.

## Dependencies
- Flutter Material 3 theming (implemented — `app_theme.dart`)
- SharedPreferences for persistence (available in `frontend/app/`)
- `fl_chart` supports dark mode via theme colors

## Integration Points
- `AppTheme` class needs a dark variant alongside the existing light theme
- All `LgCard`, `LgMetricCard`, `LgBadge` widgets must use theme colors instead of hardcoded values
- `fl_chart` charts need theme-aware color sets
- AppShell navigation needs dark variant
- Settings screen: theme toggle (System / Light / Dark)

## Estimated Scope
- Define dark color palette: 0.5 day
- Create `AppTheme.dark()` constructor: 0.5 day
- Audit all widgets for hardcoded colors: 1-2 days
- Update chart color configurations: 1 day
- Settings UI + persistence: 0.5 day
- Total: ~3-4 days

## Open Questions
- Should we support a "system follows device" option? (Recommended: yes)
- Do marketing site pages need dark mode too? (Suggested: no, separate concern)
- How to handle the onboarding checklist in dark mode (different illustration colors)?

## Suggested Approach
1. Define dark palette in `app_colors.dart` (dark backgrounds, muted card surfaces, bright text)
2. Add `AppTheme.dark()` using Material 3 `ColorScheme.dark()`
3. Add `ThemeMode` selection to `SettingsProvider` with SharedPreferences persistence
4. Wrap `MaterialApp` with `themeMode: settingsProvider.themeMode`
5. Audit all custom widgets — replace `Color(0xFF...)` with `Theme.of(context).colorScheme.xxx`
6. Update fl_chart configurations to read from theme
7. Test on web, iOS, and Android
