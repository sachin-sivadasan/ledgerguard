import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';
import '../models/analytics_model.dart';

class DashboardMetrics {
  final int mrrCents;
  final double renewalRate;
  final int revenueAtRiskCents;
  final int usageRevenueCents;
  final RiskDistribution riskDistribution;
  final RevenueMix revenueMix;
  final List<MrrSnapshot> mrrTrend;
  // Period-over-period % changes (from the backend "delta" block). Null when the backend
  // returns no delta (e.g. a single-period query) so the UI can omit the badge.
  final double? mrrDeltaPct;
  final double? renewalDeltaPct;
  final double? usageDeltaPct;
  final double? riskDeltaPct;

  DashboardMetrics({
    required this.mrrCents,
    required this.renewalRate,
    required this.revenueAtRiskCents,
    required this.usageRevenueCents,
    required this.riskDistribution,
    required this.revenueMix,
    required this.mrrTrend,
    this.mrrDeltaPct,
    this.renewalDeltaPct,
    this.usageDeltaPct,
    this.riskDeltaPct,
  });

  /// Parse from backend's period metrics response:
  /// { "period": {...}, "current": { "active_mrr_cents", "safe_count", ... }, "delta": {...} }
  factory DashboardMetrics.fromJson(Map<String, dynamic> json) {
    // Backend wraps metrics in "current" key
    final current = json['current'] as Map<String, dynamic>? ?? json;

    final mrrCents = (current['active_mrr_cents'] as num?)?.toInt() ??
        (current['mrr_cents'] as num?)?.toInt() ??
        0;
    final revenueAtRiskCents =
        (current['revenue_at_risk_cents'] as num?)?.toInt() ?? 0;
    final usageRevenueCents =
        (current['usage_revenue_cents'] as num?)?.toInt() ?? 0;
    final renewalRate =
        (current['renewal_success_rate'] as num?)?.toDouble() ??
            (current['renewal_rate'] as num?)?.toDouble() ??
            0.0;

    // Build risk distribution from flat fields
    final safe = (current['safe_count'] as num?)?.toInt() ?? 0;
    final oneCycle = (current['one_cycle_missed_count'] as num?)?.toInt() ?? 0;
    final twoCycle =
        (current['two_cycles_missed_count'] as num?)?.toInt() ?? 0;
    final churned = (current['churned_count'] as num?)?.toInt() ?? 0;

    final delta = json['delta'] as Map<String, dynamic>?;
    double? deltaPct(String key) => (delta?[key] as num?)?.toDouble();

    return DashboardMetrics(
      mrrCents: mrrCents,
      renewalRate: renewalRate * 100, // Convert 0.92 → 92.0
      revenueAtRiskCents: revenueAtRiskCents,
      usageRevenueCents: usageRevenueCents,
      riskDistribution: RiskDistribution(
        safe: safe,
        oneCycle: oneCycle,
        twoCycle: twoCycle,
        churned: churned,
      ),
      revenueMix: RevenueMix(
        recurringCents: mrrCents,
        usageCents: usageRevenueCents,
        oneTimeCents: 0,
      ),
      mrrTrend: (json['mrr_trend'] as List<dynamic>?)
              ?.map((e) => MrrSnapshot.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      mrrDeltaPct: deltaPct('active_mrr_percent'),
      renewalDeltaPct: deltaPct('renewal_success_rate_percent'),
      usageDeltaPct: deltaPct('usage_revenue_percent'),
      riskDeltaPct: deltaPct('revenue_at_risk_percent'),
    );
  }
}

class MetricsService {
  final ApiClient _client;

  MetricsService(this._client);

  static String _fmtDate(DateTime d) =>
      '${d.year}-${d.month.toString().padLeft(2, '0')}-${d.day.toString().padLeft(2, '0')}';

  Future<DashboardMetrics> fetchMetrics(String appId,
      {DateTime? from, DateTime? to, CancelToken? cancelToken}) async {
    try {
      final query = (from != null && to != null)
          ? {'start': _fmtDate(from), 'end': _fmtDate(to)}
          : null;
      final response = await _client.get('/api/v1/apps/$appId/metrics',
          queryParameters: query, cancelToken: cancelToken);
      debugPrint('[MetricsService] response: ${response.data}');
      return DashboardMetrics.fromJson(
          response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      // Let a cancellation propagate so the provider's staleness guard drops this
      // superseded load instead of overwriting fresh KPIs with a zeroed placeholder.
      if (e.type == DioExceptionType.cancel) rethrow;
      debugPrint('[MetricsService] error: ${e.response?.statusCode}');
      return DashboardMetrics(
        mrrCents: 0,
        renewalRate: 0,
        revenueAtRiskCents: 0,
        usageRevenueCents: 0,
        riskDistribution: const RiskDistribution(
            safe: 0, oneCycle: 0, twoCycle: 0, churned: 0),
        revenueMix: const RevenueMix(
            recurringCents: 0, usageCents: 0, oneTimeCents: 0),
        mrrTrend: [],
      );
    }
  }

  /// Fetches MRR trend data from the trend API with optional granularity.
  /// Returns a list of MrrSnapshots suitable for charting.
  Future<List<MrrSnapshot>> fetchMrrTrend(String appId,
      {int months = 6,
      String granularity = 'weekly',
      CancelToken? cancelToken}) async {
    try {
      final response = await _client.get(
        '/api/v1/apps/$appId/metrics/trend',
        queryParameters: {'months': months, 'granularity': granularity},
        cancelToken: cancelToken,
      );
      final data = response.data as Map<String, dynamic>;
      final snapshots = data['snapshots'] as List<dynamic>? ?? [];
      return snapshots
          .map((e) => MrrSnapshot.fromJson(e as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) return [];
      debugPrint(
          '[MetricsService] fetchMrrTrend error: ${e.response?.statusCode}');
      return [];
    }
  }

  /// Fetches a historical snapshot for a specific date.
  Future<DashboardMetrics?> fetchSnapshotForDate(String appId, DateTime date,
      {CancelToken? cancelToken}) async {
    try {
      final dateStr =
          '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')}';
      final response = await _client.get(
        '/api/v1/apps/$appId/metrics',
        queryParameters: {'start': dateStr, 'end': dateStr},
        cancelToken: cancelToken,
      );
      return DashboardMetrics.fromJson(
          response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) return null;
      debugPrint(
          '[MetricsService] fetchSnapshotForDate error: ${e.response?.statusCode}');
      return null;
    }
  }

  /// Fetches MRR movements by querying metrics for each of the last N months
  /// and computing deltas between periods.
  Future<List<MrrMovement>> fetchMrrMovements(String appId,
      {int months = 6, CancelToken? cancelToken}) async {
    try {
      final now = DateTime.now();
      final movements = <MrrMovement>[];
      int? prevMrr;

      for (var i = months; i >= 0; i--) {
        final monthStart = DateTime(now.year, now.month - i, 1);
        final monthEnd = DateTime(now.year, now.month - i + 1, 0);
        final startStr =
            '${monthStart.year}-${monthStart.month.toString().padLeft(2, '0')}-${monthStart.day.toString().padLeft(2, '0')}';
        final endStr =
            '${monthEnd.year}-${monthEnd.month.toString().padLeft(2, '0')}-${monthEnd.day.toString().padLeft(2, '0')}';

        final response = await _client.get(
          '/api/v1/apps/$appId/metrics',
          queryParameters: {'start': startStr, 'end': endStr},
          cancelToken: cancelToken,
        );
        final data = response.data as Map<String, dynamic>;
        final current = data['current'] as Map<String, dynamic>? ?? data;
        final currentMrr =
            (current['active_mrr_cents'] as num?)?.toInt() ??
                (current['mrr_cents'] as num?)?.toInt() ??
                0;

        if (prevMrr != null && i < months) {
          final delta = currentMrr - prevMrr;
          const monthNames = [
            'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
            'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'
          ];
          final monthLabel = monthNames[monthStart.month - 1];
          movements.add(MrrMovement(
            month: monthLabel,
            newCents: delta > 0 ? delta : 0,
            expansionCents: 0,
            contractionCents: delta < 0 ? -delta : 0,
            churnedCents: 0,
          ));
        }
        prevMrr = currentMrr;
      }
      return movements;
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) return [];
      debugPrint(
          '[MetricsService] fetchMrrMovements error: ${e.response?.statusCode}');
      return [];
    }
  }

  /// Returns (result, errorMessage). On success, error is null.
  /// On failure (e.g. insufficient data), result is null and error explains why.
  Future<(ForecastResult?, String?)> fetchForecast(String appId,
      {int months = 12,
      String model = 'linear',
      CancelToken? cancelToken}) async {
    try {
      final response = await _client.get(
        '/api/v1/apps/$appId/forecast',
        queryParameters: {'months': months, 'model': model},
        cancelToken: cancelToken,
      );
      return (ForecastResult.fromJson(response.data as Map<String, dynamic>), null);
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) return (null, null);
      debugPrint(
          '[MetricsService] fetchForecast error: ${e.response?.statusCode}');
      // Extract meaningful error from 422 response
      if (e.response?.statusCode == 422) {
        final data = e.response?.data;
        if (data is Map) {
          final dataPoints = data['data_points'] as int? ?? 0;
          final required = data['required'] as int? ?? 90;
          return (null, 'Need $required+ daily snapshots for forecasting (have $dataPoints). Data accumulates automatically over time.');
        }
      }
      return (null, 'Failed to load forecast data');
    }
  }

  Future<List<CohortData>> fetchCohorts(String appId,
      {int months = 6, CancelToken? cancelToken}) async {
    try {
      final response = await _client.get(
        '/api/v1/apps/$appId/cohorts',
        queryParameters: {'months': months},
        cancelToken: cancelToken,
      );
      final list = response.data['cohorts'] as List<dynamic>? ?? [];
      return list
          .map((e) => CohortData.fromJson(e as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) return [];
      debugPrint(
          '[MetricsService] fetchCohorts error: ${e.response?.statusCode}');
      return [];
    }
  }

  Future<RevenueConcentration?> fetchRevenueConcentration(String appId,
      {int top = 10, CancelToken? cancelToken}) async {
    try {
      final response = await _client.get(
        '/api/v1/apps/$appId/revenue/concentration',
        queryParameters: {'top': top},
        cancelToken: cancelToken,
      );
      return RevenueConcentration.fromJson(
          response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) return null;
      debugPrint(
          '[MetricsService] fetchRevenueConcentration error: ${e.response?.statusCode}');
      return null;
    }
  }

  Future<List<AppComparison>> fetchAppComparison(
      {CancelToken? cancelToken}) async {
    try {
      final response = await _client.get('/api/v1/metrics/aggregate',
          cancelToken: cancelToken);
      final apps = response.data['apps'] as List<dynamic>? ?? [];
      return apps
          .map((e) => AppComparison.fromJson(e as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) return [];
      debugPrint(
          '[MetricsService] fetchAppComparison error: ${e.response?.statusCode}');
      return [];
    }
  }
}
