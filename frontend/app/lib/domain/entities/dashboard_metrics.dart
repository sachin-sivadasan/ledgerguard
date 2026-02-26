import 'package:equatable/equatable.dart';

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
      ];
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
