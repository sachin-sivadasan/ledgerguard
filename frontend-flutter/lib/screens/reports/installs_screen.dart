import 'package:dio/dio.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../providers/apps_provider.dart';
import '../../providers/installs_provider.dart';
import '../../services/installs_service.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_service_unavailable.dart';

class InstallsScreen extends StatefulWidget {
  const InstallsScreen({super.key});

  @override
  State<InstallsScreen> createState() => _InstallsScreenState();
}

class _InstallsScreenState extends State<InstallsScreen> with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<InstallsProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<InstallsProvider>();
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
          'installs-${DateTime.now().toIso8601String().split('T').first}.csv';
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
      debugPrint('installs: CSV export failed: $e');
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
        title: 'Installs',
        breadcrumb: 'Reports › Growth',
        backAction: () => context.go('/reports'),
        child: LgEmptyState(
          icon: Icons.download_outlined,
          heading: 'No data yet',
          description: 'Connect your Shopify app to see install activity.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<InstallsProvider>();

    if (provider.isServiceUnavailable) {
      return LgPage(
        title: 'Installs',
        breadcrumb: 'Reports › Growth',
        backAction: () => context.go('/reports'),
        child: LgServiceUnavailable(onRetry: retryLoad),
      );
    }

    if (provider.error != null) {
      return LgPage(
        title: 'Installs',
        breadcrumb: 'Reports › Growth',
        backAction: () => context.go('/reports'),
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.report == null) {
      return LgPage(
        title: 'Installs',
        breadcrumb: 'Reports › Growth',
        backAction: () => context.go('/reports'),
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final report = provider.report ?? InstallsReport.empty();
    final appsList = appsProvider.apps;
    final showAppFilter = appsList.length > 1;
    final hasData = report.events.isNotEmpty ||
        report.trend.isNotEmpty ||
        report.installs > 0 ||
        report.uninstalls > 0;

    return LgPage(
      title: 'Installs',
      breadcrumb: 'Reports › Growth',
      subtitle: 'Install and uninstall activity — net app adoption over the period',
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
              icon: Icons.download_outlined,
              heading: 'No install activity yet',
              description:
                  'Once Shopify records install and uninstall events for your app, the trend and recent events will appear here.',
            )
          else ...[
            _HeroRow(report: report),
            const SizedBox(height: LgSpacing.s600),
            _TrendCard(report: report),
            const SizedBox(height: LgSpacing.s600),
            _EventsTable(report: report),
            const SizedBox(height: LgSpacing.s400),
            Text(
              'From RELATIONSHIP_INSTALLED / RELATIONSHIP_UNINSTALLED events. Net = installs − uninstalls.',
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
  final InstallsReport report;
  const _HeroRow({required this.report});

  @override
  Widget build(BuildContext context) {
    final netUp = report.net >= 0;
    final cards = <Widget>[
      _KpiCard(
        label: 'Installs',
        value: '${report.installs}',
        color: LgColors.success,
        footnote: 'new stores installed the app',
      ),
      _KpiCard(
        label: 'Uninstalls',
        value: '${report.uninstalls}',
        color: LgColors.critical,
        footnote: 'stores removed the app',
      ),
      _KpiCard(
        label: 'Net',
        value: '${netUp ? '+' : ''}${report.net}',
        color: netUp ? LgColors.success : LgColors.critical,
        footnote: 'installs − uninstalls',
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

// ─── Two-line trend card (installs green / uninstalls red) ──────────
class _TrendCard extends StatelessWidget {
  final InstallsReport report;
  const _TrendCard({required this.report});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final trend = report.trend;
    if (trend.isEmpty) {
      return LgCard(
        title: 'Installs vs Uninstalls — trend',
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: LgSpacing.s600),
          child: Center(
            child: Text('Trend data not yet available for this range.',
                style: theme.textTheme.bodySmall
                    ?.copyWith(color: LgColors.textSecondary)),
          ),
        ),
      );
    }

    final installSpots = <FlSpot>[
      for (var i = 0; i < trend.length; i++)
        FlSpot(i.toDouble(), trend[i].installs.toDouble()),
    ];
    final uninstallSpots = <FlSpot>[
      for (var i = 0; i < trend.length; i++)
        FlSpot(i.toDouble(), trend[i].uninstalls.toDouble()),
    ];
    final maxY = [
      ...installSpots.map((s) => s.y),
      ...uninstallSpots.map((s) => s.y),
    ].fold<double>(0, (a, b) => a > b ? a : b);

    return LgCard(
      title: 'Installs vs Uninstalls — trend',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              _LegendDot(color: LgColors.success, label: 'Installs'),
              const SizedBox(width: LgSpacing.s400),
              _LegendDot(color: LgColors.critical, label: 'Uninstalls'),
            ],
          ),
          const SizedBox(height: LgSpacing.s200),
          SizedBox(
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
                      reservedSize: 32,
                      getTitlesWidget: (value, meta) {
                        if (value == meta.min || value == meta.max) {
                          return const SizedBox.shrink();
                        }
                        return Text(value.toInt().toString(),
                            style: const TextStyle(fontSize: 10));
                      },
                    ),
                  ),
                  bottomTitles: AxisTitles(
                    sideTitles: SideTitles(
                      showTitles: true,
                      interval:
                          (trend.length / 4).clamp(1, trend.length).toDouble(),
                      getTitlesWidget: (value, meta) {
                        final i = value.toInt();
                        if (i < 0 || i >= trend.length) {
                          return const SizedBox.shrink();
                        }
                        return Padding(
                          padding: const EdgeInsets.only(top: 6),
                          child: Text(DateFormat('MMM d').format(trend[i].date),
                              style: const TextStyle(fontSize: 10)),
                        );
                      },
                    ),
                  ),
                ),
                lineBarsData: [
                  _line(installSpots, LgColors.success),
                  _line(uninstallSpots, LgColors.critical),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  LineChartBarData _line(List<FlSpot> spots, Color color) => LineChartBarData(
        spots: spots,
        isCurved: true,
        color: color,
        barWidth: 2,
        dotData: const FlDotData(show: false),
      );
}

class _LegendDot extends StatelessWidget {
  final Color color;
  final String label;
  const _LegendDot({required this.color, required this.label});

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(width: 10, height: 10, color: color),
        const SizedBox(width: LgSpacing.s100),
        Text(label,
            style: Theme.of(context)
                .textTheme
                .bodySmall
                ?.copyWith(color: LgColors.textSecondary)),
      ],
    );
  }
}

// ─── Recent events table ────────────────────────────────────────────
class _EventsTable extends StatelessWidget {
  final InstallsReport report;
  const _EventsTable({required this.report});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    // The table is capped server-side; installs + uninstalls is the true in-window
    // total, so surface "showing latest N of M" rather than silently truncating.
    final total = report.installs + report.uninstalls;
    final shown = report.events.length;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Recent install / uninstall events',
            style: theme.textTheme.titleMedium),
        if (total > shown) ...[
          const SizedBox(height: LgSpacing.s100),
          Text('Showing the latest $shown of $total events',
              style: theme.textTheme.bodySmall
                  ?.copyWith(color: LgColors.textSecondary)),
        ],
        const SizedBox(height: LgSpacing.s300),
        // Column header row: STORE | EVENT | DATE
        Padding(
          padding: const EdgeInsets.symmetric(
              horizontal: LgSpacing.s400, vertical: LgSpacing.s200),
          child: Row(
            children: [
              Expanded(
                flex: 4,
                child: Text('STORE',
                    style: theme.textTheme.bodySmall
                        ?.copyWith(color: LgColors.textSecondary)),
              ),
              Expanded(
                flex: 2,
                child: Text('EVENT',
                    style: theme.textTheme.bodySmall
                        ?.copyWith(color: LgColors.textSecondary)),
              ),
              Expanded(
                flex: 2,
                child: Text('DATE',
                    textAlign: TextAlign.end,
                    style: theme.textTheme.bodySmall
                        ?.copyWith(color: LgColors.textSecondary)),
              ),
            ],
          ),
        ),
        ...report.events.map((e) => Padding(
              padding: const EdgeInsets.only(bottom: LgSpacing.s200),
              child: _EventRow(event: e),
            )),
      ],
    );
  }
}

class _EventRow extends StatelessWidget {
  final InstallEvent event;
  const _EventRow({required this.event});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final dt = event.dateTime;
    final dateLabel = dt != null
        ? DateFormat('MMM d, yyyy').format(dt)
        : (event.date.isNotEmpty ? event.date : '—');
    final domain = event.domain.isNotEmpty ? event.domain : '—';

    return LgCard(
      child: Row(
        children: [
          Expanded(
            flex: 4,
            child: Text(domain, style: theme.textTheme.titleSmall),
          ),
          Expanded(
            flex: 2,
            child: Align(
              alignment: Alignment.centerLeft,
              child: _EventChip(event: event.event),
            ),
          ),
          Expanded(
            flex: 2,
            child: Text(dateLabel,
                textAlign: TextAlign.end,
                style: theme.textTheme.bodySmall
                    ?.copyWith(color: LgColors.textSecondary)),
          ),
        ],
      ),
    );
  }
}

/// Event pill — green for Install, amber for Uninstall.
class _EventChip extends StatelessWidget {
  final String event;
  const _EventChip({required this.event});

  @override
  Widget build(BuildContext context) {
    final (bg, fg) = switch (event.toLowerCase()) {
      'install' => (LgColors.success.withValues(alpha: 0.14), LgColors.success),
      'uninstall' => (LgColors.warning.withValues(alpha: 0.14), LgColors.warning),
      _ => (LgColors.surfaceSecondary, LgColors.textSecondary),
    };
    final label = event.isNotEmpty ? event : '—';
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration:
          BoxDecoration(color: bg, borderRadius: BorderRadius.circular(10)),
      child: Text(label,
          style: TextStyle(fontSize: 11, color: fg, fontWeight: FontWeight.w600)),
    );
  }
}
