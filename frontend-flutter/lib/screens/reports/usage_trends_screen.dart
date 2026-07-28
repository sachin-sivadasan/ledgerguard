import 'package:dio/dio.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../providers/apps_provider.dart';
import '../../providers/usage_trends_provider.dart';
import '../../services/usage_trends_service.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_service_unavailable.dart';
import '../../widgets/lg_table.dart';

class UsageTrendsScreen extends StatefulWidget {
  const UsageTrendsScreen({super.key});

  @override
  State<UsageTrendsScreen> createState() => _UsageTrendsScreenState();
}

class _UsageTrendsScreenState extends State<UsageTrendsScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<UsageTrendsProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<UsageTrendsProvider>();
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
          'usage-trends-${DateTime.now().toIso8601String().split('T').first}.csv';
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
      // Don't swallow the cause — surface a 503 with a matching
      // service-unavailable message, and log everything else so export
      // failures stay diagnosable.
      debugPrint('usage-trends: CSV export failed: $e');
      if (!mounted) return;
      final isUnavailable = e is DioException && e.response?.statusCode == 503;
      messenger.showSnackBar(
        SnackBar(
          content: Text(isUnavailable
              ? 'Service temporarily unavailable. Please try again shortly.'
              : 'Could not export CSV. Please try again.'),
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
        title: 'Usage Charge Trends',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgEmptyState(
          icon: Icons.show_chart_outlined,
          heading: 'No data yet',
          description: 'Connect your Shopify app to see usage charge trends.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<UsageTrendsProvider>();

    if (provider.isServiceUnavailable) {
      return LgPage(
        title: 'Usage Charge Trends',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgServiceUnavailable(onRetry: retryLoad),
      );
    }

    if (provider.error != null) {
      return LgPage(
        title: 'Usage Charge Trends',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.report == null) {
      return LgPage(
        title: 'Usage Charge Trends',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final report = provider.report ?? UsageTrendsReport.empty();
    final appsList = appsProvider.apps;
    final showAppFilter = appsList.isNotEmpty;
    final currency = report.currency;
    final hasData = report.stores.isNotEmpty ||
        report.usageMrrEquivCents > 0 ||
        report.activeStores > 0 ||
        report.weeklyTrend.isNotEmpty;

    return LgPage(
      title: 'Usage Charge Trends',
      breadcrumb: 'Reports › Revenue & Billing',
      subtitle:
          'Week-over-week USAGE momentum — top usage customers ranked by revenue',
      backAction: () => context.go('/reports'),
      onRefresh: refreshData,
      dateRange: provider.dateRange,
      onDateRangeChanged: provider.setDateRange,
      secondaryActions: [
        LgPageAction(label: 'Export CSV', onPressed: _exportCsv),
      ],
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (showAppFilter) ...[
            PopupMenuButton<String>(
              onSelected: provider.setSelectedApp,
              itemBuilder: (_) => appsList
                  .map((app) =>
                      PopupMenuItem(value: app.id, child: Text(app.name)))
                  .toList(),
              child: Chip(
                label: Text(appsList
                    .firstWhere((a) => a.id == provider.selectedAppId,
                        orElse: () => appsList.first)
                    .name),
              ),
            ),
            const SizedBox(height: LgSpacing.s300),
          ],
          if (!hasData)
            const LgEmptyState(
              icon: Icons.show_chart_outlined,
              heading: 'No usage activity in range',
              description:
                  'USAGE charges are tracked separately from recurring MRR. Once your app records usage-based billing in the selected window, week-over-week momentum and your top usage customers will appear here.',
            )
          else ...[
            _HeroRow(report: report, currency: currency),
            const SizedBox(height: LgSpacing.s600),
            _TrendCard(report: report, currency: currency),
            const SizedBox(height: LgSpacing.s600),
            _StoresTable(report: report, currency: currency),
            const SizedBox(height: LgSpacing.s400),
            Text(
              'USAGE strictly separated from RECURRING — never counted in MRR.',
              style: Theme.of(context)
                  .textTheme
                  .bodySmall
                  ?.copyWith(color: LgColors.textSecondary),
            ),
          ],
        ],
      ),
    );
  }
}

// ─── Hero KPI cards ─────────────────────────────────────────────────
class _HeroRow extends StatelessWidget {
  final UsageTrendsReport report;
  final String currency;
  const _HeroRow({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final cards = <Widget>[
      _KpiCard(
        label: 'Usage MRR-equiv',
        value: _money(report.usageMrrEquivCents, currency),
        color: LgColors.warning,
        footnote: 'total usage in the selected window (≈ monthly at a 30-day range)',
      ),
      _KpiCard(
        label: 'WoW Change',
        valueWidget: _WowDelta(pct: report.wowChangePct),
        footnote: 'latest vs prior weekly bucket (latest week may be partial)',
      ),
      _KpiCard(
        label: 'Active Usage Stores',
        value: '${report.activeStores}',
        color: LgColors.textPrimary,
        footnote: 'stores with usage this window',
      ),
    ];

    if (LgBreakpoints.isMobile(context)) {
      return Column(
        children: [
          for (var i = 0; i < cards.length; i++) ...[
            if (i > 0) const SizedBox(height: LgSpacing.s300),
            cards[i],
          ],
        ],
      );
    }
    return Row(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        for (var i = 0; i < cards.length; i++) ...[
          if (i > 0) const SizedBox(width: LgSpacing.s300),
          Expanded(child: cards[i]),
        ],
      ],
    );
  }
}

class _KpiCard extends StatelessWidget {
  final String label;
  final String? value;
  final Widget? valueWidget;
  final Color? color;
  final String? footnote;
  const _KpiCard({
    required this.label,
    this.value,
    this.valueWidget,
    this.color,
    this.footnote,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return LgCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label,
              style: theme.textTheme.bodySmall
                  ?.copyWith(color: LgColors.textSecondary)),
          const SizedBox(height: LgSpacing.s200),
          if (valueWidget != null)
            valueWidget!
          else
            Text(value ?? '',
                style: theme.textTheme.headlineMedium
                    ?.copyWith(color: color, fontWeight: FontWeight.w700)),
          if (footnote != null) ...[
            const SizedBox(height: LgSpacing.s100),
            Text(footnote!,
                style: theme.textTheme.bodySmall
                    ?.copyWith(color: LgColors.textSecondary)),
          ],
        ],
      ),
    );
  }
}

/// Signed WoW figure rendered as a hero-sized ▲/▼ percentage.
/// Green when the growth ratio is ≥ 0, red when negative.
class _WowDelta extends StatelessWidget {
  final double pct;
  const _WowDelta({required this.pct});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isUp = pct >= 0;
    final color = isUp ? LgColors.success : LgColors.critical;
    final arrow = isUp ? '▲' : '▼';
    final sign = isUp ? '+' : '';
    final text = '$arrow $sign${(pct * 100).toStringAsFixed(1)}%';
    return Text(
      text,
      style: theme.textTheme.headlineMedium
          ?.copyWith(color: color, fontWeight: FontWeight.w700),
    );
  }
}

// ─── Trend card ─────────────────────────────────────────────────────
class _TrendCard extends StatelessWidget {
  final UsageTrendsReport report;
  final String currency;
  const _TrendCard({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final trend = report.weeklyTrend;
    if (trend.isEmpty) {
      return LgCard(
        title: 'Usage — weekly trend',
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: LgSpacing.s600),
          child: Center(
            child: Text(
              'Trend data not yet available for this range.',
              style: Theme.of(context)
                  .textTheme
                  .bodySmall
                  ?.copyWith(color: LgColors.textSecondary),
            ),
          ),
        ),
      );
    }

    // Plot usage revenue in currency units (dollars) for readable axis labels.
    final spots = <FlSpot>[
      for (var i = 0; i < trend.length; i++)
        FlSpot(i.toDouble(), trend[i].usageCents / 100),
    ];
    final maxY = spots.map((s) => s.y).fold<double>(0, (a, b) => a > b ? a : b);
    final format = NumberFormat.compactSimpleCurrency(name: currency);

    return LgCard(
      title: 'Usage — weekly trend',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
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
                      sideTitles: SideTitles(showTitles: false)),
                  rightTitles: const AxisTitles(
                      sideTitles: SideTitles(showTitles: false)),
                  leftTitles: AxisTitles(
                    sideTitles: SideTitles(
                      showTitles: true,
                      reservedSize: 44,
                      getTitlesWidget: (value, meta) {
                        if (value == meta.min || value == meta.max) {
                          return const SizedBox.shrink();
                        }
                        return Text(format.format(value),
                            style: const TextStyle(fontSize: 10));
                      },
                    ),
                  ),
                  bottomTitles: AxisTitles(
                    sideTitles: SideTitles(
                      showTitles: true,
                      interval:
                          (trend.length / 4).clamp(1, trend.length).toDouble(),
                      getTitlesWidget: (value, meta) {
                        final i = value.toInt();
                        if (i < 0 || i >= trend.length) {
                          return const SizedBox.shrink();
                        }
                        return Padding(
                          padding: const EdgeInsets.only(top: 6),
                          child: Text(
                            DateFormat('MMM d').format(trend[i].weekStart),
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
                    color: LgColors.warning,
                    barWidth: 2,
                    dotData: const FlDotData(show: false),
                    belowBarData: BarAreaData(
                      show: true,
                      color: LgColors.warning.withValues(alpha: 0.10),
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: LgSpacing.s200),
          Text(
            'Weekly USAGE revenue — the WoW figure above compares the last two weekly buckets.',
            style: Theme.of(context)
                .textTheme
                .bodySmall
                ?.copyWith(color: LgColors.textSecondary),
          ),
        ],
      ),
    );
  }
}

// ─── Top usage customers table ──────────────────────────────────────
class _StoresTable extends StatelessWidget {
  final UsageTrendsReport report;
  final String currency;
  const _StoresTable({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    // Defensively re-sort by usage revenue descending.
    final stores = [...report.stores]
      ..sort((a, b) => b.usageCents.compareTo(a.usageCents));

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Top Usage Customers (ranked by usage revenue)',
            style: theme.textTheme.titleMedium),
        const SizedBox(height: LgSpacing.s300),
        LgTable(
          columns: const [
            LgTableColumn('STORE', flex: 3),
            LgTableColumn('USAGE \$', flex: 2, numeric: true),
            LgTableColumn('TREND', flex: 2, numeric: true),
          ],
          rows: [
            for (final s in stores)
              [
                _StoreCell(store: s),
                Text(_money(s.usageCents, currency),
                    style: theme.textTheme.titleSmall),
                _TrendDelta(pct: s.wowPct),
              ],
          ],
        ),
      ],
    );
  }
}

/// Two-line store cell: name over domain (domain hidden when absent).
class _StoreCell extends StatelessWidget {
  final UsageTrendStore store;
  const _StoreCell({required this.store});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final name = store.shopName.isNotEmpty
        ? store.shopName
        : (store.domain.isNotEmpty ? store.domain : '—');

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(name, style: theme.textTheme.titleSmall),
        if (store.domain.isNotEmpty) ...[
          const SizedBox(height: LgSpacing.s100),
          Text(
            store.domain,
            style: theme.textTheme.bodySmall
                ?.copyWith(color: LgColors.textSecondary),
          ),
        ],
      ],
    );
  }
}

/// Per-store signed WoW trend rendered as a ▲/▼ percentage.
/// Green when ≥ 0, red when negative.
class _TrendDelta extends StatelessWidget {
  final double pct;
  const _TrendDelta({required this.pct});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isUp = pct >= 0;
    final color = isUp ? LgColors.success : LgColors.critical;
    final arrow = isUp ? '▲' : '▼';
    final sign = isUp ? '+' : '';
    return Text(
      '$arrow $sign${(pct * 100).toStringAsFixed(1)}%',
      style: theme.textTheme.titleSmall
          ?.copyWith(color: color, fontWeight: FontWeight.w600),
    );
  }
}

String _money(int cents, String currency) {
  final format = NumberFormat.simpleCurrency(name: currency);
  return format.format(cents / 100);
}
