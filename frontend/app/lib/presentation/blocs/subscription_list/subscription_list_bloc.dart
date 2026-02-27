import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../domain/repositories/subscription_repository.dart';
import 'subscription_list_event.dart';
import 'subscription_list_state.dart';

/// Bloc for managing subscription list state
class SubscriptionListBloc
    extends Bloc<SubscriptionListEvent, SubscriptionListState> {
  final SubscriptionRepository _repository;

  String? _currentAppId;

  SubscriptionListBloc({
    required SubscriptionRepository repository,
  })  : _repository = repository,
        super(const SubscriptionListInitial()) {
    on<FetchSubscriptionsRequested>(_onFetchSubscriptions);
    on<RefreshSubscriptionsRequested>(_onRefreshSubscriptions);
    on<FilterByRiskStateRequested>(_onFilterByRiskState);
    on<LoadMoreSubscriptionsRequested>(_onLoadMore);
  }

  Future<void> _onFetchSubscriptions(
    FetchSubscriptionsRequested event,
    Emitter<SubscriptionListState> emit,
  ) async {
    _currentAppId = event.appId;
    emit(const SubscriptionListLoading());

    try {
      final response = await _repository.getSubscriptions(event.appId);

      if (response.subscriptions.isEmpty) {
        emit(SubscriptionListEmpty(appId: event.appId));
      } else {
        emit(SubscriptionListLoaded(
          subscriptions: response.subscriptions,
          total: response.total,
          hasMore: response.hasMore,
          appId: event.appId,
        ));
      }
    } on SubscriptionException catch (e) {
      emit(SubscriptionListError(e.message));
    } catch (e) {
      emit(SubscriptionListError('Failed to load subscriptions: $e'));
    }
  }

  Future<void> _onRefreshSubscriptions(
    RefreshSubscriptionsRequested event,
    Emitter<SubscriptionListState> emit,
  ) async {
    final currentState = state;
    if (currentState is SubscriptionListLoaded && _currentAppId != null) {
      emit(currentState.copyWith(isRefreshing: true));

      try {
        final response = await _repository.getSubscriptions(
          _currentAppId!,
          riskState: currentState.filterRiskState,
        );

        if (response.subscriptions.isEmpty) {
          emit(SubscriptionListEmpty(appId: _currentAppId!));
        } else {
          emit(SubscriptionListLoaded(
            subscriptions: response.subscriptions,
            total: response.total,
            hasMore: response.hasMore,
            filterRiskState: currentState.filterRiskState,
            appId: _currentAppId!,
          ));
        }
      } catch (e) {
        emit(currentState.copyWith(isRefreshing: false));
      }
    } else if (_currentAppId != null) {
      add(FetchSubscriptionsRequested(appId: _currentAppId!));
    }
  }

  Future<void> _onFilterByRiskState(
    FilterByRiskStateRequested event,
    Emitter<SubscriptionListState> emit,
  ) async {
    if (_currentAppId == null) return;

    final currentState = state;
    if (currentState is SubscriptionListLoaded) {
      emit(currentState.copyWith(isRefreshing: true));
    } else {
      emit(const SubscriptionListLoading());
    }

    try {
      final response = await _repository.getSubscriptions(
        _currentAppId!,
        riskState: event.riskState,
      );

      if (response.subscriptions.isEmpty) {
        emit(SubscriptionListEmpty(appId: _currentAppId!));
      } else {
        emit(SubscriptionListLoaded(
          subscriptions: response.subscriptions,
          total: response.total,
          hasMore: response.hasMore,
          filterRiskState: event.riskState,
          appId: _currentAppId!,
        ));
      }
    } on SubscriptionException catch (e) {
      emit(SubscriptionListError(e.message));
    } catch (e) {
      emit(SubscriptionListError('Failed to filter subscriptions: $e'));
    }
  }

  Future<void> _onLoadMore(
    LoadMoreSubscriptionsRequested event,
    Emitter<SubscriptionListState> emit,
  ) async {
    final currentState = state;
    if (currentState is SubscriptionListLoaded &&
        currentState.hasMore &&
        !currentState.isLoadingMore &&
        _currentAppId != null) {
      emit(currentState.copyWith(isLoadingMore: true));

      try {
        final response = await _repository.getSubscriptions(
          _currentAppId!,
          riskState: currentState.filterRiskState,
          offset: currentState.subscriptions.length,
        );

        emit(currentState.copyWith(
          subscriptions: [
            ...currentState.subscriptions,
            ...response.subscriptions,
          ],
          total: response.total,
          hasMore: response.hasMore,
          isLoadingMore: false,
        ));
      } catch (e) {
        emit(currentState.copyWith(isLoadingMore: false));
      }
    }
  }
}
