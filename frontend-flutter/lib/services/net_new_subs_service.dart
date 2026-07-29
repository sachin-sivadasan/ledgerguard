import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';

/// A single day in the net-new subscriptions trend.
class NetNewTrendPoint {
  final DateTime date;
  final int newSubs;
  final int churned;
  final int net;

  NetNewTrendPoint({
    required this.date,
    required this.newSubs,
    required this.churned,
    required this.net,
  });

  /// Returns null when the date is missing/unparseable so the caller can DROP the point
  /// rather than plot it at a sentinel date. The backend always emits a valid YYYY-MM-DD.
  static NetNewTrendPoint? tryFromJson(Map<String, dynamic> json) {
    final date = _parseDate(json['date'] as String?);
    if (date == null) return null;
    return NetNewTrendPoint(
      date: date,
      newSubs: (json['new'] as num?)?.toInt() ?? 0,
      churned: (json['churned'] as num?)?.toInt() ?? 0,
      net: (json['net'] as num?)?.toInt() ?? 0,
    );
  }
}

/// A single row in the recent-new-subscriptions table.
class NewSubRow {
  final String domain;
  final String shopName;
  final String planName;
  final int mrrCents;
  final String started; // raw YYYY-MM-DD

  NewSubRow({
    required this.domain,
    required this.shopName,
    required this.planName,
    required this.mrrCents,
    required this.started,
  });

  /// Parsed started date, or null when unparseable (UI renders raw/"—").
  DateTime? get startedDate => _parseDate(started);

  factory NewSubRow.fromJson(Map<String, dynamic> json) {
    return NewSubRow(
      domain: json['domain'] as String? ?? '',
      shopName: json['shopName'] as String? ?? '',
      planName: json['planName'] as String? ?? '',
      mrrCents: (json['mrrCents'] as num?)?.toInt() ?? 0,
      started: json['started'] as String? ?? '',
    );
  }
}

/// Full Net-New Subscriptions report payload.
class NetNewSubsReport {
  final String currency;
  final int newSubs;
  final int churned;
  final int net;
  final List<NetNewTrendPoint> trend;
  final List<NewSubRow> newStores;

  /// Full new-store-row count before ?limit/?offset paging — drives the report
  /// preview's "View all N" affordance and the detail page's pagination.
  final int newStoresTotal;

  NetNewSubsReport({
    required this.currency,
    required this.newSubs,
    required this.churned,
    required this.net,
    required this.trend,
    required this.newStores,
    required this.newStoresTotal,
  });

  factory NetNewSubsReport.fromJson(Map<String, dynamic> json) {
    final newStores = (json['newStores'] as List<dynamic>?)
            ?.map((e) => NewSubRow.fromJson(e as Map<String, dynamic>))
            .toList() ??
        [];
    return NetNewSubsReport(
      currency: json['currency'] as String? ?? 'USD',
      newSubs: (json['newSubs'] as num?)?.toInt() ?? 0,
      churned: (json['churned'] as num?)?.toInt() ?? 0,
      net: (json['net'] as num?)?.toInt() ?? 0,
      trend: (json['trend'] as List<dynamic>?)
              ?.map((e) => NetNewTrendPoint.tryFromJson(e as Map<String, dynamic>))
              .whereType<NetNewTrendPoint>()
              .toList() ??
          [],
      newStores: newStores,
      // Fall back to the visible count for older responses without the field.
      newStoresTotal: (json['newStoresTotal'] as num?)?.toInt() ?? newStores.length,
    );
  }

  static NetNewSubsReport empty() => NetNewSubsReport(
        currency: 'USD',
        newSubs: 0,
        churned: 0,
        net: 0,
        trend: const [],
        newStores: const [],
        newStoresTotal: 0,
      );
}

class NetNewSubsService {
  final ApiClient _client;

  NetNewSubsService(this._client);

  Future<NetNewSubsReport> fetchReport(
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
      '/api/v1/apps/$appId/reports/net-new-subscriptions',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
    );
    return NetNewSubsReport.fromJson(response.data as Map<String, dynamic>);
  }

  /// Fetches the CSV export bytes through the authenticated [ApiClient] (Firebase
  /// Bearer token injected by the Dio interceptor) so the download carries auth.
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
      '/api/v1/apps/$appId/reports/net-new-subscriptions',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
      options: Options(responseType: ResponseType.bytes),
    );
    return Uint8List.fromList(response.data ?? const <int>[]);
  }
}

/// Rows shown in the Net-New Subscriptions report's new-stores PREVIEW before "View all".
const int kNetNewSubsPreview = 8;

/// Rows per page on the dedicated Net-New Subscriptions detail screen.
const int kNetNewSubsPageSize = 50;

DateTime? _parseDate(String? value) {
  if (value == null || value.isEmpty) return null;
  try {
    return DateTime.parse(value);
  } catch (e) {
    debugPrint('[NetNewSubsService] bad date: $value');
    return null;
  }
}
