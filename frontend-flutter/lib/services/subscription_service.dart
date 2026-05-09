import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';
import '../models/paginated_result.dart';
import '../models/subscription_model.dart';

class SubscriptionService {
  final ApiClient _client;

  SubscriptionService(this._client);

  Future<PaginatedResult<Subscription>> fetchSubscriptions(
    String appId, {
    int page = 1,
    int pageSize = 25,
    String? search,
    CancelToken? cancelToken,
  }) async {
    try {
      final params = <String, dynamic>{'page': page, 'pageSize': pageSize};
      if (search != null && search.isNotEmpty) params['search'] = search;
      final response = await _client.get(
        '/api/v1/apps/$appId/subscriptions',
        queryParameters: params,
        cancelToken: cancelToken,
      );
      final data = response.data;
      final list = data['subscriptions'] as List<dynamic>? ?? [];
      return PaginatedResult(
        items: list
            .map((json) => Subscription.fromJson(json as Map<String, dynamic>))
            .toList(),
        total: data['total'] as int? ?? list.length,
        page: data['page'] as int? ?? 1,
        pageSize: data['pageSize'] as int? ?? pageSize,
        totalPages: data['totalPages'] as int? ?? 1,
      );
    } on DioException catch (e) {
      debugPrint('[SubscriptionService] error: ${e.response?.statusCode}');
      return const PaginatedResult(
          items: [], total: 0, page: 1, pageSize: 25, totalPages: 0);
    }
  }

  Future<Subscription?> fetchSubscription(
      String appId, String subscriptionId) async {
    try {
      final response = await _client
          .get('/api/v1/apps/$appId/subscriptions/$subscriptionId');
      final data = response.data;
      final sub = data is Map<String, dynamic> && data.containsKey('subscription')
          ? data['subscription'] as Map<String, dynamic>
          : data as Map<String, dynamic>;
      return Subscription.fromJson(sub);
    } on DioException catch (e) {
      debugPrint('[SubscriptionService] fetchSubscription error: ${e.response?.statusCode}');
      return null;
    }
  }
}
