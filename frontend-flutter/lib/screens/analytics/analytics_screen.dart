import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../providers/analytics_provider.dart';
import '../../providers/apps_provider.dart';
import '../../providers/organization_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import 'revenue_tab.dart';
import 'forecasting_tab.dart';
import 'profit_tab.dart';
import 'cohort_tab.dart';
import 'multi_app_tab.dart';

class AnalyticsScreen extends StatefulWidget {
  const AnalyticsScreen({super.key});

  @override
  State<AnalyticsScreen> createState() => _AnalyticsScreenState();
}

class _AnalyticsScreenState extends State<AnalyticsScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<AnalyticsProvider>().setSelectedApp(appId);
  }

  @override
  Widget build(BuildContext context) {
    final appsProvider = context.watch<AppsProvider>();
    final hasApps = appsProvider.apps.isNotEmpty;

    if (!hasApps) {
      return LgPage(
        title: 'Analytics',
        child: LgEmptyState(
          icon: Icons.analytics_outlined,
          heading: 'No analytics yet',
          description:
              'Connect your Shopify app to see revenue analytics.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<AnalyticsProvider>();

    if (provider.error != null) {
      return LgPage(
        title: 'Analytics',
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.mrrSnapshots.isEmpty) {
      return LgPage(
        title: 'Analytics',
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final appsList = context.watch<AppsProvider>().apps;
    final showAppFilter = appsList.length > 1;

    return LgPage(
      title: 'Analytics',
      subtitle: 'Revenue analysis and forecasting',
      onRefresh: refreshData,
      scrollable: false,
      child: DefaultTabController(
        length: 5,
        child: Column(
          children: [
            if (showAppFilter) ...[
              Align(
                alignment: Alignment.centerLeft,
                child: Builder(builder: (context) {
                  final canViewAllApps = context.watch<OrganizationProvider>().canViewAllApps;
                  return PopupMenuButton<String?>(
                    onSelected: provider.setSelectedApp,
                    itemBuilder: (_) => [
                      if (canViewAllApps)
                        const PopupMenuItem(
                            value: null, child: Text('All Apps')),
                      ...appsList.map((app) => PopupMenuItem(
                            value: app.id,
                            child: Text(app.name),
                          )),
                    ],
                    child: Chip(
                      label: Text(provider.selectedAppId != null
                          ? appsList
                              .firstWhere(
                                  (a) => a.id == provider.selectedAppId,
                                  orElse: () => appsList.first)
                              .name
                          : canViewAllApps
                              ? 'All Apps'
                              : appsList.first.name),
                      deleteIcon: canViewAllApps && provider.selectedAppId != null
                          ? const Icon(Icons.close, size: 14)
                          : null,
                      onDeleted: canViewAllApps && provider.selectedAppId != null
                          ? () => provider.setSelectedApp(null)
                          : null,
                    ),
                  );
                }),
              ),
              const SizedBox(height: LgSpacing.s300),
            ],
            // Historical snapshot date picker (live mode only)
            if (!provider.demoMode) ...[
              Row(
                children: [
                  ActionChip(
                    avatar: const Icon(Icons.calendar_today, size: 16),
                    label: Text(provider.snapshotDate != null
                        ? DateFormat('MMM d, yyyy')
                            .format(provider.snapshotDate!)
                        : 'Compare to date'),
                    onPressed: () async {
                      final picked = await showDatePicker(
                        context: context,
                        initialDate: provider.snapshotDate ??
                            DateTime.now()
                                .subtract(const Duration(days: 30)),
                        firstDate: DateTime.now()
                            .subtract(const Duration(days: 365)),
                        lastDate: DateTime.now(),
                      );
                      provider.setSnapshotDate(picked);
                    },
                  ),
                  if (provider.snapshotDate != null) ...[
                    const SizedBox(width: LgSpacing.s200),
                    IconButton(
                      icon: const Icon(Icons.close, size: 16),
                      onPressed: () =>
                          provider.setSnapshotDate(null),
                      tooltip: 'Clear comparison',
                      iconSize: 16,
                      constraints: const BoxConstraints(),
                      padding: EdgeInsets.zero,
                    ),
                  ],
                ],
              ),
              if (provider.snapshotMetrics != null) ...[
                const SizedBox(height: LgSpacing.s300),
                _SnapshotComparisonCard(
                  current: provider.mrrSnapshots.isNotEmpty
                      ? provider.mrrSnapshots.last.mrrCents
                      : 0,
                  snapshot: provider.snapshotMetrics!.mrrCents,
                  snapshotDate: provider.snapshotDate!,
                ),
              ],
              const SizedBox(height: LgSpacing.s300),
            ],
            const TabBar(
              isScrollable: true,
              tabAlignment: TabAlignment.start,
              padding: EdgeInsets.zero,
              tabs: [
                Tab(text: 'Revenue'),
                Tab(text: 'Forecasting'),
                Tab(text: 'Profit & Expense'),
                Tab(text: 'Cohorts'),
                Tab(text: 'Multi-App'),
              ],
            ),
            const SizedBox(height: LgSpacing.s400),
            Expanded(
              child: TabBarView(
                children: [
                  RevenueTab(),
                  ForecastingTab(),
                  ProfitTab(),
                  CohortTab(),
                  MultiAppTab(),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _SnapshotComparisonCard extends StatelessWidget {
  final int current;
  final int snapshot;
  final DateTime snapshotDate;

  const _SnapshotComparisonCard({
    required this.current,
    required this.snapshot,
    required this.snapshotDate,
  });

  @override
  Widget build(BuildContext context) {
    final delta = current - snapshot;
    final pctChange =
        snapshot > 0 ? (delta / snapshot * 100) : 0.0;
    final isPositive = delta >= 0;
    final theme = Theme.of(context);

    return LgCard(
      title:
          'MRR vs ${DateFormat('MMM d').format(snapshotDate)}',
      child: Row(
        children: [
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Current',
                  style: theme.textTheme.bodySmall),
              Text('\$${(current / 100).toStringAsFixed(0)}',
                  style: theme.textTheme.titleMedium),
            ],
          ),
          const SizedBox(width: LgSpacing.s600),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(DateFormat('MMM d').format(snapshotDate),
                  style: theme.textTheme.bodySmall),
              Text('\$${(snapshot / 100).toStringAsFixed(0)}',
                  style: theme.textTheme.titleMedium),
            ],
          ),
          const SizedBox(width: LgSpacing.s600),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Change',
                  style: theme.textTheme.bodySmall),
              Row(
                children: [
                  Icon(
                    isPositive
                        ? Icons.trending_up
                        : Icons.trending_down,
                    size: 16,
                    color: isPositive
                        ? LgColors.success
                        : LgColors.critical,
                  ),
                  const SizedBox(width: 4),
                  Text(
                    '${isPositive ? '+' : ''}${pctChange.toStringAsFixed(1)}%',
                    style: TextStyle(
                      fontWeight: FontWeight.w600,
                      color: isPositive
                          ? LgColors.success
                          : LgColors.critical,
                    ),
                  ),
                ],
              ),
            ],
          ),
        ],
      ),
    );
  }
}
