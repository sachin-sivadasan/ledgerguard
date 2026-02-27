import 'package:dio/dio.dart';

import '../../core/config/app_config.dart';
import '../../domain/entities/dashboard_metrics.dart';
import '../../domain/entities/time_range.dart';
import '../../domain/repositories/app_repository.dart';
import '../../domain/repositories/auth_repository.dart';
import '../../domain/repositories/dashboard_repository.dart';

/// API implementation of DashboardRepository
class ApiDashboardRepository implements DashboardRepository {
  final Dio _dio;
  final AuthRepository _authRepository;
  final AppRepository _appRepository;

  ApiDashboardRepository({
    Dio? dio,
    required AuthRepository authRepository,
    required AppRepository appRepository,
  })  : _dio = dio ?? Dio(BaseOptions(baseUrl: AppConfig.apiBaseUrl)),
        _authRepository = authRepository,
        _appRepository = appRepository;

  @override
  Future<DashboardMetrics?> fetchMetrics({TimeRange? timeRange}) async {
    return _fetchMetricsFromApi(timeRange: timeRange);
  }

  @override
  Future<DashboardMetrics?> refreshMetrics({TimeRange? timeRange}) async {
    return _fetchMetricsFromApi(timeRange: timeRange);
  }

  Future<DashboardMetrics?> _fetchMetricsFromApi({TimeRange? timeRange}) async {
    // Get the selected app
    final selectedApp = await _appRepository.getSelectedApp();
    if (selectedApp == null) {
      throw const NoAppSelectedException();
    }

    // Get auth token
    final token = await _authRepository.getIdToken();
    if (token == null) {
      throw const UnauthorizedMetricsException();
    }

    try {
      // Extract numeric ID from full GID (e.g., "gid://partners/App/4599915" -> "4599915")
      final appId = _extractNumericId(selectedApp.id);

      // Build query parameters for time range
      final queryParams = <String, String>{};
      if (timeRange != null) {
        queryParams['start'] = timeRange.startFormatted;
        queryParams['end'] = timeRange.endFormatted;
      }

      // Use the new period metrics endpoint
      final response = await _dio.get(
        '/api/v1/apps/$appId/metrics',
        queryParameters: queryParams.isNotEmpty ? queryParams : null,
        options: Options(
          headers: {'Authorization': 'Bearer $token'},
        ),
      );

      if (response.statusCode == 200) {
        final data = response.data as Map<String, dynamic>;
        return _parseMetrics(data, timeRange: timeRange);
      }

      if (response.statusCode == 204 || response.data == null) {
        // No content - empty state
        return null;
      }

      return null;
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        throw const UnauthorizedMetricsException();
      }
      if (e.response?.statusCode == 404) {
        // No metrics found - empty state
        return null;
      }
      throw DashboardException(
        e.message ?? 'Failed to fetch metrics',
        code: 'network-error',
      );
    }
  }

  /// Extracts numeric ID from Shopify GID
  /// e.g., "gid://partners/App/4599915" -> "4599915"
  String _extractNumericId(String gid) {
    final parts = gid.split('/');
    return parts.isNotEmpty ? parts.last : gid;
  }

  DashboardMetrics _parseMetrics(
    Map<String, dynamic> data, {
    TimeRange? timeRange,
  }) {
    // Parse current period data
    final current = data['current'] as Map<String, dynamic>?;
    if (current == null) {
      throw const NoMetricsException();
    }

    // Backend returns cents for monetary values
    final activeMrrCents = (current['active_mrr_cents'] as num?)?.toInt() ?? 0;
    final revenueAtRiskCents =
        (current['revenue_at_risk_cents'] as num?)?.toInt() ?? 0;
    final usageRevenueCents =
        (current['usage_revenue_cents'] as num?)?.toInt() ?? 0;
    final totalRevenueCents =
        (current['total_revenue_cents'] as num?)?.toInt() ?? 0;

    // Backend returns renewal success rate as decimal (0.0 - 1.0)
    final renewalRate =
        (current['renewal_success_rate'] as num?)?.toDouble() ?? 0;
    final renewalSuccessRatePercent = renewalRate * 100;

    // Risk distribution counts
    final safeCount = (current['safe_count'] as num?)?.toInt() ?? 0;
    final oneCycleMissedCount =
        (current['one_cycle_missed_count'] as num?)?.toInt() ?? 0;
    final twoCyclesMissedCount =
        (current['two_cycles_missed_count'] as num?)?.toInt() ?? 0;
    final churnedCount = (current['churned_count'] as num?)?.toInt() ?? 0;

    // Parse delta if available
    MetricsDelta? delta;
    final deltaData = data['delta'] as Map<String, dynamic>?;
    if (deltaData != null) {
      delta = MetricsDelta(
        activeMrrPercent: (deltaData['active_mrr_percent'] as num?)?.toDouble(),
        revenueAtRiskPercent:
            (deltaData['revenue_at_risk_percent'] as num?)?.toDouble(),
        usageRevenuePercent:
            (deltaData['usage_revenue_percent'] as num?)?.toDouble(),
        totalRevenuePercent:
            (deltaData['total_revenue_percent'] as num?)?.toDouble(),
        renewalSuccessPercent:
            (deltaData['renewal_success_rate_percent'] as num?)?.toDouble(),
        churnCountPercent:
            (deltaData['churn_count_percent'] as num?)?.toDouble(),
      );
    }

    // Parse time range from response or use provided
    TimeRange? parsedTimeRange = timeRange;
    final periodData = data['period'] as Map<String, dynamic>?;
    if (periodData != null && parsedTimeRange == null) {
      final startStr = periodData['start'] as String?;
      final endStr = periodData['end'] as String?;
      if (startStr != null && endStr != null) {
        final start = DateTime.parse(startStr);
        final end = DateTime.parse(endStr);
        parsedTimeRange = TimeRange.custom(start, end);
      }
    }

    // Calculate churned revenue (estimate based on at-risk MRR distribution)
    // In real implementation, this would come from backend
    const churnedRevenue = 0;

    return DashboardMetrics(
      renewalSuccessRate: renewalSuccessRatePercent,
      activeMrr: activeMrrCents,
      revenueAtRisk: revenueAtRiskCents,
      churnedRevenue: churnedRevenue,
      churnedCount: churnedCount,
      usageRevenue: usageRevenueCents,
      totalRevenue: totalRevenueCents,
      revenueMix: RevenueMix(
        recurring: activeMrrCents,
        usage: usageRevenueCents,
        oneTime: totalRevenueCents - activeMrrCents - usageRevenueCents,
      ),
      riskDistribution: RiskDistribution(
        safe: safeCount,
        atRisk: oneCycleMissedCount,
        critical: twoCyclesMissedCount,
        churned: churnedCount,
      ),
      delta: delta,
      timeRange: parsedTimeRange,
    );
  }
}
