import 'package:equatable/equatable.dart';

abstract class BillingEvent extends Equatable {
  const BillingEvent();

  @override
  List<Object?> get props => [];
}

/// Load the current billing status.
class LoadBillingStatusRequested extends BillingEvent {
  const LoadBillingStatusRequested();
}

/// Start a checkout session for the given plan.
class StartCheckoutRequested extends BillingEvent {
  final String plan;

  const StartCheckoutRequested({required this.plan});

  @override
  List<Object?> get props => [plan];
}
