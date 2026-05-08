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

  DashboardMetrics({
    required this.mrrCents,
    required this.renewalRate,
    required this.revenueAtRiskCents,
    required this.usageRevenueCents,
    required this.riskDistribution,
    required this.revenueMix,
    required this.mrrTrend,
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
    );
  }
}

class MetricsService {
  final ApiClient _client;

  MetricsService(this._client);

  Future<DashboardMetrics> fetchMetrics(String appId,
      {CancelToken? cancelToken}) async {
    try {
      final response = await _client.get('/api/v1/apps/$appId/metrics',
          cancelToken: cancelToken);
      debugPrint('[MetricsService] response: ${response.data}');
      return DashboardMetrics.fromJson(
          response.data as Map<String, dynamic>);
    } on DioException catch (e) {
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
}
