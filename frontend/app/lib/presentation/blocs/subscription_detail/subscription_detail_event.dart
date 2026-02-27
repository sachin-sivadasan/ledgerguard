import 'package:equatable/equatable.dart';

/// Base class for subscription detail events
abstract class SubscriptionDetailEvent extends Equatable {
  const SubscriptionDetailEvent();

  @override
  List<Object?> get props => [];
}

/// Request to fetch subscription details
class FetchSubscriptionRequested extends SubscriptionDetailEvent {
  final String appId;
  final String subscriptionId;

  const FetchSubscriptionRequested({
    required this.appId,
    required this.subscriptionId,
  });

  @override
  List<Object?> get props => [appId, subscriptionId];
}

/// Request to refresh subscription details
class RefreshSubscriptionRequested extends SubscriptionDetailEvent {
  const RefreshSubscriptionRequested();
}
