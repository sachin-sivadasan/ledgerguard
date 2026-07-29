import 'package:dio/dio.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../providers/apps_provider.dart';
import '../../providers/mrr_report_provider.dart';
import '../../services/mrr_report_service.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_service_unavailable.dart';
import '../../widgets/lg_table.dart';

class MrrReportScreen extends StatefulWidget {
  const MrrReportScreen({super.key});

  @override
  State<MrrReportScreen> createState() => _MrrReportScreenState();
}

class _MrrReportScreenState extends State<MrrReportScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<MrrReportProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<MrrReportProvider>();
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
          'mrr-${DateTime.now().toIso8601String().split('T').first}.csv';
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
      debugPrint('mrr: CSV export failed: $e');
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
        title: 'Monthly Recurring Revenue',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgEmptyState(
          icon: Icons.show_chart_outlined,
          heading: 'No data yet',
          description: 'Connect your Shopify app to see MRR.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<MrrReportProvider>();

    if (provider.isServiceUnavailable) {
      return LgPage(
        title: 'Monthly Recurring Revenue',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgServiceUnavailable(onRetry: retryLoad),
      );
    }

    if (provider.error != null) {
      return LgPage(
        title: 'Monthly Recurring Revenue',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.report == null) {
      return LgPage(
        title: 'Monthly Recurring Revenue',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final report = provider.report ?? MrrReport.empty();
    final appsList = appsProvider.apps;
    final showAppFilter = appsList.isNotEmpty;
    final currency = report.currency;
    final hasData =
        report.plans.isNotEmpty ||
        report.mrrCents > 0 ||
        report.newMrrCents > 0 ||
        report.churnedMrrCents > 0;

    return LgPage(
      title: 'Monthly Recurring Revenue',
      breadcrumb: 'Reports › Revenue & Billing',
      subtitle:
          'RECURRING revenue normalized to monthly — trend and breakdown by plan',
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
              icon: Icons.show_chart_outlined,
              heading: 'No MRR data yet',
              description:
                  'MRR is your RECURRING revenue normalized to a monthly figure. Once your app has active recurring subscriptions and a completed sync, your MRR, its trend, and the per-plan breakdown will appear here.',
            )
          : Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _TrendCard(report: report, currency: currency),
                const SizedBox(height: LgSpacing.s600),
                _PlansTable(report: report, currency: currency),
              ],
            ),
    );
  }
}

// ─── Hero KPI cards ─────────────────────────────────────────────────
class _HeroRow extends StatelessWidget {
  final MrrReport report;
  final String currency;
  const _HeroRow({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final cards = [
      _KpiCard(
        label: 'MRR',
        value: _money(report.mrrCents, currency),
        color: LgColors.textPrimary,
        // Only show the change delta when there are ≥2 snapshots to compare; with
        // 0/1 the backend returns 0 and a green "+0.0%" would be misleading.
        delta: report.trend.length >= 2
            ? _MomDelta(momChangePct: report.momChangePct)
            : null,
        footnote: 'active recurring revenue (latest snapshot)',
      ),
      _KpiCard(
        label: 'New MRR',
        value: _money(report.newMrrCents, currency),
        color: LgColors.success,
        footnote: 'from new subscriptions in range',
      ),
      _KpiCard(
        label: 'Churned MRR',
        value: '-${_money(report.churnedMrrCents, currency)}',
        color: LgColors.critical,
        footnote: 'from churned subs in range',
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
  final Widget? delta;
  const _KpiCard({
    required this.label,
    required this.value,
    required this.color,
    this.footnote,
    this.delta,
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
          if (delta != null) ...[
            const SizedBox(height: LgSpacing.s100),
            delta!,
          ],
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

/// Signed change delta line vs the start of the selected range, e.g. "▲ +6.2% vs range start".
class _MomDelta extends StatelessWidget {
  final double momChangePct;
  const _MomDelta({required this.momChangePct});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isUp = momChangePct >= 0;
    final color = isUp ? LgColors.success : LgColors.critical;
    final arrow = isUp ? '▲' : '▼';
    final sign = isUp ? '+' : '';
    final pct = '$sign${(momChangePct * 100).toStringAsFixed(1)}%';
    return Text(
      '$arrow $pct vs range start',
      style: theme.textTheme.bodySmall?.copyWith(
        color: color,
        fontWeight: FontWeight.w600,
      ),
    );
  }
}

// ─── Trend card ─────────────────────────────────────────────────────
class _TrendCard extends StatelessWidget {
  final MrrReport report;
  final String currency;
  const _TrendCard({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final trend = report.trend;
    if (trend.isEmpty) {
      return LgCard(
        title: 'MRR — trend',
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

    // Plot MRR in currency units (cents / 100) for readable axis labels.
    final spots = <FlSpot>[
      for (var i = 0; i < trend.length; i++)
        FlSpot(i.toDouble(), trend[i].mrrCents / 100),
    ];
    final maxY = spots.map((s) => s.y).fold<double>(0, (a, b) => a > b ? a : b);

    return LgCard(
      title: 'MRR — trend',
      child: SizedBox(
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
                      _compactMoney(value, currency),
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
                color: LgColors.primary,
                barWidth: 2,
                dotData: const FlDotData(show: false),
                belowBarData: BarAreaData(
                  show: true,
                  color: LgColors.primary.withValues(alpha: 0.10),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

// ─── MRR by plan table ──────────────────────────────────────────────
class _PlansTable extends StatelessWidget {
  final MrrReport report;
  final String currency;
  const _PlansTable({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final plans = [...report.plans]
      ..sort((a, b) => b.mrrCents.compareTo(a.mrrCents));

    final secondary = theme.textTheme.bodySmall?.copyWith(
      color: LgColors.textSecondary,
    );
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('MRR by Plan (ranked by MRR)', style: theme.textTheme.titleMedium),
        const SizedBox(height: LgSpacing.s300),
        LgTable(
          columns: const [
            LgTableColumn('PLAN', flex: 3),
            LgTableColumn('ACTIVE SUBS', flex: 2, numeric: true),
            LgTableColumn('MRR', flex: 2, numeric: true),
            LgTableColumn('% OF TOTAL', flex: 2, numeric: true),
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
                Text(_percent(p.pctOfTotal), style: theme.textTheme.titleSmall),
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

/// Compact currency label for chart axes (e.g. "$1.2K", "$1.2M").
String _compactMoney(double amount, String currency) {
  final format = NumberFormat.compactSimpleCurrency(name: currency);
  return format.format(amount);
}

String _percent(double rate) => '${(rate * 100).toStringAsFixed(1)}%';
