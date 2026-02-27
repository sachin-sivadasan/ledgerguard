import 'package:flutter/material.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/dashboard_metrics.dart';

/// Card widget for displaying a KPI metric
class KpiCard extends StatelessWidget {
  final String title;
  final String value;
  final String? subtitle;
  final IconData icon;
  final Color? color;
  final Color? backgroundColor;
  final bool isLarge;

  /// Optional delta indicator to show period-over-period change
  final DeltaIndicator? delta;

  const KpiCard({
    super.key,
    required this.title,
    required this.value,
    this.subtitle,
    required this.icon,
    this.color,
    this.backgroundColor,
    this.isLarge = false,
    this.delta,
  });

  @override
  Widget build(BuildContext context) {
    final effectiveColor = color ?? AppTheme.primary;
    final effectiveBgColor =
        backgroundColor ?? effectiveColor.withOpacity(0.1);

    return Container(
      padding: EdgeInsets.all(isLarge ? 24 : 20),
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
          Row(
            children: [
              Container(
                padding: EdgeInsets.all(isLarge ? 12 : 10),
                decoration: BoxDecoration(
                  color: effectiveBgColor,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Icon(
                  icon,
                  color: effectiveColor,
                  size: isLarge ? 28 : 24,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  title,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: Colors.grey[600],
                        fontWeight: FontWeight.w500,
                      ),
                ),
              ),
            ],
          ),
          SizedBox(height: isLarge ? 20 : 16),
          Row(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Expanded(
                child: Text(
                  value,
                  style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                        fontSize: isLarge ? 32 : 28,
                      ),
                ),
              ),
              if (delta != null) ...[
                const SizedBox(width: 8),
                _DeltaBadge(delta: delta!),
              ],
            ],
          ),
          if (subtitle != null) ...[
            const SizedBox(height: 4),
            Text(
              subtitle!,
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: Colors.grey[500],
                  ),
            ),
          ],
        ],
      ),
    );
  }
}

/// Delta badge showing percentage change
class _DeltaBadge extends StatelessWidget {
  final DeltaIndicator delta;

  const _DeltaBadge({required this.delta});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: delta.color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            delta.icon,
            size: 14,
            color: delta.color,
          ),
          const SizedBox(width: 2),
          Text(
            delta.formattedValue,
            style: TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w600,
              color: delta.color,
            ),
          ),
        ],
      ),
    );
  }
}

/// Compact KPI card for secondary metrics
class KpiCardCompact extends StatelessWidget {
  final String title;
  final String value;
  final IconData icon;
  final Color? color;

  /// Optional delta indicator to show period-over-period change
  final DeltaIndicator? delta;

  const KpiCardCompact({
    super.key,
    required this.title,
    required this.value,
    required this.icon,
    this.color,
    this.delta,
  });

  @override
  Widget build(BuildContext context) {
    final effectiveColor = color ?? AppTheme.primary;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey[200]!),
      ),
      child: Row(
        children: [
          Icon(icon, color: effectiveColor, size: 20),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: Colors.grey[600],
                      ),
                ),
                const SizedBox(height: 2),
                Row(
                  children: [
                    Text(
                      value,
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.bold,
                          ),
                    ),
                    if (delta != null) ...[
                      const SizedBox(width: 8),
                      _DeltaBadgeSmall(delta: delta!),
                    ],
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

/// Small delta badge for compact cards
class _DeltaBadgeSmall extends StatelessWidget {
  final DeltaIndicator delta;

  const _DeltaBadgeSmall({required this.delta});

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(
          delta.icon,
          size: 12,
          color: delta.color,
        ),
        Text(
          delta.formattedValue,
          style: TextStyle(
            fontSize: 11,
            fontWeight: FontWeight.w600,
            color: delta.color,
          ),
        ),
      ],
    );
  }
}
