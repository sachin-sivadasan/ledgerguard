import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/dashboard_preferences.dart';
import '../blocs/preferences/preferences.dart';

/// Preferences page for managing dashboard and display settings
class PreferencesPage extends StatefulWidget {
  const PreferencesPage({super.key});

  @override
  State<PreferencesPage> createState() => _PreferencesPageState();
}

class _PreferencesPageState extends State<PreferencesPage> {
  @override
  void initState() {
    super.initState();
    context.read<PreferencesBloc>().add(const LoadPreferencesRequested());
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Preferences'),
        actions: [
          BlocBuilder<PreferencesBloc, PreferencesState>(
            builder: (context, state) {
              if (state is PreferencesLoaded && state.hasUnsavedChanges) {
                return TextButton(
                  onPressed: state.isSaving
                      ? null
                      : () {
                          context
                              .read<PreferencesBloc>()
                              .add(const SavePreferencesRequested());
                        },
                  child: state.isSaving
                      ? const SizedBox(
                          width: 16,
                          height: 16,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('Save'),
                );
              }
              return const SizedBox.shrink();
            },
          ),
        ],
      ),
      body: BlocConsumer<PreferencesBloc, PreferencesState>(
        listener: (context, state) {
          if (state is PreferencesSaved) {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text('Preferences saved'),
                backgroundColor: AppTheme.success,
              ),
            );
          } else if (state is PreferencesError) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text('Error: ${state.message}'),
                backgroundColor: AppTheme.danger,
              ),
            );
          }
        },
        builder: (context, state) {
          if (state is PreferencesLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state is PreferencesError && state.preferences == null) {
            return _buildErrorState(context, state);
          }

          if (state is PreferencesLoaded) {
            return _buildContent(context, state);
          }

          // Handle error state with previous preferences
          if (state is PreferencesError && state.preferences != null) {
            return _buildContent(
              context,
              PreferencesLoaded(preferences: state.preferences!),
            );
          }

          return const SizedBox.shrink();
        },
      ),
    );
  }

  Widget _buildErrorState(BuildContext context, PreferencesError state) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.red[50],
                shape: BoxShape.circle,
              ),
              child: Icon(Icons.error_outline, size: 48, color: Colors.red[400]),
            ),
            const SizedBox(height: 16),
            Text(
              'Error loading preferences',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 8),
            Text(
              state.message,
              textAlign: TextAlign.center,
              style: TextStyle(color: Colors.grey[600]),
            ),
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: () => context
                  .read<PreferencesBloc>()
                  .add(const LoadPreferencesRequested()),
              icon: const Icon(Icons.refresh),
              label: const Text('Retry'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildContent(BuildContext context, PreferencesLoaded state) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Unsaved changes indicator
          if (state.hasUnsavedChanges)
            Container(
              margin: const EdgeInsets.only(bottom: 16),
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              decoration: BoxDecoration(
                color: AppTheme.warning.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: AppTheme.warning.withOpacity(0.3)),
              ),
              child: Row(
                children: [
                  Container(
                    width: 8,
                    height: 8,
                    decoration: const BoxDecoration(
                      color: AppTheme.warning,
                      shape: BoxShape.circle,
                    ),
                  ),
                  const SizedBox(width: 8),
                  const Text(
                    'You have unsaved changes',
                    style: TextStyle(
                      color: AppTheme.warning,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ],
              ),
            ),

          // Dashboard Section
          _buildSectionHeader(
            context,
            'Dashboard',
            Icons.dashboard_outlined,
            'Configure your dashboard layout',
          ),
          const SizedBox(height: 12),
          _buildPrimaryKpisCard(context, state.preferences),
          const SizedBox(height: 16),
          _buildSecondaryWidgetsCard(context, state.preferences),

          const SizedBox(height: 32),

          // Reset Section
          Center(
            child: TextButton.icon(
              onPressed: () {
                showDialog(
                  context: context,
                  builder: (context) => AlertDialog(
                    title: const Text('Reset to Defaults?'),
                    content: const Text(
                      'This will reset all dashboard preferences to their default values.',
                    ),
                    actions: [
                      TextButton(
                        onPressed: () => Navigator.pop(context),
                        child: const Text('Cancel'),
                      ),
                      TextButton(
                        onPressed: () {
                          Navigator.pop(context);
                          this.context
                              .read<PreferencesBloc>()
                              .add(const ResetPreferencesRequested());
                        },
                        style: TextButton.styleFrom(
                          foregroundColor: AppTheme.danger,
                        ),
                        child: const Text('Reset'),
                      ),
                    ],
                  ),
                );
              },
              icon: const Icon(Icons.restore),
              label: const Text('Reset to Defaults'),
              style: TextButton.styleFrom(
                foregroundColor: Colors.grey[600],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSectionHeader(
    BuildContext context,
    String title,
    IconData icon,
    String subtitle,
  ) {
    return Row(
      children: [
        Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: AppTheme.primary.withOpacity(0.1),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(icon, size: 20, color: AppTheme.primary),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                title,
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
              ),
              Text(
                subtitle,
                style: TextStyle(
                  color: Colors.grey[600],
                  fontSize: 12,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildPrimaryKpisCard(
      BuildContext context, DashboardPreferences preferences) {
    final availableKpis = KpiType.values
        .where((k) => !preferences.primaryKpis.contains(k))
        .toList();

    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: Colors.grey.shade200),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Icon(Icons.star, size: 18, color: AppTheme.primary),
                const SizedBox(width: 8),
                const Text(
                  'Primary KPIs',
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
                const Spacer(),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: AppTheme.primary.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    '${preferences.primaryKpis.length}/4',
                    style: const TextStyle(
                      color: AppTheme.primary,
                      fontWeight: FontWeight.w600,
                      fontSize: 12,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 4),
            Text(
              'Drag to reorder. These appear at the top of your dashboard.',
              style: TextStyle(color: Colors.grey[600], fontSize: 12),
            ),
            const SizedBox(height: 12),
            ReorderableListView.builder(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              itemCount: preferences.primaryKpis.length,
              onReorder: (oldIndex, newIndex) {
                if (newIndex > oldIndex) newIndex--;
                context.read<PreferencesBloc>().add(ReorderPrimaryKpiRequested(
                      oldIndex: oldIndex,
                      newIndex: newIndex,
                    ));
              },
              itemBuilder: (context, index) {
                final kpi = preferences.primaryKpis[index];
                return _KpiTile(
                  key: ValueKey(kpi),
                  kpi: kpi,
                  index: index,
                  onRemove: () => context
                      .read<PreferencesBloc>()
                      .add(RemovePrimaryKpiRequested(kpi)),
                );
              },
            ),
            if (availableKpis.isNotEmpty &&
                preferences.primaryKpis.length < 4) ...[
              const SizedBox(height: 12),
              _buildAddKpiButton(context, availableKpis),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildAddKpiButton(BuildContext context, List<KpiType> availableKpis) {
    return PopupMenuButton<KpiType>(
      onSelected: (kpi) {
        context.read<PreferencesBloc>().add(AddPrimaryKpiRequested(kpi));
      },
      offset: const Offset(0, 40),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      itemBuilder: (context) => availableKpis
          .map((kpi) => PopupMenuItem(
                value: kpi,
                child: Row(
                  children: [
                    Icon(_getKpiIcon(kpi), size: 18, color: AppTheme.primary),
                    const SizedBox(width: 12),
                    Text(kpi.displayName),
                  ],
                ),
              ))
          .toList(),
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          border: Border.all(color: AppTheme.primary, width: 1.5),
          borderRadius: BorderRadius.circular(10),
          color: AppTheme.primary.withOpacity(0.05),
        ),
        child: const Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.add_circle_outline, color: AppTheme.primary, size: 20),
            SizedBox(width: 8),
            Text(
              'Add KPI',
              style: TextStyle(
                color: AppTheme.primary,
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSecondaryWidgetsCard(
      BuildContext context, DashboardPreferences preferences) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: Colors.grey.shade200),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Icon(Icons.widgets, size: 18, color: AppTheme.secondary),
                const SizedBox(width: 8),
                const Text(
                  'Secondary Widgets',
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
              ],
            ),
            const SizedBox(height: 4),
            Text(
              'Toggle widgets to show or hide them on your dashboard.',
              style: TextStyle(color: Colors.grey[600], fontSize: 12),
            ),
            const SizedBox(height: 12),
            ...SecondaryWidget.values.map((widget) => _WidgetTile(
                  widget: widget,
                  isEnabled: preferences.isSecondaryWidgetEnabled(widget),
                  onToggle: () => context
                      .read<PreferencesBloc>()
                      .add(ToggleSecondaryWidgetRequested(widget)),
                )),
          ],
        ),
      ),
    );
  }

  IconData _getKpiIcon(KpiType kpi) {
    switch (kpi) {
      case KpiType.renewalSuccessRate:
        return Icons.trending_up;
      case KpiType.activeMrr:
        return Icons.attach_money;
      case KpiType.revenueAtRisk:
        return Icons.warning_amber;
      case KpiType.churned:
        return Icons.trending_down;
      case KpiType.usageRevenue:
        return Icons.data_usage;
      case KpiType.totalRevenue:
        return Icons.account_balance_wallet;
    }
  }
}

class _KpiTile extends StatelessWidget {
  final KpiType kpi;
  final int index;
  final VoidCallback onRemove;

  const _KpiTile({
    required super.key,
    required this.kpi,
    required this.index,
    required this.onRemove,
  });

  IconData _getIcon() {
    switch (kpi) {
      case KpiType.renewalSuccessRate:
        return Icons.trending_up;
      case KpiType.activeMrr:
        return Icons.attach_money;
      case KpiType.revenueAtRisk:
        return Icons.warning_amber;
      case KpiType.churned:
        return Icons.trending_down;
      case KpiType.usageRevenue:
        return Icons.data_usage;
      case KpiType.totalRevenue:
        return Icons.account_balance_wallet;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.symmetric(vertical: 4),
      decoration: BoxDecoration(
        color: Colors.grey[50],
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: Colors.grey[200]!),
      ),
      child: ListTile(
        leading: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            ReorderableDragStartListener(
              index: index,
              child: const Icon(Icons.drag_handle, color: Colors.grey),
            ),
            const SizedBox(width: 12),
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: AppTheme.primary.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Icon(_getIcon(), size: 18, color: AppTheme.primary),
            ),
          ],
        ),
        title: Text(
          kpi.displayName,
          style: const TextStyle(fontWeight: FontWeight.w500),
        ),
        trailing: IconButton(
          icon: const Icon(Icons.delete_outline),
          color: Colors.grey[400],
          onPressed: onRemove,
        ),
        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      ),
    );
  }
}

class _WidgetTile extends StatelessWidget {
  final SecondaryWidget widget;
  final bool isEnabled;
  final VoidCallback onToggle;

  const _WidgetTile({
    required this.widget,
    required this.isEnabled,
    required this.onToggle,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.symmetric(vertical: 4),
      decoration: BoxDecoration(
        color: isEnabled ? Colors.white : Colors.grey[50],
        borderRadius: BorderRadius.circular(10),
        border: Border.all(
          color:
              isEnabled ? AppTheme.primary.withOpacity(0.3) : Colors.grey[200]!,
        ),
      ),
      child: ListTile(
        leading: Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: isEnabled
                ? AppTheme.secondary.withOpacity(0.1)
                : Colors.grey[100],
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(
            widget.icon,
            size: 18,
            color: isEnabled ? AppTheme.secondary : Colors.grey[400],
          ),
        ),
        title: Text(
          widget.displayName,
          style: TextStyle(
            fontWeight: FontWeight.w500,
            color: isEnabled ? Colors.black87 : Colors.grey[600],
          ),
        ),
        trailing: Switch(
          value: isEnabled,
          onChanged: (_) => onToggle(),
          activeColor: AppTheme.primary,
        ),
        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 2),
        onTap: onToggle,
      ),
    );
  }
}
