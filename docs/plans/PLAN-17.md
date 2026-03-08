# PLAN-17: Material 3 Theme Standardization

**Date:** 2026-03-09
**Status:** Completed

## Scope
Migrate ~120 hardcoded `TextStyle(fontSize:)` declarations across 21 presentation-layer files to Material 3 semantic `Theme.of(context).textTheme.*` styles.

## Type Scale Mapping
| Hardcoded Pattern | M3 Style |
|---|---|
| `fontSize: 26, bold` | `headlineLarge` |
| `fontSize: 24, bold` | `headlineMedium` |
| `fontSize: 18, bold` | `titleLarge` |
| `fontSize: 16, w600` | `titleMedium` |
| `fontSize: 14, w600` | `titleSmall` |
| `fontSize: 16` (body) | `bodyLarge` |
| `fontSize: 14` (body) | `bodyMedium` |
| `fontSize: 13` (body) | `bodySmall` |
| `fontSize: 12` (label) | `labelMedium` |
| `fontSize: 11` (tiny) | `labelSmall` |

## Migration Pattern
**Before:** `TextStyle(fontSize: 18, fontWeight: FontWeight.bold)`
**After:** `Theme.of(context).textTheme.titleLarge`

Colors remain inline: `.copyWith(color: Colors.grey[600])`

## Files Modified
- `core/theme/app_theme.dart` — Added `textTheme` to ThemeData
- 21 presentation files — Migrated all hardcoded styles
- `test/presentation/widgets/risk_badge_test.dart` — Added AppTheme.lightTheme

## Exceptions (3)
- 2 chart tooltips (no BuildContext in fl_chart callbacks)
- 1 pagination widget (out of scope)

## Verification
- All 405 Flutter tests passing
- `grep -r 'TextStyle(fontSize:' lib/presentation/` returns only acceptable exceptions
