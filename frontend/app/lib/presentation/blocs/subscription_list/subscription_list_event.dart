import 'package:equatable/equatable.dart';

import '../../../domain/entities/subscription.dart';
import '../../../domain/entities/subscription_filter.dart';

/// Base class for subscription list events
abstract class SubscriptionListEvent extends Equatable {
  const SubscriptionListEvent();

  @override
  List<Object?> get props => [];
}

/// Request to load subscriptions with summary and price ranges
class LoadSubscriptionsRequested extends SubscriptionListEvent {
  final String appId;

  const LoadSubscriptionsRequested({required this.appId});

  @override
  List<Object?> get props => [appId];
}

/// Legacy: Request to load subscriptions (redirects to LoadSubscriptionsRequested)
class FetchSubscriptionsRequested extends SubscriptionListEvent {
  final String appId;

  const FetchSubscriptionsRequested({required this.appId});

  @override
  List<Object?> get props => [appId];
}

/// Request to refresh subscriptions
class RefreshSubscriptionsRequested extends SubscriptionListEvent {
  const RefreshSubscriptionsRequested();
}

/// Apply new filters
class ApplyFiltersRequested extends SubscriptionListEvent {
  final SubscriptionFilters filters;

  const ApplyFiltersRequested(this.filters);

  @override
  List<Object?> get props => [filters];
}

/// Change page
class ChangePageRequested extends SubscriptionListEvent {
  final int page;

  const ChangePageRequested(this.page);

  @override
  List<Object?> get props => [page];
}

/// Change page size
class ChangePageSizeRequested extends SubscriptionListEvent {
  final int pageSize;

  const ChangePageSizeRequested(this.pageSize);

  @override
  List<Object?> get props => [pageSize];
}

/// Change sort
class ChangeSortRequested extends SubscriptionListEvent {
  final SubscriptionSort sort;
  final bool ascending;

  const ChangeSortRequested(this.sort, {this.ascending = true});

  @override
  List<Object?> get props => [sort, ascending];
}

/// Search subscriptions
class SearchRequested extends SubscriptionListEvent {
  final String query;

  const SearchRequested(this.query);

  @override
  List<Object?> get props => [query];
}

/// Clear all filters
class ClearFiltersRequested extends SubscriptionListEvent {
  const ClearFiltersRequested();
}

/// Legacy: Filter by risk state (maps to ApplyFiltersRequested)
class FilterByRiskStateRequested extends SubscriptionListEvent {
  final RiskState? riskState;

  const FilterByRiskStateRequested({this.riskState});

  @override
  List<Object?> get props => [riskState];
}

/// Legacy: Load more (redirects to next page)
class LoadMoreSubscriptionsRequested extends SubscriptionListEvent {
  const LoadMoreSubscriptionsRequested();
}
