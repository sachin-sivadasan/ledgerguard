import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../domain/repositories/risk_repository.dart';
import 'risk_event.dart';
import 'risk_state.dart';

/// Bloc for managing risk summary state
class RiskBloc extends Bloc<RiskEvent, RiskState> {
  final RiskRepository _repository;

  RiskBloc({
    required RiskRepository repository,
  })  : _repository = repository,
        super(const RiskInitial()) {
    on<LoadRiskSummaryRequested>(_onLoadRiskSummary);
    on<RefreshRiskSummaryRequested>(_onRefreshRiskSummary);
  }

  Future<void> _onLoadRiskSummary(
    LoadRiskSummaryRequested event,
    Emitter<RiskState> emit,
  ) async {
    emit(const RiskLoading());

    try {
      final summary = await _repository.fetchRiskSummary();
      if (summary == null || !summary.hasData) {
        emit(const RiskEmpty());
      } else {
        emit(RiskLoaded(summary: summary));
      }
    } on RiskException catch (e) {
      emit(RiskError(e.message));
    } catch (e) {
      emit(RiskError('Failed to load risk summary: $e'));
    }
  }

  Future<void> _onRefreshRiskSummary(
    RefreshRiskSummaryRequested event,
    Emitter<RiskState> emit,
  ) async {
    final currentState = state;
    if (currentState is RiskLoaded) {
      emit(currentState.copyWith(isRefreshing: true));

      try {
        final summary = await _repository.fetchRiskSummary();
        if (summary == null || !summary.hasData) {
          emit(const RiskEmpty());
        } else {
          emit(RiskLoaded(summary: summary));
        }
      } on RiskException catch (e) {
        emit(currentState.copyWith(isRefreshing: false));
      } catch (e) {
        emit(currentState.copyWith(isRefreshing: false));
      }
    } else {
      add(const LoadRiskSummaryRequested());
    }
  }
}
