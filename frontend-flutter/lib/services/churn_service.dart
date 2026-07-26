import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';

/// A single churned store row in the Churn report.
class ChurnStore {
  final String domain;
  final String shopName;
  final int mrrLostCents;
  final DateTime? churnedDate;
  final int tenureDays;
  final String planName;

  ChurnStore({
    required this.domain,
    required this.shopName,
    required this.mrrLostCents,
    required this.churnedDate,
    required this.tenureDays,
    required this.planName,
  });

  factory ChurnStore.fromJson(Map<String, dynamic> json) {
    return ChurnStore(
      domain: json['domain'] as String? ?? '',
      shopName: json['shopName'] as String? ?? '',
      mrrLostCents: (json['mrrLostCents'] as num?)?.toInt() ?? 0,
      churnedDate: _parseDate(json['churnedDate'] as String?),
      tenureDays: (json['tenureDays'] as num?)?.toInt() ?? 0,
      planName: json['planName'] as String? ?? '',
    );
  }
}

/// A single point in the churn rate trend.
class ChurnTrendPoint {
  final DateTime date;
  final double churnRate;

  ChurnTrendPoint({required this.date, required this.churnRate});

  factory ChurnTrendPoint.fromJson(Map<String, dynamic> json) {
    return ChurnTrendPoint(
      date: _parseDate(json['date'] as String?) ?? DateTime.now(),
      churnRate: ((json['churnRate'] as num?)?.toDouble() ?? 0).clamp(0.0, 1.0),
    );
  }
}

/// Full Churn report payload.
class ChurnReport {
  final String currency;

  /// Churn rate as a 0..1 decimal (e.g. 0.042 == 4.2%). Clamped to [0,1] in
  /// [fromJson] so the invariant holds regardless of producer.
  final double churnRate;
  final int churnedMrrLostCents;
  final int churnedCount;
  final List<ChurnTrendPoint> trend;
  final List<ChurnStore> stores;

  ChurnReport({
    required this.currency,
    required this.churnRate,
    required this.churnedMrrLostCents,
    required this.churnedCount,
    required this.trend,
    required this.stores,
  });

  factory ChurnReport.fromJson(Map<String, dynamic> json) {
    return ChurnReport(
      currency: json['currency'] as String? ?? 'USD',
      churnRate: ((json['churnRate'] as num?)?.toDouble() ?? 0).clamp(0.0, 1.0),
      churnedMrrLostCents:
          (json['churnedMrrLostCents'] as num?)?.toInt() ?? 0,
      churnedCount: (json['churnedCount'] as num?)?.toInt() ?? 0,
      trend: (json['trend'] as List<dynamic>?)
              ?.map((e) =>
                  ChurnTrendPoint.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      stores: (json['stores'] as List<dynamic>?)
              ?.map((e) => ChurnStore.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  static ChurnReport empty() => ChurnReport(
        currency: 'USD',
        churnRate: 0,
        churnedMrrLostCents: 0,
        churnedCount: 0,
        trend: const [],
        stores: const [],
      );
}

class ChurnService {
  final ApiClient _client;

  ChurnService(this._client);

  Future<ChurnReport> fetchReport(
    String appId, {
    String? from,
    String? to,
    CancelToken? cancelToken,
  }) async {
    final queryParameters = <String, dynamic>{};
    if (from != null) queryParameters['from'] = from;
    if (to != null) queryParameters['to'] = to;
    final response = await _client.get(
      '/api/v1/apps/$appId/reports/churn',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
    );
    return ChurnReport.fromJson(response.data as Map<String, dynamic>);
  }

  /// Fetches the CSV export of the churned stores through the authenticated
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
      '/api/v1/apps/$appId/reports/churn',
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
    debugPrint('[ChurnService] bad date: $value');
    return null;
  }
}
