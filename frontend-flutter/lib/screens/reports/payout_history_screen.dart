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
        title: 'Payout History',
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
        backAction: () => context.go('/reports'),
        child: LgServiceUnavailable(onRetry: retryLoad),
      );
    }

    if (provider.error != null) {
      return LgPage(
        title: 'Payout History',
        backAction: () => context.go('/reports'),
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.report == null) {
      return LgPage(
        title: 'Payout History',
        backAction: () => context.go('/reports'),
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final report = provider.report ?? PayoutHistoryReport.empty();
    final appsList = appsProvider.apps;
    final showAppFilter = appsList.length > 1;
    final currency = report.currency;
    final hasData = report.rows.isNotEmpty || report.totalPaidCents > 0;

    return LgPage(
      title: 'Payout History',
      subtitle:
          'Paid earnings by month — a record of earnings Shopify has disbursed',
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
          if (!hasData)
            const LgEmptyState(
              icon: Icons.history_outlined,
              heading: 'No completed payouts yet',
              description:
                  'Once your earnings clear and Shopify pays them out, a monthly record of those payouts will appear here. Upcoming (not-yet-paid) earnings live in Payout Schedule.',
            )
          else ...[
            _HeroRow(report: report, currency: currency),
            const SizedBox(height: LgSpacing.s600),
            _PayoutLogTable(report: report, currency: currency),
            const SizedBox(height: LgSpacing.s400),
            Text(
              'Each row is one calendar month of paid earnings; amounts are net of Shopify\'s revenue share. Upcoming earnings live in Payout Schedule.',
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

// ─── Payout log timeline ────────────────────────────────────────────
class _PayoutLogTable extends StatelessWidget {
  final PayoutHistoryReport report;
  final String currency;
  const _PayoutLogTable({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Payout Log', style: theme.textTheme.titleMedium),
        const SizedBox(height: LgSpacing.s300),
        // Column header row: PERIOD | CHARGES | AMOUNT | PAID DATE
        Padding(
          padding: const EdgeInsets.symmetric(
              horizontal: LgSpacing.s400, vertical: LgSpacing.s200),
          child: Row(
            children: [
              Expanded(
                flex: 3,
                child: Text('PERIOD',
                    style: theme.textTheme.bodySmall
                        ?.copyWith(color: LgColors.textSecondary)),
              ),
              _headCell(theme, 'CHARGES'),
              _headCell(theme, 'AMOUNT'),
              _headCell(theme, 'AVAILABLE DATE'),
            ],
          ),
        ),
        ...report.rows.map((row) => Padding(
              padding: const EdgeInsets.only(bottom: LgSpacing.s200),
              child: _PayoutLogRow(row: row, currency: currency),
            )),
      ],
    );
  }

  Widget _headCell(ThemeData theme, String label) => Expanded(
        flex: 2,
        child: Text(label,
            textAlign: TextAlign.end,
            style: theme.textTheme.bodySmall
                ?.copyWith(color: LgColors.textSecondary)),
      );
}

class _PayoutLogRow extends StatelessWidget {
  final PayoutHistoryRow row;
  final String currency;
  const _PayoutLogRow({required this.row, required this.currency});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final periodDate = row.periodDate;
    final periodLabel =
        periodDate != null ? DateFormat('MMM yyyy').format(periodDate) : row.period;
    final avail = row.availableDateTime;
    final availLabel =
        avail != null ? DateFormat('MMM d, yyyy').format(avail) : '—';

    return LgCard(
      child: Row(
        children: [
          Expanded(
            flex: 3,
            child: Text(periodLabel, style: theme.textTheme.titleSmall),
          ),
          Expanded(
            flex: 2,
            child: Text('${row.chargeCount}',
                textAlign: TextAlign.end,
                style: theme.textTheme.bodySmall
                    ?.copyWith(color: LgColors.textSecondary)),
          ),
          Expanded(
            flex: 2,
            child: Text(_money(row.amountCents, currency),
                textAlign: TextAlign.end,
                style: theme.textTheme.titleSmall
                    ?.copyWith(color: LgColors.success)),
          ),
          Expanded(
            flex: 2,
            child: Text(availLabel,
                textAlign: TextAlign.end,
                style: theme.textTheme.bodySmall
                    ?.copyWith(color: LgColors.textSecondary)),
          ),
        ],
      ),
    );
  }
}

String _money(int cents, String currency) {
  final format = NumberFormat.simpleCurrency(name: currency);
  return format.format(cents / 100);
}
