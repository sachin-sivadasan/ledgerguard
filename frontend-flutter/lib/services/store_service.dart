import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';
import '../models/paginated_result.dart';
import '../models/store_model.dart';

class StoreService {
  final ApiClient _client;

  StoreService(this._client);

  Future<PaginatedResult<Store>> fetchStores(
    String appId, {
    int page = 1,
    int pageSize = 20,
    String? search,
    CancelToken? cancelToken,
  }) async {
    try {
      final params = <String, dynamic>{'page': page, 'pageSize': pageSize};
      if (search != null && search.isNotEmpty) params['search'] = search;
      final response = await _client.get(
        '/api/v1/apps/$appId/stores',
        queryParameters: params,
        cancelToken: cancelToken,
      );
      final data = response.data;
      final list = data['stores'] as List<dynamic>? ?? [];
      return PaginatedResult(
        items: list
            .map((json) => Store.fromJson(json as Map<String, dynamic>))
            .toList(),
        total: data['total'] as int? ?? list.length,
        page: data['page'] as int? ?? 1,
        pageSize: data['pageSize'] as int? ?? pageSize,
        totalPages: data['totalPages'] as int? ?? 1,
      );
    } on DioException catch (e) {
      debugPrint('[StoreService] error: ${e.response?.statusCode}');
      return const PaginatedResult(
          items: [], total: 0, page: 1, pageSize: 20, totalPages: 0);
    }
  }
}
