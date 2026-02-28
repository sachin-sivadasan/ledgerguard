import '../../core/network/api_client.dart';
import '../../domain/entities/subscription.dart';
import '../../domain/entities/subscription_filter.dart';
import '../../domain/repositories/subscription_repository.dart';

/// API implementation of SubscriptionRepository
class ApiSubscriptionRepository implements SubscriptionRepository {
  final ApiClient _apiClient;

  ApiSubscriptionRepository({required ApiClient apiClient})
      : _apiClient = apiClient;

  /// Extracts numeric ID from Shopify GID format
  /// e.g., "gid://partners/App/4599915" -> "4599915"
  String _extractNumericId(String gid) {
    final parts = gid.split('/');
    return parts.isNotEmpty ? parts.last : gid;
  }

  @override
  Future<SubscriptionListResponse> getSubscriptions(
    String appId, {
    RiskState? riskState,
    int limit = 50,
    int offset = 0,
  }) async {
    try {
      // Extract numeric ID from GID if needed
      final numericAppId = _extractNumericId(appId);

      final queryParams = <String, dynamic>{
        'limit': limit.toString(),
        'offset': offset.toString(),
      };

      if (riskState != null) {
        queryParams['risk_state'] = riskState.apiValue;
      }

      final response = await _apiClient.get(
        '/api/v1/apps/$numericAppId/subscriptions',
        queryParameters: queryParams,
      );

      if (response.statusCode == 200) {
        final data = response.data as Map<String, dynamic>;
        return SubscriptionListResponse.fromJson(data);
      }

      if (response.statusCode == 401) {
        throw const SubscriptionUnauthorizedException();
      }

      if (response.statusCode == 403) {
        throw const SubscriptionUnauthorizedException();
      }

      if (response.statusCode == 404) {
        throw const SubscriptionNotFoundException();
      }

      throw FetchSubscriptionsException(
        'Failed to fetch subscriptions: ${response.statusCode}',
      );
    } catch (e) {
      if (e is SubscriptionException) rethrow;
      throw FetchSubscriptionsException(e.toString());
    }
  }

  @override
  Future<Subscription> getSubscription(
    String appId,
    String subscriptionId,
  ) async {
    try {
      // Extract numeric ID from GID if needed
      final numericAppId = _extractNumericId(appId);

      final response = await _apiClient.get(
        '/api/v1/apps/$numericAppId/subscriptions/$subscriptionId',
      );

      if (response.statusCode == 200) {
        final data = response.data as Map<String, dynamic>;
        final subData = data['subscription'] as Map<String, dynamic>;
        return Subscription.fromJson(subData);
      }

      if (response.statusCode == 401) {
        throw const SubscriptionUnauthorizedException();
      }

      if (response.statusCode == 403) {
        throw const SubscriptionUnauthorizedException();
      }

      if (response.statusCode == 404) {
        throw const SubscriptionNotFoundException();
      }

      throw SubscriptionException(
        'Failed to fetch subscription: ${response.statusCode}',
      );
    } catch (e) {
      if (e is SubscriptionException) rethrow;
      throw SubscriptionException(e.toString());
    }
  }

  @override
  Future<PaginatedSubscriptionResponse> getSubscriptionsFiltered(
    String appId, {
    SubscriptionFilters? filters,
    int page = 1,
    int pageSize = 25,
  }) async {
    try {
      final numericAppId = _extractNumericId(appId);

      final queryParams = <String, dynamic>{
        'page': page.toString(),
        'pageSize': pageSize.toString(),
      };

      // Add filter params
      if (filters != null) {
        queryParams.addAll(filters.toQueryParams());
      }

      final response = await _apiClient.get(
        '/api/v1/apps/$numericAppId/subscriptions',
        queryParameters: queryParams,
      );

      if (response.statusCode == 200) {
        final data = response.data as Map<String, dynamic>;
        return PaginatedSubscriptionResponse.fromJson(data);
      }

      if (response.statusCode == 401 || response.statusCode == 403) {
        throw const SubscriptionUnauthorizedException();
      }

      if (response.statusCode == 404) {
        throw const SubscriptionNotFoundException();
      }

      throw FetchSubscriptionsException(
        'Failed to fetch subscriptions: ${response.statusCode}',
      );
    } catch (e) {
      if (e is SubscriptionException) rethrow;
      throw FetchSubscriptionsException(e.toString());
    }
  }

  @override
  Future<SubscriptionSummary> getSummary(String appId) async {
    try {
      final numericAppId = _extractNumericId(appId);

      final response = await _apiClient.get(
        '/api/v1/apps/$numericAppId/subscriptions/summary',
      );

      if (response.statusCode == 200) {
        final data = response.data as Map<String, dynamic>;
        return SubscriptionSummary.fromJson(data);
      }

      if (response.statusCode == 401 || response.statusCode == 403) {
        throw const SubscriptionUnauthorizedException();
      }

      if (response.statusCode == 404) {
        throw const SubscriptionNotFoundException();
      }

      throw FetchSubscriptionsException(
        'Failed to fetch summary: ${response.statusCode}',
      );
    } catch (e) {
      if (e is SubscriptionException) rethrow;
      throw FetchSubscriptionsException(e.toString());
    }
  }

  @override
  Future<PriceStats> getPriceStats(String appId) async {
    try {
      final numericAppId = _extractNumericId(appId);

      final response = await _apiClient.get(
        '/api/v1/apps/$numericAppId/subscriptions/price-stats',
      );

      if (response.statusCode == 200) {
        final data = response.data as Map<String, dynamic>;
        return PriceStats.fromJson(data);
      }

      if (response.statusCode == 401 || response.statusCode == 403) {
        throw const SubscriptionUnauthorizedException();
      }

      if (response.statusCode == 404) {
        throw const SubscriptionNotFoundException();
      }

      throw FetchSubscriptionsException(
        'Failed to fetch price stats: ${response.statusCode}',
      );
    } catch (e) {
      if (e is SubscriptionException) rethrow;
      throw FetchSubscriptionsException(e.toString());
    }
  }
}
