import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../theme/app_breakpoints.dart';
import '../theme/app_colors.dart';
import '../widgets/org_switcher.dart';

class AppShell extends StatelessWidget {
  final StatefulNavigationShell navigationShell;

  const AppShell({super.key, required this.navigationShell});

  static const _destinations = <_NavItem>[
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
  ];

  // Bottom nav shows these 4 + "More" (index 0-3 map to branch 0,1,2,7)
  static const _bottomNavBranches = [0, 1, 2, 7]; // Dashboard, Subs, Stores, Analytics
  static const _moreItems = [3, 4, 5, 6, 8, 9, 10, 11, 12]; // remaining branches

  @override
  Widget build(BuildContext context) {
    final deviceType = LgBreakpoints.deviceType(context);

    return switch (deviceType) {
      LgDeviceType.mobile => _buildMobileScaffold(context),
      LgDeviceType.tablet => _buildTabletScaffold(context),
      LgDeviceType.desktop => _buildDesktopScaffold(context),
    };
  }

  // ─── Mobile: BottomNavigationBar ──────────────────────────────────
  Widget _buildMobileScaffold(BuildContext context) {

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
      body: SafeArea(child: navigationShell),
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
                },
              );
            }).toList(),
          ),
        );
      },
    );
  }

  // ─── Tablet: AppBar + Drawer ──────────────────────────────────────
  Widget _buildTabletScaffold(BuildContext context) {
    return Scaffold(
      drawer: _buildDrawer(context),
      appBar: AppBar(
        title: const Text('LedgerGuard'),
        backgroundColor: LgColors.surface,
      ),
      body: SafeArea(child: navigationShell),
    );
  }

  // ─── Desktop: NavigationRail ──────────────────────────────────────
  Widget _buildDesktopScaffold(BuildContext context) {
    return Scaffold(
      body: Row(
        children: [
          _buildRail(context),
          const VerticalDivider(width: 1),
          Expanded(child: navigationShell),
        ],
      ),
    );
  }

  Widget _buildRail(BuildContext context) {
    final theme = Theme.of(context);
    return SizedBox(
      width: 100,
      child: Column(
        children: [
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 12),
            child: Image.asset('assets/images/logo.jpeg', height: 40),
          ),
          const Padding(
            padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
            child: OrgSwitcher(),
          ),
          Expanded(
            child: SingleChildScrollView(
              child: Column(
                children: _destinations.asMap().entries.map((e) {
                  final selected = navigationShell.currentIndex == e.key;
                  final d = e.value;
                  return InkWell(
                    onTap: () => navigationShell.goBranch(e.key),
                    child: Container(
                      width: double.infinity,
                      padding: const EdgeInsets.symmetric(vertical: 8),
                      decoration: selected
                          ? BoxDecoration(
                              color: LgColors.primary.withValues(alpha: 0.08),
                              border: Border(
                                left: BorderSide(
                                    color: LgColors.primary, width: 3),
                              ),
                            )
                          : null,
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            selected ? d.selectedIcon : d.icon,
                            size: 22,
                            color: selected
                                ? LgColors.primary
                                : theme.colorScheme.onSurface,
                          ),
                          const SizedBox(height: 2),
                          Text(
                            d.label,
                            style: TextStyle(
                              fontSize: 11,
                              fontWeight: selected
                                  ? FontWeight.w600
                                  : FontWeight.w400,
                              color: selected
                                  ? LgColors.primary
                                  : theme.colorScheme.onSurface,
                            ),
                            textAlign: TextAlign.center,
                          ),
                        ],
                      ),
                    ),
                  );
                }).toList(),
              ),
            ),
          ),
        ],
      ),
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
