import 'package:flutter/material.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/dashboard_metrics.dart';

/// Chart widget displaying revenue mix breakdown
class RevenueMixChart extends StatelessWidget {
  final RevenueMix revenueMix;

  const RevenueMixChart({
    super.key,
    required this.revenueMix,
  });

  @override
  Widget build(BuildContext context) {
    final total = revenueMix.total;
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
            'Revenue Mix',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
          ),
          const SizedBox(height: 20),
          _buildBar(context),
          const SizedBox(height: 20),
          _buildLegend(context),
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
            'Revenue Mix',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
          ),
          const SizedBox(height: 20),
          Center(
            child: Text(
              'No revenue data available',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Colors.grey[500],
                  ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildBar(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(8),
      child: SizedBox(
        height: 24,
        child: Row(
          children: [
            if (revenueMix.recurringPercent > 0)
              Expanded(
                flex: (revenueMix.recurringPercent * 10).round(),
                child: Container(
                  color: AppTheme.primary,
                ),
              ),
            if (revenueMix.usagePercent > 0)
              Expanded(
                flex: (revenueMix.usagePercent * 10).round(),
                child: Container(
                  color: AppTheme.secondary,
                ),
              ),
            if (revenueMix.oneTimePercent > 0)
              Expanded(
                flex: (revenueMix.oneTimePercent * 10).round(),
                child: Container(
                  color: Colors.orange,
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildLegend(BuildContext context) {
    return Column(
      children: [
        _buildLegendItem(
          context,
          color: AppTheme.primary,
          label: 'Recurring',
          value: _formatCurrency(revenueMix.recurring),
          percentage: revenueMix.recurringPercent,
        ),
        const SizedBox(height: 12),
        _buildLegendItem(
          context,
          color: AppTheme.secondary,
          label: 'Usage',
          value: _formatCurrency(revenueMix.usage),
          percentage: revenueMix.usagePercent,
        ),
        const SizedBox(height: 12),
        _buildLegendItem(
          context,
          color: Colors.orange,
          label: 'One-time',
          value: _formatCurrency(revenueMix.oneTime),
          percentage: revenueMix.oneTimePercent,
        ),
      ],
    );
  }

  Widget _buildLegendItem(
    BuildContext context, {
    required Color color,
    required String label,
    required String value,
    required double percentage,
  }) {
    return Row(
      children: [
        Container(
          width: 12,
          height: 12,
          decoration: BoxDecoration(
            color: color,
            borderRadius: BorderRadius.circular(3),
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Text(
            label,
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: Colors.grey[600],
                ),
          ),
        ),
        Text(
          value,
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                fontWeight: FontWeight.w600,
              ),
        ),
        const SizedBox(width: 8),
        SizedBox(
          width: 50,
          child: Text(
            '${percentage.toStringAsFixed(1)}%',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: Colors.grey[500],
                ),
            textAlign: TextAlign.right,
          ),
        ),
      ],
    );
  }

  String _formatCurrency(int cents) {
    final dollars = cents / 100;
    if (dollars >= 1000000) {
      return '\$${(dollars / 1000000).toStringAsFixed(1)}M';
    } else if (dollars >= 1000) {
      return '\$${(dollars / 1000).toStringAsFixed(0)}K';
    }
    return '\$${dollars.toStringAsFixed(0)}';
  }
}
