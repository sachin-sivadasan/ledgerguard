import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../theme/app_colors.dart';

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

  @override
  Widget build(BuildContext context) {
    final wide = MediaQuery.sizeOf(context).width > 800;

    return Scaffold(
      drawer: wide ? null : _buildDrawer(context),
      appBar: wide
          ? null
          : AppBar(
              title: const Text('LedgerGuard'),
              backgroundColor: LgColors.surface,
            ),
      body: Row(
        children: [
          if (wide) _buildRail(context),
          if (wide) const VerticalDivider(width: 1),
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
