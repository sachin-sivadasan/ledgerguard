import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../providers/apps_provider.dart';
import '../../providers/churn_provider.dart';
import '../../services/churn_service.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_service_unavailable.dart';

class ChurnScreen extends StatefulWidget {
  const ChurnScreen({super.key});

  @override
  State<ChurnScreen> createState() => _ChurnScreenState();
}

class _ChurnScreenState extends State<ChurnScreen> with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<ChurnProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<ChurnProvider>();
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
          'churn-${DateTime.now().toIso8601String().split('T').first}.csv';
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
        title: 'Churn',
        backAction: () => context.go('/reports'),
        child: LgEmptyState(
          icon: Icons.trending_down_outlined,
          heading: 'No data yet',
          description: 'Connect your Shopify app to see churn.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<ChurnProvider>();

    if (provider.isServiceUnavailable) {
      return LgPage(
        title: 'Churn',
        backAction: () => context.go('/reports'),
        child: LgServiceUnavailable(onRetry: retryLoad),
      );
    }

    if (provider.error != null) {
      return LgPage(
        title: 'Churn',
        backAction: () => context.go('/reports'),
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.report == null) {
      return LgPage(
        title: 'Churn',
        backAction: () => context.go('/reports'),
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final report = provider.report ?? ChurnReport.empty();
    final appsList = appsProvider.apps;
    final showAppFilter = appsList.length > 1;
    final currency = report.currency;
    final hasChurn = report.stores.isNotEmpty ||
        report.churnedCount > 0 ||
        report.churnedMrrLostCents > 0;

    return LgPage(
      title: 'Churn',
      subtitle:
          'Subscriptions that churned this period — MRR lost, count, and ranked stores',
      backAction: () => context.go('/reports'),
      onRefresh: refreshData,
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
          if (!hasChurn)
            const LgEmptyState(
              icon: Icons.check_circle_outline,
              heading: 'No churn this period 🎉',
              description:
                  'No subscriptions churned in this range. Retention is holding steady.',
            )
          else ...[
            _HeroRow(report: report, currency: currency),
            const SizedBox(height: LgSpacing.s600),
            _TrendCard(report: report),
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
  final ChurnReport report;
  final String currency;
  const _HeroRow({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final cards = [
      _KpiCard(
        label: 'Churn Rate',
        value: _percent(report.churnRate),
        color: LgColors.critical,
        footnote: 'churned subs ÷ active at period start',
      ),
      _KpiCard(
        label: 'Churned MRR Lost',
        value: '-${_money(report.churnedMrrLostCents, currency)}',
        color: LgColors.critical,
        footnote: 'recurring revenue removed from ledger',
      ),
      _KpiCard(
        label: 'Churned Customers',
        value: '${report.churnedCount}',
        color: LgColors.textPrimary,
        footnote: 'stores reaching risk state CHURNED',
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
          Text(label,
              style: theme.textTheme.bodySmall
                  ?.copyWith(color: LgColors.textSecondary)),
          const SizedBox(height: LgSpacing.s200),
          Text(value,
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

// ─── Trend card ─────────────────────────────────────────────────────
class _TrendCard extends StatelessWidget {
  final ChurnReport report;
  const _TrendCard({required this.report});

  @override
  Widget build(BuildContext context) {
    final trend = report.trend;
    if (trend.isEmpty) {
      return LgCard(
        title: 'Churn Rate — trend',
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

    // Plot churn rate as a percentage (0..100) for readable axis labels.
    final spots = <FlSpot>[
      for (var i = 0; i < trend.length; i++)
        FlSpot(i.toDouble(), trend[i].churnRate * 100),
    ];
    final maxY = spots.map((s) => s.y).fold<double>(0, (a, b) => a > b ? a : b);

    return LgCard(
      title: 'Churn Rate — trend',
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
                    return Text('${value.toStringAsFixed(1)}%',
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
    );
  }
}

// ─── Ranked churned-stores table ────────────────────────────────────
class _StoresTable extends StatelessWidget {
  final ChurnReport report;
  final String currency;
  const _StoresTable({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final stores = [...report.stores]
      ..sort((a, b) => b.mrrLostCents.compareTo(a.mrrLostCents));

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Churned Stores (ranked by MRR lost)',
            style: theme.textTheme.titleMedium),
        const SizedBox(height: LgSpacing.s300),
        ...stores.map((s) => Padding(
              padding: const EdgeInsets.only(bottom: LgSpacing.s200),
              child: _StoreRow(store: s, currency: currency),
            )),
      ],
    );
  }
}

class _StoreRow extends StatelessWidget {
  final ChurnStore store;
  final String currency;
  const _StoreRow({required this.store, required this.currency});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final name = store.shopName.isNotEmpty
        ? store.shopName
        : store.domain.replaceAll('.myshopify.com', '');

    return MouseRegion(
      cursor: SystemMouseCursors.click,
      child: LgCard(
        child: InkWell(
          onTap: () => context.go('/stores/${store.domain}'),
          child: Row(
            children: [
              Expanded(
                flex: 3,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(name, style: theme.textTheme.titleSmall),
                    const SizedBox(height: LgSpacing.s100),
                    Text(
                      store.planName.isNotEmpty
                          ? store.planName
                          : store.domain,
                      style: theme.textTheme.bodySmall
                          ?.copyWith(color: LgColors.textSecondary),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: LgSpacing.s300),
              Expanded(
                flex: 2,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Text('-${_money(store.mrrLostCents, currency)}',
                        style: theme.textTheme.titleSmall
                            ?.copyWith(color: LgColors.critical)),
                    Text('MRR lost',
                        style: theme.textTheme.bodySmall
                            ?.copyWith(color: LgColors.textSecondary)),
                  ],
                ),
              ),
              const SizedBox(width: LgSpacing.s300),
              Expanded(
                flex: 2,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Text(
                      store.churnedDate != null
                          ? DateFormat('MMM d').format(store.churnedDate!)
                          : '—',
                      style: theme.textTheme.bodySmall,
                    ),
                    Text(_tenure(store.tenureDays),
                        style: theme.textTheme.bodySmall
                            ?.copyWith(color: LgColors.textSecondary)),
                  ],
                ),
              ),
              const SizedBox(width: LgSpacing.s200),
              const Icon(Icons.chevron_right, color: LgColors.textSecondary),
            ],
          ),
        ),
      ),
    );
  }
}

String _money(int cents, String currency) {
  final format = NumberFormat.simpleCurrency(name: currency);
  return format.format(cents / 100);
}

String _percent(double rate) => '${(rate * 100).toStringAsFixed(1)}%';

String _tenure(int days) {
  if (days <= 0) return '—';
  final months = days / 30.0;
  return '${months.toStringAsFixed(1)} mo';
}
