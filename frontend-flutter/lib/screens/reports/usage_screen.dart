import 'package:dio/dio.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../providers/apps_provider.dart';
import '../../providers/usage_provider.dart';
import '../../services/usage_service.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_service_unavailable.dart';
import '../../widgets/lg_table.dart';

class UsageScreen extends StatefulWidget {
  const UsageScreen({super.key});

  @override
  State<UsageScreen> createState() => _UsageScreenState();
}

class _UsageScreenState extends State<UsageScreen> with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<UsageProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<UsageProvider>();
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
          'usage-${DateTime.now().toIso8601String().split('T').first}.csv';
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
      debugPrint('usage: CSV export failed: $e');
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
        title: 'Usage & One-Time Charges',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgEmptyState(
          icon: Icons.receipt_long_outlined,
          heading: 'No data yet',
          description: 'Connect your Shopify app to see usage charges.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<UsageProvider>();

    if (provider.isServiceUnavailable) {
      return LgPage(
        title: 'Usage & One-Time Charges',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgServiceUnavailable(onRetry: retryLoad),
      );
    }

    if (provider.error != null) {
      return LgPage(
        title: 'Usage & One-Time Charges',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.report == null) {
      return LgPage(
        title: 'Usage & One-Time Charges',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final report = provider.report ?? UsageReport.empty();
    final appsList = appsProvider.apps;
    final showAppFilter = appsList.isNotEmpty;
    final currency = report.currency;
    final hasData =
        report.stores.isNotEmpty ||
        report.usageCents > 0 ||
        report.oneTimeCents > 0 ||
        report.chargesCount > 0;

    return LgPage(
      title: 'Usage & One-Time Charges',
      breadcrumb: 'Reports › Revenue & Billing',
      subtitle:
          'USAGE and ONE-TIME revenue — tracked separately from recurring MRR',
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
            const SizedBox(height: LgSpacing.s300),
          ],
          if (!hasData)
            const LgEmptyState(
              icon: Icons.receipt_long_outlined,
              heading: 'No usage or one-time charges in range',
              description:
                  'USAGE and ONE-TIME charges are tracked separately from recurring MRR. Once your app records usage-based billing, setup fees, or add-ons in the selected window, they will appear here.',
            )
          else ...[
            _HeroRow(report: report, currency: currency),
            const SizedBox(height: LgSpacing.s600),
            _TrendCard(report: report, currency: currency),
            const SizedBox(height: LgSpacing.s600),
            _StoresTable(report: report, currency: currency),
            const SizedBox(height: LgSpacing.s400),
            Text(
              'USAGE never mixed into MRR — Usage Revenue = USAGE charges only.',
              style: Theme.of(
                context,
              ).textTheme.bodySmall?.copyWith(color: LgColors.textSecondary),
            ),
          ],
        ],
      ),
    );
  }
}

// ─── Hero KPI cards ─────────────────────────────────────────────────
class _HeroRow extends StatelessWidget {
  final UsageReport report;
  final String currency;
  const _HeroRow({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final cards = [
      _KpiCard(
        label: 'Usage Revenue',
        value: _money(report.usageCents, currency),
        color: LgColors.warning,
        footnote: 'USAGE charges only',
      ),
      _KpiCard(
        label: 'One-Time',
        value: _money(report.oneTimeCents, currency),
        color: LgColors.success,
        footnote: 'setup fees & add-ons',
      ),
      _KpiCard(
        label: 'Charges',
        value: '${report.chargesCount}',
        color: LgColors.textPrimary,
        footnote: 'count in selected window',
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
  final UsageReport report;
  final String currency;
  const _TrendCard({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final trend = report.trend;
    if (trend.isEmpty) {
      return LgCard(
        title: 'Usage Revenue — trend',
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

    // Plot usage revenue in currency units (dollars) for readable axis labels.
    final spots = <FlSpot>[
      for (var i = 0; i < trend.length; i++)
        FlSpot(i.toDouble(), trend[i].usageCents / 100),
    ];
    final maxY = spots.map((s) => s.y).fold<double>(0, (a, b) => a > b ? a : b);
    final format = NumberFormat.compactSimpleCurrency(name: currency);

    return LgCard(
      title: 'Usage Revenue — trend',
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
                          format.format(value),
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
            'From daily snapshots (rolling 12-month USAGE revenue) — may differ from the windowed total above.',
            style: Theme.of(
              context,
            ).textTheme.bodySmall?.copyWith(color: LgColors.textSecondary),
          ),
        ],
      ),
    );
  }
}

// ─── Top usage stores table ─────────────────────────────────────────
class _StoresTable extends StatelessWidget {
  final UsageReport report;
  final String currency;
  const _StoresTable({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final stores = [...report.stores]
      ..sort((a, b) => b.usageCents.compareTo(a.usageCents));

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Top Usage Stores (ranked by usage revenue)',
          style: theme.textTheme.titleMedium,
        ),
        const SizedBox(height: LgSpacing.s300),
        LgTable(
          columns: const [
            LgTableColumn('STORE', flex: 3),
            LgTableColumn('USAGE \$', flex: 2, numeric: true),
            LgTableColumn('ONE-TIME \$', flex: 2, numeric: true),
            LgTableColumn('CHARGES', flex: 1, numeric: true),
          ],
          rows: [
            for (final s in stores)
              [
                _StoreCell(store: s),
                Text(
                  _money(s.usageCents, currency),
                  style: theme.textTheme.titleSmall,
                ),
                Text(
                  _money(s.oneTimeCents, currency),
                  style: theme.textTheme.titleSmall,
                ),
                Text('${s.chargeCount}', style: theme.textTheme.titleSmall),
              ],
          ],
        ),
      ],
    );
  }
}

/// Two-line store cell: name over domain (domain hidden when absent).
class _StoreCell extends StatelessWidget {
  final UsageStore store;
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
            style: theme.textTheme.bodySmall?.copyWith(
              color: LgColors.textSecondary,
            ),
          ),
        ],
      ],
    );
  }
}

String _money(int cents, String currency) {
  final format = NumberFormat.simpleCurrency(name: currency);
  return format.format(cents / 100);
}
