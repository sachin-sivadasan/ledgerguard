import 'package:dio/dio.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../providers/apps_provider.dart';
import '../../providers/net_new_subs_provider.dart';
import '../../services/net_new_subs_service.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_service_unavailable.dart';
import '../../widgets/lg_table.dart';

class NetNewSubsScreen extends StatefulWidget {
  const NetNewSubsScreen({super.key});

  @override
  State<NetNewSubsScreen> createState() => _NetNewSubsScreenState();
}

class _NetNewSubsScreenState extends State<NetNewSubsScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<NetNewSubsProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<NetNewSubsProvider>();
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
          'net-new-subscriptions-${DateTime.now().toIso8601String().split('T').first}.csv';
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
      debugPrint('net-new-subs: CSV export failed: $e');
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
        title: 'Net-New Subscriptions',
        breadcrumb: 'Reports › Growth',
        backAction: () => context.go('/reports'),
        child: LgEmptyState(
          icon: Icons.group_add_outlined,
          heading: 'No data yet',
          description: 'Connect your Shopify app to see subscription growth.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<NetNewSubsProvider>();

    if (provider.isServiceUnavailable) {
      return LgPage(
        title: 'Net-New Subscriptions',
        breadcrumb: 'Reports › Growth',
        backAction: () => context.go('/reports'),
        child: LgServiceUnavailable(onRetry: retryLoad),
      );
    }

    if (provider.error != null) {
      return LgPage(
        title: 'Net-New Subscriptions',
        breadcrumb: 'Reports › Growth',
        backAction: () => context.go('/reports'),
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.report == null) {
      return LgPage(
        title: 'Net-New Subscriptions',
        breadcrumb: 'Reports › Growth',
        backAction: () => context.go('/reports'),
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final report = provider.report ?? NetNewSubsReport.empty();
    final appsList = appsProvider.apps;
    final showAppFilter = appsList.isNotEmpty;
    final hasData =
        report.newStores.isNotEmpty ||
        report.trend.isNotEmpty ||
        report.newSubs > 0 ||
        report.churned > 0;

    return LgPage(
      title: 'Net-New Subscriptions',
      breadcrumb: 'Reports › Growth',
      subtitle:
          'New vs churned subscriptions — net subscriber growth over the period',
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
          if (hasData) _HeroRow(report: report),
        ],
      ),
      child: !hasData
          ? const LgEmptyState(
              icon: Icons.group_add_outlined,
              heading: 'No new subscriptions yet',
              description:
                  'Once merchants start subscribing to your app, net-new growth and the recent new subscriptions will appear here.',
            )
          : Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _TrendCard(report: report),
                const SizedBox(height: LgSpacing.s600),
                _NewSubsTable(report: report),
                const SizedBox(height: LgSpacing.s400),
                Text(
                  'New = subscriptions started; churned = cancelled/churned. Net = new − churned.',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: LgColors.textSecondary,
                  ),
                ),
              ],
            ),
    );
  }
}

// ─── Hero KPI cards ─────────────────────────────────────────────────
class _HeroRow extends StatelessWidget {
  final NetNewSubsReport report;
  const _HeroRow({required this.report});

  @override
  Widget build(BuildContext context) {
    final netUp = report.net >= 0;
    final cards = <Widget>[
      _KpiCard(
        label: 'New Subs',
        value: '${report.newSubs}',
        color: LgColors.success,
        footnote: 'subscriptions started this period',
      ),
      _KpiCard(
        label: 'Churned',
        value: '${report.churned}',
        color: LgColors.critical,
        footnote: 'subscriptions cancelled / churned',
      ),
      _KpiCard(
        label: 'Net',
        value: '${netUp ? '+' : ''}${report.net}',
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
  final Color? color;
  final String? footnote;
  const _KpiCard({
    required this.label,
    required this.value,
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

// ─── Net-new bar trend ──────────────────────────────────────────────
class _TrendCard extends StatelessWidget {
  final NetNewSubsReport report;
  const _TrendCard({required this.report});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final trend = report.trend;
    if (trend.isEmpty) {
      return LgCard(
        title: 'Net-new subscriptions — trend',
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: LgSpacing.s600),
          child: Center(
            child: Text(
              'Trend data not yet available for this range.',
              style: theme.textTheme.bodySmall?.copyWith(
                color: LgColors.textSecondary,
              ),
            ),
          ),
        ),
      );
    }

    // Bars show net per day: green when net ≥ 0, red when negative.
    final nets = trend.map((t) => t.net.toDouble()).toList();
    final maxNet = nets.fold<double>(0, (a, b) => b > a ? b : a);
    final minNet = nets.fold<double>(0, (a, b) => b < a ? b : a);

    return LgCard(
      title: 'Net-new subscriptions — trend',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [_LegendDot(color: LgColors.success, label: 'Net-new')],
          ),
          const SizedBox(height: LgSpacing.s200),
          SizedBox(
            height: 200,
            child: BarChart(
              BarChartData(
                maxY: maxNet == 0 && minNet == 0 ? 1 : maxNet * 1.2,
                minY: minNet < 0 ? minNet * 1.2 : 0,
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
                      reservedSize: 28,
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
                barGroups: [
                  for (var i = 0; i < trend.length; i++)
                    BarChartGroupData(
                      x: i,
                      barRods: [
                        BarChartRodData(
                          toY: trend[i].net.toDouble(),
                          width: 10,
                          color: trend[i].net >= 0
                              ? LgColors.success
                              : LgColors.critical,
                          borderRadius: BorderRadius.circular(2),
                        ),
                      ],
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

// ─── Recent new subscriptions table ─────────────────────────────────
class _NewSubsTable extends StatelessWidget {
  final NetNewSubsReport report;
  const _NewSubsTable({required this.report});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    // The table is capped server-side; newSubs is the true in-window total.
    final total = report.newSubs;
    final shown = report.newStores.length;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Recent new subscriptions', style: theme.textTheme.titleMedium),
        if (total > shown) ...[
          const SizedBox(height: LgSpacing.s100),
          Text(
            'Showing the latest $shown of $total new subscriptions',
            style: theme.textTheme.bodySmall?.copyWith(
              color: LgColors.textSecondary,
            ),
          ),
        ],
        const SizedBox(height: LgSpacing.s300),
        LgTable(
          columns: const [
            LgTableColumn('STORE', flex: 4),
            LgTableColumn('PLAN', flex: 2),
            LgTableColumn('MRR', flex: 2, numeric: true),
            LgTableColumn('STARTED', flex: 2, numeric: true),
          ],
          rows: [
            for (final s in report.newStores)
              [
                Text(
                  s.domain.isNotEmpty
                      ? s.domain
                      : (s.shopName.isNotEmpty ? s.shopName : '—'),
                  style: theme.textTheme.titleSmall,
                ),
                Text(
                  s.planName.isNotEmpty ? s.planName : '—',
                  style: theme.textTheme.bodyMedium,
                ),
                Text(
                  _money(s.mrrCents, report.currency),
                  style: theme.textTheme.titleSmall?.copyWith(
                    color: LgColors.success,
                  ),
                ),
                Text(
                  s.startedDate != null
                      ? DateFormat('MMM d, yyyy').format(s.startedDate!)
                      : (s.started.isNotEmpty ? s.started : '—'),
                  style: theme.textTheme.bodySmall?.copyWith(
                    color: LgColors.textSecondary,
                  ),
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
