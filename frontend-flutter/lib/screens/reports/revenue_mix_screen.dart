import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../providers/apps_provider.dart';
import '../../providers/revenue_mix_provider.dart';
import '../../services/revenue_mix_service.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_service_unavailable.dart';
import '../../widgets/lg_table.dart';

class RevenueMixScreen extends StatefulWidget {
  const RevenueMixScreen({super.key});

  @override
  State<RevenueMixScreen> createState() => _RevenueMixScreenState();
}

class _RevenueMixScreenState extends State<RevenueMixScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<RevenueMixProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<RevenueMixProvider>();
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
          'revenue-mix-${DateTime.now().toIso8601String().split('T').first}.csv';
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
      debugPrint('revenue-mix: CSV export failed: $e');
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
        title: 'Revenue Mix',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgEmptyState(
          icon: Icons.pie_chart_outline,
          heading: 'No data yet',
          description: 'Connect your Shopify app to see your revenue mix.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<RevenueMixProvider>();

    if (provider.isServiceUnavailable) {
      return LgPage(
        title: 'Revenue Mix',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgServiceUnavailable(onRetry: retryLoad),
      );
    }

    if (provider.error != null) {
      return LgPage(
        title: 'Revenue Mix',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.report == null) {
      return LgPage(
        title: 'Revenue Mix',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final report = provider.report ?? RevenueMixReport.empty();
    final appsList = appsProvider.apps;
    final showAppFilter = appsList.isNotEmpty;
    final currency = report.currency;
    final hasData = report.grossCents > 0 || report.segments.isNotEmpty;

    return LgPage(
      title: 'Revenue Mix',
      breadcrumb: 'Reports › Revenue & Billing',
      subtitle:
          'Composition of revenue by charge type — RECURRING vs USAGE vs ONE-TIME',
      backAction: () => context.go('/reports'),
      onRefresh: refreshData,
      dateRange: provider.dateRange,
      onDateRangeChanged: provider.setDateRange,
      secondaryActions: [
        LgPageAction(label: 'Export CSV', onPressed: _exportCsv),
      ],
      // Fixed: app-selector only (no KPI hero on this report). Scrollable: the
      // composition card + breakdown table.
      pinned: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (showAppFilter)
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
        ],
      ),
      child: !hasData
          ? const LgEmptyState(
              icon: Icons.pie_chart_outline,
              heading: 'No revenue in range',
              description:
                  'Revenue mix breaks down gross revenue by charge type — recurring subscriptions, usage-based billing, and one-time charges. Once your app records charges in this range after a sync, the composition and breakdown will appear here.',
            )
          : Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _CompositionCard(report: report),
                const SizedBox(height: LgSpacing.s600),
                _BreakdownTable(report: report, currency: currency),
                const SizedBox(height: LgSpacing.s400),
                Text(
                  'RECURRING vs USAGE strictly separated — MRR uses RECURRING only, Usage Revenue uses USAGE only.',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: LgColors.textSecondary,
                    fontStyle: FontStyle.italic,
                  ),
                ),
              ],
            ),
    );
  }
}

/// Color per charge type: Recurring=primary/indigo, Usage=warning/amber,
/// One-time=success/green. Falls back to a muted color for anything else.
Color _segmentColor(String type) {
  switch (type.toLowerCase()) {
    case 'recurring':
      return LgColors.primary;
    case 'usage':
      return LgColors.warning;
    case 'one-time':
    case 'onetime':
    case 'one time':
      return LgColors.success;
    default:
      return LgColors.textSecondary;
  }
}

// ─── Composition card (stacked bar + legend) ────────────────────────
class _CompositionCard extends StatelessWidget {
  final RevenueMixReport report;
  const _CompositionCard({required this.report});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final segments = report.segments
        .where((s) => s.amountCents > 0 || s.pct > 0)
        .toList();

    return LgCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Revenue Composition', style: theme.textTheme.titleMedium),
          const SizedBox(height: LgSpacing.s400),
          // Full-width stacked bar. Each segment is weighted by its pct; a
          // min-width guard keeps tiny segments visible.
          ClipRRect(
            borderRadius: BorderRadius.circular(6),
            child: SizedBox(
              height: 28,
              child: segments.isEmpty
                  ? Container(color: LgColors.surfaceSecondary)
                  : Row(
                      children: [
                        for (final s in segments)
                          Expanded(
                            // Guard so a rounded flex of 0 still renders.
                            flex: (s.pct * 1000).round().clamp(1, 1000),
                            child: Container(
                              constraints: const BoxConstraints(minWidth: 2),
                              color: _segmentColor(s.type),
                            ),
                          ),
                      ],
                    ),
            ),
          ),
          const SizedBox(height: LgSpacing.s400),
          // Legend: each type + its % share.
          Wrap(
            spacing: LgSpacing.s500,
            runSpacing: LgSpacing.s200,
            children: [
              for (final s in segments)
                _LegendItem(
                  color: _segmentColor(s.type),
                  label: s.type,
                  pct: s.pct,
                ),
            ],
          ),
        ],
      ),
    );
  }
}

class _LegendItem extends StatelessWidget {
  final Color color;
  final String label;
  final double pct;
  const _LegendItem({
    required this.color,
    required this.label,
    required this.pct,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 12,
          height: 12,
          decoration: BoxDecoration(
            color: color,
            borderRadius: BorderRadius.circular(3),
          ),
        ),
        const SizedBox(width: LgSpacing.s200),
        Text('$label ${_pct(pct)}', style: theme.textTheme.bodySmall),
      ],
    );
  }
}

// ─── Breakdown table ────────────────────────────────────────────────
class _BreakdownTable extends StatelessWidget {
  final RevenueMixReport report;
  final String currency;
  const _BreakdownTable({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final hasRefunds = report.refundCents > 0;

    // Cell builders preserving the prior per-row styling: segment rows use
    // bodyMedium with a color swatch, Total/Net rows use titleSmall (bold),
    // "Less refunds" is muted.
    Widget typeCell(String label, {Color? swatch, TextStyle? style}) => Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        if (swatch != null) ...[
          Container(
            width: 10,
            height: 10,
            decoration: BoxDecoration(
              color: swatch,
              borderRadius: BorderRadius.circular(2),
            ),
          ),
          const SizedBox(width: LgSpacing.s200),
        ],
        Flexible(child: Text(label, style: style)),
      ],
    );

    final segmentStyle = theme.textTheme.bodyMedium?.copyWith(
      color: LgColors.textPrimary,
    );
    final mutedStyle = theme.textTheme.bodyMedium?.copyWith(
      color: LgColors.textSecondary,
    );
    final totalStyle = theme.textTheme.titleSmall;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Breakdown', style: theme.textTheme.titleMedium),
        const SizedBox(height: LgSpacing.s300),
        LgTable(
          columns: const [
            LgTableColumn('TYPE', flex: 3),
            LgTableColumn('AMOUNT', flex: 2, numeric: true),
            LgTableColumn('%', flex: 1, numeric: true),
          ],
          rows: [
            for (final s in report.segments)
              [
                typeCell(
                  s.type,
                  swatch: _segmentColor(s.type),
                  style: segmentStyle,
                ),
                Text(_money(s.amountCents, currency), style: segmentStyle),
                Text(_pct(s.pct), style: segmentStyle),
              ],
            [
              typeCell('Total', style: totalStyle),
              Text(_money(report.grossCents, currency), style: totalStyle),
              Text('100%', style: totalStyle),
            ],
            if (hasRefunds) ...[
              [
                typeCell('Less refunds', style: mutedStyle),
                Text(
                  '-${_money(report.refundCents, currency)}',
                  style: mutedStyle,
                ),
                const Text(''),
              ],
              [
                typeCell('Net', style: totalStyle),
                Text(_money(report.netCents, currency), style: totalStyle),
                const Text(''),
              ],
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

String _pct(double value) => '${(value * 100).round()}%';
