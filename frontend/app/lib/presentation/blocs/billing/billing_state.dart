import 'package:equatable/equatable.dart';

import '../../../domain/entities/billing_status.dart';

abstract class BillingState extends Equatable {
  const BillingState();

  @override
  List<Object?> get props => [];
}

class BillingInitial extends BillingState {
  const BillingInitial();
}

class BillingLoading extends BillingState {
  const BillingLoading();
}

class BillingLoaded extends BillingState {
  final BillingStatus billingStatus;
  final bool isCheckingOut;

  const BillingLoaded({
    required this.billingStatus,
    this.isCheckingOut = false,
  });

  BillingLoaded copyWith({
    BillingStatus? billingStatus,
    bool? isCheckingOut,
  }) {
    return BillingLoaded(
      billingStatus: billingStatus ?? this.billingStatus,
      isCheckingOut: isCheckingOut ?? this.isCheckingOut,
    );
  }

  @override
  List<Object?> get props => [billingStatus, isCheckingOut];
}

class BillingCheckoutReady extends BillingState {
  final CheckoutResult checkoutResult;
  final BillingStatus billingStatus;

  const BillingCheckoutReady({
    required this.checkoutResult,
    required this.billingStatus,
  });

  @override
  List<Object?> get props => [checkoutResult, billingStatus];
}

class BillingError extends BillingState {
  final String message;

  const BillingError({required this.message});

  @override
  List<Object?> get props => [message];
}
