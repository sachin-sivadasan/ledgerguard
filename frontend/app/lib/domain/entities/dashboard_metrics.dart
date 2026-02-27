import 'package:equatable/equatable.dart';
import 'package:flutter/material.dart';

import 'time_range.dart';

/// Primary KPI metrics for the executive dashboard
class DashboardMetrics extends Equatable {
  /// Renewal success rate as percentage (0-100)
  final double renewalSuccessRate;

  /// Active Monthly Recurring Revenue in cents
  final int activeMrr;

  /// Revenue at risk in cents
  final int revenueAtRisk;

  /// Churned revenue in cents
  final int churnedRevenue;

  /// Number of churned subscriptions
  final int churnedCount;

  /// Usage-based revenue in cents
  final int usageRevenue;

  /// Total revenue in cents (recurring + usage + one-time - refunds)
  final int totalRevenue;

  /// Revenue mix breakdown
  final RevenueMix revenueMix;

  /// Risk distribution
  final RiskDistribution riskDistribution;

  /// Delta changes from previous period (nullable if no comparison data)
  final MetricsDelta? delta;

  /// Time range for this metrics data
  final TimeRange? timeRange;

  const DashboardMetrics({
    required this.renewalSuccessRate,
    required this.activeMrr,
    required this.revenueAtRisk,
    required this.churnedRevenue,
    required this.churnedCount,
    required this.usageRevenue,
    required this.totalRevenue,
    required this.revenueMix,
    required this.riskDistribution,
    this.delta,
    this.timeRange,
  });

  /// Format MRR as currency string
  String get formattedMrr => _formatCurrency(activeMrr);

  /// Format revenue at risk as currency string
  String get formattedRevenueAtRisk => _formatCurrency(revenueAtRisk);

  /// Format churned revenue as currency string
  String get formattedChurnedRevenue => _formatCurrency(churnedRevenue);

  /// Format usage revenue as currency string
  String get formattedUsageRevenue => _formatCurrency(usageRevenue);

  /// Format total revenue as currency string
  String get formattedTotalRevenue => _formatCurrency(totalRevenue);

  /// Format renewal rate as percentage string
  String get formattedRenewalRate => '${renewalSuccessRate.toStringAsFixed(1)}%';

  String _formatCurrency(int cents) {
    final dollars = cents / 100;
    if (dollars >= 1000000) {
      return '\$${(dollars / 1000000).toStringAsFixed(2)}M';
    } else if (dollars >= 1000) {
      return '\$${(dollars / 1000).toStringAsFixed(1)}K';
    }
    return '\$${dollars.toStringAsFixed(2)}';
  }

  @override
  List<Object?> get props => [
        renewalSuccessRate,
        activeMrr,
        revenueAtRisk,
        churnedRevenue,
        churnedCount,
        usageRevenue,
        totalRevenue,
        revenueMix,
        riskDistribution,
        delta,
        timeRange,
      ];
}

/// Delta changes for metrics comparison between periods
class MetricsDelta extends Equatable {
  /// Active MRR percentage change (null if no previous data)
  final double? activeMrrPercent;

  /// Revenue at risk percentage change (null if no previous data)
  final double? revenueAtRiskPercent;

  /// Usage revenue percentage change (null if no previous data)
  final double? usageRevenuePercent;

  /// Total revenue percentage change (null if no previous data)
  final double? totalRevenuePercent;

  /// Renewal success rate percentage change (null if no previous data)
  final double? renewalSuccessPercent;

  /// Churn count percentage change (null if no previous data)
  final double? churnCountPercent;

  const MetricsDelta({
    this.activeMrrPercent,
    this.revenueAtRiskPercent,
    this.usageRevenuePercent,
    this.totalRevenuePercent,
    this.renewalSuccessPercent,
    this.churnCountPercent,
  });

  /// Get DeltaIndicator for Active MRR (higher is good)
  DeltaIndicator? get activeMrrIndicator =>
      activeMrrPercent != null
          ? DeltaIndicator.forMetric(activeMrrPercent!, higherIsGood: true)
          : null;

  /// Get DeltaIndicator for Revenue at Risk (lower is good)
  DeltaIndicator? get revenueAtRiskIndicator =>
      revenueAtRiskPercent != null
          ? DeltaIndicator.forMetric(revenueAtRiskPercent!, higherIsGood: false)
          : null;

  /// Get DeltaIndicator for Usage Revenue (higher is good)
  DeltaIndicator? get usageRevenueIndicator =>
      usageRevenuePercent != null
          ? DeltaIndicator.forMetric(usageRevenuePercent!, higherIsGood: true)
          : null;

  /// Get DeltaIndicator for Total Revenue (higher is good)
  DeltaIndicator? get totalRevenueIndicator =>
      totalRevenuePercent != null
          ? DeltaIndicator.forMetric(totalRevenuePercent!, higherIsGood: true)
          : null;

  /// Get DeltaIndicator for Renewal Success (higher is good)
  DeltaIndicator? get renewalSuccessIndicator =>
      renewalSuccessPercent != null
          ? DeltaIndicator.forMetric(renewalSuccessPercent!, higherIsGood: true)
          : null;

  /// Get DeltaIndicator for Churn Count (lower is good)
  DeltaIndicator? get churnCountIndicator =>
      churnCountPercent != null
          ? DeltaIndicator.forMetric(churnCountPercent!, higherIsGood: false)
          : null;

  @override
  List<Object?> get props => [
        activeMrrPercent,
        revenueAtRiskPercent,
        usageRevenuePercent,
        totalRevenuePercent,
        renewalSuccessPercent,
        churnCountPercent,
      ];
}

/// Represents a delta indicator with direction, value, and color
class DeltaIndicator extends Equatable {
  /// The percentage value
  final double value;

  /// Whether the change is positive
  final bool isPositive;

  /// Whether this change is considered "good" for the metric
  final bool isGood;

  const DeltaIndicator({
    required this.value,
    required this.isPositive,
    required this.isGood,
  });

  /// Create a DeltaIndicator for a metric
  factory DeltaIndicator.forMetric(double value, {required bool higherIsGood}) {
    final isPositive = value >= 0;
    final isGood = higherIsGood ? isPositive : !isPositive;
    return DeltaIndicator(
      value: value,
      isPositive: isPositive,
      isGood: isGood,
    );
  }

  /// Get the color for this indicator
  Color get color => isGood ? Colors.green : Colors.red;

  /// Get the icon for this indicator
  IconData get icon => isPositive ? Icons.arrow_upward : Icons.arrow_downward;

  /// Format the value as a percentage string
  String get formattedValue {
    final sign = isPositive ? '+' : '';
    return '$sign${value.toStringAsFixed(1)}%';
  }

  @override
  List<Object?> get props => [value, isPositive, isGood];
}

/// Revenue mix breakdown by charge type
class RevenueMix extends Equatable {
  final int recurring;
  final int usage;
  final int oneTime;

  const RevenueMix({
    required this.recurring,
    required this.usage,
    required this.oneTime,
  });

  int get total => recurring + usage + oneTime;

  double get recurringPercent => total > 0 ? (recurring / total) * 100 : 0;
  double get usagePercent => total > 0 ? (usage / total) * 100 : 0;
  double get oneTimePercent => total > 0 ? (oneTime / total) * 100 : 0;

  @override
  List<Object?> get props => [recurring, usage, oneTime];
}

/// Risk distribution of subscriptions
class RiskDistribution extends Equatable {
  final int safe;
  final int atRisk;
  final int critical;
  final int churned;

  const RiskDistribution({
    required this.safe,
    required this.atRisk,
    required this.critical,
    required this.churned,
  });

  int get total => safe + atRisk + critical + churned;

  double get safePercent => total > 0 ? (safe / total) * 100 : 0;
  double get atRiskPercent => total > 0 ? (atRisk / total) * 100 : 0;
  double get criticalPercent => total > 0 ? (critical / total) * 100 : 0;
  double get churnedPercent => total > 0 ? (churned / total) * 100 : 0;

  @override
  List<Object?> get props => [safe, atRisk, critical, churned];
}
