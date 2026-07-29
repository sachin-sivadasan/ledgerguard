import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';

/// A single upcoming-payout group in the schedule timeline.
class PayoutScheduleRow {
  /// Available date as raw YYYY-MM-DD, or "" when the charge is unscheduled.
  final String availableDate;
  final int amountCents;
  final int chargeCount;
  final String status; // "Available" or "Pending"

  PayoutScheduleRow({
    required this.availableDate,
    required this.amountCents,
    required this.chargeCount,
    required this.status,
  });

  /// Parsed date, or null when unscheduled/unparseable (UI renders "—").
  DateTime? get date => _parseDate(availableDate);

  factory PayoutScheduleRow.fromJson(Map<String, dynamic> json) {
    return PayoutScheduleRow(
      availableDate: json['availableDate'] as String? ?? '',
      amountCents: (json['amountCents'] as num?)?.toInt() ?? 0,
      chargeCount: (json['chargeCount'] as num?)?.toInt() ?? 0,
      status: json['status'] as String? ?? '',
    );
  }
}

/// Full Payout Schedule report payload.
class PayoutScheduleReport {
  final String currency;
  final int upcomingPayoutCents;
  final int pendingCents;

  /// Earliest scheduled payout date as raw YYYY-MM-DD, or "" when none.
  final String nextPayoutDate;
  final List<PayoutScheduleRow> rows;

  /// Full payout-row count before ?limit/?offset paging — drives the report
  /// preview's "View all N" affordance and the detail page's pagination.
  final int rowsTotal;

  PayoutScheduleReport({
    required this.currency,
    required this.upcomingPayoutCents,
    required this.pendingCents,
    required this.nextPayoutDate,
    required this.rows,
    required this.rowsTotal,
  });

  /// Parsed next-payout date, or null when none (UI renders "—").
  DateTime? get nextPayoutDateTime => _parseDate(nextPayoutDate);

  factory PayoutScheduleReport.fromJson(Map<String, dynamic> json) {
    final rows = (json['rows'] as List<dynamic>?)
            ?.map((e) => PayoutScheduleRow.fromJson(e as Map<String, dynamic>))
            .toList() ??
        [];
    return PayoutScheduleReport(
      currency: json['currency'] as String? ?? 'USD',
      upcomingPayoutCents: (json['upcomingPayoutCents'] as num?)?.toInt() ?? 0,
      pendingCents: (json['pendingCents'] as num?)?.toInt() ?? 0,
      nextPayoutDate: json['nextPayoutDate'] as String? ?? '',
      rows: rows,
      // Fall back to the visible count for older responses without the field.
      rowsTotal: (json['rowsTotal'] as num?)?.toInt() ?? rows.length,
    );
  }

  static PayoutScheduleReport empty() => PayoutScheduleReport(
        currency: 'USD',
        upcomingPayoutCents: 0,
        pendingCents: 0,
        nextPayoutDate: '',
        rows: const [],
        rowsTotal: 0,
      );
}

class PayoutScheduleService {
  final ApiClient _client;

  PayoutScheduleService(this._client);

  Future<PayoutScheduleReport> fetchReport(
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
      '/api/v1/apps/$appId/reports/payout-schedule',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
    );
    return PayoutScheduleReport.fromJson(response.data as Map<String, dynamic>);
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
      '/api/v1/apps/$appId/reports/payout-schedule',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
      options: Options(responseType: ResponseType.bytes),
    );
    return Uint8List.fromList(response.data ?? const <int>[]);
  }
}

/// Rows shown in the Payout Schedule preview before "View all".
const int kPayoutSchedulePreview = 8;

/// Rows per page on the dedicated Payout Schedule payouts detail screen.
const int kPayoutSchedulePageSize = 50;

DateTime? _parseDate(String? value) {
  if (value == null || value.isEmpty) return null;
  try {
    return DateTime.parse(value);
  } catch (e) {
    debugPrint('[PayoutScheduleService] bad date: $value');
    return null;
  }
}
