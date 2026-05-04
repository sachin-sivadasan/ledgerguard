import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../providers/analytics_provider.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';

class RevenueTab extends StatelessWidget {
  const RevenueTab({super.key});

  @override
  Widget build(BuildContext context) {
    final provider = context.watch<AnalyticsProvider>();
    final mix = provider.revenueMix;
    final movements = provider.mrrMovements;

    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Revenue mix
          LgCard(
            title: 'Revenue Breakdown',
            child: LgBreakpoints.isMobile(context)
                ? Column(
                    children: [
                      _MixBar('Recurring', mix.recurringCents / 100, mix.recurringPct, LgColors.success),
                      const SizedBox(height: LgSpacing.s400),
                      _MixBar('Usage', mix.usageCents / 100, mix.usagePct, LgColors.info),
                      const SizedBox(height: LgSpacing.s400),
                      _MixBar('One-Time', mix.oneTimeCents / 100, mix.oneTimePct, LgColors.warning),
                    ],
                  )
                : Row(
                    children: [
                      _MixBar('Recurring', mix.recurringCents / 100, mix.recurringPct, LgColors.success),
                      const SizedBox(width: LgSpacing.s400),
                      _MixBar('Usage', mix.usageCents / 100, mix.usagePct, LgColors.info),
                      const SizedBox(width: LgSpacing.s400),
                      _MixBar('One-Time', mix.oneTimeCents / 100, mix.oneTimePct, LgColors.warning),
                    ],
                  ),
          ),
          const SizedBox(height: LgSpacing.s600),

          // MRR movement chart
          LgCard(
            title: 'MRR Movement',
            child: Column(
              children: [
                SizedBox(
                  height: 250,
                  child: BarChart(
                    BarChartData(
                      gridData: const FlGridData(show: false),
                      titlesData: FlTitlesData(
                        leftTitles: AxisTitles(
                          sideTitles: SideTitles(
                            showTitles: true,
                            reservedSize: 46,
                            getTitlesWidget: (value, meta) {
                              if (value == meta.min || value == meta.max) {
                                return const SizedBox.shrink();
                              }
                              final label = value >= 0
                                  ? '\$${value.toInt()}'
                                  : '-\$${value.abs().toInt()}';
                              return Text(label, style: const TextStyle(fontSize: 10));
                            },
                          ),
                        ),
                        topTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
                        rightTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
                        bottomTitles: AxisTitles(
                          sideTitles: SideTitles(
                            showTitles: true,
                            getTitlesWidget: (value, meta) {
                              if (value.toInt() < movements.length) {
                                return Padding(
                                  padding: const EdgeInsets.only(top: 8),
                                  child: Text(movements[value.toInt()].month, style: const TextStyle(fontSize: 11)),
                                );
                              }
                              return const SizedBox.shrink();
                            },
                          ),
                        ),
                      ),
                      borderData: FlBorderData(show: false),
                      barGroups: movements.asMap().entries.map((entry) {
                        final m = entry.value;
                        return BarChartGroupData(
                          x: entry.key,
                          barRods: [
                            BarChartRodData(toY: m.newCents / 100, color: LgColors.success, width: 12),
                            BarChartRodData(toY: m.expansionCents / 100, color: LgColors.info, width: 12),
                            BarChartRodData(toY: -m.contractionCents / 100, color: LgColors.warning, width: 12),
                            BarChartRodData(toY: -m.churnedCents / 100, color: LgColors.critical, width: 12),
                          ],
                        );
                      }).toList(),
                    ),
                  ),
                ),
                const SizedBox(height: LgSpacing.s300),
                Wrap(
                  spacing: LgSpacing.s400,
                  children: [
                    _Legend('New', LgColors.success),
                    _Legend('Expansion', LgColors.info),
                    _Legend('Contraction', LgColors.warning),
                    _Legend('Churned', LgColors.critical),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _MixBar extends StatelessWidget {
  final String label;
  final double amount;
  final double pct;
  final Color color;
  const _MixBar(this.label, this.amount, this.pct, this.color);

  @override
  Widget build(BuildContext context) {
    final content = Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: Theme.of(context).textTheme.bodySmall),
        const SizedBox(height: LgSpacing.s100),
        Text('\$${amount.toStringAsFixed(0)}',
            style: TextStyle(fontSize: 18, fontWeight: FontWeight.w600, color: color)),
        const SizedBox(height: LgSpacing.s100),
        LinearProgressIndicator(
          value: pct / 100,
          backgroundColor: color.withValues(alpha: 0.15),
          color: color,
        ),
        const SizedBox(height: LgSpacing.s100),
        Text('${pct.toStringAsFixed(1)}%', style: Theme.of(context).textTheme.bodySmall),
      ],
    );
    // When inside a Row parent, Expanded is needed; in Column, it's not
    if (LgBreakpoints.isMobile(context)) return content;
    return Expanded(child: content);
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
