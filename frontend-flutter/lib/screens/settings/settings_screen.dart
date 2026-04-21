import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../providers/apps_provider.dart';
import '../../providers/auth_provider.dart';
import '../../providers/settings_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_page.dart';

class SettingsScreen extends StatelessWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final provider = context.watch<SettingsProvider>();
    final appsProvider = context.watch<AppsProvider>();
    final auth = context.watch<AuthProvider>();

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
              subtitle: const Text('Show sample data to preview all features'),
              value: appsProvider.demoMode,
              onChanged: (v) => appsProvider.setDemoMode(v),
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
