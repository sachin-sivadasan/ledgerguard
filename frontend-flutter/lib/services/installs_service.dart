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

  /// Returns null when the date is missing/unparseable so the caller can DROP the point
  /// rather than plot it at a sentinel date (an invented DateTime(1970) would silently
  /// mislabel the chart axis). The backend always emits a valid YYYY-MM-DD.
  static InstallTrendPoint? tryFromJson(Map<String, dynamic> json) {
    final date = _parseDate(json['date'] as String?);
    if (date == null) return null;
    return InstallTrendPoint(
      date: date,
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

/// All-time install-lifecycle snapshot (APPS-1b tiles) — distinct-shop counts.
class InstallLifecycle {
  final int active; // currently installed
  final int installed; // lifetime install base
  final int uninstalled; // currently uninstalled
  final int reactivated; // returning shops (ever reactivated)
  final int deactivated;

  const InstallLifecycle({
    this.active = 0,
    this.installed = 0,
    this.uninstalled = 0,
    this.reactivated = 0,
    this.deactivated = 0,
  });

  factory InstallLifecycle.fromJson(Map<String, dynamic> json) => InstallLifecycle(
        active: (json['active'] as num?)?.toInt() ?? 0,
        installed: (json['installed'] as num?)?.toInt() ?? 0,
        uninstalled: (json['uninstalled'] as num?)?.toInt() ?? 0,
        reactivated: (json['reactivated'] as num?)?.toInt() ?? 0,
        deactivated: (json['deactivated'] as num?)?.toInt() ?? 0,
      );
}

/// Install→paid conversion headline (APPS-1b). The full funnel is the Activation report.
class InstallConversion {
  final int installs; // lifetime install base (denominator)
  final int paid; // distinct shops that ever paid
  final double rate; // paid / installs, 0..1

  const InstallConversion({this.installs = 0, this.paid = 0, this.rate = 0});

  /// Whole-percent string for display, e.g. 0.238 → "24%".
  String get ratePercent => '${(rate * 100).round()}%';

  factory InstallConversion.fromJson(Map<String, dynamic> json) => InstallConversion(
        installs: (json['installs'] as num?)?.toInt() ?? 0,
        paid: (json['paid'] as num?)?.toInt() ?? 0,
        rate: (json['rate'] as num?)?.toDouble() ?? 0,
      );
}

/// Full Installs report payload.
class InstallsReport {
  final int installs;
  final int uninstalls;
  final int net;
  final List<InstallTrendPoint> trend;
  final List<InstallEvent> events;
  final InstallLifecycle lifecycle;
  final InstallConversion conversion;

  /// Full event-row count before ?limit/?offset paging — drives the report
  /// preview's "View all N" affordance and the detail page's pagination.
  final int eventsTotal;

  InstallsReport({
    required this.installs,
    required this.uninstalls,
    required this.net,
    required this.trend,
    required this.events,
    required this.eventsTotal,
    this.lifecycle = const InstallLifecycle(),
    this.conversion = const InstallConversion(),
  });

  factory InstallsReport.fromJson(Map<String, dynamic> json) {
    final events = (json['events'] as List<dynamic>?)
            ?.map((e) => InstallEvent.fromJson(e as Map<String, dynamic>))
            .toList() ??
        [];
    return InstallsReport(
      installs: (json['installs'] as num?)?.toInt() ?? 0,
      uninstalls: (json['uninstalls'] as num?)?.toInt() ?? 0,
      net: (json['net'] as num?)?.toInt() ?? 0,
      trend: (json['trend'] as List<dynamic>?)
              ?.map((e) => InstallTrendPoint.tryFromJson(e as Map<String, dynamic>))
              .whereType<InstallTrendPoint>()
              .toList() ??
          [],
      events: events,
      lifecycle: json['lifecycle'] is Map<String, dynamic>
          ? InstallLifecycle.fromJson(json['lifecycle'] as Map<String, dynamic>)
          : const InstallLifecycle(),
      conversion: json['conversion'] is Map<String, dynamic>
          ? InstallConversion.fromJson(json['conversion'] as Map<String, dynamic>)
          : const InstallConversion(),
      // Fall back to the visible count for older responses without the field.
      eventsTotal: (json['eventsTotal'] as num?)?.toInt() ?? events.length,
    );
  }

  static InstallsReport empty() => InstallsReport(
        installs: 0,
        uninstalls: 0,
        net: 0,
        trend: const [],
        events: const [],
        eventsTotal: 0,
      );
}

class InstallsService {
  final ApiClient _client;

  InstallsService(this._client);

  Future<InstallsReport> fetchReport(
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

/// Rows shown in the Installs report's events PREVIEW before "View all".
const int kInstallsEventsPreview = 8;

/// Rows per page on the dedicated Installs events detail screen.
const int kInstallsEventsPageSize = 50;

DateTime? _parseDate(String? value) {
  if (value == null || value.isEmpty) return null;
  try {
    return DateTime.parse(value);
  } catch (e) {
    debugPrint('[InstallsService] bad date: $value');
    return null;
  }
}
