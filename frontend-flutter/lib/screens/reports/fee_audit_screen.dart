import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../providers/apps_provider.dart';
import '../../providers/fee_audit_provider.dart';
import '../../services/fee_audit_service.dart';
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

class FeeAuditScreen extends StatefulWidget {
  const FeeAuditScreen({super.key});

  @override
  State<FeeAuditScreen> createState() => _FeeAuditScreenState();
}

class _FeeAuditScreenState extends State<FeeAuditScreen> with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<FeeAuditProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<FeeAuditProvider>();
    if (provider.selectedAppId == null) return;
    try {
      final bytes = await provider.fetchCsvBytes();
      if (bytes == null || bytes.isEmpty) {
        if (!mounted) return;
        messenger.showSnackBar(
            const SnackBar(content: Text('CSV export returned no data.')));
        return;
      }
      final ok = downloadBytes(bytes, 'fee-audit.csv', 'text/csv');
      if (!mounted) return;
      if (!ok) {
        messenger.showSnackBar(const SnackBar(
            content: Text('CSV export is only available on the web app.')));
      }
    } catch (e) {
      debugPrint('fee-audit: CSV export failed: $e');
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
        title: 'Fee Audit',
        breadcrumb: 'Reports › Guard',
        backAction: () => context.go('/reports'),
        child: child,
      );

  @override
  Widget build(BuildContext context) {
    final appsProvider = context.watch<AppsProvider>();
    if (appsProvider.apps.isEmpty) {
      return _shell(LgEmptyState(
        icon: Icons.verified_user_outlined,
        heading: 'No data yet',
        description: 'Connect your Shopify app to audit its fees.',
        actionLabel: 'Go to Apps',
        onAction: () => context.go('/apps'),
      ));
    }

    final provider = context.watch<FeeAuditProvider>();
    if (provider.isServiceUnavailable) {
      return _shell(LgServiceUnavailable(onRetry: retryLoad));
    }
    if (provider.error != null) {
      return _shell(LgErrorState(message: provider.error!, onRetry: retryLoad));
    }
    if (provider.isLoading && provider.report == null) {
      return _shell(const Center(child: CircularProgressIndicator()));
    }

    final report = provider.report ?? FeeAuditReport.empty();
    final appsList = appsProvider.apps;
    final hasData = report.monthsAudited > 0 || report.totalGrossCents > 0;

    return LgPage(
      title: 'Fee Audit',
      breadcrumb: 'Reports › Guard',
      subtitle:
          'Verify Shopify charged the right revenue share — actual vs expected.',
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
              icon: Icons.verified_user_outlined,
              heading: 'No fee data yet',
              description:
                  'Once transactions with Shopify fees are synced, the audit appears here.',
            )
          else ...[
            _VerdictBanner(report: report),
            const SizedBox(height: LgSpacing.s400),
            if (!report.tierMatches) ...[
              _TierMismatchNote(report: report),
              const SizedBox(height: LgSpacing.s400),
            ],
            _KpiRow(report: report),
            const SizedBox(height: LgSpacing.s600),
            _AuditTable(report: report),
            const SizedBox(height: LgSpacing.s400),
            Text(
              'Expected = gross × the detected revenue-share rate (${report.detectedFeePct.toStringAsFixed(0)}%). '
              'A month is flagged when the actual cut deviates by more than 1% of gross.',
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

// ─── Verdict banner ─────────────────────────────────────────────────
class _VerdictBanner extends StatelessWidget {
  final FeeAuditReport report;
  const _VerdictBanner({required this.report});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final ok = report.allClear;
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
                      ? 'All clear — Shopify charged the expected rate'
                      : '${report.flaggedMonths} of ${report.monthsAudited} months flagged',
                  style: theme.textTheme.titleMedium
                      ?.copyWith(fontWeight: FontWeight.w700, color: color),
                ),
                const SizedBox(height: LgSpacing.s100),
                Text(
                  ok
                      ? 'Every audited month matches the detected ${report.detectedFeePct.toStringAsFixed(0)}% revenue share.'
                      : 'Some months deviate from the detected ${report.detectedFeePct.toStringAsFixed(0)}% rate — review the table below (a tier transition can cause a one-off).',
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

// ─── Tier mismatch note ─────────────────────────────────────────────
class _TierMismatchNote extends StatelessWidget {
  final FeeAuditReport report;
  const _TierMismatchNote({required this.report});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.all(LgSpacing.s400),
      decoration: BoxDecoration(
        color: LgColors.warningBg.withValues(alpha: 0.4),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: LgColors.warning.withValues(alpha: 0.5)),
      ),
      child: Row(
        children: [
          const Icon(Icons.info_outline, size: 18, color: LgColors.warning),
          const SizedBox(width: LgSpacing.s300),
          Expanded(
            child: Text(
              'Configured tier is ${report.configuredFeePct.toStringAsFixed(0)}%, but Shopify actually retains ~${report.detectedFeePct.toStringAsFixed(0)}%. '
              'Update the app tier so the configured rate matches reality.',
              style: theme.textTheme.bodySmall,
            ),
          ),
        ],
      ),
    );
  }
}

// ─── KPI row ────────────────────────────────────────────────────────
class _KpiRow extends StatelessWidget {
  final FeeAuditReport report;
  const _KpiRow({required this.report});

  @override
  Widget build(BuildContext context) {
    final cards = [
      _Kpi(
          label: 'Effective fee rate',
          value: '${report.effectiveFeePct.toStringAsFixed(1)}%',
          color: LgColors.primary,
          footnote: 'Shopify cut ÷ gross'),
      _Kpi(
          label: 'Total Shopify cut',
          value: _money(report.totalCutCents),
          color: LgColors.warning,
          footnote: 'across ${report.monthsAudited} months'),
      _Kpi(
          label: 'Saved vs 20% plan',
          value: _money(report.savingsVsDefaultCents),
          color: LgColors.success,
          footnote: 'reduced-share savings'),
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

// ─── Per-month audit table ──────────────────────────────────────────
class _AuditTable extends StatelessWidget {
  final FeeAuditReport report;
  const _AuditTable({required this.report});

  @override
  Widget build(BuildContext context) {
    return LgCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Monthly audit', style: Theme.of(context).textTheme.titleSmall),
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
        0: FlexColumnWidth(1),
        1: FlexColumnWidth(1.2),
        2: FlexColumnWidth(1.2),
        3: FlexColumnWidth(1),
        4: FlexColumnWidth(1.2),
        5: FlexColumnWidth(1),
      },
      children: [
        TableRow(
          decoration: BoxDecoration(color: LgColors.surfaceSecondary),
          children: [
            'Month',
            'Gross',
            'Shopify Cut',
            'Rate',
            'Expected',
            'Guard'
          ]
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
                _cell(_money(m.shopifyCutCents)),
                _cell('${m.effectiveFeePct.toStringAsFixed(1)}%'),
                _cell(_money(m.expectedCutCents)),
                Padding(
                  padding: const EdgeInsets.all(8),
                  child: m.feeGuardOk
                      ? const Icon(Icons.check_circle_outline,
                          size: 16, color: LgColors.success)
                      : Row(mainAxisSize: MainAxisSize.min, children: [
                          const Icon(Icons.warning_amber_rounded,
                              size: 16, color: LgColors.warning),
                          const SizedBox(width: 4),
                          Text('${m.feeVarianceCents >= 0 ? '+' : '−'}${_money(m.feeVarianceCents.abs())}',
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
