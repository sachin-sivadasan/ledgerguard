import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';

/// A single day in the install/uninstall trend.
class InstallTrendPoint {
  final DateTime date;
  final int installs;
  final int uninstalls;

  InstallTrendPoint({
    required this.date,
    required this.installs,
    required this.uninstalls,
  });

  factory InstallTrendPoint.fromJson(Map<String, dynamic> json) {
    return InstallTrendPoint(
      date: _parseDate(json['date'] as String?) ?? DateTime(1970),
      installs: (json['installs'] as num?)?.toInt() ?? 0,
      uninstalls: (json['uninstalls'] as num?)?.toInt() ?? 0,
    );
  }
}

/// A single row in the recent install/uninstall events table.
class InstallEvent {
  final String domain;
  final String event; // "Install" or "Uninstall"
  final String date; // raw YYYY-MM-DD

  InstallEvent({required this.domain, required this.event, required this.date});

  /// Parsed date, or null when unparseable (UI renders raw/"—").
  DateTime? get dateTime => _parseDate(date);

  factory InstallEvent.fromJson(Map<String, dynamic> json) {
    return InstallEvent(
      domain: json['domain'] as String? ?? '',
      event: json['event'] as String? ?? '',
      date: json['date'] as String? ?? '',
    );
  }
}

/// Full Installs report payload.
class InstallsReport {
  final int installs;
  final int uninstalls;
  final int net;
  final List<InstallTrendPoint> trend;
  final List<InstallEvent> events;

  InstallsReport({
    required this.installs,
    required this.uninstalls,
    required this.net,
    required this.trend,
    required this.events,
  });

  factory InstallsReport.fromJson(Map<String, dynamic> json) {
    return InstallsReport(
      installs: (json['installs'] as num?)?.toInt() ?? 0,
      uninstalls: (json['uninstalls'] as num?)?.toInt() ?? 0,
      net: (json['net'] as num?)?.toInt() ?? 0,
      trend: (json['trend'] as List<dynamic>?)
              ?.map((e) => InstallTrendPoint.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      events: (json['events'] as List<dynamic>?)
              ?.map((e) => InstallEvent.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  static InstallsReport empty() => InstallsReport(
        installs: 0,
        uninstalls: 0,
        net: 0,
        trend: const [],
        events: const [],
      );
}

class InstallsService {
  final ApiClient _client;

  InstallsService(this._client);

  Future<InstallsReport> fetchReport(
    String appId, {
    String? from,
    String? to,
    CancelToken? cancelToken,
  }) async {
    final queryParameters = <String, dynamic>{};
    if (from != null) queryParameters['from'] = from;
    if (to != null) queryParameters['to'] = to;
    final response = await _client.get(
      '/api/v1/apps/$appId/reports/installs',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
    );
    return InstallsReport.fromJson(response.data as Map<String, dynamic>);
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
      '/api/v1/apps/$appId/reports/installs',
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
    debugPrint('[InstallsService] bad date: $value');
    return null;
  }
}
