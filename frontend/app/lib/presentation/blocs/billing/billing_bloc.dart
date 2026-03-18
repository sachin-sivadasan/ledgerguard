import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../domain/repositories/billing_repository.dart';
import 'billing_event.dart';
import 'billing_state.dart';

class BillingBloc extends Bloc<BillingEvent, BillingState> {
  final BillingRepository _billingRepository;

  BillingBloc({required BillingRepository billingRepository})
      : _billingRepository = billingRepository,
        super(const BillingInitial()) {
    on<LoadBillingStatusRequested>(_onLoadBillingStatus);
    on<StartCheckoutRequested>(_onStartCheckout);
  }

  Future<void> _onLoadBillingStatus(
    LoadBillingStatusRequested event,
    Emitter<BillingState> emit,
  ) async {
    emit(const BillingLoading());
    try {
      final status = await _billingRepository.getBillingStatus();
      emit(BillingLoaded(billingStatus: status));
    } catch (e) {
      emit(BillingError(message: e.toString()));
    }
  }

  Future<void> _onStartCheckout(
    StartCheckoutRequested event,
    Emitter<BillingState> emit,
  ) async {
    final currentState = state;
    final currentStatus = currentState is BillingLoaded
        ? currentState.billingStatus
        : null;

    if (currentStatus != null) {
      emit(BillingLoaded(
        billingStatus: currentStatus,
        isCheckingOut: true,
      ));
    }

    try {
      final result = await _billingRepository.createCheckout(event.plan);
      final status = currentStatus ??
          await _billingRepository.getBillingStatus();
      emit(BillingCheckoutReady(
        checkoutResult: result,
        billingStatus: status,
      ));
    } catch (e) {
      emit(BillingError(message: e.toString()));
    }
  }
}
