import 'package:equatable/equatable.dart';

/// Risk states matching backend risk engine
enum RiskLevel {
  safe('SAFE', 'Safe'),
  oneCycleMissed('ONE_CYCLE_MISSED', 'One Cycle Missed'),
  twoCyclesMissed('TWO_CYCLES_MISSED', 'Two Cycles Missed'),
  churned('CHURNED', 'Churned');

  final String id;
  final String displayName;

  const RiskLevel(this.id, this.displayName);
}

/// Summary of subscription risk distribution
class RiskSummary extends Equatable {
  final int safeCount;
  final int oneCycleMissedCount;
  final int twoCyclesMissedCount;
  final int churnedCount;
  final int revenueAtRiskCents;

  const RiskSummary({
    required this.safeCount,
    required this.oneCycleMissedCount,
    required this.twoCyclesMissedCount,
    required this.churnedCount,
    this.revenueAtRiskCents = 0,
  });

  int get totalSubscriptions =>
      safeCount + oneCycleMissedCount + twoCyclesMissedCount + churnedCount;

  int countFor(RiskLevel state) {
    switch (state) {
      case RiskLevel.safe:
        return safeCount;
      case RiskLevel.oneCycleMissed:
        return oneCycleMissedCount;
      case RiskLevel.twoCyclesMissed:
        return twoCyclesMissedCount;
      case RiskLevel.churned:
        return churnedCount;
    }
  }

  double percentFor(RiskLevel state) {
    if (totalSubscriptions == 0) return 0;
    return (countFor(state) / totalSubscriptions) * 100;
  }

  int get atRiskCount => oneCycleMissedCount + twoCyclesMissedCount;

  double get safePercent => percentFor(RiskLevel.safe);

  String get formattedRevenueAtRisk {
    final dollars = revenueAtRiskCents / 100;
    if (dollars >= 1000000) {
      return '\$${(dollars / 1000000).toStringAsFixed(2)}M';
    } else if (dollars >= 1000) {
      return '\$${(dollars / 1000).toStringAsFixed(1)}K';
    }
    return '\$${dollars.toStringAsFixed(2)}';
  }

  bool get hasData => totalSubscriptions > 0;

  @override
  List<Object?> get props => [
        safeCount,
        oneCycleMissedCount,
        twoCyclesMissedCount,
        churnedCount,
        revenueAtRiskCents,
      ];
}
