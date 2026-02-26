import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/dashboard_preferences.dart';
import '../blocs/preferences/preferences.dart';

/// Dialog for configuring dashboard preferences
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
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 500, maxHeight: 600),
        child: BlocConsumer<PreferencesBloc, PreferencesState>(
          listener: (context, state) {
            if (state is PreferencesSaved) {
              // Could show a snackbar, but transition is quick
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
              return Padding(
                padding: const EdgeInsets.all(24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.error_outline, size: 48, color: Colors.red[300]),
                    const SizedBox(height: 16),
                    Text('Error: ${state.message}'),
                    const SizedBox(height: 16),
                    ElevatedButton(
                      onPressed: () => context
                          .read<PreferencesBloc>()
                          .add(const LoadPreferencesRequested()),
                      child: const Text('Retry'),
                    ),
                  ],
                ),
              );
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

  Widget _buildContent(BuildContext context, PreferencesLoaded state) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        _buildHeader(context, state),
        Flexible(
          child: SingleChildScrollView(
            padding: const EdgeInsets.symmetric(horizontal: 24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const SizedBox(height: 16),
                _buildPrimaryKpisSection(context, state.preferences),
                const SizedBox(height: 24),
                _buildSecondaryWidgetsSection(context, state.preferences),
                const SizedBox(height: 24),
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
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppTheme.primary.withOpacity(0.1),
        border: Border(
          bottom: BorderSide(color: Colors.grey[300]!),
        ),
      ),
      child: Row(
        children: [
          const Icon(Icons.settings, color: AppTheme.primary),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Dashboard Configuration',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                if (state.hasUnsavedChanges)
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
          IconButton(
            icon: const Icon(Icons.close),
            onPressed: () => Navigator.of(context).pop(),
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
            const Text(
              'Primary KPIs',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(width: 8),
            Text(
              '(${preferences.primaryKpis.length}/4)',
              style: TextStyle(
                color: Colors.grey[600],
                fontSize: 14,
              ),
            ),
          ],
        ),
        const SizedBox(height: 8),
        Text(
          'Drag to reorder. Maximum 4 KPIs.',
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
            return _KpiListTile(
              key: ValueKey(kpi),
              kpi: kpi,
              onRemove: () => context
                  .read<PreferencesBloc>()
                  .add(RemovePrimaryKpiRequested(kpi)),
            );
          },
        ),
        if (availableKpis.isNotEmpty && preferences.primaryKpis.length < 4) ...[
          const SizedBox(height: 12),
          _buildAddKpiDropdown(context, availableKpis),
        ],
      ],
    );
  }

  Widget _buildAddKpiDropdown(
      BuildContext context, List<KpiType> availableKpis) {
    return PopupMenuButton<KpiType>(
      onSelected: (kpi) {
        context.read<PreferencesBloc>().add(AddPrimaryKpiRequested(kpi));
      },
      itemBuilder: (context) => availableKpis
          .map((kpi) => PopupMenuItem(
                value: kpi,
                child: Text(kpi.displayName),
              ))
          .toList(),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
        decoration: BoxDecoration(
          border: Border.all(color: AppTheme.primary),
          borderRadius: BorderRadius.circular(8),
        ),
        child: const Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.add, color: AppTheme.primary, size: 20),
            SizedBox(width: 8),
            Text(
              'Add KPI',
              style: TextStyle(color: AppTheme.primary),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSecondaryWidgetsSection(
      BuildContext context, DashboardPreferences preferences) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          'Secondary Widgets',
          style: TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.bold,
          ),
        ),
        const SizedBox(height: 8),
        Text(
          'Toggle widgets to show or hide them.',
          style: TextStyle(color: Colors.grey[600], fontSize: 13),
        ),
        const SizedBox(height: 12),
        ...SecondaryWidget.values.map((widget) => _SecondaryWidgetTile(
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
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.grey[50],
        border: Border(
          top: BorderSide(color: Colors.grey[300]!),
        ),
      ),
      child: Wrap(
        alignment: WrapAlignment.end,
        spacing: 12,
        runSpacing: 8,
        children: [
          TextButton(
            onPressed: () {
              context
                  .read<PreferencesBloc>()
                  .add(const ResetPreferencesRequested());
            },
            child: const Text('Reset to Default'),
          ),
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: state.isSaving || !state.hasUnsavedChanges
                ? null
                : () {
                    context
                        .read<PreferencesBloc>()
                        .add(const SavePreferencesRequested());
                    Navigator.of(context).pop();
                  },
            child: state.isSaving
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      key: Key('save-progress'),
                    ),
                  )
                : const Text('Save'),
          ),
        ],
      ),
    );
  }
}

class _KpiListTile extends StatelessWidget {
  final KpiType kpi;
  final VoidCallback onRemove;

  const _KpiListTile({
    required super.key,
    required this.kpi,
    required this.onRemove,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.symmetric(vertical: 4),
      child: ListTile(
        leading: const Icon(Icons.drag_handle),
        title: Text(kpi.displayName),
        trailing: IconButton(
          icon: const Icon(Icons.remove_circle_outline),
          color: AppTheme.danger,
          onPressed: onRemove,
        ),
      ),
    );
  }
}

class _SecondaryWidgetTile extends StatelessWidget {
  final SecondaryWidget widget;
  final bool isEnabled;
  final VoidCallback onToggle;

  const _SecondaryWidgetTile({
    required this.widget,
    required this.isEnabled,
    required this.onToggle,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.symmetric(vertical: 4),
      child: SwitchListTile(
        title: Text(widget.displayName),
        value: isEnabled,
        onChanged: (_) => onToggle(),
        activeColor: AppTheme.primary,
      ),
    );
  }
}
