import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../providers/analytics_provider.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';

class ForecastingTab extends StatelessWidget {
  const ForecastingTab({super.key});

  @override
  Widget build(BuildContext context) {
    final provider = context.watch<AnalyticsProvider>();
    final forecast = provider.forecast;
    final theme = Theme.of(context);
    final monthFmt = DateFormat('MMM');

    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Summary card
          LgResponsive(
            mobile: Column(
              children: [
                LgCard(
                  title: 'Next Month Expected MRR',
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('\$${forecast.first.expected.toStringAsFixed(0)}', style: theme.textTheme.headlineSmall),
                      const SizedBox(height: LgSpacing.s100),
                      Text('Range: \$${forecast.first.pessimistic.toStringAsFixed(0)} – \$${forecast.first.optimistic.toStringAsFixed(0)}', style: theme.textTheme.bodySmall),
                    ],
                  ),
                ),
                const SizedBox(height: LgSpacing.s400),
                LgCard(
                  title: '12-Month Projected MRR',
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('\$${forecast.last.expected.toStringAsFixed(0)}', style: theme.textTheme.headlineSmall),
                      const SizedBox(height: LgSpacing.s100),
                      Text('Range: \$${forecast.last.pessimistic.toStringAsFixed(0)} – \$${forecast.last.optimistic.toStringAsFixed(0)}', style: theme.textTheme.bodySmall),
                    ],
                  ),
                ),
              ],
            ),
            desktop: Row(
              children: [
                Expanded(
                  child: LgCard(
                    title: 'Next Month Expected MRR',
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text('\$${forecast.first.expected.toStringAsFixed(0)}', style: theme.textTheme.headlineSmall),
                        const SizedBox(height: LgSpacing.s100),
                        Text('Range: \$${forecast.first.pessimistic.toStringAsFixed(0)} – \$${forecast.first.optimistic.toStringAsFixed(0)}', style: theme.textTheme.bodySmall),
                      ],
                    ),
                  ),
                ),
                const SizedBox(width: LgSpacing.s400),
                Expanded(
                  child: LgCard(
                    title: '12-Month Projected MRR',
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text('\$${forecast.last.expected.toStringAsFixed(0)}', style: theme.textTheme.headlineSmall),
                        const SizedBox(height: LgSpacing.s100),
                        Text('Range: \$${forecast.last.pessimistic.toStringAsFixed(0)} – \$${forecast.last.optimistic.toStringAsFixed(0)}', style: theme.textTheme.bodySmall),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: LgSpacing.s600),

          // Forecast chart with bands
          LgCard(
            title: '12-Month Revenue Forecast',
            child: SizedBox(
              height: 300,
              child: LineChart(
                LineChartData(
                  gridData: const FlGridData(show: false),
                  titlesData: FlTitlesData(
                    leftTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
                    topTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
                    rightTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
                    bottomTitles: AxisTitles(
                      sideTitles: SideTitles(
                        showTitles: true,
                        interval: 2,
                        getTitlesWidget: (value, meta) {
                          final idx = value.toInt();
                          if (idx >= 0 && idx < forecast.length) {
                            return Padding(
                              padding: const EdgeInsets.only(top: 8),
                              child: Text(monthFmt.format(forecast[idx].date), style: const TextStyle(fontSize: 10)),
                            );
                          }
                          return const SizedBox.shrink();
                        },
                      ),
                    ),
                  ),
                  borderData: FlBorderData(show: false),
                  lineBarsData: [
                    // Optimistic
                    LineChartBarData(
                      spots: forecast.asMap().entries.map((e) => FlSpot(e.key.toDouble(), e.value.optimistic)).toList(),
                      isCurved: true,
                      color: LgColors.success.withValues(alpha: 0.4),
                      barWidth: 1,
                      dotData: const FlDotData(show: false),
                    ),
                    // Expected
                    LineChartBarData(
                      spots: forecast.asMap().entries.map((e) => FlSpot(e.key.toDouble(), e.value.expected)).toList(),
                      isCurved: true,
                      color: LgColors.primary,
                      barWidth: 3,
                      dotData: const FlDotData(show: false),
                    ),
                    // Pessimistic
                    LineChartBarData(
                      spots: forecast.asMap().entries.map((e) => FlSpot(e.key.toDouble(), e.value.pessimistic)).toList(),
                      isCurved: true,
                      color: LgColors.critical.withValues(alpha: 0.4),
                      barWidth: 1,
                      dotData: const FlDotData(show: false),
                    ),
                  ],
                  betweenBarsData: [
                    BetweenBarsData(
                      fromIndex: 0,
                      toIndex: 2,
                      color: LgColors.primary.withValues(alpha: 0.05),
                    ),
                  ],
                ),
              ),
            ),
          ),
          const SizedBox(height: LgSpacing.s200),
          Wrap(
            spacing: LgSpacing.s400,
            children: const [
              _Legend('Optimistic', LgColors.success),
              _Legend('Expected', LgColors.primary),
              _Legend('Pessimistic', LgColors.critical),
            ],
          ),
        ],
      ),
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
