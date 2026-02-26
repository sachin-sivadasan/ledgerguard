import '../../domain/entities/dashboard_metrics.dart';
import '../../domain/repositories/dashboard_repository.dart';

/// Mock implementation of DashboardRepository for development
class MockDashboardRepository implements DashboardRepository {
  /// Simulated delay for API calls
  final Duration delay;

  MockDashboardRepository({
    this.delay = const Duration(milliseconds: 800),
  });

  /// Mock dashboard metrics
  static const _mockMetrics = DashboardMetrics(
    renewalSuccessRate: 94.2,
    activeMrr: 12450000, // $124,500.00
    revenueAtRisk: 1850000, // $18,500.00
    churnedRevenue: 320000, // $3,200.00
    churnedCount: 12,
    usageRevenue: 2340000, // $23,400.00
    revenueMix: RevenueMix(
      recurring: 12450000, // $124,500
      usage: 2340000, // $23,400
      oneTime: 450000, // $4,500
    ),
    riskDistribution: RiskDistribution(
      safe: 842,
      atRisk: 45,
      critical: 18,
      churned: 12,
    ),
  );

  @override
  Future<DashboardMetrics> fetchMetrics() async {
    await Future.delayed(delay);
    return _mockMetrics;
  }

  @override
  Future<DashboardMetrics> refreshMetrics() async {
    await Future.delayed(delay);
    return _mockMetrics;
  }
}
