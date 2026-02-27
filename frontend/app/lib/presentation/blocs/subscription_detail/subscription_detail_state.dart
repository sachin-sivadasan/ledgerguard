import 'package:equatable/equatable.dart';

import '../../../domain/entities/subscription.dart';

/// Base class for subscription detail states
abstract class SubscriptionDetailState extends Equatable {
  const SubscriptionDetailState();

  @override
  List<Object?> get props => [];
}

/// Initial state before loading
class SubscriptionDetailInitial extends SubscriptionDetailState {
  const SubscriptionDetailInitial();
}

/// Loading subscription details
class SubscriptionDetailLoading extends SubscriptionDetailState {
  const SubscriptionDetailLoading();
}

/// Subscription details loaded successfully
class SubscriptionDetailLoaded extends SubscriptionDetailState {
  final Subscription subscription;
  final bool isRefreshing;

  const SubscriptionDetailLoaded({
    required this.subscription,
    this.isRefreshing = false,
  });

  SubscriptionDetailLoaded copyWith({
    Subscription? subscription,
    bool? isRefreshing,
  }) {
    return SubscriptionDetailLoaded(
      subscription: subscription ?? this.subscription,
      isRefreshing: isRefreshing ?? this.isRefreshing,
    );
  }

  @override
  List<Object?> get props => [subscription, isRefreshing];
}

/// Error loading subscription details
class SubscriptionDetailError extends SubscriptionDetailState {
  final String message;

  const SubscriptionDetailError(this.message);

  @override
  List<Object?> get props => [message];
}
