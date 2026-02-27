import 'package:equatable/equatable.dart';

import '../../../domain/entities/subscription.dart';

/// Base class for subscription list states
abstract class SubscriptionListState extends Equatable {
  const SubscriptionListState();

  @override
  List<Object?> get props => [];
}

/// Initial state before loading
class SubscriptionListInitial extends SubscriptionListState {
  const SubscriptionListInitial();
}

/// Loading subscriptions
class SubscriptionListLoading extends SubscriptionListState {
  const SubscriptionListLoading();
}

/// Subscriptions loaded successfully
class SubscriptionListLoaded extends SubscriptionListState {
  final List<Subscription> subscriptions;
  final int total;
  final bool hasMore;
  final bool isRefreshing;
  final bool isLoadingMore;
  final RiskState? filterRiskState;
  final String appId;

  const SubscriptionListLoaded({
    required this.subscriptions,
    required this.total,
    required this.appId,
    this.hasMore = false,
    this.isRefreshing = false,
    this.isLoadingMore = false,
    this.filterRiskState,
  });

  SubscriptionListLoaded copyWith({
    List<Subscription>? subscriptions,
    int? total,
    bool? hasMore,
    bool? isRefreshing,
    bool? isLoadingMore,
    RiskState? filterRiskState,
    bool clearFilter = false,
    String? appId,
  }) {
    return SubscriptionListLoaded(
      subscriptions: subscriptions ?? this.subscriptions,
      total: total ?? this.total,
      hasMore: hasMore ?? this.hasMore,
      isRefreshing: isRefreshing ?? this.isRefreshing,
      isLoadingMore: isLoadingMore ?? this.isLoadingMore,
      filterRiskState: clearFilter ? null : (filterRiskState ?? this.filterRiskState),
      appId: appId ?? this.appId,
    );
  }

  @override
  List<Object?> get props => [
        subscriptions,
        total,
        hasMore,
        isRefreshing,
        isLoadingMore,
        filterRiskState,
        appId,
      ];
}

/// No subscriptions found
class SubscriptionListEmpty extends SubscriptionListState {
  final String message;
  final String appId;

  const SubscriptionListEmpty({
    this.message = 'No subscriptions found',
    required this.appId,
  });

  @override
  List<Object?> get props => [message, appId];
}

/// Error loading subscriptions
class SubscriptionListError extends SubscriptionListState {
  final String message;

  const SubscriptionListError(this.message);

  @override
  List<Object?> get props => [message];
}
