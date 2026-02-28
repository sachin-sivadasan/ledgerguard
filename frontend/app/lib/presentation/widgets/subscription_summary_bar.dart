import 'package:flutter/material.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/subscription_filter.dart';

/// A horizontal bar displaying subscription summary statistics
class SubscriptionSummaryBar extends StatelessWidget {
  final SubscriptionSummary summary;
  final bool isLoading;

  const SubscriptionSummaryBar({
    super.key,
    required this.summary,
    this.isLoading = false,
  });

  String _formatCurrency(int cents) {
    final dollars = cents / 100;
    if (dollars >= 1000) {
      return '\$${(dollars / 1000).toStringAsFixed(1)}k';
    }
    return '\$${dollars.toStringAsFixed(0)}';
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        children: [
          _SummaryCard(
            icon: Icons.check_circle_outline,
            label: 'Active',
            value: summary.activeCount.toString(),
            color: AppTheme.success,
            isLoading: isLoading,
          ),
          const SizedBox(width: 12),
          _SummaryCard(
            icon: Icons.warning_amber_outlined,
            label: 'At Risk',
            value: summary.atRiskCount.toString(),
            color: AppTheme.warning,
            isLoading: isLoading,
          ),
          const SizedBox(width: 12),
          _SummaryCard(
            icon: Icons.cancel_outlined,
            label: 'Churned',
            value: summary.churnedCount.toString(),
            color: AppTheme.danger,
            isLoading: isLoading,
          ),
          const SizedBox(width: 12),
          _SummaryCard(
            icon: Icons.attach_money,
            label: 'Avg Price',
            value: _formatCurrency(summary.avgPriceCents),
            color: AppTheme.primary,
            isLoading: isLoading,
          ),
        ],
      ),
    );
  }
}

class _SummaryCard extends StatelessWidget {
  final IconData icon;
  final String label;
  final String value;
  final Color color;
  final bool isLoading;

  const _SummaryCard({
    required this.icon,
    required this.label,
    required this.value,
    required this.color,
    this.isLoading = false,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      constraints: const BoxConstraints(minWidth: 120),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withOpacity(0.2)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: color.withOpacity(0.1),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Icon(icon, color: color, size: 20),
          ),
          const SizedBox(width: 12),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                label,
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: Colors.grey[600],
                      fontWeight: FontWeight.w500,
                    ),
              ),
              const SizedBox(height: 2),
              if (isLoading)
                SizedBox(
                  width: 40,
                  height: 20,
                  child: LinearProgressIndicator(
                    backgroundColor: Colors.grey[200],
                    valueColor: AlwaysStoppedAnimation(color.withOpacity(0.5)),
                  ),
                )
              else
                Text(
                  value,
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                        color: color,
                      ),
                ),
            ],
          ),
        ],
      ),
    );
  }
}
