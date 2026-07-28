import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../providers/apps_provider.dart';
import '../../providers/uninstall_context_provider.dart';
import '../../services/uninstall_context_service.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_service_unavailable.dart';
import '../../widgets/lg_table.dart';

class UninstallContextScreen extends StatefulWidget {
  const UninstallContextScreen({super.key});

  @override
  State<UninstallContextScreen> createState() => _UninstallContextScreenState();
}

class _UninstallContextScreenState extends State<UninstallContextScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<UninstallContextProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<UninstallContextProvider>();
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
          'uninstall-context-${DateTime.now().toIso8601String().split('T').first}.csv';
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
      debugPrint('uninstall-context: CSV export failed: $e');
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
        title: 'Uninstall Context',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: LgEmptyState(
          icon: Icons.link_off_outlined,
          heading: 'No data yet',
          description: 'Connect your Shopify app to see uninstall context.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<UninstallContextProvider>();

    if (provider.isServiceUnavailable) {
      return LgPage(
        title: 'Uninstall Context',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: LgServiceUnavailable(onRetry: retryLoad),
      );
    }

    if (provider.error != null) {
      return LgPage(
        title: 'Uninstall Context',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.report == null) {
      return LgPage(
        title: 'Uninstall Context',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final report = provider.report ?? UninstallContextReport.empty();
    final appsList = appsProvider.apps;
    final showAppFilter = appsList.isNotEmpty;
    final hasData = report.stores.isNotEmpty || report.uninstalls > 0;

    return LgPage(
      title: 'Uninstall Context',
      subtitle:
          'What state stores were in just before they uninstalled — inferred from risk signals',
      breadcrumb: 'Reports › Retention & Risk',
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
              icon: Icons.link_off_outlined,
              heading: 'No uninstalls in range',
              description:
                  'When a store uninstalls your app, LedgerGuard infers the state it was in beforehand from risk signals. LedgerGuard is read-only, so there is no self-reported reason — once uninstalls occur in this range, they will appear here.',
            )
          else ...[
            _HeroRow(report: report),
            const SizedBox(height: LgSpacing.s300),
            const _CaveatLine(),
            const SizedBox(height: LgSpacing.s600),
            _StoresTable(report: report),
          ],
        ],
      ),
    );
  }
}

// ─── Caveat line ────────────────────────────────────────────────────
class _CaveatLine extends StatelessWidget {
  const _CaveatLine();

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    // Amber info banner (matches the wireframe callout) rather than a plain inline line.
    return Container(
      padding: const EdgeInsets.symmetric(
          horizontal: LgSpacing.s300, vertical: LgSpacing.s200),
      decoration: BoxDecoration(
        color: LgColors.warning.withValues(alpha: 0.10),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: LgColors.warning.withValues(alpha: 0.30)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Icon(Icons.info_outline, size: 16, color: LgColors.warning),
          const SizedBox(width: LgSpacing.s200),
          Expanded(
            child: Text(
              'Inferred state, NOT a self-reported reason — LedgerGuard is read-only (no uninstall survey).',
              style: theme.textTheme.bodySmall?.copyWith(
                color: LgColors.textPrimary,
                fontStyle: FontStyle.italic,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

// ─── Hero KPI cards ─────────────────────────────────────────────────
class _HeroRow extends StatelessWidget {
  final UninstallContextReport report;
  const _HeroRow({required this.report});

  @override
  Widget build(BuildContext context) {
    final cards = [
      _KpiCard(
        label: 'Uninstalls',
        value: '${report.uninstalls}',
        color: LgColors.textPrimary,
        footnote: 'distinct stores uninstalled in range',
      ),
      _KpiCard(
        label: 'Were At-Risk First',
        value: _percent(report.wereAtRiskPct),
        color: LgColors.warning,
        footnote:
            'in a risk/frozen state pre-uninstall (of correlated uninstalls)',
      ),
      _KpiCard(
        label: 'Median Tenure',
        value: _months(report.medianTenureMonths),
        color: LgColors.textPrimary,
        footnote: 'install → uninstall, median',
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

// ─── Uninstalled stores table ───────────────────────────────────────
class _StoresTable extends StatelessWidget {
  final UninstallContextReport report;
  const _StoresTable({required this.report});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final stores = [...report.stores]
      ..sort((a, b) => (b.uninstalledDate ?? DateTime(0))
          .compareTo(a.uninstalledDate ?? DateTime(0)));

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Uninstalled Stores (pre-churn state)',
            style: theme.textTheme.titleMedium),
        const SizedBox(height: LgSpacing.s300),
        LgTable(
          columns: const [
            LgTableColumn('STORE', flex: 4),
            LgTableColumn('STATE BEFORE UNINSTALL', flex: 3),
            LgTableColumn('PLAN', flex: 2, numeric: true),
            LgTableColumn('TENURE', flex: 2, numeric: true),
            LgTableColumn('UNINSTALLED', flex: 2, numeric: true),
          ],
          rows: [
            for (final s in stores)
              [
                Text(s.domain.isNotEmpty ? s.domain : '—',
                    style: theme.textTheme.titleSmall),
                _StateBadge(state: s.stateBeforeUninstall),
                Text(s.planName.isNotEmpty ? s.planName : '—',
                    style: theme.textTheme.titleSmall),
                Text(_months(s.tenureMonths), style: theme.textTheme.titleSmall),
                Text(
                  s.uninstalledDate != null
                      ? DateFormat('MMM dd').format(s.uninstalledDate!)
                      : '—',
                  style: theme.textTheme.titleSmall,
                ),
              ],
          ],
        ),
      ],
    );
  }
}

/// Colored chip for the inferred pre-uninstall state:
/// Healthy=green, At-Risk=amber, Frozen=orange, Unknown=grey.
class _StateBadge extends StatelessWidget {
  final String state;
  const _StateBadge({required this.state});

  @override
  Widget build(BuildContext context) {
    final (fg, bg) = switch (state.toLowerCase()) {
      'healthy' => (LgColors.success, LgColors.successBg),
      'at-risk' => (LgColors.warning, LgColors.warningBg),
      'frozen' => (LgColors.riskTwoCycle, LgColors.warningBg),
      _ => (LgColors.textSecondary, LgColors.surfaceSecondary),
    };
    final label = state.isNotEmpty ? state : 'Unknown';

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

String _months(double months) => '${months.toStringAsFixed(1)} mo';

String _percent(double rate) => '${(rate * 100).toStringAsFixed(0)}%';
