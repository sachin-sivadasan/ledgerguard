import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';

/// A single charge-type segment in the Revenue Mix report.
class RevenueMixSegment {
  final String type;
  final int amountCents;

  /// Share of gross revenue as a 0..1 decimal (clamped).
  final double pct;

  RevenueMixSegment({
    required this.type,
    required this.amountCents,
    required this.pct,
  });

  factory RevenueMixSegment.fromJson(Map<String, dynamic> json) {
    return RevenueMixSegment(
      type: json['type'] as String? ?? '',
      amountCents: (json['amountCents'] as num?)?.toInt() ?? 0,
      // Backend sends pct as a 0..1 decimal — clamp defensively.
      pct: ((json['pct'] as num?)?.toDouble() ?? 0).clamp(0.0, 1.0),
    );
  }
}

/// Full Revenue Mix report payload — composition of revenue by charge type.
class RevenueMixReport {
  final String currency;
  final int recurringCents;
  final int usageCents;
  final int oneTimeCents;
  final int refundCents;
  final int grossCents;
  final int netCents;
  final List<RevenueMixSegment> segments;

  RevenueMixReport({
    required this.currency,
    required this.recurringCents,
    required this.usageCents,
    required this.oneTimeCents,
    required this.refundCents,
    required this.grossCents,
    required this.netCents,
    required this.segments,
  });

  factory RevenueMixReport.fromJson(Map<String, dynamic> json) {
    return RevenueMixReport(
      currency: json['currency'] as String? ?? 'USD',
      recurringCents: (json['recurringCents'] as num?)?.toInt() ?? 0,
      usageCents: (json['usageCents'] as num?)?.toInt() ?? 0,
      oneTimeCents: (json['oneTimeCents'] as num?)?.toInt() ?? 0,
      refundCents: (json['refundCents'] as num?)?.toInt() ?? 0,
      grossCents: (json['grossCents'] as num?)?.toInt() ?? 0,
      netCents: (json['netCents'] as num?)?.toInt() ?? 0,
      segments: (json['segments'] as List<dynamic>?)
              ?.map((e) => RevenueMixSegment.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  static RevenueMixReport empty() => RevenueMixReport(
        currency: 'USD',
        recurringCents: 0,
        usageCents: 0,
        oneTimeCents: 0,
        refundCents: 0,
        grossCents: 0,
        netCents: 0,
        segments: const [],
      );
}

class RevenueMixService {
  final ApiClient _client;

  RevenueMixService(this._client);

  Future<RevenueMixReport> fetchReport(
    String appId, {
    String? from,
    String? to,
    CancelToken? cancelToken,
  }) async {
    final queryParameters = <String, dynamic>{};
    if (from != null) queryParameters['from'] = from;
    if (to != null) queryParameters['to'] = to;
    final response = await _client.get(
      '/api/v1/apps/$appId/reports/revenue-mix',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
    );
    return RevenueMixReport.fromJson(response.data as Map<String, dynamic>);
  }

  /// Fetches the CSV export of the revenue mix report through the authenticated
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
      '/api/v1/apps/$appId/reports/revenue-mix',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
      options: Options(responseType: ResponseType.bytes),
    );
    return Uint8List.fromList(response.data ?? const <int>[]);
  }
}
