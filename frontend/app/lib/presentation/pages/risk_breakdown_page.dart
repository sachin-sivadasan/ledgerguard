import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/risk_summary.dart';
import '../blocs/risk/risk.dart';
import '../widgets/shared.dart';

/// Risk Breakdown page displaying subscription risk distribution
class RiskBreakdownPage extends StatelessWidget {
  const RiskBreakdownPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey[50],
      appBar: AppBar(
        title: const Text('Risk Breakdown'),
        actions: [
          BlocBuilder<RiskBloc, RiskState>(
            builder: (context, state) {
              final isRefreshing = state is RiskLoaded && state.isRefreshing;
              return IconButton(
                icon: isRefreshing
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: Colors.white,
                        ),
                      )
                    : const Icon(Icons.refresh),
                onPressed: isRefreshing
                    ? null
                    : () => context
                        .read<RiskBloc>()
                        .add(const RefreshRiskSummaryRequested()),
              );
            },
          ),
        ],
      ),
      body: BlocBuilder<RiskBloc, RiskState>(
        builder: (context, state) {
          if (state is RiskInitial) {
            context.read<RiskBloc>().add(const LoadRiskSummaryRequested());
            return const Center(child: CircularProgressIndicator());
          }

          if (state is RiskLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state is RiskEmpty) {
            return _buildEmptyState(context, state.message);
          }

          if (state is RiskError) {
            return _buildErrorState(context, state.message);
          }

          if (state is RiskLoaded) {
            return _buildContent(context, state.summary);
          }

          return const SizedBox.shrink();
        },
      ),
    );
  }

  Widget _buildEmptyState(BuildContext context, String message) {
    return EmptyStateWidget(
      title: 'No Risk Data',
      message: message,
      icon: Icons.pie_chart_outline,
    );
  }

  Widget _buildErrorState(BuildContext context, String message) {
    return ErrorStateWidget(
      title: 'Failed to load risk data',
      message: message,
      onRetry: () =>
          context.read<RiskBloc>().add(const LoadRiskSummaryRequested()),
    );
  }

  Widget _buildContent(BuildContext context, RiskSummary summary) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildSummaryCard(context, summary),
          const SizedBox(height: 24),
          _buildChartSection(context, summary),
          const SizedBox(height: 24),
          _buildBreakdownList(context, summary),
        ],
      ),
    );
  }

  Widget _buildSummaryCard(BuildContext context, RiskSummary summary) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Total Subscriptions',
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: Colors.grey[600],
                      ),
                ),
                const SizedBox(height: 4),
                Text(
                  summary.totalSubscriptions.toString(),
                  style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                ),
              ],
            ),
          ),
          Container(
            width: 1,
            height: 50,
            color: Colors.grey[300],
          ),
          Expanded(
            child: Padding(
              padding: const EdgeInsets.only(left: 20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Revenue at Risk',
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                          color: Colors.grey[600],
                        ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    summary.formattedRevenueAtRisk,
                    style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                          fontWeight: FontWeight.bold,
                          color: AppTheme.warning,
                        ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildChartSection(BuildContext context, RiskSummary summary) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Distribution',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
          ),
          const SizedBox(height: 24),
          Center(
            child: SizedBox(
              width: 200,
              height: 200,
              child: CustomPaint(
                painter: _PieChartPainter(summary: summary),
                child: Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(
                        '${summary.safePercent.toStringAsFixed(0)}%',
                        style:
                            Theme.of(context).textTheme.headlineSmall?.copyWith(
                                  fontWeight: FontWeight.bold,
                                  color: _RiskColors.safe,
                                ),
                      ),
                      Text(
                        'Safe',
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: Colors.grey[600],
                            ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
          const SizedBox(height: 24),
          _buildLegend(context),
        ],
      ),
    );
  }

  Widget _buildLegend(BuildContext context) {
    return Wrap(
      spacing: 24,
      runSpacing: 12,
      alignment: WrapAlignment.center,
      children: [
        _buildLegendItem(context, _RiskColors.safe, 'Safe'),
        _buildLegendItem(context, _RiskColors.oneCycle, 'One Cycle Missed'),
        _buildLegendItem(context, _RiskColors.twoCycles, 'Two Cycles Missed'),
        _buildLegendItem(context, _RiskColors.churned, 'Churned'),
      ],
    );
  }

  Widget _buildLegendItem(BuildContext context, Color color, String label) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 12,
          height: 12,
          decoration: BoxDecoration(
            color: color,
            shape: BoxShape.circle,
          ),
        ),
        const SizedBox(width: 6),
        Text(
          label,
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Colors.grey[700],
              ),
        ),
      ],
    );
  }

  Widget _buildBreakdownList(BuildContext context, RiskSummary summary) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Breakdown by State',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
          ),
          const SizedBox(height: 16),
          _buildRiskRow(
            context,
            color: _RiskColors.safe,
            label: 'SAFE',
            description: 'Active and healthy subscriptions',
            count: summary.safeCount,
            percent: summary.percentFor(RiskLevel.safe),
          ),
          const Divider(height: 24),
          _buildRiskRow(
            context,
            color: _RiskColors.oneCycle,
            label: 'ONE_CYCLE_MISSED',
            description: 'Missed 1 billing cycle (31-60 days)',
            count: summary.oneCycleMissedCount,
            percent: summary.percentFor(RiskLevel.oneCycleMissed),
          ),
          const Divider(height: 24),
          _buildRiskRow(
            context,
            color: _RiskColors.twoCycles,
            label: 'TWO_CYCLES_MISSED',
            description: 'Missed 2 billing cycles (61-90 days)',
            count: summary.twoCyclesMissedCount,
            percent: summary.percentFor(RiskLevel.twoCyclesMissed),
          ),
          const Divider(height: 24),
          _buildRiskRow(
            context,
            color: _RiskColors.churned,
            label: 'CHURNED',
            description: 'Inactive for 90+ days',
            count: summary.churnedCount,
            percent: summary.percentFor(RiskLevel.churned),
          ),
        ],
      ),
    );
  }

  Widget _buildRiskRow(
    BuildContext context, {
    required Color color,
    required String label,
    required String description,
    required int count,
    required double percent,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Container(
              width: 10,
              height: 10,
              decoration: BoxDecoration(
                color: color,
                shape: BoxShape.circle,
              ),
            ),
            const SizedBox(width: 8),
            Text(
              label,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                    fontFamily: 'monospace',
                  ),
            ),
            const Spacer(),
            Text(
              count.toString(),
              style: Theme.of(context).textTheme.titleLarge?.copyWith(
                    fontWeight: FontWeight.bold,
                    color: color,
                  ),
            ),
            const SizedBox(width: 8),
            Text(
              '(${percent.toStringAsFixed(1)}%)',
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: Colors.grey[500],
                  ),
            ),
          ],
        ),
        const SizedBox(height: 8),
        Text(
          description,
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Colors.grey[600],
              ),
        ),
        const SizedBox(height: 8),
        ClipRRect(
          borderRadius: BorderRadius.circular(4),
          child: LinearProgressIndicator(
            value: percent / 100,
            backgroundColor: Colors.grey[200],
            valueColor: AlwaysStoppedAnimation<Color>(color),
            minHeight: 6,
          ),
        ),
      ],
    );
  }
}

/// Colors for risk states
class _RiskColors {
  static const Color safe = Color(0xFF22C55E);
  static const Color oneCycle = Color(0xFFF59E0B);
  static const Color twoCycles = Color(0xFFEF4444);
  static const Color churned = Color(0xFF6B7280);
}

/// Custom painter for pie chart
class _PieChartPainter extends CustomPainter {
  final RiskSummary summary;

  _PieChartPainter({required this.summary});

  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);
    final radius = math.min(size.width, size.height) / 2;
    final innerRadius = radius * 0.6;

    final total = summary.totalSubscriptions;
    if (total == 0) return;

    final segments = [
      _Segment(_RiskColors.safe, summary.safeCount / total),
      _Segment(_RiskColors.oneCycle, summary.oneCycleMissedCount / total),
      _Segment(_RiskColors.twoCycles, summary.twoCyclesMissedCount / total),
      _Segment(_RiskColors.churned, summary.churnedCount / total),
    ];

    double startAngle = -math.pi / 2;

    for (final segment in segments) {
      if (segment.fraction == 0) continue;

      final sweepAngle = segment.fraction * 2 * math.pi;
      final paint = Paint()
        ..color = segment.color
        ..style = PaintingStyle.fill;

      final path = Path()
        ..moveTo(
          center.dx + innerRadius * math.cos(startAngle),
          center.dy + innerRadius * math.sin(startAngle),
        )
        ..lineTo(
          center.dx + radius * math.cos(startAngle),
          center.dy + radius * math.sin(startAngle),
        )
        ..arcTo(
          Rect.fromCircle(center: center, radius: radius),
          startAngle,
          sweepAngle,
          false,
        )
        ..lineTo(
          center.dx + innerRadius * math.cos(startAngle + sweepAngle),
          center.dy + innerRadius * math.sin(startAngle + sweepAngle),
        )
        ..arcTo(
          Rect.fromCircle(center: center, radius: innerRadius),
          startAngle + sweepAngle,
          -sweepAngle,
          false,
        )
        ..close();

      canvas.drawPath(path, paint);
      startAngle += sweepAngle;
    }
  }

  @override
  bool shouldRepaint(covariant _PieChartPainter oldDelegate) {
    return oldDelegate.summary != summary;
  }
}

class _Segment {
  final Color color;
  final double fraction;

  _Segment(this.color, this.fraction);
}
