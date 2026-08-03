import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../providers/analytics_provider.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';

class ProfitTab extends StatelessWidget {
  const ProfitTab({super.key});

  @override
  Widget build(BuildContext context) {
    final provider = context.watch<AnalyticsProvider>();
    final expenses = provider.expenses;
    final theme = Theme.of(context);

    if (expenses.isEmpty) {
      return const LgEmptyState(
        icon: Icons.account_balance,
        heading: 'Profit data not yet available',
        description:
            'Profit and expense breakdown will appear once transactions are processed.',
      );
    }

    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Summary
          LgResponsive(
            mobile: Column(
              children: [
                LgCard(title: 'Average Profit Margin', child: Text('${provider.avgProfitMargin.toStringAsFixed(1)}%', style: theme.textTheme.headlineSmall)),
                const SizedBox(height: LgSpacing.s400),
                LgCard(title: 'This Month Net Profit', child: Text('\$${(expenses.last.netProfitCents / 100).toStringAsFixed(0)}', style: theme.textTheme.headlineSmall)),
              ],
            ),
            desktop: Row(
              children: [
                Expanded(child: LgCard(title: 'Average Profit Margin', child: Text('${provider.avgProfitMargin.toStringAsFixed(1)}%', style: theme.textTheme.headlineSmall))),
                const SizedBox(width: LgSpacing.s400),
                Expanded(child: LgCard(title: 'This Month Net Profit', child: Text('\$${(expenses.last.netProfitCents / 100).toStringAsFixed(0)}', style: theme.textTheme.headlineSmall))),
              ],
            ),
          ),
          const SizedBox(height: LgSpacing.s600),

          // P&L breakdown chart
          LgCard(
            title: 'Monthly P&L',
            child: SizedBox(
              height: 280,
              child: BarChart(
                BarChartData(
                  gridData: const FlGridData(show: false),
                  titlesData: FlTitlesData(
                    leftTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
                    topTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
                    rightTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
                    bottomTitles: AxisTitles(
                      sideTitles: SideTitles(
                        showTitles: true,
                        getTitlesWidget: (value, meta) {
                          if (value.toInt() < expenses.length) {
                            return Padding(
                              padding: const EdgeInsets.only(top: 8),
                              child: Text(expenses[value.toInt()].month, style: const TextStyle(fontSize: 11)),
                            );
                          }
                          return const SizedBox.shrink();
                        },
                      ),
                    ),
                  ),
                  borderData: FlBorderData(show: false),
                  barGroups: expenses.asMap().entries.map((entry) {
                    final e = entry.value;
                    return BarChartGroupData(
                      x: entry.key,
                      barRods: [
                        BarChartRodData(
                          toY: e.grossRevenueCents / 100,
                          width: 20,
                          rodStackItems: [
                            BarChartRodStackItem(0, e.netProfitCents / 100, LgColors.success),
                            BarChartRodStackItem(e.netProfitCents / 100, (e.netProfitCents + e.shopifyCutCents) / 100, LgColors.warning),
                            BarChartRodStackItem(
                              (e.netProfitCents + e.shopifyCutCents) / 100,
                              (e.netProfitCents + e.shopifyCutCents + e.infraCostCents + e.paymentFeesCents) / 100,
                              LgColors.critical.withValues(alpha: 0.6),
                            ),
                          ],
                          color: Colors.transparent,
                          borderRadius: BorderRadius.circular(4),
                        ),
                      ],
                    );
                  }).toList(),
                ),
              ),
            ),
          ),
          const SizedBox(height: LgSpacing.s200),
          Wrap(
            spacing: LgSpacing.s400,
            children: const [
              _Legend('Net Profit', LgColors.success),
              _Legend('Shopify Cut', LgColors.warning),
              _Legend('Infra + Fees', LgColors.critical),
            ],
          ),
          const SizedBox(height: LgSpacing.s600),

          // Expense table
          LgCard(
            title: 'Expense Details',
            child: LgBreakpoints.isMobile(context)
                ? SingleChildScrollView(
                    scrollDirection: Axis.horizontal,
                    child: _buildExpenseTable(expenses),
                  )
                : _buildExpenseTable(expenses),
          ),
        ],
      ),
    );
  }

  Widget _buildExpenseTable(List expenses) {
    return Table(
      columnWidths: const {
        0: FlexColumnWidth(1),
        1: FlexColumnWidth(1),
        2: FlexColumnWidth(1),
        3: FlexColumnWidth(1),
        4: FlexColumnWidth(1),
        5: FlexColumnWidth(1.2),
      },
      children: [
        TableRow(
          decoration: BoxDecoration(color: LgColors.surfaceSecondary),
          children: ['Month', 'Gross', 'Shopify Cut', 'Net Profit', 'Fee Guard']
              .map((h) => Padding(
                    padding: const EdgeInsets.all(8),
                    child: Text(h, style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: LgColors.textSecondary)),
                  ))
              .toList(),
        ),
        ...expenses.map((e) => TableRow(
              children: [
                _Cell(e.month),
                _Cell('\$${(e.grossRevenueCents / 100).toStringAsFixed(0)}'),
                _Cell('\$${(e.shopifyCutCents / 100).toStringAsFixed(0)}'),
                _Cell('\$${(e.netProfitCents / 100).toStringAsFixed(0)}'),
                _FeeGuardCell(ok: e.feeGuardOk, varianceCents: e.feeVarianceCents),
              ],
            )),
      ],
    );
  }
}

/// Fee Guard status cell: ✓ when Shopify's retained cut matches the app's expected
/// revenue-share tier, else a warning with the signed variance ("Shopify may have
/// charged the wrong rate").
class _FeeGuardCell extends StatelessWidget {
  final bool ok;
  final int varianceCents;
  const _FeeGuardCell({required this.ok, required this.varianceCents});

  @override
  Widget build(BuildContext context) {
    final dollars = (varianceCents.abs() / 100).toStringAsFixed(0);
    final sign = varianceCents >= 0 ? '+' : '−';
    return Padding(
      padding: const EdgeInsets.all(8),
      child: ok
          ? const Row(mainAxisSize: MainAxisSize.min, children: [
              Icon(Icons.verified_outlined, size: 15, color: LgColors.success),
              SizedBox(width: 4),
              Text('OK', style: TextStyle(fontSize: 13, color: LgColors.success)),
            ])
          : Row(mainAxisSize: MainAxisSize.min, children: [
              const Icon(Icons.warning_amber_rounded, size: 15, color: LgColors.warning),
              const SizedBox(width: 4),
              Text('$sign\$$dollars', style: const TextStyle(fontSize: 13, color: LgColors.warning)),
            ]),
    );
  }
}

class _Cell extends StatelessWidget {
  final String text;
  const _Cell(this.text);
  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.all(8),
        child: Text(text, style: const TextStyle(fontSize: 13)),
      );
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
