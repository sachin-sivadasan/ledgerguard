import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../domain/repositories/subscription_repository.dart';
import 'subscription_detail_event.dart';
import 'subscription_detail_state.dart';

/// Bloc for managing subscription detail state
class SubscriptionDetailBloc
    extends Bloc<SubscriptionDetailEvent, SubscriptionDetailState> {
  final SubscriptionRepository _repository;

  String? _currentAppId;
  String? _currentSubscriptionId;

  SubscriptionDetailBloc({
    required SubscriptionRepository repository,
  })  : _repository = repository,
        super(const SubscriptionDetailInitial()) {
    on<FetchSubscriptionRequested>(_onFetchSubscription);
    on<RefreshSubscriptionRequested>(_onRefreshSubscription);
  }

  Future<void> _onFetchSubscription(
    FetchSubscriptionRequested event,
    Emitter<SubscriptionDetailState> emit,
  ) async {
    _currentAppId = event.appId;
    _currentSubscriptionId = event.subscriptionId;
    emit(const SubscriptionDetailLoading());

    try {
      final subscription = await _repository.getSubscription(
        event.appId,
        event.subscriptionId,
      );
      emit(SubscriptionDetailLoaded(subscription: subscription));
    } on SubscriptionException catch (e) {
      emit(SubscriptionDetailError(e.message));
    } catch (e) {
      emit(SubscriptionDetailError('Failed to load subscription: $e'));
    }
  }

  Future<void> _onRefreshSubscription(
    RefreshSubscriptionRequested event,
    Emitter<SubscriptionDetailState> emit,
  ) async {
    final currentState = state;
    if (currentState is SubscriptionDetailLoaded &&
        _currentAppId != null &&
        _currentSubscriptionId != null) {
      emit(currentState.copyWith(isRefreshing: true));

      try {
        final subscription = await _repository.getSubscription(
          _currentAppId!,
          _currentSubscriptionId!,
        );
        emit(SubscriptionDetailLoaded(subscription: subscription));
      } catch (e) {
        emit(currentState.copyWith(isRefreshing: false));
      }
    } else if (_currentAppId != null && _currentSubscriptionId != null) {
      add(FetchSubscriptionRequested(
        appId: _currentAppId!,
        subscriptionId: _currentSubscriptionId!,
      ));
    }
  }
}
