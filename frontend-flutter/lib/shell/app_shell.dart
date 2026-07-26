import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../core/navigation/navigation_refresh_notifier.dart';
import '../providers/organization_provider.dart';
import '../theme/app_breakpoints.dart';
import '../theme/app_colors.dart';
import '../widgets/lg_service_unavailable.dart';
import '../widgets/org_switcher.dart';

// Nav items, indexed by branch index (order must match the router branches).
const _destinations = <_NavItem>[
  _NavItem(icon: Icons.dashboard_outlined, selectedIcon: Icons.dashboard, label: 'Dashboard'),
  _NavItem(icon: Icons.subscriptions_outlined, selectedIcon: Icons.subscriptions, label: 'Subscriptions'),
  _NavItem(icon: Icons.store_outlined, selectedIcon: Icons.store, label: 'Stores'),
  _NavItem(icon: Icons.receipt_long_outlined, selectedIcon: Icons.receipt_long, label: 'Transactions'),
  _NavItem(icon: Icons.event_outlined, selectedIcon: Icons.event, label: 'Events'),
  _NavItem(icon: Icons.webhook_outlined, selectedIcon: Icons.webhook, label: 'Webhooks'),
  _NavItem(icon: Icons.warning_amber_outlined, selectedIcon: Icons.warning_amber, label: 'Risk'),
  _NavItem(icon: Icons.analytics_outlined, selectedIcon: Icons.analytics, label: 'Analytics'),
  _NavItem(icon: Icons.account_balance_wallet_outlined, selectedIcon: Icons.account_balance_wallet, label: 'Earnings'),
  _NavItem(icon: Icons.apps_outlined, selectedIcon: Icons.apps, label: 'Apps'),
  _NavItem(icon: Icons.vpn_key_outlined, selectedIcon: Icons.vpn_key, label: 'API Keys'),
  _NavItem(icon: Icons.auto_awesome_outlined, selectedIcon: Icons.auto_awesome, label: 'AI Insights'),
  _NavItem(icon: Icons.settings_outlined, selectedIcon: Icons.settings, label: 'Settings'),
  _NavItem(icon: Icons.assessment_outlined, selectedIcon: Icons.assessment, label: 'Reports'),
];

// Desktop sidebar grouping (values are branch indices into _destinations).
const _navGroups = <_NavGroup>[
  _NavGroup('CORE', [0, 1, 2, 3, 4, 5, 6]),
  _NavGroup('ANALYTICS', [7, 8, 11, 13]),
  _NavGroup('ADMIN', [9, 10, 12]),
];

// Bottom nav shows these 4 + "More" (index 0-3 map to branch 0,1,2,7)
const _bottomNavBranches = [0, 1, 2, 7]; // Dashboard, Subs, Stores, Analytics
const _moreItems = [3, 4, 5, 6, 8, 9, 10, 11, 12, 13]; // remaining branches

class AppShell extends StatefulWidget {
  final StatefulNavigationShell navigationShell;

  const AppShell({super.key, required this.navigationShell});

  @override
  State<AppShell> createState() => _AppShellState();
}

class _AppShellState extends State<AppShell> {
  static const _railCollapsedKey = 'rail_collapsed';

  bool _railCollapsed = false;

  StatefulNavigationShell get navigationShell => widget.navigationShell;

  @override
  void initState() {
    super.initState();
    SharedPreferences.getInstance().then((p) {
      final saved = p.getBool(_railCollapsedKey);
      if (saved != null && mounted) {
        setState(() => _railCollapsed = saved);
      }
    });
  }

  void _toggleRail() {
    setState(() => _railCollapsed = !_railCollapsed);
    SharedPreferences.getInstance()
        .then((p) => p.setBool(_railCollapsedKey, _railCollapsed));
  }

  @override
  Widget build(BuildContext context) {
    final orgProvider = context.watch<OrganizationProvider>();
    final deviceType = LgBreakpoints.deviceType(context);

    // When service is unavailable, show the unavailable UI in place of
    // every screen's content, while keeping the navigation chrome.
    final Widget content = orgProvider.isServiceUnavailable
        ? LgServiceUnavailable(
            onRetry: () => orgProvider.loadOrganizations(),
          )
        : navigationShell;

    return switch (deviceType) {
      LgDeviceType.mobile => _buildMobileScaffold(context, content),
      LgDeviceType.tablet => _buildTabletScaffold(context, content),
      LgDeviceType.desktop => _buildDesktopScaffold(context, content),
    };
  }

  // ─── Mobile: BottomNavigationBar ──────────────────────────────────
  Widget _buildMobileScaffold(BuildContext context, Widget content) {
    final currentIndex = navigationShell.currentIndex;
    // Map current branch to bottom nav index
    int bottomIndex;
    final branchIdx = _bottomNavBranches.indexOf(currentIndex);
    if (branchIdx >= 0) {
      bottomIndex = branchIdx;
    } else if (_moreItems.contains(currentIndex)) {
      bottomIndex = 4; // "More" tab
    } else {
      bottomIndex = 0;
    }

    return Scaffold(
      body: SafeArea(child: content),
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: bottomIndex,
        type: BottomNavigationBarType.fixed,
        selectedItemColor: LgColors.primary,
        unselectedItemColor: LgColors.textSecondary,
        selectedFontSize: 11,
        unselectedFontSize: 11,
        onTap: (index) {
          if (index < 4) {
            navigationShell.goBranch(_bottomNavBranches[index]);
            context.read<NavigationRefreshNotifier>().triggerRefreshCheck();
          } else {
            _showMoreSheet(context);
          }
        },
        items: [
          BottomNavigationBarItem(
            icon: Icon(_destinations[0].icon),
            activeIcon: Icon(_destinations[0].selectedIcon),
            label: 'Dashboard',
          ),
          BottomNavigationBarItem(
            icon: Icon(_destinations[1].icon),
            activeIcon: Icon(_destinations[1].selectedIcon),
            label: 'Subs',
          ),
          BottomNavigationBarItem(
            icon: Icon(_destinations[2].icon),
            activeIcon: Icon(_destinations[2].selectedIcon),
            label: 'Stores',
          ),
          BottomNavigationBarItem(
            icon: Icon(_destinations[7].icon),
            activeIcon: Icon(_destinations[7].selectedIcon),
            label: 'Analytics',
          ),
          BottomNavigationBarItem(
            icon: Icon(
              Icons.more_horiz,
              color: _moreItems.contains(currentIndex) ? LgColors.primary : null,
            ),
            label: 'More',
          ),
        ],
      ),
    );
  }

  void _showMoreSheet(BuildContext context) {
    showModalBottomSheet(
      context: context,
      builder: (ctx) {
        return SafeArea(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: _moreItems.map((branchIndex) {
              final dest = _destinations[branchIndex];
              final isSelected = navigationShell.currentIndex == branchIndex;
              return ListTile(
                leading: Icon(
                  isSelected ? dest.selectedIcon : dest.icon,
                  color: isSelected ? LgColors.primary : null,
                ),
                title: Text(dest.label),
                selected: isSelected,
                onTap: () {
                  Navigator.pop(ctx);
                  navigationShell.goBranch(branchIndex);
                  context.read<NavigationRefreshNotifier>().triggerRefreshCheck();
                },
              );
            }).toList(),
          ),
        );
      },
    );
  }

  // ─── Tablet: AppBar + Drawer ──────────────────────────────────────
  Widget _buildTabletScaffold(BuildContext context, Widget content) {
    return Scaffold(
      drawer: _buildDrawer(context),
      appBar: AppBar(
        title: const Text('LedgerGuard'),
        backgroundColor: LgColors.surface,
      ),
      body: SafeArea(child: content),
    );
  }

  // ─── Desktop: NavigationRail ──────────────────────────────────────
  Widget _buildDesktopScaffold(BuildContext context, Widget content) {
    return Scaffold(
      body: Row(
        children: [
          _buildRail(context),
          const VerticalDivider(width: 1),
          Expanded(child: content),
        ],
      ),
    );
  }

  Widget _buildRail(BuildContext context) {
    final collapsed = _railCollapsed;
    return AnimatedContainer(
      duration: const Duration(milliseconds: 160),
      width: collapsed ? 72 : 220,
      color: LgColors.surface,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // Header: logo + collapse toggle
          Padding(
            padding: const EdgeInsets.fromLTRB(12, 12, 8, 8),
            child: Row(
              mainAxisAlignment: collapsed
                  ? MainAxisAlignment.center
                  : MainAxisAlignment.spaceBetween,
              children: [
                if (!collapsed) Image.asset('assets/images/logo.jpeg', height: 36),
                IconButton(
                  tooltip: collapsed ? 'Expand sidebar' : 'Collapse sidebar',
                  visualDensity: VisualDensity.compact,
                  icon: Icon(collapsed ? Icons.menu : Icons.menu_open, size: 20),
                  onPressed: _toggleRail,
                ),
              ],
            ),
          ),
          if (!collapsed)
            const Padding(
              padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              child: OrgSwitcher(),
            ),
          Expanded(
            child: SingleChildScrollView(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  for (final group in _navGroups) ...[
                    if (collapsed)
                      const Padding(
                        padding: EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                        child: Divider(height: 1),
                      )
                    else
                      Padding(
                        padding: const EdgeInsets.fromLTRB(20, 16, 16, 6),
                        child: Text(
                          group.title,
                          style: TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.w600,
                            letterSpacing: 0.6,
                            color: group.branchIndices
                                    .contains(navigationShell.currentIndex)
                                ? LgColors.primary
                                : LgColors.textSecondary,
                          ),
                        ),
                      ),
                    ...group.branchIndices.map((i) => _buildRailItem(context, i)),
                  ],
                  const SizedBox(height: 12),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildRailItem(BuildContext context, int branchIndex) {
    final theme = Theme.of(context);
    final collapsed = _railCollapsed;
    final d = _destinations[branchIndex];
    final selected = navigationShell.currentIndex == branchIndex;
    final color = selected ? LgColors.primary : theme.colorScheme.onSurface;

    final Widget row = Container(
      height: 44,
      padding: EdgeInsets.symmetric(horizontal: collapsed ? 0 : 16),
      decoration: BoxDecoration(
        color: selected ? LgColors.primary.withValues(alpha: 0.08) : null,
        border: selected
            ? const Border(left: BorderSide(color: LgColors.primary, width: 3))
            : null,
      ),
      child: Row(
        mainAxisAlignment:
            collapsed ? MainAxisAlignment.center : MainAxisAlignment.start,
        children: [
          Icon(selected ? d.selectedIcon : d.icon, size: 22, color: color),
          if (!collapsed) ...[
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                d.label,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: selected ? FontWeight.w600 : FontWeight.w400,
                  color: color,
                ),
              ),
            ),
          ],
        ],
      ),
    );

    return InkWell(
      onTap: () {
        navigationShell.goBranch(branchIndex);
        context.read<NavigationRefreshNotifier>().triggerRefreshCheck();
      },
      child: collapsed ? Tooltip(message: d.label, child: row) : row,
    );
  }

  Drawer _buildDrawer(BuildContext context) {
    return Drawer(
      child: ListView(
        padding: EdgeInsets.zero,
        children: [
          DrawerHeader(
            decoration: const BoxDecoration(color: LgColors.primary),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                Image.asset('assets/images/logo.jpeg', height: 48),
                const SizedBox(height: 8),
                const Text(
                  'LedgerGuard',
                  style: TextStyle(color: Colors.white, fontSize: 20, fontWeight: FontWeight.w600),
                ),
              ],
            ),
          ),
          const Padding(
            padding: EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            child: OrgSwitcher(),
          ),
          ..._destinations.asMap().entries.map((e) => ListTile(
                leading: Icon(
                  navigationShell.currentIndex == e.key ? e.value.selectedIcon : e.value.icon,
                  color: navigationShell.currentIndex == e.key ? LgColors.primary : null,
                ),
                title: Text(e.value.label),
                selected: navigationShell.currentIndex == e.key,
                onTap: () {
                  Navigator.pop(context);
                  navigationShell.goBranch(e.key);
                  context.read<NavigationRefreshNotifier>().triggerRefreshCheck();
                },
              )),
        ],
      ),
    );
  }
}

class _NavItem {
  final IconData icon;
  final IconData selectedIcon;
  final String label;

  const _NavItem({required this.icon, required this.selectedIcon, required this.label});
}

class _NavGroup {
  final String title;
  final List<int> branchIndices;

  const _NavGroup(this.title, this.branchIndices);
}
