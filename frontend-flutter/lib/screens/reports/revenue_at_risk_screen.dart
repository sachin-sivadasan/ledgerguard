import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../providers/apps_provider.dart';
import '../../providers/revenue_at_risk_provider.dart';
import '../../services/revenue_at_risk_service.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_risk_badge.dart';
import '../../widgets/lg_service_unavailable.dart';
import '../../widgets/lg_table.dart';

class RevenueAtRiskScreen extends StatefulWidget {
  const RevenueAtRiskScreen({super.key});

  @override
  State<RevenueAtRiskScreen> createState() => _RevenueAtRiskScreenState();
}

class _RevenueAtRiskScreenState extends State<RevenueAtRiskScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<RevenueAtRiskProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<RevenueAtRiskProvider>();
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
          'revenue-at-risk-${DateTime.now().toIso8601String().split('T').first}.csv';
      final ok = downloadBytes(bytes, filename, 'text/csv');
      if (!mounted) return;
      if (!ok) {
        messenger.showSnackBar(
          const SnackBar(
            content: Text('CSV export is only available on the web app.'),
          ),
        );
      }
    } catch (_) {
      if (!mounted) return;
      messenger.showSnackBar(
        const SnackBar(content: Text('Could not export CSV. Please try again.')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final appsProvider = context.watch<AppsProvider>();
    final hasApps = appsProvider.apps.isNotEmpty;

    if (!hasApps) {
      return LgPage(
        title: 'Revenue at Risk',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: LgEmptyState(
          icon: Icons.shield_outlined,
          heading: 'No data yet',
          description: 'Connect your Shopify app to see revenue at risk.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<RevenueAtRiskProvider>();

    if (provider.isServiceUnavailable) {
      return LgPage(
        title: 'Revenue at Risk',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: LgServiceUnavailable(onRetry: retryLoad),
      );
    }

    if (provider.error != null) {
      return LgPage(
        title: 'Revenue at Risk',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.report == null) {
      return LgPage(
        title: 'Revenue at Risk',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final report = provider.report ?? RevenueAtRiskReport.empty();
    final appsList = appsProvider.apps;
    final showAppFilter = appsList.isNotEmpty;
    final currency = report.currency;
    final hasRisk = report.stores.isNotEmpty || report.totalAtRiskCents > 0;

    return LgPage(
      title: 'Revenue at Risk',
      subtitle: 'At-risk MRR, recoverable revenue, and ranked stores',
      breadcrumb: 'Reports › Retention & Risk',
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
          if (!hasRisk)
            const LgEmptyState(
              icon: Icons.check_circle_outline,
              heading: 'No revenue at risk 🎉',
              description:
                  'All subscriptions are on track. Nothing needs recovery right now.',
            )
          else ...[
            _HeroRow(report: report, currency: currency),
            const SizedBox(height: LgSpacing.s600),
            _TrendCard(report: report, currency: currency),
            const SizedBox(height: LgSpacing.s600),
            _StoresTable(report: report, currency: currency),
          ],
        ],
      ),
    );
  }
}

// ─── Hero KPI cards ─────────────────────────────────────────────────
class _HeroRow extends StatelessWidget {
  final RevenueAtRiskReport report;
  final String currency;
  const _HeroRow({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final cards = [
      _KpiCard(
        label: 'Total at Risk',
        value: _money(report.totalAtRiskCents, currency),
        color: LgColors.critical,
      ),
      _KpiCard(
        label: 'Recoverable',
        value: _money(report.recoverableCents, currency),
        color: LgColors.success,
      ),
      _KpiCard(
        label: 'At-Risk Stores',
        value: '${report.atRiskStoreCount}',
        color: LgColors.primary,
        footnoteWidget: Wrap(
          spacing: LgSpacing.s100,
          runSpacing: LgSpacing.s100,
          children: [
            _CyclePill(
              color: LgColors.warning,
              label: '${report.oneCycleCount} × 1-cycle',
            ),
            _CyclePill(
              color: LgColors.riskTwoCycle,
              label: '${report.twoCycleCount} × 2-cycle',
            ),
          ],
        ),
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
  final String value;
  final Color color;
  final Widget? footnoteWidget;
  const _KpiCard({
    required this.label,
    required this.value,
    required this.color,
    this.footnoteWidget,
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
          Text(value,
              style: theme.textTheme.headlineMedium
                  ?.copyWith(color: color, fontWeight: FontWeight.w700)),
          if (footnoteWidget != null) ...[
            const SizedBox(height: LgSpacing.s100),
            footnoteWidget!,
          ],
        ],
      ),
    );
  }
}

/// Small rounded cycle-breakdown pill (same style as report status chips).
class _CyclePill extends StatelessWidget {
  final Color color;
  final String label;
  const _CyclePill({required this.color, required this.label});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.14),
        borderRadius: BorderRadius.circular(10),
      ),
      child: Text(label,
          style: TextStyle(
              fontSize: 11, color: color, fontWeight: FontWeight.w600)),
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
        Text(label,
            style: Theme.of(context)
                .textTheme
                .bodySmall
                ?.copyWith(color: LgColors.textSecondary)),
      ],
    );
  }
}

// ─── Trend card ─────────────────────────────────────────────────────
class _TrendCard extends StatelessWidget {
  final RevenueAtRiskReport report;
  final String currency;
  const _TrendCard({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final trend = report.trend;
    if (trend.isEmpty) {
      return LgCard(
        title: 'At-Risk MRR Trend',
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

    final spots = <FlSpot>[
      for (var i = 0; i < trend.length; i++)
        FlSpot(i.toDouble(), trend[i].atRiskCents / 100),
    ];
    final maxY = spots.map((s) => s.y).fold<double>(0, (a, b) => a > b ? a : b);

    return LgCard(
      title: 'At-Risk MRR Trend',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              _LegendDot(color: LgColors.critical, label: 'At risk'),
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
                    return Text('\$${value.toInt()}',
                        style: const TextStyle(fontSize: 10));
                  },
                ),
              ),
              bottomTitles: AxisTitles(
                sideTitles: SideTitles(
                  showTitles: true,
                  interval: (trend.length / 4).clamp(1, trend.length).toDouble(),
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
                color: LgColors.critical,
                barWidth: 2,
                dotData: const FlDotData(show: false),
                belowBarData: BarAreaData(
                  show: true,
                  color: LgColors.critical.withValues(alpha: 0.10),
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

// ─── Ranked stores table ────────────────────────────────────────────
class _StoresTable extends StatelessWidget {
  final RevenueAtRiskReport report;
  final String currency;
  const _StoresTable({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final stores = [...report.stores]
      ..sort((a, b) => b.mrrCents.compareTo(a.mrrCents));

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Ranked Stores (${stores.length})',
            style: theme.textTheme.titleMedium),
        const SizedBox(height: LgSpacing.s300),
        LgTable(
          columns: const [
            LgTableColumn('STORE', flex: 4),
            LgTableColumn('MRR', flex: 2, numeric: true),
            LgTableColumn('RISK', flex: 3),
            LgTableColumn('DAYS LATE', flex: 2, numeric: true),
            LgTableColumn('EXPECTED CHARGE', flex: 2, numeric: true),
            LgTableColumn('RECOVERABLE', flex: 2, numeric: true),
          ],
          rows: [
            for (final s in stores)
              [
                _StoreCell(store: s),
                Text('${_money(s.mrrCents, currency)}/mo',
                    style: theme.textTheme.titleSmall),
                LgRiskBadge(riskState: s.riskState),
                Text('${s.daysLate}d', style: theme.textTheme.bodySmall),
                Text(
                  s.expectedChargeDate != null
                      ? DateFormat('MMM d').format(s.expectedChargeDate!)
                      : '—',
                  style: theme.textTheme.bodySmall,
                ),
                Text(_money(s.recoverableCents, currency),
                    style: theme.textTheme.titleSmall
                        ?.copyWith(color: LgColors.success)),
              ],
          ],
        ),
      ],
    );
  }
}

/// Two-line store cell: shop name over plan — links to store detail.
class _StoreCell extends StatelessWidget {
  final RevenueAtRiskStore store;
  const _StoreCell({required this.store});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final name = store.shopName.isNotEmpty
        ? store.shopName
        : store.domain.replaceAll('.myshopify.com', '');

    return MouseRegion(
      cursor: SystemMouseCursors.click,
      child: InkWell(
        onTap: () => context.go('/stores/${store.domain}'),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(name, style: theme.textTheme.titleSmall),
            if (store.planName.isNotEmpty) ...[
              const SizedBox(height: LgSpacing.s100),
              Text(
                store.planName,
                style: theme.textTheme.bodySmall
                    ?.copyWith(color: LgColors.textSecondary),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

String _money(int cents, String currency) {
  final format = NumberFormat.simpleCurrency(name: currency);
  return format.format(cents / 100);
}
