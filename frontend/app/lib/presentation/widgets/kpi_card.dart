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

    return LayoutBuilder(
      builder: (context, constraints) {
        // Responsive sizing based on card width
        final isCompact = constraints.maxWidth < 200;
        final iconSize = isCompact ? 20.0 : (isLarge ? 28.0 : 24.0);
        final iconPadding = isCompact ? 8.0 : (isLarge ? 12.0 : 10.0);
        final cardPadding = isCompact ? 12.0 : (isLarge ? 24.0 : 20.0);
        final valueFontSize = isCompact ? 20.0 : (isLarge ? 28.0 : 24.0);

        return Container(
          padding: EdgeInsets.all(cardPadding),
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
                    padding: EdgeInsets.all(iconPadding),
                    decoration: BoxDecoration(
                      color: effectiveBgColor,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Icon(
                      icon,
                      color: effectiveColor,
                      size: iconSize,
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Text(
                      title,
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: Colors.grey[600],
                            fontWeight: FontWeight.w500,
                            fontSize: isCompact ? 12 : null,
                          ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                ],
              ),
              SizedBox(height: isCompact ? 12 : (isLarge ? 20 : 16)),
              Row(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Expanded(
                    child: FittedBox(
                      fit: BoxFit.scaleDown,
                      alignment: Alignment.centerLeft,
                      child: Text(
                        value,
                        style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                              fontWeight: FontWeight.bold,
                              fontSize: valueFontSize,
                            ),
                      ),
                    ),
                  ),
                  if (delta != null) ...[
                    const SizedBox(width: 6),
                    _DeltaBadge(delta: delta!, isCompact: isCompact),
                  ],
                ],
              ),
              if (subtitle != null) ...[
                const SizedBox(height: 4),
                Text(
                  subtitle!,
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: Colors.grey[500],
                        fontSize: isCompact ? 10 : null,
                      ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ],
          ),
        );
      },
    );
  }
}

/// Delta badge showing percentage change
class _DeltaBadge extends StatelessWidget {
  final DeltaIndicator delta;
  final bool isCompact;

  const _DeltaBadge({required this.delta, this.isCompact = false});

  @override
  Widget build(BuildContext context) {
    final padding = isCompact
        ? const EdgeInsets.symmetric(horizontal: 4, vertical: 2)
        : const EdgeInsets.symmetric(horizontal: 8, vertical: 4);
    final iconSize = isCompact ? 10.0 : 14.0;
    final fontSize = isCompact ? 9.0 : 12.0;

    return Container(
      padding: padding,
      decoration: BoxDecoration(
        color: delta.color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            delta.icon,
            size: iconSize,
            color: delta.color,
          ),
          const SizedBox(width: 2),
          Text(
            delta.formattedValue,
            style: TextStyle(
              fontSize: fontSize,
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
                    Flexible(
                      child: Text(
                        value,
                        style: Theme.of(context).textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.bold,
                            ),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    if (delta != null) ...[
                      const SizedBox(width: 6),
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
