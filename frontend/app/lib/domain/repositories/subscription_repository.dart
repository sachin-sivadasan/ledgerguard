import '../entities/subscription.dart';
import '../entities/subscription_filter.dart';

/// Repository interface for subscription operations
abstract class SubscriptionRepository {
  /// Fetch subscriptions for an app with optional filtering
  @Deprecated('Use getSubscriptionsFiltered instead')
  Future<SubscriptionListResponse> getSubscriptions(
    String appId, {
    RiskState? riskState,
    int limit = 50,
    int offset = 0,
  });

  /// Fetch subscriptions with advanced filtering and pagination
  Future<PaginatedSubscriptionResponse> getSubscriptionsFiltered(
    String appId, {
    SubscriptionFilters? filters,
    int page = 1,
    int pageSize = 25,
  });

  /// Get subscription summary statistics
  Future<SubscriptionSummary> getSummary(String appId);

  /// Get price statistics for filtering
  Future<PriceStats> getPriceStats(String appId);

  /// Get a single subscription by ID
  Future<Subscription> getSubscription(String appId, String subscriptionId);
}

/// Exception for subscription-related errors
class SubscriptionException implements Exception {
  final String message;
  final String? code;

  const SubscriptionException(this.message, {this.code});

  @override
  String toString() => message;
}

/// Subscription not found
class SubscriptionNotFoundException extends SubscriptionException {
  const SubscriptionNotFoundException()
      : super('Subscription not found');
}

/// Failed to fetch subscriptions
class FetchSubscriptionsException extends SubscriptionException {
  const FetchSubscriptionsException([String message = 'Failed to fetch subscriptions'])
      : super(message);
}

/// Not authorized to access subscriptions
class SubscriptionUnauthorizedException extends SubscriptionException {
  const SubscriptionUnauthorizedException()
      : super('Not authorized to access this subscription');
}
