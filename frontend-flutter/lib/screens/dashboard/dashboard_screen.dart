import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../core/dashboard_registry.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../providers/apps_provider.dart';
import '../../providers/dashboard_provider.dart';
import '../../providers/earnings_provider.dart';
import '../../providers/organization_provider.dart';
import '../../services/mixpanel_service.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../theme/app_breakpoints.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_metric_card.dart';
import '../../widgets/lg_metric_grid.dart';
import '../../widgets/lg_onboarding_checklist.dart';
import '../../widgets/lg_page.dart';

class DashboardScreen extends StatefulWidget {
  const DashboardScreen({super.key});

  @override
  State<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends State<DashboardScreen>
    with DataLoadingMixin {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) {
        context.read<MixpanelService>().trackDashboardViewed();
        context.read<DashboardProvider>().loadDashboardPreferences();
      }
    });
  }

  @override
  void loadData(String appId) {
    context.read<DashboardProvider>().setSelectedApp(appId);
  }

  @override
  Widget build(BuildContext context) {
    final dp = context.watch<DashboardProvider>();
    final appsProvider = context.watch<AppsProvider>();
    final appsList = appsProvider.apps;
    final hasApps = appsList.isNotEmpty;
    final theme = Theme.of(context);
    final showAppFilter = appsList.length > 1;
    final activeWidgets = dp.secondaryWidgets;

    if (!hasApps) {
      return LgPage(
        title: 'Dashboard',
        subtitle: 'Revenue intelligence overview',
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const LgOnboardingChecklist(),
            const SizedBox(height: LgSpacing.s600),
            LgMetricGrid(
              children: [
                LgMetricCard(label: 'Monthly Recurring Revenue', value: '\$0', icon: Icons.attach_money),
                LgMetricCard(label: 'Renewal Rate', value: '0%', icon: Icons.autorenew),
                LgMetricCard(label: 'Revenue at Risk', value: '\$0', icon: Icons.warning_amber),
                LgMetricCard(label: 'Usage Revenue', value: '\$0', icon: Icons.trending_up),
              ],
            ),
          ],
        ),
      );
    }

    if (dp.error != null) {
      return LgPage(
        title: 'Dashboard',
        subtitle: 'Revenue intelligence overview',
        child: LgErrorState(message: dp.error!, onRetry: retryLoad),
      );
    }

    if (dp.isLoading && dp.mrrCents == 0 && !dp.demoMode) {
      return LgPage(
        title: 'Dashboard',
        subtitle: 'Revenue intelligence overview',
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    return LgPage(
      title: 'Dashboard',
      subtitle: 'Revenue intelligence overview',
      onRefresh: refreshData,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Filters row: App chip + Time range toggle
          Wrap(
            spacing: LgSpacing.s300,
            runSpacing: LgSpacing.s200,
            crossAxisAlignment: WrapCrossAlignment.center,
            children: [
              if (showAppFilter) Builder(builder: (context) {
                final canViewAllApps = context.watch<OrganizationProvider>().canViewAllApps;
                return PopupMenuButton<String?>(
                  onSelected: dp.setSelectedApp,
                  itemBuilder: (_) => [
                    if (canViewAllApps)
                      const PopupMenuItem(
                          value: null, child: Text('All Apps')),
                    ...appsList.map((app) => PopupMenuItem(
                          value: app.id,
                          child: Text(app.name),
                        )),
                  ],
                  child: Chip(
                    label: Text(dp.selectedAppId != null
                        ? appsList
                            .firstWhere(
                                (a) => a.id == dp.selectedAppId,
                                orElse: () => appsList.first)
                            .name
                        : canViewAllApps
                            ? 'All Apps'
                            : appsList.first.name),
                    deleteIcon: canViewAllApps && dp.selectedAppId != null
                        ? const Icon(Icons.close, size: 14)
                        : null,
                    onDeleted: canViewAllApps && dp.selectedAppId != null
                        ? () => dp.setSelectedApp(null)
                        : null,
                  ),
                );
              }),
              SingleChildScrollView(
                scrollDirection: Axis.horizontal,
                child: SegmentedButton<DashboardTimeRange>(
                  segments: [
                    ButtonSegment(
                        value: DashboardTimeRange.thisWeek,
                        label: Text(LgBreakpoints.isMobile(context) ? '1W' : 'This Week')),
                    ButtonSegment(
                        value: DashboardTimeRange.thisMonth,
                        label: Text(LgBreakpoints.isMobile(context) ? '1M' : 'This Month')),
                    ButtonSegment(
                        value: DashboardTimeRange.lastMonth,
                        label: Text(LgBreakpoints.isMobile(context) ? 'Last' : 'Last Month')),
                    ButtonSegment(
                        value: DashboardTimeRange.threeMonths,
                        label: Text(LgBreakpoints.isMobile(context) ? '3M' : '3 Months')),
                  ],
                  selected: {dp.timeRange},
                  onSelectionChanged: (s) => dp.setTimeRange(s.first),
                ),
              ),
            ],
          ),
          const SizedBox(height: LgSpacing.s400),

          // KPI cards (dynamic from preferences)
          LgMetricGrid(
            children: [
              for (final kpiId in dp.primaryKpis)
                if (lookupKpi(kpiId) case final def?)
                  LgMetricCard(
                    label: def.label,
                    value: def.valueGetter(dp),
                    trend: def.trend,
                    trendPositive: def.trendPositive,
                    icon: def.icon,
                  ),
            ],
          ),
          const SizedBox(height: LgSpacing.s600),

          // Row 1: MRR Trend + Risk Distribution (only if active)
          if (activeWidgets.contains('mrr_trend') || activeWidgets.contains('risk_distribution_chart'))
            ...[
              LgResponsive(
                mobile: Column(
                  children: [
                    if (activeWidgets.contains('mrr_trend'))
                      LgCard(
                        title: 'MRR Trend (12 months)',
                        child: SizedBox(height: 200, child: _MrrChart(snapshots: dp.mrrTrend)),
                      ),
                    if (activeWidgets.contains('mrr_trend') && activeWidgets.contains('risk_distribution_chart'))
                      const SizedBox(height: LgSpacing.s400),
                    if (activeWidgets.contains('risk_distribution_chart'))
                      LgCard(
                        title: 'Risk Distribution',
                        child: SizedBox(height: 200, child: _RiskDonut(dist: dp.riskDistribution)),
                      ),
                  ],
                ),
                desktop: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    if (activeWidgets.contains('mrr_trend'))
                      Expanded(
                        flex: 2,
                        child: LgCard(
                          title: 'MRR Trend (12 months)',
                          child: SizedBox(height: 250, child: _MrrChart(snapshots: dp.mrrTrend)),
                        ),
                      ),
                    if (activeWidgets.contains('mrr_trend') && activeWidgets.contains('risk_distribution_chart'))
                      const SizedBox(width: LgSpacing.s400),
                    if (activeWidgets.contains('risk_distribution_chart'))
                      Expanded(
                        child: LgCard(
                          title: 'Risk Distribution',
                          child: SizedBox(height: 250, child: _RiskDonut(dist: dp.riskDistribution)),
                        ),
                      ),
                  ],
                ),
              ),
              const SizedBox(height: LgSpacing.s600),
            ],

          // Row 2: Forecast + Revenue Mix + Weekly Activity + Earnings (only if active)
          if (activeWidgets.contains('forecast') || activeWidgets.contains('revenue_mix_chart') || activeWidgets.contains('weekly_activity') || activeWidgets.contains('earnings_timeline'))
            Builder(builder: (context) {
              final row2Widgets = <Widget>[];
              if (activeWidgets.contains('forecast')) {
                row2Widgets.add(_ForecastCard(dp: dp, theme: theme));
              }
              if (activeWidgets.contains('revenue_mix_chart')) {
                row2Widgets.add(_RevenueMixCard(dp: dp));
              }
              if (activeWidgets.contains('weekly_activity')) {
                row2Widgets.add(_WeeklyActivityCard(title: dp.activityTitle, activity: dp.activity));
              }
              if (activeWidgets.contains('earnings_timeline')) {
                row2Widgets.add(_EarningsOverviewCard(ep: context.watch<EarningsProvider>()));
              }

              return LgResponsive(
                mobile: Column(
                  children: [
                    for (int i = 0; i < row2Widgets.length; i++) ...[
                      row2Widgets[i],
                      if (i < row2Widgets.length - 1) const SizedBox(height: LgSpacing.s400),
                    ],
                  ],
                ),
                desktop: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    for (int i = 0; i < row2Widgets.length; i++) ...[
                      Expanded(child: row2Widgets[i]),
                      if (i < row2Widgets.length - 1) const SizedBox(width: LgSpacing.s400),
                    ],
                  ],
                ),
              );
            }),
        ],
      ),
    );
  }
}

class _MixRow extends StatelessWidget {
  final String label;
  final double pct;
  final Color color;
  const _MixRow(this.label, this.pct, this.color);

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Container(width: 8, height: 8, decoration: BoxDecoration(color: color, shape: BoxShape.circle)),
        const SizedBox(width: LgSpacing.s200),
        Expanded(child: Text(label, style: Theme.of(context).textTheme.bodyMedium)),
        Text('${pct.toStringAsFixed(1)}%', style: Theme.of(context).textTheme.bodySmall),
      ],
    );
  }
}

class _MrrChart extends StatelessWidget {
  final List snapshots;
  const _MrrChart({required this.snapshots});

  @override
  Widget build(BuildContext context) {
    // Sample every 7 days for performance
    final sampled = <FlSpot>[];
    for (int i = 0; i < snapshots.length; i += 7) {
      sampled.add(FlSpot(i.toDouble(), snapshots[i].mrrDollars));
    }

    return LineChart(
      LineChartData(
        gridData: const FlGridData(show: false),
        titlesData: const FlTitlesData(
          leftTitles: AxisTitles(sideTitles: SideTitles(showTitles: false)),
          topTitles: AxisTitles(sideTitles: SideTitles(showTitles: false)),
          rightTitles: AxisTitles(sideTitles: SideTitles(showTitles: false)),
          bottomTitles: AxisTitles(sideTitles: SideTitles(showTitles: false)),
        ),
        borderData: FlBorderData(show: false),
        lineBarsData: [
          LineChartBarData(
            spots: sampled,
            isCurved: true,
            color: LgColors.primary,
            barWidth: 2,
            dotData: const FlDotData(show: false),
            belowBarData: BarAreaData(
              show: true,
              color: LgColors.primary.withValues(alpha: 0.1),
            ),
          ),
        ],
      ),
    );
  }
}

class _RiskDonut extends StatelessWidget {
  final dynamic dist;
  const _RiskDonut({required this.dist});

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Expanded(
          child: PieChart(
            PieChartData(
              sectionsSpace: 2,
              centerSpaceRadius: 40,
              sections: [
                PieChartSectionData(value: dist.safe.toDouble(), color: LgColors.riskSafe, title: '${dist.safe}', titleStyle: const TextStyle(fontSize: 12, color: Colors.white, fontWeight: FontWeight.w600), radius: 30),
                PieChartSectionData(value: dist.oneCycle.toDouble(), color: LgColors.riskOneCycle, title: '${dist.oneCycle}', titleStyle: const TextStyle(fontSize: 12, color: Colors.white, fontWeight: FontWeight.w600), radius: 30),
                PieChartSectionData(value: dist.twoCycle.toDouble(), color: LgColors.riskTwoCycle, title: '${dist.twoCycle}', titleStyle: const TextStyle(fontSize: 12, color: Colors.white, fontWeight: FontWeight.w600), radius: 30),
                PieChartSectionData(value: dist.churned.toDouble(), color: LgColors.riskChurned, title: '${dist.churned}', titleStyle: const TextStyle(fontSize: 12, color: Colors.white, fontWeight: FontWeight.w600), radius: 30),
              ],
            ),
          ),
        ),
        const SizedBox(height: LgSpacing.s200),
        Wrap(
          spacing: LgSpacing.s400,
          children: [
            _Legend('Safe', LgColors.riskSafe),
            _Legend('1 Cycle', LgColors.riskOneCycle),
            _Legend('2 Cycles', LgColors.riskTwoCycle),
            _Legend('Churned', LgColors.riskChurned),
          ],
        ),
      ],
    );
  }
}

class _Legend extends StatelessWidget {
  final String label;
  final Color color;
  const _Legend(this.label, this.color);

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(width: 8, height: 8, decoration: BoxDecoration(color: color, shape: BoxShape.circle)),
        const SizedBox(width: 4),
        Text(label, style: const TextStyle(fontSize: 11)),
      ],
    );
  }
}

class _ForecastCard extends StatelessWidget {
  final DashboardProvider dp;
  final ThemeData theme;
  const _ForecastCard({required this.dp, required this.theme});

  @override
  Widget build(BuildContext context) {
    return LgCard(
      title: 'Forecast (Next Month)',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Expected MRR', style: theme.textTheme.bodySmall),
          const SizedBox(height: LgSpacing.s100),
          Text(
            '\$${dp.nextMonthForecast.expected.toStringAsFixed(0)}',
            style: theme.textTheme.headlineSmall,
          ),
          const SizedBox(height: LgSpacing.s200),
          Text(
            'Range: \$${dp.nextMonthForecast.pessimistic.toStringAsFixed(0)} – \$${dp.nextMonthForecast.optimistic.toStringAsFixed(0)}',
            style: theme.textTheme.bodySmall,
          ),
        ],
      ),
    );
  }
}

class _RevenueMixCard extends StatelessWidget {
  final DashboardProvider dp;
  const _RevenueMixCard({required this.dp});

  @override
  Widget build(BuildContext context) {
    return LgCard(
      title: 'Revenue Mix',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _MixRow('Recurring', dp.revenueMix.recurringPct, LgColors.success),
          const SizedBox(height: LgSpacing.s200),
          _MixRow('Usage', dp.revenueMix.usagePct, LgColors.info),
          const SizedBox(height: LgSpacing.s200),
          _MixRow('One-Time', dp.revenueMix.oneTimePct, LgColors.warning),
        ],
      ),
    );
  }
}

class _WeeklyActivityCard extends StatelessWidget {
  final String title;
  final Map<String, int> activity;
  const _WeeklyActivityCard({required this.title, required this.activity});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return LgCard(
      title: title,
      child: Column(
        children: activity.entries.map((entry) {
          return Padding(
            padding: const EdgeInsets.only(bottom: LgSpacing.s200),
            child: Row(
              children: [
                Expanded(
                  child: Text(entry.key, style: theme.textTheme.bodyMedium),
                ),
                Text(
                  '${entry.value}',
                  style: theme.textTheme.titleSmall
                      ?.copyWith(fontWeight: FontWeight.w600),
                ),
              ],
            ),
          );
        }).toList(),
      ),
    );
  }
}

class _EarningsOverviewCard extends StatelessWidget {
  final EarningsProvider ep;
  const _EarningsOverviewCard({required this.ep});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return LgCard(
      title: 'Earnings Overview',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _EarningsRow('Total Earned', ep.totalEarned, theme),
          const SizedBox(height: LgSpacing.s200),
          _EarningsRow('Pending', ep.pendingAmount, theme),
          const SizedBox(height: LgSpacing.s200),
          _EarningsRow('Available', ep.availableAmount, theme),
        ],
      ),
    );
  }
}

class _EarningsRow extends StatelessWidget {
  final String label;
  final String value;
  final ThemeData theme;
  const _EarningsRow(this.label, this.value, this.theme);

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(child: Text(label, style: theme.textTheme.bodyMedium)),
        Text(value, style: theme.textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w600)),
      ],
    );
  }
}
