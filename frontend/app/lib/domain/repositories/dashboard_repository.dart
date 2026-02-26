import '../entities/dashboard_metrics.dart';

/// Repository interface for dashboard metrics
abstract class DashboardRepository {
  /// Fetch dashboard metrics for the selected app
  Future<DashboardMetrics> fetchMetrics();

  /// Refresh metrics (force fetch)
  Future<DashboardMetrics> refreshMetrics();
}

/// Exception for dashboard-related errors
class DashboardException implements Exception {
  final String message;
  final String? code;

  const DashboardException(this.message, {this.code});

  @override
  String toString() => message;
}
