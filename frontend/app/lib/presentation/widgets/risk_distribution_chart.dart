import 'package:flutter/material.dart';

import '../../domain/entities/dashboard_metrics.dart';

/// Chart widget displaying risk distribution breakdown
class RiskDistributionChart extends StatelessWidget {
  final RiskDistribution riskDistribution;

  const RiskDistributionChart({
    super.key,
    required this.riskDistribution,
  });

  static const Color _safeColor = Color(0xFF22C55E);
  static const Color _atRiskColor = Color(0xFFF59E0B);
  static const Color _criticalColor = Color(0xFFEF4444);
  static const Color _churnedColor = Color(0xFF6B7280);

  @override
  Widget build(BuildContext context) {
    final total = riskDistribution.total;
    if (total == 0) {
      return _buildEmptyState(context);
    }

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
            'Risk Distribution',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
          ),
          const SizedBox(height: 20),
          _buildDistributionGrid(context),
        ],
      ),
    );
  }

  Widget _buildEmptyState(BuildContext context) {
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
            'Risk Distribution',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
          ),
          const SizedBox(height: 20),
          Center(
            child: Text(
              'No subscription data available',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Colors.grey[500],
                  ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDistributionGrid(BuildContext context) {
    return Column(
      children: [
        Row(
          children: [
            Expanded(
              child: _buildRiskItem(
                context,
                color: _safeColor,
                label: 'Safe',
                count: riskDistribution.safe,
                percentage: riskDistribution.safePercent,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: _buildRiskItem(
                context,
                color: _atRiskColor,
                label: 'At Risk',
                count: riskDistribution.atRisk,
                percentage: riskDistribution.atRiskPercent,
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        Row(
          children: [
            Expanded(
              child: _buildRiskItem(
                context,
                color: _criticalColor,
                label: 'Critical',
                count: riskDistribution.critical,
                percentage: riskDistribution.criticalPercent,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: _buildRiskItem(
                context,
                color: _churnedColor,
                label: 'Churned',
                count: riskDistribution.churned,
                percentage: riskDistribution.churnedPercent,
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildRiskItem(
    BuildContext context, {
    required Color color,
    required String label,
    required int count,
    required double percentage,
  }) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withOpacity(0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                width: 8,
                height: 8,
                decoration: BoxDecoration(
                  color: color,
                  shape: BoxShape.circle,
                ),
              ),
              const SizedBox(width: 8),
              Text(
                label,
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: Colors.grey[600],
                      fontWeight: FontWeight.w500,
                    ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          FittedBox(
            fit: BoxFit.scaleDown,
            alignment: Alignment.centerLeft,
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Text(
                  count.toString(),
                  style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                        fontWeight: FontWeight.bold,
                        color: color,
                      ),
                ),
                const SizedBox(width: 4),
                Padding(
                  padding: const EdgeInsets.only(bottom: 4),
                  child: Text(
                    '(${percentage.toStringAsFixed(1)}%)',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: Colors.grey[500],
                        ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
