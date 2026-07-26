import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';
import '../models/analytics_model.dart';

/// Service for the Retention Cohorts report. Reuses [CohortData] from
/// analytics_model.dart (snake_case `cohort_month`/`initial_stores`/
/// `retention_pcts`).
class CohortsService {
  final ApiClient _client;

  CohortsService(this._client);

  /// Fetches the cohort retention heatmap for [appId]. Parses
  /// `response.data['cohorts']` into a list of [CohortData].
  Future<List<CohortData>> fetchCohorts(
    String appId, {
    int months = 6,
    CancelToken? cancelToken,
  }) async {
    final response = await _client.get(
      '/api/v1/apps/$appId/reports/cohorts',
      queryParameters: {'months': months},
      cancelToken: cancelToken,
    );
    final list = (response.data as Map<String, dynamic>)['cohorts']
            as List<dynamic>? ??
        const [];
    return list
        .map((e) => CohortData.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  /// Fetches the CSV export of the cohort heatmap through the authenticated
  /// [ApiClient] (Firebase Bearer token injected by the Dio interceptor).
  ///
  /// Returns the raw response bytes so the caller can trigger a client-side
  /// download without relying on an external browser navigation (which would
  /// 401 because it carries no auth header).
  Future<Uint8List> fetchCsvBytes(
    String appId, {
    int months = 6,
    CancelToken? cancelToken,
  }) async {
    final response = await _client.get<List<int>>(
      '/api/v1/apps/$appId/reports/cohorts',
      queryParameters: {'format': 'csv', 'months': months},
      cancelToken: cancelToken,
      options: Options(responseType: ResponseType.bytes),
    );
    return Uint8List.fromList(response.data ?? const <int>[]);
  }
}
