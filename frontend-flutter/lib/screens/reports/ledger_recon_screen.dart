import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../providers/apps_provider.dart';
import '../../providers/ledger_recon_provider.dart';
import '../../services/ledger_recon_service.dart';
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

class LedgerReconScreen extends StatefulWidget {
  const LedgerReconScreen({super.key});

  @override
  State<LedgerReconScreen> createState() => _LedgerReconScreenState();
}

class _LedgerReconScreenState extends State<LedgerReconScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<LedgerReconProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<LedgerReconProvider>();
    if (provider.selectedAppId == null) return;
    try {
      final bytes = await provider.fetchCsvBytes();
      if (bytes == null || bytes.isEmpty) {
        if (!mounted) return;
        messenger.showSnackBar(
            const SnackBar(content: Text('CSV export returned no data.')));
        return;
      }
      final ok = downloadBytes(bytes, 'ledger-reconciliation.csv', 'text/csv');
      if (!mounted) return;
      if (!ok) {
        messenger.showSnackBar(const SnackBar(
            content: Text('CSV export is only available on the web app.')));
      }
    } catch (e) {
      debugPrint('ledger-recon: CSV export failed: $e');
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
        title: 'Ledger Reconciliation',
        breadcrumb: 'Reports › Guard',
        backAction: () => context.go('/reports'),
        child: child,
      );

  @override
  Widget build(BuildContext context) {
    final appsProvider = context.watch<AppsProvider>();
    if (appsProvider.apps.isEmpty) {
      return _shell(LgEmptyState(
        icon: Icons.balance_outlined,
        heading: 'No data yet',
        description: 'Connect your Shopify app to reconcile its ledger.',
        actionLabel: 'Go to Apps',
        onAction: () => context.go('/apps'),
      ));
    }

    final provider = context.watch<LedgerReconProvider>();
    if (provider.isServiceUnavailable) {
      return _shell(LgServiceUnavailable(onRetry: retryLoad));
    }
    if (provider.error != null) {
      return _shell(LgErrorState(message: provider.error!, onRetry: retryLoad));
    }
    if (provider.isLoading && provider.report == null) {
      return _shell(const Center(child: CircularProgressIndicator()));
    }

    final report = provider.report ?? ReconReport.empty();
    final appsList = appsProvider.apps;
    final hasData = report.monthsAudited > 0;

    return LgPage(
      title: 'Ledger Reconciliation',
      breadcrumb: 'Reports › Guard',
      subtitle: 'Does the money add up? Gross = net + revenue share + processing.',
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
              icon: Icons.balance_outlined,
              heading: 'Nothing to reconcile yet',
              description:
                  'Once transactions are synced, the monthly reconciliation appears here.',
            )
          else ...[
            _Verdict(report: report),
            const SizedBox(height: LgSpacing.s400),
            _KpiRow(report: report),
            const SizedBox(height: LgSpacing.s600),
            _ReconTable(report: report),
            const SizedBox(height: LgSpacing.s400),
            Text(
              'Reconciled when net + revenue share + processing accounts for gross (within 1%). Processing (~3%) is Shopify\'s payment fee, derived per sale. A residual is money those buckets don\'t explain — usually a refund whose fee reversal hasn\'t synced.',
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: LgColors.textSecondary,
                  ),
            ),
          ],
        ],
      ),
    );
  }
}

class _Verdict extends StatelessWidget {
  final ReconReport report;
  const _Verdict({required this.report});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final ok = report.reconciled;
    final color = ok ? LgColors.success : LgColors.warning;
    return LgCard(
      child: Row(
        children: [
          Icon(ok ? Icons.verified_outlined : Icons.warning_amber_rounded,
              color: color, size: 28),
          const SizedBox(width: LgSpacing.s400),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  ok
                      ? 'Reconciled — every dollar accounted for'
                      : '${report.monthsFlagged} of ${report.monthsAudited} months don\'t reconcile',
                  style: theme.textTheme.titleMedium
                      ?.copyWith(fontWeight: FontWeight.w700, color: color),
                ),
                const SizedBox(height: LgSpacing.s100),
                Text(
                  ok
                      ? 'Net + revenue share + processing accounts for gross in every audited month.'
                      : 'Some months leave a residual the buckets don\'t explain — usually an unsynced refund fee; review the table.',
                  style: theme.textTheme.bodySmall
                      ?.copyWith(color: LgColors.textSecondary),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _KpiRow extends StatelessWidget {
  final ReconReport report;
  const _KpiRow({required this.report});

  @override
  Widget build(BuildContext context) {
    final cards = [
      _Kpi(label: 'Gross', value: _money(report.totalGrossCents), color: LgColors.textPrimary, footnote: 'merchants paid'),
      _Kpi(label: 'Net', value: _money(report.totalNetCents), color: LgColors.success, footnote: 'your payout'),
      _Kpi(label: 'Revenue share', value: _money(report.totalRevenueShareCents), color: LgColors.warning, footnote: "Shopify's cut"),
      _Kpi(label: 'Processing', value: _money(report.totalProcessingCents), color: LgColors.warning, footnote: 'derived'),
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

class _ReconTable extends StatelessWidget {
  final ReconReport report;
  const _ReconTable({required this.report});

  @override
  Widget build(BuildContext context) {
    return LgCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Monthly reconciliation',
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
        0: FlexColumnWidth(0.9),
        1: FlexColumnWidth(1.1),
        2: FlexColumnWidth(1.1),
        3: FlexColumnWidth(1.1),
        4: FlexColumnWidth(1.1),
        5: FlexColumnWidth(1),
      },
      children: [
        TableRow(
          decoration: BoxDecoration(color: LgColors.surfaceSecondary),
          children: ['Month', 'Gross', 'Net', 'Revenue share', 'Processing', 'Status']
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
        ...report.months.map((m) => TableRow(
              children: [
                _cell(m.month),
                _cell(_money(m.grossCents)),
                _cell(_money(m.netCents)),
                _cell(_money(m.revenueShareCents)),
                _cell(_money(m.processingCents)),
                Padding(
                  padding: const EdgeInsets.all(8),
                  child: m.reconciled
                      ? const Icon(Icons.check_circle_outline,
                          size: 16, color: LgColors.success)
                      : Row(mainAxisSize: MainAxisSize.min, children: [
                          const Icon(Icons.warning_amber_rounded,
                              size: 16, color: LgColors.warning),
                          const SizedBox(width: 4),
                          Text(_money(m.residualCents),
                              style: const TextStyle(
                                  fontSize: 12, color: LgColors.warning)),
                        ]),
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
