import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../providers/apps_provider.dart';
import '../../providers/earnings_report_provider.dart';
import '../../services/earnings_report_service.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_table.dart';
import '../../widgets/lg_service_unavailable.dart';

class EarningsReportScreen extends StatefulWidget {
  const EarningsReportScreen({super.key});

  @override
  State<EarningsReportScreen> createState() => _EarningsReportScreenState();
}

class _EarningsReportScreenState extends State<EarningsReportScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<EarningsReportProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<EarningsReportProvider>();
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
          'earnings-${DateTime.now().toIso8601String().split('T').first}.csv';
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
      debugPrint('earnings: CSV export failed: $e');
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
        title: 'Earnings',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgEmptyState(
          icon: Icons.account_balance_wallet_outlined,
          heading: 'No data yet',
          description: 'Connect your Shopify app to see earnings.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<EarningsReportProvider>();

    if (provider.isServiceUnavailable) {
      return LgPage(
        title: 'Earnings',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgServiceUnavailable(onRetry: retryLoad),
      );
    }

    if (provider.error != null) {
      return LgPage(
        title: 'Earnings',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.report == null) {
      return LgPage(
        title: 'Earnings',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final report = provider.report ?? EarningsReport.empty();
    final appsList = appsProvider.apps;
    final showAppFilter = appsList.isNotEmpty;
    final currency = report.currency;
    final hasData = report.charges.isNotEmpty ||
        report.netEarningsCents > 0 ||
        report.pendingCents > 0 ||
        report.availableCents > 0 ||
        report.paidOutCents > 0;

    return LgPage(
      title: 'Earnings',
      breadcrumb: 'Reports › Revenue & Billing',
      subtitle:
          'Developer earnings after Shopify revenue share — net, pending, available, and paid out',
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
              icon: Icons.account_balance_wallet_outlined,
              heading: 'No earnings in range',
              description:
                  'Earnings are developer payouts after Shopify takes its revenue share. Once your app records charges in this range, net earnings, pending/available/paid-out balances, and the charge breakdown will appear here.',
            )
          else ...[
            _HeroRow(report: report, currency: currency),
            const SizedBox(height: LgSpacing.s600),
            _ChargesTable(report: report, currency: currency),
          ],
        ],
      ),
    );
  }
}

// ─── Hero KPI cards ─────────────────────────────────────────────────
class _HeroRow extends StatelessWidget {
  final EarningsReport report;
  final String currency;
  const _HeroRow({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final cards = [
      _KpiCard(
        label: 'Net Earnings',
        value: _money(report.netEarningsCents, currency),
        color: LgColors.success,
        footnote: 'net after Shopify share — pending + available + paid',
      ),
      _KpiCard(
        label: 'Pending',
        value: _money(report.pendingCents, currency),
        color: LgColors.warning,
        footnote: 'not yet cleared payout window',
      ),
      _KpiCard(
        label: 'Available',
        value: _money(report.availableCents, currency),
        color: LgColors.success,
        footnote: 'cleared & withdrawable now',
      ),
      _KpiCard(
        label: 'Paid Out',
        value: _money(report.paidOutCents, currency),
        color: LgColors.textSecondary,
        footnote: 'already disbursed',
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

// ─── Charges table ──────────────────────────────────────────────────
class _ChargesTable extends StatelessWidget {
  final EarningsReport report;
  final String currency;
  const _ChargesTable({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final charges = [...report.charges]
      ..sort((a, b) => (b.date ?? DateTime(0)).compareTo(a.date ?? DateTime(0)));

    final secondary =
        theme.textTheme.bodySmall?.copyWith(color: LgColors.textSecondary);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Charges', style: theme.textTheme.titleMedium),
        const SizedBox(height: LgSpacing.s300),
        LgTable(
          columns: const [
            LgTableColumn('DATE', flex: 2),
            LgTableColumn('STORE', flex: 4),
            LgTableColumn('GROSS', flex: 2, numeric: true),
            LgTableColumn('NET', flex: 2, numeric: true),
            LgTableColumn('STATUS', flex: 2),
            LgTableColumn('AVAILABLE DATE', flex: 3, numeric: true),
          ],
          rows: [
            for (final c in charges)
              [
                Text(_date(c.date), style: secondary),
                Text(
                  c.shopName.isNotEmpty
                      ? c.shopName
                      : (c.domain.isNotEmpty ? c.domain : '—'),
                  style: theme.textTheme.titleSmall,
                ),
                Text(_money(c.grossCents, currency), style: secondary),
                Text(_money(c.netCents, currency),
                    style: theme.textTheme.titleSmall),
                _StatusChip(status: c.status),
                Text(_date(c.availableDate), style: secondary),
              ],
          ],
        ),
      ],
    );
  }
}

class _StatusChip extends StatelessWidget {
  final String status;
  const _StatusChip({required this.status});

  @override
  Widget build(BuildContext context) {
    final (bg, fg) = switch (status.toLowerCase()) {
      'pending' => (LgColors.warning.withValues(alpha: 0.14), LgColors.warning),
      'available' => (
          LgColors.success.withValues(alpha: 0.14),
          LgColors.success
        ),
      'paid' => (LgColors.surfaceSecondary, LgColors.textSecondary),
      _ => (LgColors.surfaceSecondary, LgColors.textSecondary),
    };
    final label = status.isNotEmpty ? status : '—';
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(10),
      ),
      child: Text(
        label,
        style: TextStyle(fontSize: 11, color: fg, fontWeight: FontWeight.w600),
      ),
    );
  }
}

String _money(int cents, String currency) {
  final format = NumberFormat.simpleCurrency(name: currency);
  return format.format(cents / 100);
}

String _date(DateTime? value) =>
    value == null ? '—' : DateFormat('MMM d, yyyy').format(value);
