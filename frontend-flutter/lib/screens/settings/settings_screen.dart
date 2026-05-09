import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import '../../core/dashboard_registry.dart';
import '../../core/demo_mode_coordinator.dart';
import '../../providers/apps_provider.dart';
import '../../providers/auth_provider.dart';
import '../../providers/dashboard_provider.dart';
import '../../providers/organization_provider.dart';
import '../../providers/settings_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_page.dart';

class SettingsScreen extends StatefulWidget {
  const SettingsScreen({super.key});

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<SettingsProvider>().loadPreferences();
    });
  }

  @override
  Widget build(BuildContext context) {
    final provider = context.watch<SettingsProvider>();
    final appsProvider = context.watch<AppsProvider>();
    final auth = context.watch<AuthProvider>();
    final dp = context.watch<DashboardProvider>();

    return LgPage(
      title: 'Settings',
      subtitle: 'Notifications, sync, and workspace configuration',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Demo mode
          LgCard(
            title: 'Developer',
            child: SwitchListTile(
              title: const Text('Demo Mode'),
              subtitle: const Text('Enable sample data to preview features without a Shopify connection'),
              value: appsProvider.demoMode,
              onChanged: (v) async {
                final coordinator = context.read<DemoModeCoordinator>();
                if (v) {
                  coordinator.setDemoMode(true);
                } else {
                  await coordinator.switchToLiveMode();
                }
              },
            ),
          ),
          const SizedBox(height: LgSpacing.s600),

          // Dashboard customization
          LgCard(
            title: 'Dashboard',
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const ListTile(
                  title: Text('Primary KPIs'),
                  subtitle: Text('Choose up to 4 metrics'),
                ),
                for (final kpi in kAllKpis)
                  CheckboxListTile(
                    title: Text(kpi.label),
                    value: dp.primaryKpis.contains(kpi.id),
                    onChanged: (v) {
                      final current = List<String>.from(dp.primaryKpis);
                      if (v == true) {
                        if (current.length >= 4) {
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(content: Text('Maximum 4 KPIs')),
                          );
                          return;
                        }
                        current.add(kpi.id);
                      } else {
                        current.remove(kpi.id);
                      }
                      dp.saveDashboardPreferences(current, dp.secondaryWidgets);
                    },
                  ),
                const Divider(),
                const ListTile(
                  title: Text('Secondary Widgets'),
                  subtitle: Text('Toggle dashboard widgets'),
                ),
                for (final w in kAllWidgets)
                  CheckboxListTile(
                    title: Text(w.label),
                    value: dp.secondaryWidgets.contains(w.id),
                    onChanged: (v) {
                      final current = List<String>.from(dp.secondaryWidgets);
                      if (v == true) {
                        current.add(w.id);
                      } else {
                        current.remove(w.id);
                      }
                      dp.saveDashboardPreferences(dp.primaryKpis, current);
                    },
                  ),
              ],
            ),
          ),
          const SizedBox(height: LgSpacing.s600),

          // Organization / Team
          LgCard(
            title: 'Organization',
            child: Column(
              children: [
                ListTile(
                  leading: const Icon(Icons.group_outlined),
                  title: const Text('Team Members'),
                  subtitle: Text(
                    context.watch<OrganizationProvider>().currentOrg != null
                        ? '${context.watch<OrganizationProvider>().members.length} members'
                        : 'Manage your team',
                  ),
                  trailing: const Icon(Icons.chevron_right),
                  onTap: () => context.go('/settings/team'),
                ),
                ListTile(
                  leading: const Icon(Icons.history_outlined),
                  title: const Text('Audit Log'),
                  subtitle: const Text('View organization activity'),
                  trailing: const Icon(Icons.chevron_right),
                  onTap: () => context.go('/settings/audit-log'),
                ),
              ],
            ),
          ),
          const SizedBox(height: LgSpacing.s600),

          // Shopify Integration
          LgCard(
            title: 'Shopify Integration',
            child: ListTile(
              leading: const Icon(Icons.link),
              title: const Text('Connect Partner Account'),
              subtitle: const Text('Link your Shopify Partner API token'),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => context.go('/settings/connect-shopify'),
            ),
          ),
          const SizedBox(height: LgSpacing.s600),

          // Notification preferences
          LgCard(
            title: 'Notification Preferences',
            child: Column(
              children: [
                SwitchListTile(
                  title: const Text('Email Alerts'),
                  subtitle: const Text('Receive alerts via email'),
                  value: provider.emailAlerts,
                  onChanged: provider.setEmailAlerts,
                ),
                SwitchListTile(
                  title: const Text('Slack Alerts'),
                  subtitle: const Text('Receive alerts in Slack'),
                  value: provider.slackAlerts,
                  onChanged: provider.setSlackAlerts,
                ),
                const Divider(),
                SwitchListTile(
                  title: const Text('Churn Risk Alerts'),
                  subtitle: const Text('When a subscription enters at-risk state'),
                  value: provider.churnAlerts,
                  onChanged: provider.setChurnAlerts,
                ),
                SwitchListTile(
                  title: const Text('Revenue Alerts'),
                  subtitle: const Text('Daily revenue summaries'),
                  value: provider.revenueAlerts,
                  onChanged: provider.setRevenueAlerts,
                ),
                SwitchListTile(
                  title: const Text('Review Alerts'),
                  subtitle: const Text('New app store reviews'),
                  value: provider.reviewAlerts,
                  onChanged: provider.setReviewAlerts,
                ),
                const Divider(),
                ListTile(
                  title: const Text('Risk Threshold'),
                  subtitle: Text('Alert when ${provider.riskThresholdDays} days overdue'),
                  trailing: SizedBox(
                    width: 200,
                    child: Slider(
                      value: provider.riskThresholdDays.toDouble(),
                      min: 7,
                      max: 90,
                      divisions: 83,
                      label: '${provider.riskThresholdDays} days',
                      onChanged: (v) => provider.setRiskThresholdDays(v.round()),
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: LgSpacing.s600),

          // Sync schedule
          LgCard(
            title: 'Sync Schedule',
            child: Column(
              children: [
                SwitchListTile(
                  title: const Text('Auto Sync'),
                  subtitle: const Text('Automatically sync partner data'),
                  value: provider.autoSync,
                  onChanged: provider.setAutoSync,
                ),
                ListTile(
                  title: const Text('Sync Frequency'),
                  trailing: DropdownButton<String>(
                    value: provider.syncFrequency,
                    items: ['Every hour', 'Every 6 hours', 'Every 12 hours', 'Daily']
                        .map((f) => DropdownMenuItem(value: f, child: Text(f)))
                        .toList(),
                    onChanged: (v) {
                      if (v != null) provider.setSyncFrequency(v);
                    },
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: LgSpacing.s600),

          // Workspace
          LgCard(
            title: 'Workspace',
            child: Column(
              children: [
                ListTile(
                  title: const Text('Workspace Name'),
                  trailing: SizedBox(
                    width: 200,
                    child: TextField(
                      controller: TextEditingController(text: provider.workspaceName),
                      onSubmitted: provider.setWorkspaceName,
                    ),
                  ),
                ),
                ListTile(
                  title: const Text('Currency'),
                  trailing: DropdownButton<String>(
                    value: provider.currency,
                    items: ['USD', 'EUR', 'GBP', 'CAD', 'AUD']
                        .map((c) => DropdownMenuItem(value: c, child: Text(c)))
                        .toList(),
                    onChanged: (v) {
                      if (v != null) provider.setCurrency(v);
                    },
                  ),
                ),
                ListTile(
                  title: const Text('Timezone'),
                  trailing: DropdownButton<String>(
                    value: provider.timezone,
                    items: ['America/New_York', 'America/Chicago', 'America/Los_Angeles', 'Europe/London', 'Asia/Tokyo']
                        .map((tz) => DropdownMenuItem(value: tz, child: Text(tz)))
                        .toList(),
                    onChanged: (v) {
                      if (v != null) provider.setTimezone(v);
                    },
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: LgSpacing.s600),

          // Account
          LgCard(
            title: 'Account',
            child: Column(
              children: [
                ListTile(
                  leading: const Icon(Icons.person_outline),
                  title: Text(auth.user?.name ?? 'User'),
                  subtitle: Text(auth.user?.email ?? ''),
                ),
                const Divider(),
                Padding(
                  padding: const EdgeInsets.symmetric(
                    horizontal: LgSpacing.s400,
                    vertical: LgSpacing.s200,
                  ),
                  child: Align(
                    alignment: Alignment.centerLeft,
                    child: OutlinedButton.icon(
                      onPressed: () => auth.signOut(),
                      icon: const Icon(Icons.logout, size: 18),
                      label: const Text('Sign Out'),
                      style: OutlinedButton.styleFrom(
                        foregroundColor: LgColors.critical,
                        side: const BorderSide(color: LgColors.critical),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
