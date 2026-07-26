import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';

/// A single uninstalled store row in the Uninstall Context report.
class UninstallStore {
  final String domain;

  /// Inferred risk state just before uninstall: Healthy, At-Risk, Frozen, or
  /// Unknown. This is inferred from risk signals — NOT a self-reported reason.
  final String stateBeforeUninstall;
  final String planName;

  /// Tenure in months from install to uninstall.
  final double tenureMonths;
  final DateTime? uninstalledDate;

  UninstallStore({
    required this.domain,
    required this.stateBeforeUninstall,
    required this.planName,
    required this.tenureMonths,
    required this.uninstalledDate,
  });

  factory UninstallStore.fromJson(Map<String, dynamic> json) {
    return UninstallStore(
      domain: json['domain'] as String? ?? '',
      stateBeforeUninstall:
          json['stateBeforeUninstall'] as String? ?? 'Unknown',
      planName: json['planName'] as String? ?? '',
      tenureMonths: (json['tenureMonths'] as num?)?.toDouble() ?? 0,
      uninstalledDate: _parseDate(json['uninstalledDate'] as String?),
    );
  }
}

/// Full Uninstall Context report payload.
class UninstallContextReport {
  final int uninstalls;

  /// Share of correlated uninstalls that were in a risk/frozen state before
  /// uninstalling, as a 0..1 decimal (e.g. 0.71 == 71%). Clamped to [0,1] in
  /// [fromJson] so the invariant holds regardless of producer.
  final double wereAtRiskPct;
  final double medianTenureMonths;
  final List<UninstallStore> stores;

  UninstallContextReport({
    required this.uninstalls,
    required this.wereAtRiskPct,
    required this.medianTenureMonths,
    required this.stores,
  });

  factory UninstallContextReport.fromJson(Map<String, dynamic> json) {
    return UninstallContextReport(
      uninstalls: (json['uninstalls'] as num?)?.toInt() ?? 0,
      wereAtRiskPct:
          ((json['wereAtRiskPct'] as num?)?.toDouble() ?? 0).clamp(0.0, 1.0),
      medianTenureMonths:
          (json['medianTenureMonths'] as num?)?.toDouble() ?? 0,
      stores: (json['stores'] as List<dynamic>?)
              ?.map((e) => UninstallStore.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  static UninstallContextReport empty() => UninstallContextReport(
        uninstalls: 0,
        wereAtRiskPct: 0,
        medianTenureMonths: 0,
        stores: const [],
      );
}

class UninstallContextService {
  final ApiClient _client;

  UninstallContextService(this._client);

  Future<UninstallContextReport> fetchReport(
    String appId, {
    String? from,
    String? to,
    CancelToken? cancelToken,
  }) async {
    final queryParameters = <String, dynamic>{};
    if (from != null) queryParameters['from'] = from;
    if (to != null) queryParameters['to'] = to;
    final response = await _client.get(
      '/api/v1/apps/$appId/reports/uninstall-context',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
    );
    return UninstallContextReport.fromJson(
        response.data as Map<String, dynamic>);
  }

  /// Fetches the CSV export of the uninstall context report through the
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
      '/api/v1/apps/$appId/reports/uninstall-context',
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
    debugPrint('[UninstallContextService] bad date: $value');
    return null;
  }
}
