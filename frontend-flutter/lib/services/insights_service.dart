import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';
import '../models/insight_model.dart';

class InsightsService {
  final ApiClient _client;

  InsightsService(this._client);

  Future<List<AiInsight>> fetchInsights(String appId,
      {CancelToken? cancelToken}) async {
    try {
      final response = await _client.get('/api/v1/apps/$appId/insights/daily',
          cancelToken: cancelToken);
      final data = response.data;
      if (data is Map<String, dynamic>) {
        // Single daily insight → wrap in list
        return [AiInsight.fromJson(data)];
      }
      final list = (data as List<dynamic>?) ?? [];
      return list
          .map((json) => AiInsight.fromJson(json as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      debugPrint('[InsightsService] fetchInsights error: ${e.response?.statusCode}');
      return [];
    }
  }
}
