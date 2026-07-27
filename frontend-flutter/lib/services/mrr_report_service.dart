import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';

/// A single plan row in the MRR report.
class MrrPlan {
  final String planName;
  final int activeSubs;
  final int mrrCents;

  /// Share of total MRR as a 0..1 decimal. Clamped to [0,1] in [fromJson].
  final double pctOfTotal;

  MrrPlan({
    required this.planName,
    required this.activeSubs,
    required this.mrrCents,
    required this.pctOfTotal,
  });

  factory MrrPlan.fromJson(Map<String, dynamic> json) {
    return MrrPlan(
      planName: json['planName'] as String? ?? '',
      activeSubs: (json['activeSubs'] as num?)?.toInt() ?? 0,
      mrrCents: (json['mrrCents'] as num?)?.toInt() ?? 0,
      pctOfTotal:
          ((json['pctOfTotal'] as num?)?.toDouble() ?? 0).clamp(0.0, 1.0),
    );
  }
}

/// A single point in the MRR trend.
class MrrTrendPoint {
  final DateTime date;
  final int mrrCents;

  MrrTrendPoint({required this.date, required this.mrrCents});

  factory MrrTrendPoint.fromJson(Map<String, dynamic> json) {
    return MrrTrendPoint(
      date: _parseDate(json['date'] as String?) ?? DateTime.now(),
      mrrCents: (json['mrrCents'] as num?)?.toInt() ?? 0,
    );
  }
}

/// Full MRR report payload.
class MrrReport {
  final String currency;
  final int mrrCents;

  /// Signed month-over-month growth ratio (e.g. 0.062 == +6.2%). Can be
  /// negative or >1 — NOT clamped, since it represents a growth rate.
  final double momChangePct;
  final int newMrrCents;
  final int churnedMrrCents;
  final List<MrrTrendPoint> trend;
  final List<MrrPlan> plans;

  MrrReport({
    required this.currency,
    required this.mrrCents,
    required this.momChangePct,
    required this.newMrrCents,
    required this.churnedMrrCents,
    required this.trend,
    required this.plans,
  });

  factory MrrReport.fromJson(Map<String, dynamic> json) {
    return MrrReport(
      currency: json['currency'] as String? ?? 'USD',
      mrrCents: (json['mrrCents'] as num?)?.toInt() ?? 0,
      // Signed growth ratio — do NOT clamp.
      momChangePct: (json['momChangePct'] as num?)?.toDouble() ?? 0,
      newMrrCents: (json['newMrrCents'] as num?)?.toInt() ?? 0,
      churnedMrrCents: (json['churnedMrrCents'] as num?)?.toInt() ?? 0,
      trend: (json['trend'] as List<dynamic>?)
              ?.map((e) => MrrTrendPoint.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      plans: (json['plans'] as List<dynamic>?)
              ?.map((e) => MrrPlan.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  static MrrReport empty() => MrrReport(
        currency: 'USD',
        mrrCents: 0,
        momChangePct: 0,
        newMrrCents: 0,
        churnedMrrCents: 0,
        trend: const [],
        plans: const [],
      );
}

class MrrReportService {
  final ApiClient _client;

  MrrReportService(this._client);

  Future<MrrReport> fetchReport(
    String appId, {
    String? from,
    String? to,
    CancelToken? cancelToken,
  }) async {
    final queryParameters = <String, dynamic>{};
    if (from != null) queryParameters['from'] = from;
    if (to != null) queryParameters['to'] = to;
    final response = await _client.get(
      '/api/v1/apps/$appId/reports/mrr',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
    );
    return MrrReport.fromJson(response.data as Map<String, dynamic>);
  }

  /// Fetches the CSV export of the MRR report through the authenticated
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
      '/api/v1/apps/$appId/reports/mrr',
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
    debugPrint('[MrrReportService] bad date: $value');
    return null;
  }
}
