import 'package:flutter/material.dart';

import '../../domain/entities/subscription.dart';

/// Badge widget showing risk state with appropriate color
class RiskBadge extends StatelessWidget {
  final RiskState riskState;
  final bool isCompact;

  const RiskBadge({
    super.key,
    required this.riskState,
    this.isCompact = false,
  });

  Color get _color {
    switch (riskState) {
      case RiskState.safe:
        return const Color(0xFF22C55E); // Green
      case RiskState.oneCycleMissed:
        return const Color(0xFFEAB308); // Yellow
      case RiskState.twoCyclesMissed:
        return const Color(0xFFF97316); // Orange
      case RiskState.churned:
        return const Color(0xFFEF4444); // Red
    }
  }

  Color get _backgroundColor => _color.withOpacity(0.1);

  IconData get _icon {
    switch (riskState) {
      case RiskState.safe:
        return Icons.check_circle_outline;
      case RiskState.oneCycleMissed:
        return Icons.warning_amber_outlined;
      case RiskState.twoCyclesMissed:
        return Icons.error_outline;
      case RiskState.churned:
        return Icons.cancel_outlined;
    }
  }

  @override
  Widget build(BuildContext context) {
    if (isCompact) {
      return _buildCompact();
    }
    return _buildFull();
  }

  Widget _buildFull() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(
        color: _backgroundColor,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: _color.withOpacity(0.3)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(_icon, size: 16, color: _color),
          const SizedBox(width: 6),
          Text(
            riskState.displayName,
            style: TextStyle(
              color: _color,
              fontSize: 13,
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCompact() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: _backgroundColor,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(
        riskState.displayName,
        style: TextStyle(
          color: _color,
          fontSize: 11,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }
}

/// Larger risk state indicator for detail views
class RiskStateIndicator extends StatelessWidget {
  final RiskState riskState;

  const RiskStateIndicator({
    super.key,
    required this.riskState,
  });

  Color get _color {
    switch (riskState) {
      case RiskState.safe:
        return const Color(0xFF22C55E);
      case RiskState.oneCycleMissed:
        return const Color(0xFFEAB308);
      case RiskState.twoCyclesMissed:
        return const Color(0xFFF97316);
      case RiskState.churned:
        return const Color(0xFFEF4444);
    }
  }

  IconData get _icon {
    switch (riskState) {
      case RiskState.safe:
        return Icons.verified;
      case RiskState.oneCycleMissed:
        return Icons.warning;
      case RiskState.twoCyclesMissed:
        return Icons.error;
      case RiskState.churned:
        return Icons.cancel;
    }
  }

  String get _description {
    switch (riskState) {
      case RiskState.safe:
        return 'Subscription is healthy with recent payments';
      case RiskState.oneCycleMissed:
        return 'One billing cycle missed - monitor closely';
      case RiskState.twoCyclesMissed:
        return 'Two billing cycles missed - action needed';
      case RiskState.churned:
        return 'Subscription has churned';
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: _color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _color.withOpacity(0.3)),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              color: _color.withOpacity(0.2),
              borderRadius: BorderRadius.circular(10),
            ),
            child: Icon(_icon, size: 24, color: _color),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  riskState.displayName,
                  style: TextStyle(
                    color: _color,
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  _description,
                  style: TextStyle(
                    color: Colors.grey[700],
                    fontSize: 13,
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
