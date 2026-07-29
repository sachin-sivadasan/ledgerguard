import 'package:dio/dio.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../providers/apps_provider.dart';
import '../../providers/retention_provider.dart';
import '../../services/retention_service.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_service_unavailable.dart';
import '../../widgets/lg_table.dart';

class RetentionScreen extends StatefulWidget {
  const RetentionScreen({super.key});

  @override
  State<RetentionScreen> createState() => _RetentionScreenState();
}

class _RetentionScreenState extends State<RetentionScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<RetentionProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<RetentionProvider>();
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
          'retention-${DateTime.now().toIso8601String().split('T').first}.csv';
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
      debugPrint('retention: CSV export failed: $e');
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
        title: 'Retention',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: LgEmptyState(
          icon: Icons.favorite_outline,
          heading: 'No data yet',
          description: 'Connect your Shopify app to see retention.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<RetentionProvider>();

    if (provider.isServiceUnavailable) {
      return LgPage(
        title: 'Retention',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: LgServiceUnavailable(onRetry: retryLoad),
      );
    }

    if (provider.error != null) {
      return LgPage(
        title: 'Retention',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.report == null) {
      return LgPage(
        title: 'Retention',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final report = provider.report ?? RetentionReport.empty();
    final appsList = appsProvider.apps;
    final showAppFilter = appsList.isNotEmpty;
    final currency = report.currency;
    final hasData =
        report.plans.isNotEmpty ||
        report.renewalRate > 0 ||
        report.retainedMrrCents > 0 ||
        report.reactivations > 0;

    return LgPage(
      title: 'Retention',
      subtitle:
          'How well recurring revenue is retained. Renewal rate & trend follow the selected range; retained MRR and the plan table are current-state; reactivations are counted in-range.',
      breadcrumb: 'Reports › Retention & Risk',
      backAction: () => context.go('/reports'),
      onRefresh: refreshData,
      dateRange: provider.dateRange,
      onDateRangeChanged: provider.setDateRange,
      secondaryActions: [
        LgPageAction(label: 'Export CSV', onPressed: _exportCsv),
      ],
      // Fixed: app-selector + KPI hero. Scrollable: the tables/charts.
      pinned: Column(
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
          if (hasData) _HeroRow(report: report, currency: currency),
        ],
      ),
      child: !hasData
          ? const LgEmptyState(
              icon: Icons.favorite_outline,
              heading: 'No active subscriptions yet',
              description:
                  'Retention measures how well recurring revenue carries into the next cycle. Once you have active subscriptions, renewal rate, retained MRR, and reactivations will appear here.',
            )
          : Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _TrendCard(report: report),
                const SizedBox(height: LgSpacing.s600),
                _PlansTable(report: report, currency: currency),
              ],
            ),
    );
  }
}

// ─── Hero KPI cards ─────────────────────────────────────────────────
class _HeroRow extends StatelessWidget {
  final RetentionReport report;
  final String currency;
  const _HeroRow({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final cards = [
      _KpiCard(
        label: 'Renewal Rate',
        value: _percent(report.renewalRate),
        color: LgColors.success,
        footnote: 'from daily snapshots (RenewalSuccessRate)',
      ),
      _KpiCard(
        label: 'Retained MRR',
        value: _money(report.retainedMrrCents, currency),
        color: LgColors.textPrimary,
        footnote:
            'MRR of currently-active (SAFE) subs — current-state, not date-filtered',
      ),
      _KpiCard(
        label: 'Reactivations',
        value: '${report.reactivations}',
        color: LgColors.success,
        footnote: 'previously-churned stores that resumed this period',
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
  final RetentionReport report;
  const _TrendCard({required this.report});

  @override
  Widget build(BuildContext context) {
    final trend = report.trend;
    if (trend.isEmpty) {
      return LgCard(
        title: 'Renewal Success Rate — trend',
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: LgSpacing.s600),
          child: Center(
            child: Text(
              'Trend data not yet available for this range.',
              style: Theme.of(
                context,
              ).textTheme.bodySmall?.copyWith(color: LgColors.textSecondary),
            ),
          ),
        ),
      );
    }

    // Plot renewal rate as a percentage (0..100) for readable axis labels.
    final spots = <FlSpot>[
      for (var i = 0; i < trend.length; i++)
        FlSpot(i.toDouble(), trend[i].renewalRate * 100),
    ];
    final maxY = spots.map((s) => s.y).fold<double>(0, (a, b) => a > b ? a : b);

    return LgCard(
      title: 'Renewal Success Rate — trend',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              _LegendDot(color: LgColors.success, label: 'Renewal rate'),
            ],
          ),
          const SizedBox(height: LgSpacing.s200),
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
                      reservedSize: 44,
                      getTitlesWidget: (value, meta) {
                        if (value == meta.min || value == meta.max) {
                          return const SizedBox.shrink();
                        }
                        return Text(
                          '${value.toStringAsFixed(1)}%',
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

class _LegendDot extends StatelessWidget {
  final Color color;
  final String label;
  const _LegendDot({required this.color, required this.label});

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(width: 10, height: 10, color: color),
        const SizedBox(width: LgSpacing.s100),
        Text(
          label,
          style: Theme.of(
            context,
          ).textTheme.bodySmall?.copyWith(color: LgColors.textSecondary),
        ),
      ],
    );
  }
}

// ─── Retention by plan table ────────────────────────────────────────
class _PlansTable extends StatelessWidget {
  final RetentionReport report;
  final String currency;
  const _PlansTable({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final plans = [...report.plans]
      ..sort((a, b) => b.retainedMrrCents.compareTo(a.retainedMrrCents));

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Retention by Plan (ranked by retained MRR)',
          style: theme.textTheme.titleMedium,
        ),
        const SizedBox(height: LgSpacing.s300),
        LgTable(
          columns: const [
            LgTableColumn('PLAN', flex: 3),
            LgTableColumn('ACTIVE SUBS', flex: 2, numeric: true),
            LgTableColumn('RENEWAL RATE', flex: 2, numeric: true),
            LgTableColumn('RETAINED MRR', flex: 2, numeric: true),
          ],
          rows: [
            for (final p in plans)
              [
                Text(
                  p.planName.isNotEmpty ? p.planName : '—',
                  style: theme.textTheme.titleSmall,
                ),
                Text('${p.activeSubs}', style: theme.textTheme.titleSmall),
                Text(
                  _percent(p.renewalRate),
                  style: theme.textTheme.titleSmall?.copyWith(
                    color: LgColors.success,
                  ),
                ),
                Text(
                  _money(p.retainedMrrCents, currency),
                  style: theme.textTheme.titleSmall,
                ),
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
