import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';
import '../models/subscription_model.dart';

class SubscriptionService {
  final ApiClient _client;

  SubscriptionService(this._client);

  Future<List<Subscription>> fetchSubscriptions(String appId) async {
    try {
      final response = await _client.get('/api/v1/apps/$appId/subscriptions');
      final list = response.data['subscriptions'] as List<dynamic>? ?? [];
      return list
          .map((json) => Subscription.fromJson(json as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      debugPrint('[SubscriptionService] error: ${e.response?.statusCode}');
      return [];
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
