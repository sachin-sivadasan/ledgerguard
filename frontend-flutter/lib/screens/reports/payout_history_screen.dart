import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../providers/apps_provider.dart';
import '../../providers/payout_history_provider.dart';
import '../../services/payout_history_service.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_service_unavailable.dart';
import '../../widgets/lg_table.dart';

class PayoutHistoryScreen extends StatefulWidget {
  const PayoutHistoryScreen({super.key});

  @override
  State<PayoutHistoryScreen> createState() => _PayoutHistoryScreenState();
}

class _PayoutHistoryScreenState extends State<PayoutHistoryScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<PayoutHistoryProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<PayoutHistoryProvider>();
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
          'payout-history-${DateTime.now().toIso8601String().split('T').first}.csv';
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
      debugPrint('payout-history: CSV export failed: $e');
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
        title: 'Payout History',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgEmptyState(
          icon: Icons.history_outlined,
          heading: 'No data yet',
          description: 'Connect your Shopify app to see completed payouts.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<PayoutHistoryProvider>();

    if (provider.isServiceUnavailable) {
      return LgPage(
        title: 'Payout History',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgServiceUnavailable(onRetry: retryLoad),
      );
    }

    if (provider.error != null) {
      return LgPage(
        title: 'Payout History',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.report == null) {
      return LgPage(
        title: 'Payout History',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final report = provider.report ?? PayoutHistoryReport.empty();
    final appsList = appsProvider.apps;
    final showAppFilter = appsList.isNotEmpty;
    final currency = report.currency;
    final hasData = report.rows.isNotEmpty || report.totalPaidCents > 0;

    return LgPage(
      title: 'Payout History',
      breadcrumb: 'Reports › Revenue & Billing',
      subtitle: 'Earnings marked paid out, grouped by charge month',
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
              icon: Icons.history_outlined,
              heading: 'No completed payouts yet',
              description:
                  'Once earnings are marked paid out, a monthly summary appears here. Upcoming (not-yet-paid) earnings live in Payout Schedule.',
            )
          else ...[
            _HeroRow(report: report, currency: currency),
            const SizedBox(height: LgSpacing.s600),
            _PayoutLogTable(report: report, currency: currency),
            const SizedBox(height: LgSpacing.s400),
            Text(
              'Each row is one calendar month of paid earnings; amounts are net of Shopify\'s revenue share. Upcoming earnings live in Payout Schedule.',
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
  final PayoutHistoryReport report;
  final String currency;
  const _HeroRow({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final cards = <Widget>[
      _KpiCard(
        label: 'Total Paid',
        value: _money(report.totalPaidCents, currency),
        color: LgColors.success,
        footnote: 'paid earnings in range',
      ),
      _KpiCard(
        label: 'Payouts',
        value: '${report.payoutCount}',
        color: LgColors.textPrimary,
        footnote: 'months with paid earnings',
      ),
      _KpiCard(
        label: 'Avg Payout',
        value: _money(report.avgPayoutCents, currency),
        color: LgColors.textPrimary,
        footnote: 'total paid ÷ payout count',
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

// ─── Payout log timeline ────────────────────────────────────────────
class _PayoutLogTable extends StatelessWidget {
  final PayoutHistoryReport report;
  final String currency;
  const _PayoutLogTable({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final secondary = theme.textTheme.bodySmall?.copyWith(
      color: LgColors.textSecondary,
    );
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Payout Log', style: theme.textTheme.titleMedium),
        const SizedBox(height: LgSpacing.s300),
        LgTable(
          columns: const [
            LgTableColumn('PERIOD', flex: 3),
            LgTableColumn('CHARGES', flex: 2, numeric: true),
            LgTableColumn('AMOUNT', flex: 2, numeric: true),
            LgTableColumn('AVAILABLE DATE', flex: 2, numeric: true),
          ],
          rows: [
            for (final row in report.rows)
              [
                Text(
                  row.periodDate != null
                      ? DateFormat('MMM yyyy').format(row.periodDate!)
                      : row.period,
                  style: theme.textTheme.titleSmall,
                ),
                Text('${row.chargeCount}', style: secondary),
                Text(
                  _money(row.amountCents, currency),
                  style: theme.textTheme.titleSmall?.copyWith(
                    color: LgColors.success,
                  ),
                ),
                Text(
                  row.availableDateTime != null
                      ? DateFormat('MMM d, yyyy').format(row.availableDateTime!)
                      : '—',
                  style: secondary,
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
