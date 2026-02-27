import 'package:equatable/equatable.dart';

import '../../../domain/entities/subscription.dart';

/// Base class for subscription list events
abstract class SubscriptionListEvent extends Equatable {
  const SubscriptionListEvent();

  @override
  List<Object?> get props => [];
}

/// Request to load subscriptions
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

/// Request to filter by risk state
class FilterByRiskStateRequested extends SubscriptionListEvent {
  final RiskState? riskState;

  const FilterByRiskStateRequested({this.riskState});

  @override
  List<Object?> get props => [riskState];
}

/// Request to load more subscriptions (pagination)
class LoadMoreSubscriptionsRequested extends SubscriptionListEvent {
  const LoadMoreSubscriptionsRequested();
}
