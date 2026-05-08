import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../providers/apps_provider.dart';
import '../../providers/subscription_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_data_table.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_metric_card.dart';
import '../../widgets/lg_metric_grid.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_risk_badge.dart';
import '../../widgets/lg_search_field.dart';
import '../../widgets/lg_status_badge.dart';

class SubscriptionListScreen extends StatefulWidget {
  const SubscriptionListScreen({super.key});

  @override
  State<SubscriptionListScreen> createState() => _SubscriptionListScreenState();
}

class _SubscriptionListScreenState extends State<SubscriptionListScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<SubscriptionProvider>().loadSubscriptions(appId);
  }

  @override
  Widget build(BuildContext context) {
    final appsProvider = context.watch<AppsProvider>();
    final hasApps = appsProvider.apps.isNotEmpty;

    if (!hasApps) {
      return LgPage(
        title: 'Subscriptions',
        child: LgEmptyState(
          icon: Icons.receipt_long,
          heading: 'No subscriptions yet',
          description:
              'Connect your Shopify app to start tracking subscriptions.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<SubscriptionProvider>();

    if (provider.error != null) {
      return LgPage(
        title: 'Subscriptions',
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    final subs = provider.subscriptions;
    final dateFmt = DateFormat('MMM d, y');
    final appsList = context.watch<AppsProvider>().apps;
    final showAppFilter = appsList.length > 1;

    return LgPage(
      title: 'Subscriptions',
      subtitle: '${subs.length} subscriptions',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (showAppFilter) ...[
            PopupMenuButton<String?>(
              onSelected: provider.setAppFilter,
              itemBuilder: (_) => [
                const PopupMenuItem(value: null, child: Text('All Apps')),
                ...appsList.map((app) => PopupMenuItem(
                      value: app.id,
                      child: Text(app.name),
                    )),
              ],
              child: Chip(
                label: Text(provider.appFilter != null
                    ? appsList.firstWhere((a) => a.id == provider.appFilter, orElse: () => appsList.first).name
                    : 'All Apps'),
                deleteIcon: provider.appFilter != null
                    ? const Icon(Icons.close, size: 14)
                    : null,
                onDeleted: provider.appFilter != null
                    ? () => provider.setAppFilter(null)
                    : null,
              ),
            ),
            const SizedBox(height: LgSpacing.s300),
          ],

          // KPI summary
          LgMetricGrid(
            children: [
              LgMetricCard(label: 'Active', value: '${provider.activeCount}', icon: Icons.check_circle_outline),
              LgMetricCard(label: 'At Risk', value: '${provider.atRiskCount}', icon: Icons.warning_amber),
              LgMetricCard(label: 'Churned', value: '${provider.churnedCount}', icon: Icons.cancel_outlined),
              LgMetricCard(label: 'Avg Price', value: provider.avgPrice, icon: Icons.attach_money),
            ],
          ),
          const SizedBox(height: LgSpacing.s400),

          // Filter bar
          Wrap(
            spacing: LgSpacing.s300,
            runSpacing: LgSpacing.s200,
            crossAxisAlignment: WrapCrossAlignment.center,
            children: [
              LgSearchField(
                value: provider.searchQuery,
                onChanged: provider.setSearch,
                hintText: 'Search by store or plan...',
              ),
              _FilterChip(
                label: 'Status',
                value: provider.statusFilter?.name,
                items: SubscriptionStatus.values.map((s) => s.name).toList(),
                onSelected: (v) => provider.setStatusFilter(
                    v == null ? null : SubscriptionStatus.values.firstWhere((s) => s.name == v)),
              ),
              _FilterChip(
                label: 'Risk',
                value: provider.riskFilter?.name,
                items: RiskState.values.map((r) => r.name).toList(),
                onSelected: (v) => provider.setRiskFilter(
                    v == null ? null : RiskState.values.firstWhere((r) => r.name == v)),
              ),
              _FilterChip(
                label: 'Plan',
                value: provider.planFilter,
                items: const ['Basic', 'Pro', 'Enterprise'],
                onSelected: provider.setPlanFilter,
              ),
              if (provider.searchQuery.isNotEmpty ||
                  provider.statusFilter != null ||
                  provider.riskFilter != null ||
                  provider.planFilter != null ||
                  provider.appFilter != null)
                TextButton(
                  onPressed: provider.clearFilters,
                  child: const Text('Clear all'),
                ),
            ],
          ),
          const SizedBox(height: LgSpacing.s400),

          // Empty filter state
          if (subs.isEmpty)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: LgSpacing.s800),
              child: Center(
                child: Column(
                  children: [
                    Icon(Icons.filter_list_off, size: 40, color: LgColors.textDisabled),
                    const SizedBox(height: LgSpacing.s300),
                    Text('No subscriptions match your filters',
                        style: TextStyle(fontSize: 14, color: LgColors.textSecondary)),
                  ],
                ),
              ),
            ),

          // Table
          if (subs.isNotEmpty)
          Card(
            child: LgDataTable(
              columns: const [
                LgColumn(title: 'Store'),
                LgColumn(title: 'Plan'),
                LgColumn(title: 'Price', numeric: true),
                LgColumn(title: 'Status'),
                LgColumn(title: 'Risk'),
                LgColumn(title: 'Next Charge'),
              ],
              rows: subs.map((sub) {
                return DataRow(
                  onSelectChanged: (_) => context.go('/subscriptions/${sub.id}'),
                  cells: [
                    DataCell(Row(
                      children: [
                        CircleAvatar(
                          radius: 14,
                          backgroundColor: LgColors.primary,
                          child: Text(
                            sub.shopDomain.isNotEmpty ? sub.shopDomain[0].toUpperCase() : '?',
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 12,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ),
                        const SizedBox(width: LgSpacing.s200),
                        Text(sub.shopDomain.replaceAll('.myshopify.com', '')),
                      ],
                    )),
                    DataCell(Text(sub.planName)),
                    DataCell(Text(sub.priceFormatted)),
                    DataCell(LgStatusBadge(status: sub.status)),
                    DataCell(LgRiskBadge(riskState: sub.riskState)),
                    DataCell(Text(dateFmt.format(sub.expectedNextCharge))),
                  ],
                );
              }).toList(),
            ),
          ),
        ],
      ),
    );
  }
}

class _FilterChip extends StatelessWidget {
  final String label;
  final String? value;
  final List<String> items;
  final ValueChanged<String?> onSelected;

  const _FilterChip({
    required this.label,
    this.value,
    required this.items,
    required this.onSelected,
  });

  @override
  Widget build(BuildContext context) {
    return PopupMenuButton<String?>(
      onSelected: onSelected,
      itemBuilder: (ctx) => [
        const PopupMenuItem(value: null, child: Text('All')),
        ...items.map((i) => PopupMenuItem(value: i, child: Text(i))),
      ],
      child: Chip(
        label: Text(value ?? label),
        deleteIcon: value != null ? const Icon(Icons.close, size: 14) : null,
        onDeleted: value != null ? () => onSelected(null) : null,
      ),
    );
  }
}
