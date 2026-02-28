import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/dashboard_preferences.dart';
import '../blocs/preferences/preferences.dart';

/// Premium dialog for configuring dashboard preferences
class DashboardConfigDialog extends StatelessWidget {
  const DashboardConfigDialog({super.key});

  static Future<void> show(BuildContext context) {
    return showDialog(
      context: context,
      builder: (context) => const DashboardConfigDialog(),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 520, maxHeight: 680),
        child: BlocConsumer<PreferencesBloc, PreferencesState>(
          listener: (context, state) {
            if (state is PreferencesSaved) {
              Navigator.of(context).pop();
            }
          },
          builder: (context, state) {
            if (state is PreferencesLoading) {
              return const Padding(
                padding: EdgeInsets.all(48),
                child: Center(child: CircularProgressIndicator()),
              );
            }

            if (state is PreferencesError) {
              return _buildErrorState(context, state);
            }

            if (state is PreferencesLoaded) {
              return _buildContent(context, state);
            }

            return const SizedBox.shrink();
          },
        ),
      ),
    );
  }

  Widget _buildErrorState(BuildContext context, PreferencesError state) {
    return Padding(
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
            'Error: ${state.message}',
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
    );
  }

  Widget _buildContent(BuildContext context, PreferencesLoaded state) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        _buildHeader(context, state),
        Flexible(
          child: SingleChildScrollView(
            padding: const EdgeInsets.fromLTRB(24, 8, 24, 0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildPrimaryKpisSection(context, state.preferences),
                const SizedBox(height: 28),
                _buildWidgetsSection(context, state.preferences),
                const SizedBox(height: 16),
              ],
            ),
          ),
        ),
        _buildFooter(context, state),
      ],
    );
  }

  Widget _buildHeader(BuildContext context, PreferencesLoaded state) {
    return Container(
      padding: const EdgeInsets.fromLTRB(24, 20, 16, 16),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [AppTheme.primary.withOpacity(0.1), Colors.transparent],
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
        ),
        borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              color: AppTheme.primary.withOpacity(0.1),
              borderRadius: BorderRadius.circular(10),
            ),
            child: const Icon(Icons.dashboard_customize, color: AppTheme.primary),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Dashboard Configuration',
                  style: TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                if (state.hasUnsavedChanges)
                  Padding(
                    padding: const EdgeInsets.only(top: 4),
                    child: Row(
                      children: [
                        Container(
                          width: 6,
                          height: 6,
                          decoration: const BoxDecoration(
                            color: AppTheme.warning,
                            shape: BoxShape.circle,
                          ),
                        ),
                        const SizedBox(width: 6),
                        const Text(
                          'Unsaved changes',
                          style: TextStyle(
                            fontSize: 12,
                            color: AppTheme.warning,
                          ),
                        ),
                      ],
                    ),
                  ),
              ],
            ),
          ),
          IconButton(
            icon: const Icon(Icons.close),
            onPressed: () => Navigator.of(context).pop(),
            style: IconButton.styleFrom(
              backgroundColor: Colors.grey[100],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPrimaryKpisSection(
      BuildContext context, DashboardPreferences preferences) {
    final availableKpis = KpiType.values
        .where((k) => !preferences.primaryKpis.contains(k))
        .toList();

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            const Icon(Icons.star, size: 18, color: AppTheme.primary),
            const SizedBox(width: 8),
            Text(
              'Primary KPIs',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const Spacer(),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(
                color: AppTheme.primary.withOpacity(0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Text(
                '(${preferences.primaryKpis.length}/4)',
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
          style: TextStyle(color: Colors.grey[600], fontSize: 13),
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
            return _PremiumKpiTile(
              key: ValueKey(kpi),
              kpi: kpi,
              index: index,
              onRemove: () => context
                  .read<PreferencesBloc>()
                  .add(RemovePrimaryKpiRequested(kpi)),
            );
          },
        ),
        if (availableKpis.isNotEmpty && preferences.primaryKpis.length < 4) ...[
          const SizedBox(height: 12),
          _buildAddKpiButton(context, availableKpis),
        ],
      ],
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
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          border: Border.all(color: AppTheme.primary, width: 1.5),
          borderRadius: BorderRadius.circular(10),
          color: AppTheme.primary.withOpacity(0.05),
        ),
        child: const Row(
          mainAxisSize: MainAxisSize.min,
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

  Widget _buildWidgetsSection(
      BuildContext context, DashboardPreferences preferences) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            const Icon(Icons.widgets, size: 18, color: AppTheme.secondary),
            const SizedBox(width: 8),
            Text(
              'Secondary Widgets',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
          ],
        ),
        const SizedBox(height: 4),
        Text(
          'Toggle widgets to show or hide them on your dashboard.',
          style: TextStyle(color: Colors.grey[600], fontSize: 13),
        ),
        const SizedBox(height: 12),
        ...SecondaryWidget.values.map((widget) => _PremiumWidgetTile(
              widget: widget,
              isEnabled: preferences.isSecondaryWidgetEnabled(widget),
              onToggle: () => context
                  .read<PreferencesBloc>()
                  .add(ToggleSecondaryWidgetRequested(widget)),
            )),
      ],
    );
  }

  Widget _buildFooter(BuildContext context, PreferencesLoaded state) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.grey[50],
        borderRadius: const BorderRadius.vertical(bottom: Radius.circular(16)),
        border: Border(
          top: BorderSide(color: Colors.grey[200]!),
        ),
      ),
      child: Row(
        children: [
          TextButton(
            onPressed: () {
              context
                  .read<PreferencesBloc>()
                  .add(const ResetPreferencesRequested());
            },
            style: TextButton.styleFrom(foregroundColor: Colors.grey[700]),
            child: const Text('Reset to Default'),
          ),
          const Spacer(),
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Cancel'),
          ),
          const SizedBox(width: 12),
          ElevatedButton(
            onPressed: state.isSaving || !state.hasUnsavedChanges
                ? null
                : () {
                    context
                        .read<PreferencesBloc>()
                        .add(const SavePreferencesRequested());
                  },
            child: state.isSaving
                ? const SizedBox(
                    key: Key('save-progress'),
                    width: 18,
                    height: 18,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text('Save'),
          ),
        ],
      ),
    );
  }
}

class _PremiumKpiTile extends StatelessWidget {
  final KpiType kpi;
  final int index;
  final VoidCallback onRemove;

  const _PremiumKpiTile({
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
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: Colors.grey[200]!),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
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
          icon: const Icon(Icons.remove_circle_outline),
          color: Colors.grey[400],
          onPressed: onRemove,
        ),
        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      ),
    );
  }
}

class _PremiumWidgetTile extends StatelessWidget {
  final SecondaryWidget widget;
  final bool isEnabled;
  final VoidCallback onToggle;

  const _PremiumWidgetTile({
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
          color: isEnabled ? AppTheme.primary.withOpacity(0.3) : Colors.grey[200]!,
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
