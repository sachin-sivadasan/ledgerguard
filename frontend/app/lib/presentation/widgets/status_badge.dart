import 'package:flutter/material.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/user_profile.dart';
import '../../domain/entities/risk_summary.dart';

/// Generic status badge widget
class StatusBadge extends StatelessWidget {
  /// Badge text
  final String label;

  /// Badge color
  final Color color;

  /// Optional icon
  final IconData? icon;

  /// Text size
  final double fontSize;

  /// Icon size
  final double iconSize;

  /// Horizontal padding
  final double horizontalPadding;

  /// Vertical padding
  final double verticalPadding;

  const StatusBadge({
    super.key,
    required this.label,
    required this.color,
    this.icon,
    this.fontSize = 12,
    this.iconSize = 14,
    this.horizontalPadding = 12,
    this.verticalPadding = 4,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: EdgeInsets.symmetric(
        horizontal: horizontalPadding,
        vertical: verticalPadding,
      ),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: color.withOpacity(0.3)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (icon != null) ...[
            Icon(icon, size: iconSize, color: color),
            const SizedBox(width: 4),
          ],
          Text(
            label,
            style: TextStyle(
              color: color,
              fontSize: fontSize,
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }
}

/// Badge for user roles
class RoleBadge extends StatelessWidget {
  final UserRole role;

  const RoleBadge({super.key, required this.role});

  @override
  Widget build(BuildContext context) {
    final color = role == UserRole.owner ? AppTheme.primary : AppTheme.secondary;
    return StatusBadge(
      label: role.displayName,
      color: color,
    );
  }
}

/// Badge for plan tiers
class PlanBadge extends StatelessWidget {
  final PlanTier tier;

  const PlanBadge({super.key, required this.tier});

  @override
  Widget build(BuildContext context) {
    final color = tier.isPro ? AppTheme.warning : Colors.grey;
    final icon = tier.isPro ? Icons.star : Icons.star_border;
    return StatusBadge(
      label: tier.displayName,
      color: color,
      icon: icon,
    );
  }
}

/// Badge for risk levels
class RiskBadge extends StatelessWidget {
  final RiskLevel level;

  const RiskBadge({super.key, required this.level});

  @override
  Widget build(BuildContext context) {
    return StatusBadge(
      label: level.displayName,
      color: level.badgeColor,
    );
  }
}

/// Extension to add color to RiskLevel
extension RiskLevelColor on RiskLevel {
  Color get badgeColor {
    switch (this) {
      case RiskLevel.safe:
        return AppTheme.success;
      case RiskLevel.oneCycleMissed:
        return AppTheme.warning;
      case RiskLevel.twoCyclesMissed:
        return AppTheme.danger;
      case RiskLevel.churned:
        return Colors.grey;
    }
  }
}
