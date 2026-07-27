import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';

/// A single store row in the Usage Charge Trends report.
class UsageTrendStore {
  final String domain;
  final String shopName;
  final int usageCents;

  /// Signed week-over-week growth ratio for this store (can be negative or
  /// greater than 1). Never clamped.
  final double wowPct;

  UsageTrendStore({
    required this.domain,
    required this.shopName,
    required this.usageCents,
    required this.wowPct,
  });

  factory UsageTrendStore.fromJson(Map<String, dynamic> json) {
    return UsageTrendStore(
      domain: json['domain'] as String? ?? '',
      shopName: json['shopName'] as String? ?? '',
      usageCents: (json['usageCents'] as num?)?.toInt() ?? 0,
      // Signed growth ratio — do NOT clamp.
      wowPct: (json['wowPct'] as num?)?.toDouble() ?? 0,
    );
  }
}

/// A single weekly point in the usage momentum trend.
class UsageWeekPoint {
  final DateTime weekStart;
  final int usageCents;

  UsageWeekPoint({required this.weekStart, required this.usageCents});

  factory UsageWeekPoint.fromJson(Map<String, dynamic> json) {
    return UsageWeekPoint(
      weekStart: _parseDate(json['weekStart'] as String?) ?? DateTime.now(),
      usageCents: (json['usageCents'] as num?)?.toInt() ?? 0,
    );
  }
}

/// Full Usage Charge Trends report payload.
class UsageTrendsReport {
  final String currency;
  final int usageMrrEquivCents;

  /// Signed week-over-week usage growth ratio (can be negative or greater
  /// than 1). Never clamped.
  final double wowChangePct;
  final int activeStores;
  final List<UsageWeekPoint> weeklyTrend;
  final List<UsageTrendStore> stores;

  UsageTrendsReport({
    required this.currency,
    required this.usageMrrEquivCents,
    required this.wowChangePct,
    required this.activeStores,
    required this.weeklyTrend,
    required this.stores,
  });

  factory UsageTrendsReport.fromJson(Map<String, dynamic> json) {
    return UsageTrendsReport(
      currency: json['currency'] as String? ?? 'USD',
      usageMrrEquivCents: (json['usageMrrEquivCents'] as num?)?.toInt() ?? 0,
      // Signed growth ratio — do NOT clamp.
      wowChangePct: (json['wowChangePct'] as num?)?.toDouble() ?? 0,
      activeStores: (json['activeStores'] as num?)?.toInt() ?? 0,
      weeklyTrend: (json['weeklyTrend'] as List<dynamic>?)
              ?.map((e) => UsageWeekPoint.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      stores: (json['stores'] as List<dynamic>?)
              ?.map((e) => UsageTrendStore.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  static UsageTrendsReport empty() => UsageTrendsReport(
        currency: 'USD',
        usageMrrEquivCents: 0,
        wowChangePct: 0,
        activeStores: 0,
        weeklyTrend: const [],
        stores: const [],
      );
}

class UsageTrendsService {
  final ApiClient _client;

  UsageTrendsService(this._client);

  Future<UsageTrendsReport> fetchReport(
    String appId, {
    String? from,
    String? to,
    CancelToken? cancelToken,
  }) async {
    final queryParameters = <String, dynamic>{};
    if (from != null) queryParameters['from'] = from;
    if (to != null) queryParameters['to'] = to;
    final response = await _client.get(
      '/api/v1/apps/$appId/reports/usage-trends',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
    );
    return UsageTrendsReport.fromJson(response.data as Map<String, dynamic>);
  }

  /// Fetches the CSV export of the usage trends report through the
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
      '/api/v1/apps/$appId/reports/usage-trends',
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
    debugPrint('[UsageTrendsService] bad date: $value');
    return null;
  }
}
