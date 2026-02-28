import 'package:equatable/equatable.dart';

import '../../../domain/entities/subscription.dart';
import '../../../domain/entities/subscription_filter.dart';

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
  final SubscriptionSummary summary;
  final PriceStats? priceStats;
  final SubscriptionFilters filters;
  final int page;
  final int pageSize;
  final int total;
  final int totalPages;
  final String appId;
  final bool isLoading;
  final bool isRefreshing;

  const SubscriptionListLoaded({
    required this.subscriptions,
    required this.summary,
    this.priceStats,
    required this.filters,
    required this.page,
    required this.pageSize,
    required this.total,
    required this.totalPages,
    required this.appId,
    this.isLoading = false,
    this.isRefreshing = false,
  });

  /// For backward compatibility
  bool get hasMore => page < totalPages;
  RiskState? get filterRiskState =>
      filters.riskStates.length == 1 ? filters.riskStates.first : null;
  bool get isLoadingMore => isLoading && page > 1;

  SubscriptionListLoaded copyWith({
    List<Subscription>? subscriptions,
    SubscriptionSummary? summary,
    PriceStats? priceStats,
    SubscriptionFilters? filters,
    int? page,
    int? pageSize,
    int? total,
    int? totalPages,
    String? appId,
    bool? isLoading,
    bool? isRefreshing,
  }) {
    return SubscriptionListLoaded(
      subscriptions: subscriptions ?? this.subscriptions,
      summary: summary ?? this.summary,
      priceStats: priceStats ?? this.priceStats,
      filters: filters ?? this.filters,
      page: page ?? this.page,
      pageSize: pageSize ?? this.pageSize,
      total: total ?? this.total,
      totalPages: totalPages ?? this.totalPages,
      appId: appId ?? this.appId,
      isLoading: isLoading ?? this.isLoading,
      isRefreshing: isRefreshing ?? this.isRefreshing,
    );
  }

  @override
  List<Object?> get props => [
        subscriptions,
        summary,
        priceStats,
        filters,
        page,
        pageSize,
        total,
        totalPages,
        appId,
        isLoading,
        isRefreshing,
      ];
}

/// No subscriptions found
class SubscriptionListEmpty extends SubscriptionListState {
  final String message;
  final String appId;
  final SubscriptionSummary? summary;
  final PriceStats? priceStats;
  final SubscriptionFilters filters;

  const SubscriptionListEmpty({
    this.message = 'No subscriptions found',
    required this.appId,
    this.summary,
    this.priceStats,
    this.filters = const SubscriptionFilters(),
  });

  @override
  List<Object?> get props => [message, appId, summary, priceStats, filters];
}

/// Error loading subscriptions
class SubscriptionListError extends SubscriptionListState {
  final String message;

  const SubscriptionListError(this.message);

  @override
  List<Object?> get props => [message];
}
