import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';

/// A single charge row in the Earnings report.
class EarningsCharge {
  final DateTime? date;
  final String domain;
  final String shopName;
  final int grossCents;
  final int netCents;
  final String status;
  final DateTime? availableDate;

  EarningsCharge({
    required this.date,
    required this.domain,
    required this.shopName,
    required this.grossCents,
    required this.netCents,
    required this.status,
    required this.availableDate,
  });

  factory EarningsCharge.fromJson(Map<String, dynamic> json) {
    return EarningsCharge(
      date: _parseDate(json['date'] as String?),
      domain: json['domain'] as String? ?? '',
      shopName: json['shopName'] as String? ?? '',
      grossCents: (json['grossCents'] as num?)?.toInt() ?? 0,
      netCents: (json['netCents'] as num?)?.toInt() ?? 0,
      status: json['status'] as String? ?? '',
      availableDate: _parseDate(json['availableDate'] as String?),
    );
  }
}

/// Full Earnings report payload.
class EarningsReport {
  final String currency;
  final int netEarningsCents;
  final int pendingCents;
  final int availableCents;
  final int paidOutCents;
  final List<EarningsCharge> charges;

  EarningsReport({
    required this.currency,
    required this.netEarningsCents,
    required this.pendingCents,
    required this.availableCents,
    required this.paidOutCents,
    required this.charges,
  });

  factory EarningsReport.fromJson(Map<String, dynamic> json) {
    return EarningsReport(
      currency: json['currency'] as String? ?? 'USD',
      netEarningsCents: (json['netEarningsCents'] as num?)?.toInt() ?? 0,
      pendingCents: (json['pendingCents'] as num?)?.toInt() ?? 0,
      availableCents: (json['availableCents'] as num?)?.toInt() ?? 0,
      paidOutCents: (json['paidOutCents'] as num?)?.toInt() ?? 0,
      charges: (json['charges'] as List<dynamic>?)
              ?.map((e) => EarningsCharge.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  static EarningsReport empty() => EarningsReport(
        currency: 'USD',
        netEarningsCents: 0,
        pendingCents: 0,
        availableCents: 0,
        paidOutCents: 0,
        charges: const [],
      );
}

class EarningsReportService {
  final ApiClient _client;

  EarningsReportService(this._client);

  Future<EarningsReport> fetchReport(
    String appId, {
    String? from,
    String? to,
    CancelToken? cancelToken,
  }) async {
    final queryParameters = <String, dynamic>{};
    if (from != null) queryParameters['from'] = from;
    if (to != null) queryParameters['to'] = to;
    final response = await _client.get(
      '/api/v1/apps/$appId/reports/earnings',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
    );
    return EarningsReport.fromJson(response.data as Map<String, dynamic>);
  }

  /// Fetches the CSV export of the earnings report through the authenticated
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
      '/api/v1/apps/$appId/reports/earnings',
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
    debugPrint('[EarningsReportService] bad date: $value');
    return null;
  }
}
