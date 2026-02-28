import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../domain/repositories/store_health_repository.dart';
import 'store_health_event.dart';
import 'store_health_state.dart';

/// BLoC for managing store health data
class StoreHealthBloc extends Bloc<StoreHealthEvent, StoreHealthState> {
  final StoreHealthRepository _storeHealthRepository;

  String? _currentAppId;
  String? _currentDomain;

  StoreHealthBloc(this._storeHealthRepository)
      : super(const StoreHealthInitial()) {
    on<LoadStoreHealthRequested>(_onLoadStoreHealth);
    on<RefreshStoreHealthRequested>(_onRefreshStoreHealth);
  }

  Future<void> _onLoadStoreHealth(
    LoadStoreHealthRequested event,
    Emitter<StoreHealthState> emit,
  ) async {
    _currentAppId = event.appId;
    _currentDomain = event.domain;

    emit(const StoreHealthLoading());

    try {
      final storeHealth = await _storeHealthRepository.getStoreHealth(
        event.appId,
        event.domain,
      );
      emit(StoreHealthLoaded(storeHealth: storeHealth));
    } on StoreNotFoundException catch (e) {
      emit(StoreHealthNotFound(e.domain));
    } catch (e) {
      emit(StoreHealthError(e.toString()));
    }
  }

  Future<void> _onRefreshStoreHealth(
    RefreshStoreHealthRequested event,
    Emitter<StoreHealthState> emit,
  ) async {
    if (_currentAppId == null || _currentDomain == null) {
      return;
    }

    final currentState = state;
    if (currentState is StoreHealthLoaded) {
      emit(currentState.copyWith(isRefreshing: true));
    }

    try {
      final storeHealth = await _storeHealthRepository.getStoreHealth(
        _currentAppId!,
        _currentDomain!,
      );
      emit(StoreHealthLoaded(storeHealth: storeHealth));
    } on StoreNotFoundException catch (e) {
      emit(StoreHealthNotFound(e.domain));
    } catch (e) {
      emit(StoreHealthError(e.toString()));
    }
  }
}
