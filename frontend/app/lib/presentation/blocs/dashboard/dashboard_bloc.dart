import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../domain/entities/time_range.dart';
import '../../../domain/repositories/dashboard_repository.dart';
import 'dashboard_event.dart';
import 'dashboard_state.dart';

/// Bloc for managing dashboard state
class DashboardBloc extends Bloc<DashboardEvent, DashboardState> {
  final DashboardRepository _repository;

  /// Current time range (defaults to this month)
  TimeRange _currentTimeRange = TimeRange.thisMonth();

  DashboardBloc({
    required DashboardRepository repository,
  })  : _repository = repository,
        super(const DashboardInitial()) {
    on<LoadDashboardRequested>(_onLoadDashboard);
    on<RefreshDashboardRequested>(_onRefreshDashboard);
    on<TimeRangeChanged>(_onTimeRangeChanged);
  }

  /// Get the current time range
  TimeRange get currentTimeRange => _currentTimeRange;

  Future<void> _onLoadDashboard(
    LoadDashboardRequested event,
    Emitter<DashboardState> emit,
  ) async {
    emit(DashboardLoading(timeRange: _currentTimeRange));

    try {
      final metrics = await _repository.fetchMetrics(
        timeRange: _currentTimeRange,
      );
      if (metrics == null) {
        emit(const DashboardEmpty());
      } else {
        emit(DashboardLoaded(
          metrics: metrics,
          timeRange: _currentTimeRange,
        ));
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
        final metrics = await _repository.refreshMetrics(
          timeRange: _currentTimeRange,
        );
        if (metrics == null) {
          emit(const DashboardEmpty());
        } else {
          emit(DashboardLoaded(
            metrics: metrics,
            timeRange: _currentTimeRange,
          ));
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

  Future<void> _onTimeRangeChanged(
    TimeRangeChanged event,
    Emitter<DashboardState> emit,
  ) async {
    _currentTimeRange = event.timeRange;

    final currentState = state;
    if (currentState is DashboardLoaded) {
      // Show loading overlay while fetching new data
      emit(currentState.copyWith(isRefreshing: true));

      try {
        final metrics = await _repository.fetchMetrics(
          timeRange: _currentTimeRange,
        );
        if (metrics == null) {
          emit(const DashboardEmpty());
        } else {
          emit(DashboardLoaded(
            metrics: metrics,
            timeRange: _currentTimeRange,
          ));
        }
      } on DashboardException catch (e) {
        emit(currentState.copyWith(
          isRefreshing: false,
          timeRange: _currentTimeRange,
        ));
      } catch (e) {
        emit(currentState.copyWith(
          isRefreshing: false,
          timeRange: _currentTimeRange,
        ));
      }
    } else {
      // If not loaded, trigger initial load with new time range
      add(const LoadDashboardRequested());
    }
  }
}
