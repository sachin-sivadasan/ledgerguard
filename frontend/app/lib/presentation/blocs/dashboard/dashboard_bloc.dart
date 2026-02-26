import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../domain/repositories/dashboard_repository.dart';
import 'dashboard_event.dart';
import 'dashboard_state.dart';

/// Bloc for managing dashboard state
class DashboardBloc extends Bloc<DashboardEvent, DashboardState> {
  final DashboardRepository _repository;

  DashboardBloc({
    required DashboardRepository repository,
  })  : _repository = repository,
        super(const DashboardInitial()) {
    on<LoadDashboardRequested>(_onLoadDashboard);
    on<RefreshDashboardRequested>(_onRefreshDashboard);
  }

  Future<void> _onLoadDashboard(
    LoadDashboardRequested event,
    Emitter<DashboardState> emit,
  ) async {
    emit(const DashboardLoading());

    try {
      final metrics = await _repository.fetchMetrics();
      if (metrics == null) {
        emit(const DashboardEmpty());
      } else {
        emit(DashboardLoaded(metrics: metrics));
      }
    } on DashboardException catch (e) {
      emit(DashboardError(e.message));
    } catch (e) {
      emit(DashboardError('Failed to load dashboard: $e'));
    }
  }

  Future<void> _onRefreshDashboard(
    RefreshDashboardRequested event,
    Emitter<DashboardState> emit,
  ) async {
    final currentState = state;
    if (currentState is DashboardLoaded) {
      emit(currentState.copyWith(isRefreshing: true));

      try {
        final metrics = await _repository.refreshMetrics();
        if (metrics == null) {
          emit(const DashboardEmpty());
        } else {
          emit(DashboardLoaded(metrics: metrics));
        }
      } on DashboardException catch (e) {
        emit(currentState.copyWith(isRefreshing: false));
        // Could also emit error, but we keep showing current data
      } catch (e) {
        emit(currentState.copyWith(isRefreshing: false));
      }
    } else {
      // If not loaded, treat as initial load
      add(const LoadDashboardRequested());
    }
  }
}
