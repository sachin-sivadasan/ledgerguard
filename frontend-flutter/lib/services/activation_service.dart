import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';

/// A single stage in the activation funnel.
class ActivationStage {
  final String key;
  final String label;
  final int count;

  /// Conversion from the prior stage as a 0..1 decimal. Clamped to [0,1] in
  /// [fromJson]. The first stage is 1.0.
  final double pctOfPrior;

  ActivationStage({
    required this.key,
    required this.label,
    required this.count,
    required this.pctOfPrior,
  });

  factory ActivationStage.fromJson(Map<String, dynamic> json) {
    return ActivationStage(
      key: json['key'] as String? ?? '',
      label: json['label'] as String? ?? '',
      count: (json['count'] as num?)?.toInt() ?? 0,
      pctOfPrior: ((json['pctOfPrior'] as num?)?.toDouble() ?? 0).clamp(
        0.0,
        1.0,
      ),
    );
  }
}

/// Full Activation report payload: the install-to-paid conversion funnel.
class ActivationReport {
  final int installs;
  final int started;
  final int paid;

  /// Overall install → paid conversion, 0..1. Clamped in [fromJson].
  final double overallPct;

  /// Install → subscription conversion, 0..1. Clamped in [fromJson].
  final double installToSubPct;

  /// Subscription → paid conversion, 0..1. Clamped in [fromJson].
  final double subToPaidPct;

  final List<ActivationStage> stages;

  ActivationReport({
    required this.installs,
    required this.started,
    required this.paid,
    required this.overallPct,
    required this.installToSubPct,
    required this.subToPaidPct,
    required this.stages,
  });

  factory ActivationReport.fromJson(Map<String, dynamic> json) {
    return ActivationReport(
      installs: (json['installs'] as num?)?.toInt() ?? 0,
      started: (json['started'] as num?)?.toInt() ?? 0,
      paid: (json['paid'] as num?)?.toInt() ?? 0,
      overallPct: ((json['overallPct'] as num?)?.toDouble() ?? 0).clamp(
        0.0,
        1.0,
      ),
      installToSubPct: ((json['installToSubPct'] as num?)?.toDouble() ?? 0)
          .clamp(0.0, 1.0),
      subToPaidPct: ((json['subToPaidPct'] as num?)?.toDouble() ?? 0).clamp(
        0.0,
        1.0,
      ),
      stages:
          (json['stages'] as List<dynamic>?)
              ?.map((e) => ActivationStage.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  static ActivationReport empty() => ActivationReport(
    installs: 0,
    started: 0,
    paid: 0,
    overallPct: 0,
    installToSubPct: 0,
    subToPaidPct: 0,
    stages: const [],
  );
}

class ActivationService {
  final ApiClient _client;

  ActivationService(this._client);

  Future<ActivationReport> fetchReport(
    String appId, {
    String? from,
    String? to,
    CancelToken? cancelToken,
  }) async {
    final queryParameters = <String, dynamic>{};
    if (from != null) queryParameters['from'] = from;
    if (to != null) queryParameters['to'] = to;
    final response = await _client.get(
      '/api/v1/apps/$appId/reports/activation',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
    );
    return ActivationReport.fromJson(response.data as Map<String, dynamic>);
  }

  /// Fetches the CSV export of the Activation report through the authenticated
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
      '/api/v1/apps/$appId/reports/activation',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
      options: Options(responseType: ResponseType.bytes),
    );
    return Uint8List.fromList(response.data ?? const <int>[]);
  }
}
