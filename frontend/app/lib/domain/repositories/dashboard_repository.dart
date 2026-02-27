import '../entities/dashboard_metrics.dart';
import '../entities/time_range.dart';

/// Repository interface for dashboard metrics
abstract class DashboardRepository {
  /// Fetch dashboard metrics for the selected app
  /// Returns null if no metrics are available (empty state)
  /// If [timeRange] is provided, fetches metrics for that period
  Future<DashboardMetrics?> fetchMetrics({TimeRange? timeRange});

  /// Refresh metrics (force fetch)
  /// Returns null if no metrics are available (empty state)
  /// If [timeRange] is provided, fetches metrics for that period
  Future<DashboardMetrics?> refreshMetrics({TimeRange? timeRange});
}

/// Exception for dashboard-related errors
class DashboardException implements Exception {
  final String message;
  final String? code;

  const DashboardException(this.message, {this.code});

  @override
  String toString() => message;
}

/// No app selected - cannot fetch metrics
class NoAppSelectedException extends DashboardException {
  const NoAppSelectedException()
      : super('No app selected. Please select an app first.',
            code: 'no-app-selected');
}

/// No metrics available for the selected app
class NoMetricsException extends DashboardException {
  const NoMetricsException()
      : super('No metrics available. Sync your app data to see metrics.',
            code: 'no-metrics');
}

/// Unauthorized to access metrics
class UnauthorizedMetricsException extends DashboardException {
  const UnauthorizedMetricsException()
      : super('Please sign in to view metrics.', code: 'unauthorized');
}
