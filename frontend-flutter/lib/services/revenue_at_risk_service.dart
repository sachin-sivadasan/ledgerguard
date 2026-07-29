import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';
import '../widgets/lg_risk_badge.dart';

/// A single at-risk store row in the Revenue at Risk report.
class RevenueAtRiskStore {
  final String domain;
  final String shopName;
  final int mrrCents;
  final RiskState riskState;
  final int daysLate;
  final DateTime? expectedChargeDate;
  final String planName;
  final int recoverableCents;

  RevenueAtRiskStore({
    required this.domain,
    required this.shopName,
    required this.mrrCents,
    required this.riskState,
    required this.daysLate,
    required this.expectedChargeDate,
    required this.planName,
    required this.recoverableCents,
  });

  factory RevenueAtRiskStore.fromJson(Map<String, dynamic> json) {
    return RevenueAtRiskStore(
      domain: json['domain'] as String? ?? '',
      shopName: json['shopName'] as String? ?? '',
      mrrCents: (json['mrrCents'] as num?)?.toInt() ?? 0,
      riskState: _parseRiskState(json['riskState'] as String?),
      daysLate: (json['daysLate'] as num?)?.toInt() ?? 0,
      expectedChargeDate: _parseDate(json['expectedChargeDate'] as String?),
      planName: json['planName'] as String? ?? '',
      recoverableCents: (json['recoverableCents'] as num?)?.toInt() ?? 0,
    );
  }
}

/// A single point in the at-risk MRR trend.
class RevenueAtRiskTrendPoint {
  final DateTime date;
  final int atRiskCents;

  RevenueAtRiskTrendPoint({required this.date, required this.atRiskCents});

  factory RevenueAtRiskTrendPoint.fromJson(Map<String, dynamic> json) {
    return RevenueAtRiskTrendPoint(
      date: _parseDate(json['date'] as String?) ?? DateTime.now(),
      atRiskCents: (json['atRiskCents'] as num?)?.toInt() ?? 0,
    );
  }
}

/// Full Revenue at Risk report payload.
class RevenueAtRiskReport {
  final String currency;
  final int totalAtRiskCents;
  final int recoverableCents;
  final int oneCycleCents;
  final int twoCycleCents;
  final int oneCycleCount;
  final int twoCycleCount;
  final List<RevenueAtRiskTrendPoint> trend;
  final List<RevenueAtRiskStore> stores;

  /// Full at-risk-store count before ?limit/?offset paging — drives the report
  /// preview's "View all N" affordance and the detail page's pagination.
  final int storesTotal;

  RevenueAtRiskReport({
    required this.currency,
    required this.totalAtRiskCents,
    required this.recoverableCents,
    required this.oneCycleCents,
    required this.twoCycleCents,
    required this.oneCycleCount,
    required this.twoCycleCount,
    required this.trend,
    required this.stores,
    required this.storesTotal,
  });

  int get atRiskStoreCount => oneCycleCount + twoCycleCount;

  factory RevenueAtRiskReport.fromJson(Map<String, dynamic> json) {
    final byState = json['byState'] as Map<String, dynamic>? ?? const {};
    final counts = json['counts'] as Map<String, dynamic>? ?? const {};
    final stores = (json['stores'] as List<dynamic>?)
            ?.map((e) =>
                RevenueAtRiskStore.fromJson(e as Map<String, dynamic>))
            .toList() ??
        [];
    return RevenueAtRiskReport(
      currency: json['currency'] as String? ?? 'USD',
      totalAtRiskCents: (json['totalAtRiskCents'] as num?)?.toInt() ?? 0,
      recoverableCents: (json['recoverableCents'] as num?)?.toInt() ?? 0,
      oneCycleCents: (byState['oneCycleCents'] as num?)?.toInt() ?? 0,
      twoCycleCents: (byState['twoCycleCents'] as num?)?.toInt() ?? 0,
      oneCycleCount: (counts['oneCycle'] as num?)?.toInt() ?? 0,
      twoCycleCount: (counts['twoCycle'] as num?)?.toInt() ?? 0,
      trend: (json['trend'] as List<dynamic>?)
              ?.map((e) => RevenueAtRiskTrendPoint.fromJson(
                  e as Map<String, dynamic>))
              .toList() ??
          [],
      stores: stores,
      // Fall back to the visible count for older responses without the field.
      storesTotal: (json['storesTotal'] as num?)?.toInt() ?? stores.length,
    );
  }

  static RevenueAtRiskReport empty() => RevenueAtRiskReport(
        currency: 'USD',
        totalAtRiskCents: 0,
        recoverableCents: 0,
        oneCycleCents: 0,
        twoCycleCents: 0,
        oneCycleCount: 0,
        twoCycleCount: 0,
        trend: const [],
        stores: const [],
        storesTotal: 0,
      );
}

class RevenueAtRiskService {
  final ApiClient _client;

  RevenueAtRiskService(this._client);

  Future<RevenueAtRiskReport> fetchReport(
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
      '/api/v1/apps/$appId/reports/revenue-at-risk',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
    );
    return RevenueAtRiskReport.fromJson(response.data as Map<String, dynamic>);
  }

  /// Fetches the CSV export of the ranked stores through the authenticated
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
      '/api/v1/apps/$appId/reports/revenue-at-risk',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
      options: Options(responseType: ResponseType.bytes),
    );
    return Uint8List.fromList(response.data ?? const <int>[]);
  }
}

/// Rows shown in the Revenue at Risk report's store PREVIEW before "View all".
const int kRevenueAtRiskStoresPreview = 8;

/// Rows per page on the dedicated Revenue at Risk stores detail screen.
const int kRevenueAtRiskStoresPageSize = 50;

RiskState _parseRiskState(String? value) {
  switch (value) {
    case 'ONE_CYCLE_MISSED':
    case 'ONE_CYCLE':
      return RiskState.oneCycleMissed;
    case 'TWO_CYCLES_MISSED':
    case 'TWO_CYCLE_MISSED':
    case 'TWO_CYCLE':
      return RiskState.twoCycleMissed;
    case 'CHURNED':
      return RiskState.churned;
    case 'SAFE':
      return RiskState.safe;
    default:
      // Safety net: this report only ever contains at-risk stores, so an
      // unrecognized token must never render green. Fall back to the most
      // severe at-risk state rather than turning an at-risk store safe.
      debugPrint('[RevenueAtRiskService] unknown riskState: $value');
      return RiskState.churned;
  }
}

DateTime? _parseDate(String? value) {
  if (value == null || value.isEmpty) return null;
  try {
    return DateTime.parse(value);
  } catch (e) {
    debugPrint('[RevenueAtRiskService] bad date: $value');
    return null;
  }
}
