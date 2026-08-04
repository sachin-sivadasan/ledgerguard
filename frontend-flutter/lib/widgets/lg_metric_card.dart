import 'package:flutter/material.dart';
import '../theme/app_colors.dart';
import '../theme/app_spacing.dart';
import 'lg_card.dart';

class LgMetricCard extends StatelessWidget {
  final String label;
  final String value;
  final String? trend;
  final bool? trendPositive;
  final IconData? icon;

  const LgMetricCard({
    super.key,
    required this.label,
    required this.value,
    this.trend,
    this.trendPositive,
    this.icon,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return LgCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              if (icon != null) ...[
                Icon(icon, size: 16, color: LgColors.textSecondary),
                const SizedBox(width: LgSpacing.s200),
              ],
              Flexible(child: Text(label, style: theme.textTheme.bodySmall, overflow: TextOverflow.ellipsis)),
            ],
          ),
          const SizedBox(height: LgSpacing.s200),
          Text(value, style: theme.textTheme.headlineSmall),
          if (trend != null) ...[
            const SizedBox(height: LgSpacing.s100),
            Row(
              children: [
                Icon(
                  // Arrow follows the numeric direction (a leading "-" = down); colour
                  // follows whether the move is GOOD (trendPositive) — so a metric that
                  // dropped for the better shows a green down-arrow, not a green up-arrow.
                  trend!.trimLeft().startsWith('-')
                      ? Icons.arrow_downward
                      : Icons.arrow_upward,
                  size: 14,
                  color: trendPositive == true
                      ? LgColors.success
                      : LgColors.critical,
                ),
                const SizedBox(width: 2),
                Text(
                  trend!,
                  style: TextStyle(
                    fontSize: 12,
                    color: trendPositive == true
                        ? LgColors.success
                        : LgColors.critical,
                  ),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }
}
