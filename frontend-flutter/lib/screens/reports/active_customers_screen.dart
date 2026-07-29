import 'package:dio/dio.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../providers/apps_provider.dart';
import '../../providers/active_customers_provider.dart';
import '../../services/active_customers_service.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_service_unavailable.dart';
import '../../widgets/lg_table.dart';

class ActiveCustomersScreen extends StatefulWidget {
  const ActiveCustomersScreen({super.key});

  @override
  State<ActiveCustomersScreen> createState() => _ActiveCustomersScreenState();
}

class _ActiveCustomersScreenState extends State<ActiveCustomersScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<ActiveCustomersProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<ActiveCustomersProvider>();
    final appId = provider.selectedAppId;
    if (appId == null) return;

    try {
      final bytes = await provider.fetchCsvBytes();
      if (bytes == null || bytes.isEmpty) {
        if (!mounted) return;
        messenger.showSnackBar(
          const SnackBar(content: Text('CSV export returned no data.')),
        );
        return;
      }
      final filename =
          'active-customers-${DateTime.now().toIso8601String().split('T').first}.csv';
      final ok = downloadBytes(bytes, filename, 'text/csv');
      if (!mounted) return;
      if (!ok) {
        messenger.showSnackBar(
          const SnackBar(
            content: Text('CSV export is only available on the web app.'),
          ),
        );
      }
    } catch (e) {
      // Don't swallow the cause — surface a 503 with the same "service
      // unavailable" copy the report body uses, and log everything else so
      // export failures stay diagnosable.
      debugPrint('active-customers: CSV export failed: $e');
      if (!mounted) return;
      final isUnavailable = e is DioException && e.response?.statusCode == 503;
      messenger.showSnackBar(
        SnackBar(
          content: Text(
            isUnavailable
                ? 'Service temporarily unavailable. Please try again shortly.'
                : 'Could not export CSV. Please try again.',
          ),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final appsProvider = context.watch<AppsProvider>();
    final hasApps = appsProvider.apps.isNotEmpty;

    if (!hasApps) {
      return LgPage(
        title: 'Active Customers',
        breadcrumb: 'Reports › Customers',
        backAction: () => context.go('/reports'),
        child: LgEmptyState(
          icon: Icons.people_outline,
          heading: 'No data yet',
          description: 'Connect your Shopify app to see active customers.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<ActiveCustomersProvider>();

    if (provider.isServiceUnavailable) {
      return LgPage(
        title: 'Active Customers',
        breadcrumb: 'Reports › Customers',
        backAction: () => context.go('/reports'),
        child: LgServiceUnavailable(onRetry: retryLoad),
      );
    }

    if (provider.error != null) {
      return LgPage(
        title: 'Active Customers',
        breadcrumb: 'Reports › Customers',
        backAction: () => context.go('/reports'),
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.report == null) {
      return LgPage(
        title: 'Active Customers',
        breadcrumb: 'Reports › Customers',
        backAction: () => context.go('/reports'),
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final report = provider.report ?? ActiveCustomersReport.empty();
    final appsList = appsProvider.apps;
    final showAppFilter = appsList.isNotEmpty;
    final currency = report.currency;
    final hasData =
        report.plans.isNotEmpty ||
        report.activeCustomers > 0 ||
        report.newCount > 0 ||
        report.churnedCount > 0;

    return LgPage(
      title: 'Active Customers',
      breadcrumb: 'Reports › Customers',
      subtitle: 'Active paying customers over time',
      backAction: () => context.go('/reports'),
      onRefresh: refreshData,
      dateRange: provider.dateRange,
      onDateRangeChanged: provider.setDateRange,
      secondaryActions: [
        LgPageAction(label: 'Export CSV', onPressed: _exportCsv),
      ],
      // Normal page scroll: app-selector + KPI hero at top, then chart/table.
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (showAppFilter) ...[
            PopupMenuButton<String>(
              onSelected: provider.setSelectedApp,
              itemBuilder: (_) => appsList
                  .map(
                    (app) =>
                        PopupMenuItem(value: app.id, child: Text(app.name)),
                  )
                  .toList(),
              child: Chip(
                label: Text(
                  appsList
                      .firstWhere(
                        (a) => a.id == provider.selectedAppId,
                        orElse: () => appsList.first,
                      )
                      .name,
                ),
              ),
            ),
            if (hasData) const SizedBox(height: LgSpacing.s300),
          ],
          if (hasData) _HeroRow(report: report),
          const SizedBox(height: LgSpacing.s600),
          if (!hasData)
            const LgEmptyState(
              icon: Icons.people_outline,
              heading: 'No active customers yet',
              description:
                  'Active Customers is the count of your paying subscriptions (Safe plus at-risk). Once your app has active subscriptions and a completed sync, the count, its trend, and the per-plan breakdown will appear here.',
            )
          else ...[
            _TrendCard(report: report),
            const SizedBox(height: LgSpacing.s600),
            _PlansTable(report: report, currency: currency),
          ],
        ],
      ),
    );
  }
}

// ─── Hero KPI cards ─────────────────────────────────────────────────
class _HeroRow extends StatelessWidget {
  final ActiveCustomersReport report;
  const _HeroRow({required this.report});

  @override
  Widget build(BuildContext context) {
    final netUp = report.netChange >= 0;
    final netSign = netUp ? '+' : '';
    final cards = [
      _KpiCard(
        label: 'Active Customers',
        value: '${report.activeCustomers}',
        color: LgColors.success,
        footnote: 'paying subscriptions (Safe + at-risk)',
      ),
      _KpiCard(
        label: 'New',
        value: '+${report.newCount}',
        color: LgColors.textPrimary,
        footnote: 'newly-activated subscriptions',
      ),
      _KpiCard(
        label: 'Net Change',
        value: '$netSign${report.netChange}',
        color: netUp ? LgColors.success : LgColors.critical,
        footnote: 'new − churned',
      ),
    ];

    if (LgBreakpoints.isMobile(context)) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          for (var i = 0; i < cards.length; i++) ...[
            if (i > 0) const SizedBox(height: LgSpacing.s300),
            cards[i],
          ],
        ],
      );
    }
    return IntrinsicHeight(
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          for (var i = 0; i < cards.length; i++) ...[
            if (i > 0) const SizedBox(width: LgSpacing.s300),
            Expanded(child: cards[i]),
          ],
        ],
      ),
    );
  }
}

class _KpiCard extends StatelessWidget {
  final String label;
  final String value;
  final Color color;
  final String? footnote;
  const _KpiCard({
    required this.label,
    required this.value,
    required this.color,
    this.footnote,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return LgCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            label,
            style: theme.textTheme.bodySmall?.copyWith(
              color: LgColors.textSecondary,
            ),
          ),
          const SizedBox(height: LgSpacing.s200),
          Text(
            value,
            style: theme.textTheme.headlineMedium?.copyWith(
              color: color,
              fontWeight: FontWeight.w700,
            ),
          ),
          if (footnote != null) ...[
            const SizedBox(height: LgSpacing.s100),
            Text(
              footnote!,
              style: theme.textTheme.bodySmall?.copyWith(
                color: LgColors.textSecondary,
              ),
            ),
          ],
        ],
      ),
    );
  }
}

// ─── Trend card ─────────────────────────────────────────────────────
class _TrendCard extends StatelessWidget {
  final ActiveCustomersReport report;
  const _TrendCard({required this.report});

  @override
  Widget build(BuildContext context) {
    final trend = report.trend;
    if (trend.isEmpty) {
      return LgCard(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const _TrendHeader(),
            const SizedBox(height: LgSpacing.s300),
            Padding(
              padding: const EdgeInsets.symmetric(vertical: LgSpacing.s600),
              child: Center(
                child: Text(
                  'Trend data not yet available for this range.',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: LgColors.textSecondary,
                      ),
                ),
              ),
            ),
          ],
        ),
      );
    }

    // Plot the integer active-customer count directly.
    final spots = <FlSpot>[
      for (var i = 0; i < trend.length; i++)
        FlSpot(i.toDouble(), trend[i].activeCustomers.toDouble()),
    ];
    final maxY = spots.map((s) => s.y).fold<double>(0, (a, b) => a > b ? a : b);

    return LgCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const _TrendHeader(),
          const SizedBox(height: LgSpacing.s300),
          SizedBox(
            height: 200,
            child: LineChart(
          LineChartData(
            minY: 0,
            maxY: maxY == 0 ? 1 : maxY * 1.2,
            gridData: const FlGridData(show: false),
            borderData: FlBorderData(show: false),
            titlesData: FlTitlesData(
              topTitles: const AxisTitles(
                sideTitles: SideTitles(showTitles: false),
              ),
              rightTitles: const AxisTitles(
                sideTitles: SideTitles(showTitles: false),
              ),
              leftTitles: AxisTitles(
                sideTitles: SideTitles(
                  showTitles: true,
                  reservedSize: 52,
                  getTitlesWidget: (value, meta) {
                    if (value == meta.min || value == meta.max) {
                      return const SizedBox.shrink();
                    }
                    return Text(
                      value.toInt().toString(),
                      style: const TextStyle(fontSize: 10),
                    );
                  },
                ),
              ),
              bottomTitles: AxisTitles(
                sideTitles: SideTitles(
                  showTitles: true,
                  interval: (trend.length / 4)
                      .clamp(1, trend.length)
                      .toDouble(),
                  getTitlesWidget: (value, meta) {
                    final i = value.toInt();
                    if (i < 0 || i >= trend.length) {
                      return const SizedBox.shrink();
                    }
                    return Padding(
                      padding: const EdgeInsets.only(top: 6),
                      child: Text(
                        DateFormat('MMM d').format(trend[i].date),
                        style: const TextStyle(fontSize: 10),
                      ),
                    );
                  },
                ),
              ),
            ),
            lineBarsData: [
              LineChartBarData(
                spots: spots,
                isCurved: true,
                color: LgColors.success,
                barWidth: 2,
                dotData: const FlDotData(show: false),
                belowBarData: BarAreaData(
                  show: true,
                  color: LgColors.success.withValues(alpha: 0.10),
                ),
              ),
            ],
          ),
            ),
          ),
        ],
      ),
    );
  }
}

/// Title + subtitle header for the trend card.
class _TrendHeader extends StatelessWidget {
  const _TrendHeader();

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Active customers — trend', style: theme.textTheme.titleSmall),
        const SizedBox(height: LgSpacing.s100),
        Text(
          'from daily snapshots',
          style: theme.textTheme.bodySmall?.copyWith(
            color: LgColors.textSecondary,
          ),
        ),
      ],
    );
  }
}

// ─── Active by plan table ───────────────────────────────────────────
class _PlansTable extends StatelessWidget {
  final ActiveCustomersReport report;
  final String currency;
  const _PlansTable({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    // Plans are already sorted by MRR desc server-side.
    final plans = report.plans;

    final secondary = theme.textTheme.bodySmall?.copyWith(
      color: LgColors.textSecondary,
    );
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Active by plan', style: theme.textTheme.titleMedium),
        const SizedBox(height: LgSpacing.s300),
        LgTable(
          columns: const [
            LgTableColumn('PLAN', flex: 3),
            LgTableColumn('ACTIVE', flex: 2, numeric: true),
            LgTableColumn('MRR', flex: 2, numeric: true),
            LgTableColumn('% OF ACTIVE', flex: 2, numeric: true),
          ],
          rows: [
            for (final p in plans)
              [
                Text(
                  p.planName.isNotEmpty ? p.planName : '—',
                  style: theme.textTheme.titleSmall,
                ),
                Text('${p.activeSubs}', style: secondary),
                Text(
                  _money(p.mrrCents, currency),
                  style: theme.textTheme.titleSmall,
                ),
                Text(_percent(p.pctOfActive), style: theme.textTheme.titleSmall),
              ],
          ],
        ),
      ],
    );
  }
}

String _money(int cents, String currency) {
  final format = NumberFormat.simpleCurrency(name: currency);
  return format.format(cents / 100);
}

String _percent(double rate) => '${(rate * 100).toStringAsFixed(1)}%';
