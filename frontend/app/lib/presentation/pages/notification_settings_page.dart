import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/theme/app_theme.dart';
import '../blocs/notification_preferences/notification_preferences.dart';

/// Notification Settings page for managing notification preferences
class NotificationSettingsPage extends StatelessWidget {
  const NotificationSettingsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey[50],
      appBar: AppBar(
        title: const Text('Notification Settings'),
      ),
      body: BlocConsumer<NotificationPreferencesBloc, NotificationPreferencesState>(
        listener: (context, state) {
          if (state is NotificationPreferencesSaved) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: const Row(
                  children: [
                    Icon(Icons.check_circle, color: Colors.white),
                    SizedBox(width: 8),
                    Text('Settings saved successfully'),
                  ],
                ),
                backgroundColor: AppTheme.success,
                behavior: SnackBarBehavior.floating,
                duration: const Duration(seconds: 2),
              ),
            );
          } else if (state is NotificationPreferencesError) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Row(
                  children: [
                    const Icon(Icons.error_outline, color: Colors.white),
                    const SizedBox(width: 8),
                    Expanded(child: Text(state.message)),
                  ],
                ),
                backgroundColor: AppTheme.danger,
                behavior: SnackBarBehavior.floating,
              ),
            );
          }
        },
        builder: (context, state) {
          if (state is NotificationPreferencesInitial) {
            context.read<NotificationPreferencesBloc>().add(
                  const LoadNotificationPreferencesRequested(),
                );
            return const Center(child: CircularProgressIndicator());
          }

          if (state is NotificationPreferencesLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state is NotificationPreferencesError && state.previousPreferences == null) {
            return _buildErrorState(context, state.message);
          }

          if (state is NotificationPreferencesLoaded) {
            return _buildContent(context, state);
          }

          return const SizedBox.shrink();
        },
      ),
    );
  }

  Widget _buildErrorState(BuildContext context, String message) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.error_outline,
              size: 64,
              color: Colors.red[300],
            ),
            const SizedBox(height: 16),
            Text(
              'Failed to load settings',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 8),
            Text(
              message,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Colors.grey[600],
                  ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: () {
                context.read<NotificationPreferencesBloc>().add(
                      const LoadNotificationPreferencesRequested(),
                    );
              },
              icon: const Icon(Icons.refresh),
              label: const Text('Retry'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildContent(BuildContext context, NotificationPreferencesLoaded state) {
    final preferences = state.preferences;

    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Critical Alerts Section
          _buildSection(
            context,
            title: 'Critical Alerts',
            description:
                'Receive immediate notifications when subscription statuses change to at-risk or churned.',
            child: _buildSwitchTile(
              context,
              title: 'Enable Critical Alerts',
              subtitle: 'Get notified about urgent subscription issues',
              value: preferences.criticalAlertsEnabled,
              onChanged: (value) {
                context.read<NotificationPreferencesBloc>().add(
                      ToggleCriticalAlertsRequested(enabled: value),
                    );
              },
            ),
          ),

          const SizedBox(height: 24),

          // Daily Summary Section
          _buildSection(
            context,
            title: 'Daily Summary',
            description:
                'Receive a daily digest of your revenue metrics and subscription health.',
            child: Column(
              children: [
                _buildSwitchTile(
                  context,
                  title: 'Enable Daily Summary',
                  subtitle: 'Get a daily overview of your metrics',
                  value: preferences.dailySummaryEnabled,
                  onChanged: (value) {
                    context.read<NotificationPreferencesBloc>().add(
                          ToggleDailySummaryRequested(enabled: value),
                        );
                  },
                ),
                const Divider(height: 1),
                _buildTimePicker(
                  context,
                  title: 'Summary Time',
                  subtitle: 'When to receive your daily summary',
                  time: preferences.dailySummaryTime,
                  enabled: preferences.dailySummaryEnabled,
                  onChanged: (time) {
                    context.read<NotificationPreferencesBloc>().add(
                          UpdateDailySummaryTimeRequested(time: time),
                        );
                  },
                ),
              ],
            ),
          ),

          const SizedBox(height: 32),

          // Save Button
          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              onPressed: state.isSaving || !state.hasUnsavedChanges
                  ? null
                  : () {
                      context.read<NotificationPreferencesBloc>().add(
                            const SaveNotificationPreferencesRequested(),
                          );
                    },
              style: ElevatedButton.styleFrom(
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
              child: state.isSaving
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        color: Colors.white,
                      ),
                    )
                  : const Text('Save Changes'),
            ),
          ),

          if (state.hasUnsavedChanges) ...[
            const SizedBox(height: 8),
            Center(
              child: Text(
                'You have unsaved changes',
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: AppTheme.warning,
                    ),
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildSection(
    BuildContext context, {
    required String title,
    required String description,
    required Widget child,
  }) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: Colors.grey[200]!),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                ),
                const SizedBox(height: 4),
                Text(
                  description,
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: Colors.grey[600],
                      ),
                ),
              ],
            ),
          ),
          const Divider(height: 1),
          child,
        ],
      ),
    );
  }

  Widget _buildSwitchTile(
    BuildContext context, {
    required String title,
    required String subtitle,
    required bool value,
    required ValueChanged<bool> onChanged,
  }) {
    return SwitchListTile(
      title: Text(title),
      subtitle: Text(
        subtitle,
        style: Theme.of(context).textTheme.bodySmall?.copyWith(
              color: Colors.grey[600],
            ),
      ),
      value: value,
      onChanged: onChanged,
      activeColor: AppTheme.primary,
    );
  }

  Widget _buildTimePicker(
    BuildContext context, {
    required String title,
    required String subtitle,
    required TimeOfDay time,
    required bool enabled,
    required ValueChanged<TimeOfDay> onChanged,
  }) {
    return ListTile(
      enabled: enabled,
      title: Text(
        title,
        style: TextStyle(
          color: enabled ? null : Colors.grey,
        ),
      ),
      subtitle: Text(
        subtitle,
        style: Theme.of(context).textTheme.bodySmall?.copyWith(
              color: enabled ? Colors.grey[600] : Colors.grey[400],
            ),
      ),
      trailing: TextButton(
        onPressed: enabled
            ? () async {
                final selectedTime = await showTimePicker(
                  context: context,
                  initialTime: time,
                  builder: (context, child) {
                    return Theme(
                      data: Theme.of(context).copyWith(
                        colorScheme: Theme.of(context).colorScheme.copyWith(
                              primary: AppTheme.primary,
                            ),
                      ),
                      child: child!,
                    );
                  },
                );
                if (selectedTime != null) {
                  onChanged(selectedTime);
                }
              }
            : null,
        child: Text(
          _formatTime(time),
          style: TextStyle(
            color: enabled ? AppTheme.primary : Colors.grey,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
    );
  }

  String _formatTime(TimeOfDay time) {
    final hour = time.hourOfPeriod == 0 ? 12 : time.hourOfPeriod;
    final minute = time.minute.toString().padLeft(2, '0');
    final period = time.period == DayPeriod.am ? 'AM' : 'PM';
    return '$hour:$minute $period';
  }
}
