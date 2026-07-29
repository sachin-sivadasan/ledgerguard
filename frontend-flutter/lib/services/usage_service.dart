import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';

/// A single store row in the Usage & One-Time report.
class UsageStore {
  final String domain;
  final String shopName;
  final int usageCents;
  final int oneTimeCents;
  final int chargeCount;

  UsageStore({
    required this.domain,
    required this.shopName,
    required this.usageCents,
    required this.oneTimeCents,
    required this.chargeCount,
  });

  factory UsageStore.fromJson(Map<String, dynamic> json) {
    return UsageStore(
      domain: json['domain'] as String? ?? '',
      shopName: json['shopName'] as String? ?? '',
      usageCents: (json['usageCents'] as num?)?.toInt() ?? 0,
      oneTimeCents: (json['oneTimeCents'] as num?)?.toInt() ?? 0,
      chargeCount: (json['chargeCount'] as num?)?.toInt() ?? 0,
    );
  }
}

/// A single point in the usage revenue trend.
class UsageTrendPoint {
  final DateTime date;
  final int usageCents;

  UsageTrendPoint({required this.date, required this.usageCents});

  factory UsageTrendPoint.fromJson(Map<String, dynamic> json) {
    return UsageTrendPoint(
      date: _parseDate(json['date'] as String?) ?? DateTime.now(),
      usageCents: (json['usageCents'] as num?)?.toInt() ?? 0,
    );
  }
}

/// Full Usage & One-Time Charges report payload.
class UsageReport {
  final String currency;
  final int usageCents;
  final int oneTimeCents;
  final int chargesCount;
  final List<UsageTrendPoint> trend;
  final List<UsageStore> stores;

  /// Full store-row count before ?limit/?offset paging — drives the report
  /// preview's "View all N" affordance and the detail page's pagination.
  final int storesTotal;

  UsageReport({
    required this.currency,
    required this.usageCents,
    required this.oneTimeCents,
    required this.chargesCount,
    required this.trend,
    required this.stores,
    required this.storesTotal,
  });

  factory UsageReport.fromJson(Map<String, dynamic> json) {
    final stores = (json['stores'] as List<dynamic>?)
            ?.map((e) => UsageStore.fromJson(e as Map<String, dynamic>))
            .toList() ??
        [];
    return UsageReport(
      currency: json['currency'] as String? ?? 'USD',
      usageCents: (json['usageCents'] as num?)?.toInt() ?? 0,
      oneTimeCents: (json['oneTimeCents'] as num?)?.toInt() ?? 0,
      chargesCount: (json['chargesCount'] as num?)?.toInt() ?? 0,
      trend: (json['trend'] as List<dynamic>?)
              ?.map((e) => UsageTrendPoint.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      stores: stores,
      // Fall back to the visible count for older responses without the field.
      storesTotal: (json['storesTotal'] as num?)?.toInt() ?? stores.length,
    );
  }

  static UsageReport empty() => UsageReport(
        currency: 'USD',
        usageCents: 0,
        oneTimeCents: 0,
        chargesCount: 0,
        trend: const [],
        stores: const [],
        storesTotal: 0,
      );
}

class UsageService {
  final ApiClient _client;

  UsageService(this._client);

  Future<UsageReport> fetchReport(
    String appId, {
    String? from,
    String? to,
    int? limit,
    int? offset,
    CancelToken? cancelToken,
  }) async {
    final queryParameters = <String, dynamic>{};
    if (from != null) queryParameters['from'] = from;
    if (to != null) queryParameters['to'] = to;
    if (limit != null) queryParameters['limit'] = limit;
    if (offset != null) queryParameters['offset'] = offset;
    final response = await _client.get(
      '/api/v1/apps/$appId/reports/usage',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
    );
    return UsageReport.fromJson(response.data as Map<String, dynamic>);
  }

  /// Fetches the CSV export of the usage report through the authenticated
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
      '/api/v1/apps/$appId/reports/usage',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
      options: Options(responseType: ResponseType.bytes),
    );
    return Uint8List.fromList(response.data ?? const <int>[]);
  }
}

/// Rows shown in the Usage report's stores PREVIEW before "View all".
const int kUsageStoresPreview = 8;

/// Rows per page on the dedicated Usage stores detail screen.
const int kUsageStoresPageSize = 50;

DateTime? _parseDate(String? value) {
  if (value == null || value.isEmpty) return null;
  try {
    return DateTime.parse(value);
  } catch (e) {
    debugPrint('[UsageService] bad date: $value');
    return null;
  }
}
