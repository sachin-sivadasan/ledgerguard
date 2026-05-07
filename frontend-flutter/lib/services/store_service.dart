import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';
import '../models/store_model.dart';

class StoreService {
  final ApiClient _client;

  StoreService(this._client);

  Future<List<Store>> fetchStores(String appId) async {
    try {
      final response = await _client.get('/api/v1/apps/$appId/stores');
      final list = response.data['stores'] as List<dynamic>? ?? [];
      return list
          .map((json) => Store.fromJson(json as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      debugPrint('[StoreService] error: ${e.response?.statusCode}');
      return [];
    }
  }
}
