import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../domain/repositories/insight_repository.dart';
import 'insight_event.dart';
import 'insight_state.dart';

/// Bloc for managing AI insight state
class InsightBloc extends Bloc<InsightEvent, InsightState> {
  final InsightRepository _repository;

  InsightBloc({
    required InsightRepository repository,
  })  : _repository = repository,
        super(const InsightInitial()) {
    on<LoadInsightRequested>(_onLoadInsight);
    on<RefreshInsightRequested>(_onRefreshInsight);
  }

  Future<void> _onLoadInsight(
    LoadInsightRequested event,
    Emitter<InsightState> emit,
  ) async {
    emit(const InsightLoading());

    try {
      final insight = await _repository.fetchDailyInsight();
      if (insight == null) {
        emit(const InsightEmpty());
      } else {
        emit(InsightLoaded(insight: insight));
      }
    } on InsightException catch (e) {
      emit(InsightError(e.message));
    } catch (e) {
      emit(InsightError('Failed to load insight: $e'));
    }
  }

  Future<void> _onRefreshInsight(
    RefreshInsightRequested event,
    Emitter<InsightState> emit,
  ) async {
    final currentState = state;
    if (currentState is InsightLoaded) {
      emit(currentState.copyWith(isRefreshing: true));

      try {
        final insight = await _repository.fetchDailyInsight();
        if (insight == null) {
          emit(const InsightEmpty());
        } else {
          emit(InsightLoaded(insight: insight));
        }
      } on InsightException {
        emit(currentState.copyWith(isRefreshing: false));
      } catch (e) {
        emit(currentState.copyWith(isRefreshing: false));
      }
    } else {
      add(const LoadInsightRequested());
    }
  }
}
