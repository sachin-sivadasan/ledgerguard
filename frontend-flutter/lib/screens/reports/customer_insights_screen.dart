import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../providers/apps_provider.dart';
import '../../providers/customer_insights_provider.dart';
import '../../services/customer_insights_service.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_service_unavailable.dart';

String _money(int cents) =>
    NumberFormat.currency(symbol: '\$', decimalDigits: 0).format(cents / 100);

String _riskLabel(String state) {
  switch (state) {
    case 'SAFE':
      return 'Safe';
    case 'AT_RISK':
    case 'ONE_CYCLE_MISSED':
    case 'TWO_CYCLES_MISSED':
      return 'At risk';
    case 'CHURNED':
      return 'Churned';
    default:
      return state;
  }
}

Color _riskColor(String state) {
  switch (state) {
    case 'SAFE':
      return LgColors.success;
    case 'CHURNED':
      return LgColors.textSecondary;
    default:
      return LgColors.warning;
  }
}

class CustomerInsightsScreen extends StatefulWidget {
  const CustomerInsightsScreen({super.key});

  @override
  State<CustomerInsightsScreen> createState() => _CustomerInsightsScreenState();
}

class _CustomerInsightsScreenState extends State<CustomerInsightsScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<CustomerInsightsProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<CustomerInsightsProvider>();
    if (provider.selectedAppId == null) return;
    try {
      final bytes = await provider.fetchCsvBytes();
      if (bytes == null || bytes.isEmpty) {
        if (!mounted) return;
        messenger.showSnackBar(
            const SnackBar(content: Text('CSV export returned no data.')));
        return;
      }
      final ok = downloadBytes(bytes, 'customer-insights.csv', 'text/csv');
      if (!mounted) return;
      if (!ok) {
        messenger.showSnackBar(const SnackBar(
            content: Text('CSV export is only available on the web app.')));
      }
    } catch (e) {
      debugPrint('customer-insights: CSV export failed: $e');
      if (!mounted) return;
      final unavailable = e is DioException && e.response?.statusCode == 503;
      messenger.showSnackBar(SnackBar(
        content: Text(unavailable
            ? 'Service temporarily unavailable. Please try again shortly.'
            : 'Could not export CSV. Please try again.'),
      ));
    }
  }

  LgPage _shell(Widget child) => LgPage(
        title: 'Customer Insights',
        breadcrumb: 'Reports › Customers',
        backAction: () => context.go('/reports'),
        child: child,
      );

  @override
  Widget build(BuildContext context) {
    final appsProvider = context.watch<AppsProvider>();
    if (appsProvider.apps.isEmpty) {
      return _shell(LgEmptyState(
        icon: Icons.groups_outlined,
        heading: 'No data yet',
        description: 'Connect your Shopify app to segment its customers.',
        actionLabel: 'Go to Apps',
        onAction: () => context.go('/apps'),
      ));
    }

    final provider = context.watch<CustomerInsightsProvider>();
    if (provider.isServiceUnavailable) {
      return _shell(LgServiceUnavailable(onRetry: retryLoad));
    }
    if (provider.error != null) {
      return _shell(LgErrorState(message: provider.error!, onRetry: retryLoad));
    }
    if (provider.isLoading && provider.report == null) {
      return _shell(const Center(child: CircularProgressIndicator()));
    }

    final report = provider.report ?? CustomerInsights.empty();
    final appsList = appsProvider.apps;
    final hasData = report.totalCustomers > 0;

    return LgPage(
      title: 'Customer Insights',
      breadcrumb: 'Reports › Customers',
      subtitle: 'Slice your customer base by revenue, risk and plan.',
      backAction: () => context.go('/reports'),
      onRefresh: refreshData,
      secondaryActions: [
        LgPageAction(label: 'Export CSV', onPressed: _exportCsv),
      ],
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          PopupMenuButton<String>(
            onSelected: provider.setSelectedApp,
            itemBuilder: (_) => appsList
                .map((a) => PopupMenuItem(value: a.id, child: Text(a.name)))
                .toList(),
            child: Chip(
              label: Text(appsList
                  .firstWhere((a) => a.id == provider.selectedAppId,
                      orElse: () => appsList.first)
                  .name),
            ),
          ),
          const SizedBox(height: LgSpacing.s400),
          if (!hasData)
            const LgEmptyState(
              icon: Icons.groups_outlined,
              heading: 'No active customers yet',
              description:
                  'Once paying subscriptions sync, the segments appear here.',
            )
          else ...[
            _KpiRow(report: report),
            const SizedBox(height: LgSpacing.s600),
            _RevenueBands(report: report),
            const SizedBox(height: LgSpacing.s600),
            _PlanRiskTable(report: report),
            const SizedBox(height: LgSpacing.s600),
            _TopCustomersTable(report: report),
          ],
        ],
      ),
    );
  }
}

class _KpiRow extends StatelessWidget {
  final CustomerInsights report;
  const _KpiRow({required this.report});

  @override
  Widget build(BuildContext context) {
    final cards = [
      _Kpi(label: 'Customers', value: NumberFormat.decimalPattern().format(report.totalCustomers), color: LgColors.textPrimary, footnote: 'active (non-churned)'),
      _Kpi(label: 'Active MRR', value: _money(report.activeMrrCents), color: LgColors.success, footnote: 'monthly recurring'),
      _Kpi(label: 'At-risk customers', value: NumberFormat.decimalPattern().format(report.atRiskCustomers), color: LgColors.warning, footnote: 'missed a cycle'),
      _Kpi(label: 'At-risk MRR', value: _money(report.atRiskMrrCents), color: LgColors.warning, footnote: 'exposed'),
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

class _Kpi extends StatelessWidget {
  final String label;
  final String value;
  final Color color;
  final String footnote;
  const _Kpi(
      {required this.label,
      required this.value,
      required this.color,
      required this.footnote});

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
          const SizedBox(height: LgSpacing.s100),
          Text(footnote,
              style: theme.textTheme.bodySmall
                  ?.copyWith(color: LgColors.textSecondary)),
        ],
      ),
    );
  }
}

/// Horizontal bars showing how customers distribute across MRR bands.
class _RevenueBands extends StatelessWidget {
  final CustomerInsights report;
  const _RevenueBands({required this.report});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final maxCustomers = report.revenueBands.fold<int>(
        1, (m, b) => b.customers > m ? b.customers : m);
    return LgCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Customers by revenue band',
              style: theme.textTheme.titleSmall),
          const SizedBox(height: LgSpacing.s400),
          for (final b in report.revenueBands) ...[
            Row(
              children: [
                SizedBox(
                  width: 84,
                  child: Text(b.label, style: theme.textTheme.bodySmall),
                ),
                Expanded(
                  child: ClipRRect(
                    borderRadius: BorderRadius.circular(4),
                    child: LinearProgressIndicator(
                      value: maxCustomers == 0
                          ? 0
                          : b.customers / maxCustomers,
                      minHeight: 18,
                      backgroundColor: LgColors.surfaceSecondary,
                      valueColor: const AlwaysStoppedAnimation(LgColors.success),
                    ),
                  ),
                ),
                const SizedBox(width: LgSpacing.s300),
                SizedBox(
                  width: 150,
                  child: Text(
                    '${NumberFormat.decimalPattern().format(b.customers)}  ·  ${_money(b.mrrCents)}',
                    textAlign: TextAlign.right,
                    style: theme.textTheme.bodySmall,
                  ),
                ),
              ],
            ),
            const SizedBox(height: LgSpacing.s300),
          ],
        ],
      ),
    );
  }
}

/// Plan × risk crosstab: where the at-risk revenue concentrates.
class _PlanRiskTable extends StatelessWidget {
  final CustomerInsights report;
  const _PlanRiskTable({required this.report});

  @override
  Widget build(BuildContext context) {
    return LgCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Plan breakdown (risk split)',
              style: Theme.of(context).textTheme.titleSmall),
          const SizedBox(height: LgSpacing.s300),
          LgBreakpoints.isMobile(context)
              ? SingleChildScrollView(
                  scrollDirection: Axis.horizontal, child: _table())
              : _table(),
        ],
      ),
    );
  }

  Widget _table() {
    return Table(
      columnWidths: const {
        0: FlexColumnWidth(1.3),
        1: FlexColumnWidth(1),
        2: FlexColumnWidth(1),
        3: FlexColumnWidth(1.1),
        4: FlexColumnWidth(1.2),
      },
      children: [
        TableRow(
          decoration: BoxDecoration(color: LgColors.surfaceSecondary),
          children: ['Plan', 'Customers', 'At risk', 'MRR', 'At-risk MRR']
              .map((h) => Padding(
                    padding: const EdgeInsets.all(8),
                    child: Text(h,
                        style: const TextStyle(
                            fontSize: 12,
                            fontWeight: FontWeight.w600,
                            color: LgColors.textSecondary)),
                  ))
              .toList(),
        ),
        ...report.planRisk.map((p) => TableRow(
              children: [
                _cell(p.planName),
                _cell(NumberFormat.decimalPattern().format(p.customers)),
                Padding(
                  padding: const EdgeInsets.all(8),
                  child: Text(
                    p.atRiskCount == 0
                        ? '—'
                        : NumberFormat.decimalPattern().format(p.atRiskCount),
                    style: TextStyle(
                        fontSize: 13,
                        color: p.atRiskCount == 0
                            ? LgColors.textSecondary
                            : LgColors.warning),
                  ),
                ),
                _cell(_money(p.mrrCents)),
                Padding(
                  padding: const EdgeInsets.all(8),
                  child: Text(
                    p.atRiskMrrCents == 0 ? '—' : _money(p.atRiskMrrCents),
                    style: TextStyle(
                        fontSize: 13,
                        color: p.atRiskMrrCents == 0
                            ? LgColors.textSecondary
                            : LgColors.warning),
                  ),
                ),
              ],
            )),
      ],
    );
  }

  Widget _cell(String text) => Padding(
        padding: const EdgeInsets.all(8),
        child: Text(text, style: const TextStyle(fontSize: 13)),
      );
}

/// The highest-MRR customers (the whales), with a risk pill.
class _TopCustomersTable extends StatelessWidget {
  final CustomerInsights report;
  const _TopCustomersTable({required this.report});

  @override
  Widget build(BuildContext context) {
    return LgCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Top customers by MRR',
              style: Theme.of(context).textTheme.titleSmall),
          const SizedBox(height: LgSpacing.s300),
          for (final c in report.topCustomers)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: LgSpacing.s200),
              child: Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(c.shopName,
                            style: const TextStyle(
                                fontSize: 13, fontWeight: FontWeight.w600)),
                        Text(c.planName,
                            style: const TextStyle(
                                fontSize: 11, color: LgColors.textSecondary)),
                      ],
                    ),
                  ),
                  _RiskPill(state: c.riskState),
                  const SizedBox(width: LgSpacing.s400),
                  SizedBox(
                    width: 80,
                    child: Text('${_money(c.mrrCents)}/mo',
                        textAlign: TextAlign.right,
                        style: const TextStyle(
                            fontSize: 13, fontWeight: FontWeight.w600)),
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }
}

class _RiskPill extends StatelessWidget {
  final String state;
  const _RiskPill({required this.state});

  @override
  Widget build(BuildContext context) {
    final color = _riskColor(state);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(10),
      ),
      child: Text(_riskLabel(state),
          style: TextStyle(
              fontSize: 11, color: color, fontWeight: FontWeight.w600)),
    );
  }
}
