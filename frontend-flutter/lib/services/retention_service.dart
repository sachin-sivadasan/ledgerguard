import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';

/// A single plan row in the Retention report.
class RetentionPlan {
  final String planName;
  final int activeSubs;

  /// Renewal rate as a 0..1 decimal. Clamped to [0,1] in [fromJson].
  final double renewalRate;
  final int retainedMrrCents;

  RetentionPlan({
    required this.planName,
    required this.activeSubs,
    required this.renewalRate,
    required this.retainedMrrCents,
  });

  factory RetentionPlan.fromJson(Map<String, dynamic> json) {
    return RetentionPlan(
      planName: json['planName'] as String? ?? '',
      activeSubs: (json['activeSubs'] as num?)?.toInt() ?? 0,
      renewalRate:
          ((json['renewalRate'] as num?)?.toDouble() ?? 0).clamp(0.0, 1.0),
      retainedMrrCents: (json['retainedMrrCents'] as num?)?.toInt() ?? 0,
    );
  }
}

/// A single point in the renewal success rate trend.
class RetentionTrendPoint {
  final DateTime date;
  final double renewalRate;

  RetentionTrendPoint({required this.date, required this.renewalRate});

  factory RetentionTrendPoint.fromJson(Map<String, dynamic> json) {
    return RetentionTrendPoint(
      date: _parseDate(json['date'] as String?) ?? DateTime.now(),
      renewalRate:
          ((json['renewalRate'] as num?)?.toDouble() ?? 0).clamp(0.0, 1.0),
    );
  }
}

/// Full Retention report payload.
class RetentionReport {
  final String currency;

  /// Renewal rate as a 0..1 decimal (e.g. 0.92 == 92%). Clamped to [0,1] in
  /// [fromJson] so the invariant holds regardless of producer.
  final double renewalRate;
  final int retainedMrrCents;
  final int reactivations;
  final List<RetentionTrendPoint> trend;
  final List<RetentionPlan> plans;

  RetentionReport({
    required this.currency,
    required this.renewalRate,
    required this.retainedMrrCents,
    required this.reactivations,
    required this.trend,
    required this.plans,
  });

  factory RetentionReport.fromJson(Map<String, dynamic> json) {
    return RetentionReport(
      currency: json['currency'] as String? ?? 'USD',
      renewalRate:
          ((json['renewalRate'] as num?)?.toDouble() ?? 0).clamp(0.0, 1.0),
      retainedMrrCents: (json['retainedMrrCents'] as num?)?.toInt() ?? 0,
      reactivations: (json['reactivations'] as num?)?.toInt() ?? 0,
      trend: (json['trend'] as List<dynamic>?)
              ?.map((e) =>
                  RetentionTrendPoint.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      plans: (json['plans'] as List<dynamic>?)
              ?.map((e) => RetentionPlan.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  static RetentionReport empty() => RetentionReport(
        currency: 'USD',
        renewalRate: 0,
        retainedMrrCents: 0,
        reactivations: 0,
        trend: const [],
        plans: const [],
      );
}

class RetentionService {
  final ApiClient _client;

  RetentionService(this._client);

  Future<RetentionReport> fetchReport(
    String appId, {
    String? from,
    String? to,
    CancelToken? cancelToken,
  }) async {
    final queryParameters = <String, dynamic>{};
    if (from != null) queryParameters['from'] = from;
    if (to != null) queryParameters['to'] = to;
    final response = await _client.get(
      '/api/v1/apps/$appId/reports/retention',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
    );
    return RetentionReport.fromJson(response.data as Map<String, dynamic>);
  }

  /// Fetches the CSV export of the retention report through the authenticated
  /// [ApiClient] (Firebase Bearer token injected by the Dio interceptor).
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
      '/api/v1/apps/$appId/reports/retention',
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
    debugPrint('[RetentionService] bad date: $value');
    return null;
  }
}
