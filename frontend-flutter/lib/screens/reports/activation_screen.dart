import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../providers/apps_provider.dart';
import '../../providers/activation_provider.dart';
import '../../services/activation_service.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_service_unavailable.dart';

class ActivationScreen extends StatefulWidget {
  const ActivationScreen({super.key});

  @override
  State<ActivationScreen> createState() => _ActivationScreenState();
}

class _ActivationScreenState extends State<ActivationScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<ActivationProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<ActivationProvider>();
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
          'activation-${DateTime.now().toIso8601String().split('T').first}.csv';
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
      debugPrint('activation: CSV export failed: $e');
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
        title: 'Activation',
        breadcrumb: 'Reports › Growth',
        backAction: () => context.go('/reports'),
        child: LgEmptyState(
          icon: Icons.filter_alt_outlined,
          heading: 'No data yet',
          description:
              'Connect your Shopify app to see your install-to-paid funnel.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<ActivationProvider>();

    if (provider.isServiceUnavailable) {
      return LgPage(
        title: 'Activation',
        breadcrumb: 'Reports › Growth',
        backAction: () => context.go('/reports'),
        child: LgServiceUnavailable(onRetry: retryLoad),
      );
    }

    if (provider.error != null) {
      return LgPage(
        title: 'Activation',
        breadcrumb: 'Reports › Growth',
        backAction: () => context.go('/reports'),
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.report == null) {
      return LgPage(
        title: 'Activation',
        breadcrumb: 'Reports › Growth',
        backAction: () => context.go('/reports'),
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final report = provider.report ?? ActivationReport.empty();
    final appsList = appsProvider.apps;
    final showAppFilter = appsList.isNotEmpty;
    final hasData =
        report.installs > 0 || report.started > 0 || report.paid > 0;

    return LgPage(
      title: 'Activation',
      breadcrumb: 'Reports › Growth',
      subtitle:
          'Install-to-paid conversion funnel — how installs become recurring revenue',
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
            if (hasData) const SizedBox(height: LgSpacing.s300),
          ],
          if (hasData) _HeroRow(report: report),
          const SizedBox(height: LgSpacing.s600),
          if (!hasData)
            const LgEmptyState(
              icon: Icons.filter_alt_outlined,
              heading: 'No activation data yet',
              description:
                  'Activation tracks how many installs become recurring-revenue customers. Once your app has install events, subscriptions, and a completed sync, the conversion funnel will appear here.',
            )
          else
            _FunnelCard(report: report),
        ],
      ),
    );
  }
}

// ─── Hero KPI cards ─────────────────────────────────────────────────
class _HeroRow extends StatelessWidget {
  final ActivationReport report;
  const _HeroRow({required this.report});

  @override
  Widget build(BuildContext context) {
    final cards = [
      _KpiCard(
        label: 'Overall Install → Paid',
        value: _percent(report.overallPct),
        color: LgColors.success,
        footnote: '${report.paid} paid of ${report.installs} installs',
      ),
      _KpiCard(
        label: 'Install → Subscription',
        value: _percent(report.installToSubPct),
        color: LgColors.textPrimary,
        footnote: '${report.started} started a subscription',
      ),
      _KpiCard(
        label: 'Subscription → Paid',
        value: _percent(report.subToPaidPct),
        color: LgColors.textPrimary,
        footnote: '${report.paid} reached recurring charge',
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

// ─── Funnel card ────────────────────────────────────────────────────
class _FunnelCard extends StatelessWidget {
  final ActivationReport report;
  const _FunnelCard({required this.report});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    // Stage lookups keyed off the backend contract, with a graceful fallback
    // to the top-level scalar fields so the funnel still renders if `stages`
    // is missing/short.
    ActivationStage? stageFor(String key) {
      for (final s in report.stages) {
        if (s.key == key) return s;
      }
      return null;
    }

    final installs = report.installs;
    final started = report.started;
    final paid = report.paid;

    // Widths are proportional to count / installs, so each bar is narrower
    // than the one above it. Guard against divide-by-zero.
    double frac(int count) {
      if (installs <= 0) return 0;
      return (count / installs).clamp(0.0, 1.0);
    }

    return LgCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Activation funnel', style: theme.textTheme.titleSmall),
          const SizedBox(height: LgSpacing.s100),
          Text(
            'joins install events ↔ subscriptions',
            style: theme.textTheme.bodySmall?.copyWith(
              color: LgColors.textSecondary,
            ),
          ),
          const SizedBox(height: LgSpacing.s400),
          // Stage 1 — Installs (widest, purple).
          _FunnelBar(
            widthFactor: frac(installs),
            fillColor: const Color(0xFFF0F1FF),
            borderColor: const Color(0xFF5C6AC4),
            countColor: const Color(0xFF5C6AC4),
            title: stageFor('installs')?.label ?? 'Installs',
            subtitle: 'Stores that installed the app',
            count: installs,
          ),
          _ConversionLabel(
            text: '↓  ${_percent(report.installToSubPct)} convert',
          ),
          // Stage 2 — Started Subscription (green, narrower).
          _FunnelBar(
            widthFactor: frac(started),
            fillColor: const Color(0xFFECFDF5),
            borderColor: LgColors.success,
            countColor: LgColors.success,
            title: stageFor('started')?.label ?? 'Started Subscription',
            subtitle: 'Selected a plan / began billing',
            count: started,
            note: '${_percent(report.installToSubPct)} of installs',
          ),
          _ConversionLabel(text: '↓  ${_percent(report.subToPaidPct)} convert'),
          // Stage 3 — Paid / Recurring (green, narrowest).
          _FunnelBar(
            widthFactor: frac(paid),
            fillColor: const Color(0xFFECFDF5),
            borderColor: LgColors.success,
            countColor: LgColors.success,
            title: stageFor('paid')?.label ?? 'Paid / Recurring',
            subtitle: 'Reached first recurring charge',
            count: paid,
            note: '${_percent(report.subToPaidPct)} of subs',
          ),
          const SizedBox(height: LgSpacing.s400),
          Text(
            'End-to-end: $installs installs → $paid paid = '
            '${_percent(report.overallPct)} overall activation.',
            style: theme.textTheme.bodySmall?.copyWith(
              color: LgColors.textSecondary,
            ),
          ),
        ],
      ),
    );
  }
}

/// A single horizontal funnel bar whose width is proportional to
/// [widthFactor] (0..1) of the available row width. Left-aligned so the
/// narrowing is visible as a funnel.
class _FunnelBar extends StatelessWidget {
  final double widthFactor;
  final Color fillColor;
  final Color borderColor;
  final Color countColor;
  final String title;
  final String subtitle;
  final int count;
  final String? note;

  const _FunnelBar({
    required this.widthFactor,
    required this.fillColor,
    required this.borderColor,
    required this.countColor,
    required this.title,
    required this.subtitle,
    required this.count,
    this.note,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    // A minimum visible width keeps the narrowest bar readable even when the
    // conversion rate is very low.
    final factor = widthFactor <= 0 ? 0.0 : widthFactor.clamp(0.28, 1.0);

    return Align(
      alignment: Alignment.centerLeft,
      child: FractionallySizedBox(
        widthFactor: factor == 0 ? 1.0 : factor,
        child: Container(
          padding: const EdgeInsets.symmetric(
            horizontal: LgSpacing.s400,
            vertical: LgSpacing.s300,
          ),
          decoration: BoxDecoration(
            color: fillColor,
            border: Border.all(color: borderColor),
            borderRadius: BorderRadius.circular(6),
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.center,
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      title,
                      style: theme.textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w600,
                        color: LgColors.textPrimary,
                      ),
                    ),
                    const SizedBox(height: LgSpacing.s100),
                    Text(
                      subtitle,
                      style: theme.textTheme.bodySmall?.copyWith(
                        color: LgColors.textSecondary,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: LgSpacing.s300),
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    '$count',
                    style: theme.textTheme.headlineSmall?.copyWith(
                      fontWeight: FontWeight.w700,
                      color: countColor,
                    ),
                  ),
                  if (note != null) ...[
                    const SizedBox(height: LgSpacing.s100),
                    Text(
                      note!,
                      style: theme.textTheme.bodySmall?.copyWith(
                        color: LgColors.textSecondary,
                      ),
                    ),
                  ],
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

/// The "↓ N% convert" annotation shown between two funnel bars.
class _ConversionLabel extends StatelessWidget {
  final String text;
  const _ConversionLabel({required this.text});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: LgSpacing.s200),
      child: Center(
        child: Text(
          text,
          style: theme.textTheme.bodySmall?.copyWith(
            color: LgColors.textSecondary,
            fontWeight: FontWeight.w500,
          ),
        ),
      ),
    );
  }
}

String _percent(double rate) => '${(rate * 100).round()}%';
