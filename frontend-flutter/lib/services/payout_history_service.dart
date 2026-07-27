import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';

/// A single historical payout period (a calendar month of paid earnings).
class PayoutHistoryRow {
  /// Charge month as raw YYYY-MM.
  final String period;
  final int amountCents;
  final int chargeCount;

  /// Latest availability date in the period as raw YYYY-MM-DD, or "". This is the
  /// estimated earnings-available date, not Shopify's authoritative disbursement date.
  final String availableDate;

  PayoutHistoryRow({
    required this.period,
    required this.amountCents,
    required this.chargeCount,
    required this.availableDate,
  });

  /// Parsed period (first of the month), or null when unparseable.
  DateTime? get periodDate => _parseMonth(period);

  /// Parsed availability date, or null when unset/unparseable (UI renders "—").
  DateTime? get availableDateTime => _parseDate(availableDate);

  factory PayoutHistoryRow.fromJson(Map<String, dynamic> json) {
    return PayoutHistoryRow(
      period: json['period'] as String? ?? '',
      amountCents: (json['amountCents'] as num?)?.toInt() ?? 0,
      chargeCount: (json['chargeCount'] as num?)?.toInt() ?? 0,
      availableDate: json['availableDate'] as String? ?? '',
    );
  }
}

/// Full Payout History report payload.
class PayoutHistoryReport {
  final String currency;
  final int totalPaidCents;
  final int payoutCount;
  final int avgPayoutCents;
  final List<PayoutHistoryRow> rows;

  PayoutHistoryReport({
    required this.currency,
    required this.totalPaidCents,
    required this.payoutCount,
    required this.avgPayoutCents,
    required this.rows,
  });

  factory PayoutHistoryReport.fromJson(Map<String, dynamic> json) {
    return PayoutHistoryReport(
      currency: json['currency'] as String? ?? 'USD',
      totalPaidCents: (json['totalPaidCents'] as num?)?.toInt() ?? 0,
      payoutCount: (json['payoutCount'] as num?)?.toInt() ?? 0,
      avgPayoutCents: (json['avgPayoutCents'] as num?)?.toInt() ?? 0,
      rows: (json['rows'] as List<dynamic>?)
              ?.map((e) => PayoutHistoryRow.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  static PayoutHistoryReport empty() => PayoutHistoryReport(
        currency: 'USD',
        totalPaidCents: 0,
        payoutCount: 0,
        avgPayoutCents: 0,
        rows: const [],
      );
}

class PayoutHistoryService {
  final ApiClient _client;

  PayoutHistoryService(this._client);

  Future<PayoutHistoryReport> fetchReport(
    String appId, {
    String? from,
    String? to,
    CancelToken? cancelToken,
  }) async {
    final queryParameters = <String, dynamic>{};
    if (from != null) queryParameters['from'] = from;
    if (to != null) queryParameters['to'] = to;
    final response = await _client.get(
      '/api/v1/apps/$appId/reports/payout-history',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
    );
    return PayoutHistoryReport.fromJson(response.data as Map<String, dynamic>);
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
      '/api/v1/apps/$appId/reports/payout-history',
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
    debugPrint('[PayoutHistoryService] bad date: $value');
    return null;
  }
}

DateTime? _parseMonth(String? value) {
  if (value == null || value.isEmpty) return null;
  try {
    // "YYYY-MM" → first of the month.
    return DateTime.parse('$value-01');
  } catch (e) {
    debugPrint('[PayoutHistoryService] bad period: $value');
    return null;
  }
}
