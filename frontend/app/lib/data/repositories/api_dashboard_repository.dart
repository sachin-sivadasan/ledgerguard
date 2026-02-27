import 'package:dio/dio.dart';

import '../../core/config/app_config.dart';
import '../../domain/entities/dashboard_metrics.dart';
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
  Future<DashboardMetrics?> fetchMetrics() async {
    return _fetchMetricsFromApi();
  }

  @override
  Future<DashboardMetrics?> refreshMetrics() async {
    return _fetchMetricsFromApi();
  }

  Future<DashboardMetrics?> _fetchMetricsFromApi() async {
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

      final response = await _dio.get(
        '/api/v1/apps/$appId/metrics/latest',
        options: Options(
          headers: {'Authorization': 'Bearer $token'},
        ),
      );

      if (response.statusCode == 200) {
        final data = response.data as Map<String, dynamic>;
        return _parseMetrics(data);
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

  DashboardMetrics _parseMetrics(Map<String, dynamic> data) {
    // Backend returns cents for monetary values
    final activeMrrCents = (data['active_mrr_cents'] as num?)?.toInt() ?? 0;
    final revenueAtRiskCents =
        (data['revenue_at_risk_cents'] as num?)?.toInt() ?? 0;
    final usageRevenueCents =
        (data['usage_revenue_cents'] as num?)?.toInt() ?? 0;
    final totalRevenueCents =
        (data['total_revenue_cents'] as num?)?.toInt() ?? 0;

    // Backend returns renewal success rate as decimal (0.0 - 1.0)
    final renewalRate = (data['renewal_success_rate'] as num?)?.toDouble() ?? 0;
    final renewalSuccessRatePercent = renewalRate * 100;

    // Risk distribution counts
    final safeCount = (data['safe_count'] as num?)?.toInt() ?? 0;
    final oneCycleMissedCount =
        (data['one_cycle_missed_count'] as num?)?.toInt() ?? 0;
    final twoCyclesMissedCount =
        (data['two_cycles_missed_count'] as num?)?.toInt() ?? 0;
    final churnedCount = (data['churned_count'] as num?)?.toInt() ?? 0;

    // Calculate churned revenue (estimate based on at-risk MRR distribution)
    // In real implementation, this would come from backend
    final churnedRevenue = 0;

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
    );
  }
}
