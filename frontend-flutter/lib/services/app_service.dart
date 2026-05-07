import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';
import '../models/app_model.dart';
import '../models/review_model.dart';

class AppService {
  final ApiClient _client;

  AppService(this._client);

  Future<List<ShopifyApp>> fetchApps() async {
    final response = await _client.get('/api/v1/apps');
    debugPrint('[AppService] fetchApps response: ${response.data}');
    final list = response.data['apps'] as List<dynamic>? ?? [];
    debugPrint('[AppService] parsed ${list.length} raw apps');
    return list
        .map((json) => ShopifyApp.fromJson(json as Map<String, dynamic>))
        .toList();
  }

  Future<List<AppReview>> fetchReviews(String appId) async {
    try {
      final response = await _client.get('/api/v1/apps/$appId/reviews');
      final list = response.data['reviews'] as List<dynamic>? ?? [];
      return list
          .map((json) => AppReview.fromJson(json as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      debugPrint('[AppService] fetchReviews error: ${e.response?.statusCode}');
      return [];
    }
  }
}
