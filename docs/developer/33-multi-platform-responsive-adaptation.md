# 33 – Multi-Platform Responsive Adaptation

> Enables Android, iOS, macOS, and tablet support with responsive layouts for the `frontend-flutter/` Provider-based prototype.

---

## Overview

The `frontend-flutter/` app was originally web-only with desktop-first layouts. This feature adds:

1. **Platform scaffolding** — Android, iOS, macOS directories
2. **Responsive breakpoints** — Mobile (<600px), Tablet (600–900px), Desktop (>900px)
3. **Adaptive navigation** — BottomNav (mobile), Drawer (tablet), NavigationRail (desktop)
4. **Per-screen responsive layouts** — All 13 screens adapt to narrow viewports

---

## Breakpoint Strategy

| Range | Device Type | Navigation | LgPage Padding | Metric Columns |
|-------|------------|------------|----------------|----------------|
| <600px | Mobile | BottomNavigationBar (5 tabs + More sheet) | 12px | 2 |
| 600–900px | Tablet | AppBar + Drawer | 16px | 3 |
| >900px | Desktop | NavigationRail (100px) | 24px | 4 |

### LgBreakpoints Utility

```dart
enum LgDeviceType { mobile, tablet, desktop }

class LgBreakpoints {
  static const double mobile = 600;
  static const double tablet = 900;

  static LgDeviceType deviceType(BuildContext context) {
    final width = MediaQuery.sizeOf(context).width;
    if (width < mobile) return LgDeviceType.mobile;
    if (width < tablet) return LgDeviceType.tablet;
    return LgDeviceType.desktop;
  }

  static bool isMobile(BuildContext context) =>
      deviceType(context) == LgDeviceType.mobile;

  static int metricColumns(BuildContext context) => switch (deviceType(context)) {
    LgDeviceType.mobile => 2,
    LgDeviceType.tablet => 3,
    LgDeviceType.desktop => 4,
  };
}
```

### LgResponsive Widget

```dart
class LgResponsive extends StatelessWidget {
  final Widget mobile;
  final Widget? tablet;
  final Widget desktop;

  // Renders mobile, tablet (fallback to desktop), or desktop
  // based on LgBreakpoints.deviceType(context)
}
```

---

## Navigation Changes

### Mobile (<600px)

**BottomNavigationBar** with 5 items:

| Index | Icon | Label |
|-------|------|-------|
| 0 | dashboard | Dashboard |
| 1 | subscriptions | Subscriptions |
| 2 | store | Stores |
| 3 | analytics | Analytics |
| 4 | more_horiz | More |

**"More" tap** → `showModalBottomSheet` with remaining items:
- Transactions, Events, Webhooks, Risk, Earnings, Apps, API Keys, AI Insights, Settings

Each item calls `navigationShell.goBranch(correctIndex)`.

### Tablet (600–900px)

Existing Drawer behavior (AppBar hamburger → 13 drawer items).

### Desktop (>900px)

Existing NavigationRail (100px sidebar with 13 icons).

---

## Widget Adaptations

### LgMetricCard

**Before:** Always wrapped in `Expanded` — forced parent to be `Row`.
**After:** Returns `LgCard` directly — can be placed in `Row`, `Wrap`, or `Column`.

### LgMetricGrid

New widget that arranges N metric cards in a responsive `Wrap`:
- Mobile: 2 columns
- Tablet: 3 columns
- Desktop: 4 columns

### LgDataTable

**Mobile:** Wrapped in `SingleChildScrollView(scrollDirection: Axis.horizontal)`.
**Desktop:** Unchanged (`SizedBox(width: double.infinity)`).

### LgPage

Adaptive padding:
- Mobile: 12px
- Tablet: 16px
- Desktop: 24px (unchanged)

### LgSearchField

**Mobile:** Full-width (removes `SizedBox(width: 300)`).
**Desktop:** Unchanged (300px).

---

## Per-Screen Summary

| Screen | Adaptation |
|--------|-----------|
| Dashboard | KPI Row → LgMetricGrid; Chart Row → LgResponsive(Column/Row); Bottom Row → LgResponsive |
| Subscription List | KPI Row → LgMetricGrid; DataTable auto-scrolls |
| Subscription Detail | Payment/Risk KPI Rows → LgMetricGrid; Payment entries → stacked on mobile |
| Store List | Fixed 350px cards → LayoutBuilder responsive grid |
| Store Detail | Info label widths reduced on mobile |
| Transactions | Summary Row → LgResponsive(Column/Row) |
| Events | KPI Row → LgMetricGrid |
| Webhooks | KPI Row → LgMetricGrid |
| Earnings | KPI Row → LgMetricGrid; Period inner Row → LgResponsive |
| Analytics Revenue | 3-bar Row → LgResponsive(Column/Row) |
| Analytics Forecasting | 2-card Row → LgResponsive |
| Analytics Profit | Summary Row → LgResponsive; Expense table → horizontal scroll |
| Analytics Multi-App | N-card Row → Wrap grid |
| Insights | Side-by-side Row → LgResponsive(Column/Row) |

---

## Diagrams

- **Navigation flow:** [`frontend-flutter/docs/MOBILE_NAVIGATION.puml`](../../frontend-flutter/docs/MOBILE_NAVIGATION.puml)
- **Architecture:** [`frontend-flutter/docs/responsive-architecture.excalidraw`](../../frontend-flutter/docs/responsive-architecture.excalidraw)

---

## Key Files

| File | Purpose |
|------|---------|
| `lib/theme/app_breakpoints.dart` | LgBreakpoints, LgDeviceType, LgResponsive |
| `lib/widgets/lg_metric_grid.dart` | Responsive metric card grid |
| `lib/widgets/lg_metric_card.dart` | Removed Expanded wrapper |
| `lib/widgets/lg_data_table.dart` | Mobile horizontal scroll |
| `lib/widgets/lg_page.dart` | Adaptive padding |
| `lib/widgets/lg_search_field.dart` | Adaptive width |
| `lib/shell/app_shell.dart` | Three-tier navigation (BottomNav/Drawer/Rail) |
| `lib/app.dart` | ScrollBehavior for desktop touch+mouse |
