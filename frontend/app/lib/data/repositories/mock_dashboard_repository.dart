import '../../domain/entities/dashboard_metrics.dart';
import '../../domain/entities/time_range.dart';
import '../../domain/repositories/dashboard_repository.dart';

/// Mock implementation of DashboardRepository for development
class MockDashboardRepository implements DashboardRepository {
  /// Simulated delay for API calls
  final Duration delay;

  /// Whether to return empty state (for testing)
  final bool returnEmpty;

  MockDashboardRepository({
    this.delay = const Duration(milliseconds: 800),
    this.returnEmpty = false,
  });

  /// Mock delta metrics
  static const _mockDelta = MetricsDelta(
    activeMrrPercent: 5.93,
    revenueAtRiskPercent: -8.5,
    usageRevenuePercent: 12.3,
    totalRevenuePercent: 8.7,
    renewalSuccessPercent: 2.1,
    churnCountPercent: -15.0,
  );

  /// Generate mock metrics with optional time range
  DashboardMetrics _generateMockMetrics({TimeRange? timeRange}) {
    return DashboardMetrics(
      renewalSuccessRate: 94.2,
      activeMrr: 12450000, // $124,500.00
      revenueAtRisk: 1850000, // $18,500.00
      churnedRevenue: 320000, // $3,200.00
      churnedCount: 12,
      usageRevenue: 2340000, // $23,400.00
      totalRevenue: 15240000, // $152,400.00 (recurring + usage + one-time)
      revenueMix: const RevenueMix(
        recurring: 12450000, // $124,500
        usage: 2340000, // $23,400
        oneTime: 450000, // $4,500
      ),
      riskDistribution: const RiskDistribution(
        safe: 842,
        atRisk: 45,
        critical: 18,
        churned: 12,
      ),
      delta: _mockDelta,
      timeRange: timeRange,
    );
  }

  @override
  Future<DashboardMetrics?> fetchMetrics({TimeRange? timeRange}) async {
    await Future.delayed(delay);
    return returnEmpty ? null : _generateMockMetrics(timeRange: timeRange);
  }

  @override
  Future<DashboardMetrics?> refreshMetrics({TimeRange? timeRange}) async {
    await Future.delayed(delay);
    return returnEmpty ? null : _generateMockMetrics(timeRange: timeRange);
  }
}
