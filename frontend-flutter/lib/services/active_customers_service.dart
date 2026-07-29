import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';

/// A single plan row in the Active Customers report.
class ActivePlan {
  final String planName;
  final int activeSubs;
  final int mrrCents;

  /// Share of active customers as a 0..1 decimal. Clamped to [0,1] in [fromJson].
  final double pctOfActive;

  ActivePlan({
    required this.planName,
    required this.activeSubs,
    required this.mrrCents,
    required this.pctOfActive,
  });

  factory ActivePlan.fromJson(Map<String, dynamic> json) {
    return ActivePlan(
      planName: json['planName'] as String? ?? '',
      activeSubs: (json['activeSubs'] as num?)?.toInt() ?? 0,
      mrrCents: (json['mrrCents'] as num?)?.toInt() ?? 0,
      pctOfActive:
          ((json['pctOfActive'] as num?)?.toDouble() ?? 0).clamp(0.0, 1.0),
    );
  }
}

/// A single point in the active-customers trend.
class ActiveTrendPoint {
  final DateTime date;
  final int activeCustomers;

  ActiveTrendPoint({required this.date, required this.activeCustomers});

  factory ActiveTrendPoint.fromJson(Map<String, dynamic> json) {
    return ActiveTrendPoint(
      date: _parseDate(json['date'] as String?) ?? DateTime.now(),
      activeCustomers: (json['activeCustomers'] as num?)?.toInt() ?? 0,
    );
  }
}

/// Full Active Customers report payload.
class ActiveCustomersReport {
  final String currency;
  final int activeCustomers;
  final int newCount;
  final int churnedCount;
  final int netChange;
  final String interval;
  final List<ActiveTrendPoint> trend;
  final List<ActivePlan> plans;

  ActiveCustomersReport({
    required this.currency,
    required this.activeCustomers,
    required this.newCount,
    required this.churnedCount,
    required this.netChange,
    required this.interval,
    required this.trend,
    required this.plans,
  });

  factory ActiveCustomersReport.fromJson(Map<String, dynamic> json) {
    return ActiveCustomersReport(
      currency: json['currency'] as String? ?? 'USD',
      activeCustomers: (json['activeCustomers'] as num?)?.toInt() ?? 0,
      newCount: (json['newCount'] as num?)?.toInt() ?? 0,
      churnedCount: (json['churnedCount'] as num?)?.toInt() ?? 0,
      netChange: (json['netChange'] as num?)?.toInt() ?? 0,
      interval: json['interval'] as String? ?? 'week',
      trend: (json['trend'] as List<dynamic>?)
              ?.map((e) => ActiveTrendPoint.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      plans: (json['plans'] as List<dynamic>?)
              ?.map((e) => ActivePlan.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  static ActiveCustomersReport empty() => ActiveCustomersReport(
        currency: 'USD',
        activeCustomers: 0,
        newCount: 0,
        churnedCount: 0,
        netChange: 0,
        interval: 'week',
        trend: const [],
        plans: const [],
      );
}

class ActiveCustomersService {
  final ApiClient _client;

  ActiveCustomersService(this._client);

  Future<ActiveCustomersReport> fetchReport(
    String appId, {
    String? from,
    String? to,
    CancelToken? cancelToken,
  }) async {
    final queryParameters = <String, dynamic>{};
    if (from != null) queryParameters['from'] = from;
    if (to != null) queryParameters['to'] = to;
    final response = await _client.get(
      '/api/v1/apps/$appId/reports/active-customers',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
    );
    return ActiveCustomersReport.fromJson(response.data as Map<String, dynamic>);
  }

  /// Fetches the CSV export of the Active Customers report through the
  /// authenticated [ApiClient] (Firebase Bearer token injected by the Dio
  /// interceptor).
  ///
  /// Returns the raw response bytes so the caller can trigger a client-side
  /// download without relying on an external browser navigation (which would
  /// 401 because it carries no auth header).
  Future<Uint8List> fetchCsvBytes(
    String appId, {
    String? from,
    String? to,
    CancelToken? cancelToken,
  }) async {
    final queryParameters = <String, dynamic>{'format': 'csv'};
    if (from != null) queryParameters['from'] = from;
    if (to != null) queryParameters['to'] = to;
    final response = await _client.get<List<int>>(
      '/api/v1/apps/$appId/reports/active-customers',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
      options: Options(responseType: ResponseType.bytes),
    );
    return Uint8List.fromList(response.data ?? const <int>[]);
  }
}

DateTime? _parseDate(String? value) {
  if (value == null || value.isEmpty) return null;
  try {
    return DateTime.parse(value);
  } catch (e) {
    debugPrint('[ActiveCustomersService] bad date: $value');
    return null;
  }
}
