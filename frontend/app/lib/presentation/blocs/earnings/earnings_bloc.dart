import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../domain/entities/earnings_timeline.dart';
import '../../../domain/entities/time_range.dart';
import '../../../domain/repositories/earnings_repository.dart';
import 'earnings_event.dart';
import 'earnings_state.dart';

/// BLoC for managing earnings timeline state
class EarningsBloc extends Bloc<EarningsEvent, EarningsState> {
  final EarningsRepository _earningsRepository;

  int _currentYear = DateTime.now().year;
  int _currentMonth = DateTime.now().month;
  EarningsMode _currentMode = EarningsMode.combined;

  EarningsBloc({
    required EarningsRepository earningsRepository,
  })  : _earningsRepository = earningsRepository,
        super(const EarningsInitial()) {
    on<LoadEarningsRequested>(_onLoadEarningsRequested);
    on<EarningsTimeRangeChanged>(_onTimeRangeChanged);
    on<PreviousMonthRequested>(_onPreviousMonthRequested);
    on<NextMonthRequested>(_onNextMonthRequested);
    on<EarningsModeChanged>(_onEarningsModeChanged);
  }

  /// Extract target month from TimeRange preset
  (int year, int month) _getTargetMonth(TimeRange timeRange) {
    switch (timeRange.preset) {
      case TimeRangePreset.lastMonth:
        // Use start date for "Last Month"
        return (timeRange.start.year, timeRange.start.month);
      case TimeRangePreset.thisMonth:
      case TimeRangePreset.last30Days:
      case TimeRangePreset.last90Days:
      case TimeRangePreset.custom:
        // Use end date for all others
        return (timeRange.end.year, timeRange.end.month);
    }
  }

  Future<void> _onLoadEarningsRequested(
    LoadEarningsRequested event,
    Emitter<EarningsState> emit,
  ) async {
    final (year, month) = _getTargetMonth(event.timeRange);
    _currentYear = year;
    _currentMonth = month;

    await _loadMonth(emit);
  }

  Future<void> _onTimeRangeChanged(
    EarningsTimeRangeChanged event,
    Emitter<EarningsState> emit,
  ) async {
    final (year, month) = _getTargetMonth(event.timeRange);
    _currentYear = year;
    _currentMonth = month;

    await _loadMonth(emit);
  }

  Future<void> _onPreviousMonthRequested(
    PreviousMonthRequested event,
    Emitter<EarningsState> emit,
  ) async {
    if (_currentMonth == 1) {
      _currentMonth = 12;
      _currentYear--;
    } else {
      _currentMonth--;
    }

    await _loadMonth(emit);
  }

  Future<void> _onNextMonthRequested(
    NextMonthRequested event,
    Emitter<EarningsState> emit,
  ) async {
    final now = DateTime.now();

    // Don't allow going past current month
    if (_currentYear == now.year && _currentMonth >= now.month) {
      return;
    }

    if (_currentMonth == 12) {
      _currentMonth = 1;
      _currentYear++;
    } else {
      _currentMonth++;
    }

    await _loadMonth(emit);
  }

  Future<void> _onEarningsModeChanged(
    EarningsModeChanged event,
    Emitter<EarningsState> emit,
  ) async {
    _currentMode = event.mode;
    await _loadMonth(emit);
  }

  Future<void> _loadMonth(Emitter<EarningsState> emit) async {
    emit(EarningsLoading(year: _currentYear, month: _currentMonth));

    try {
      // Create date range for the month
      final startDate = DateTime(_currentYear, _currentMonth, 1);
      final endDate = DateTime(_currentYear, _currentMonth + 1, 0); // Last day of month

      final timeline = await _earningsRepository.fetchEarnings(
        startDate: startDate,
        endDate: endDate,
        mode: _currentMode,
      );

      final now = DateTime.now();
      final canGoNext = _currentYear < now.year || _currentMonth < now.month;

      if (timeline.earnings.isEmpty) {
        emit(EarningsEmpty(
          message: 'No earnings data for this month.',
          year: _currentYear,
          month: _currentMonth,
          canGoNext: canGoNext,
          canGoPrevious: true,
        ));
        return;
      }

      emit(EarningsLoaded(
        timeline: timeline,
        mode: _currentMode,
        year: _currentYear,
        month: _currentMonth,
        canGoNext: canGoNext,
        canGoPrevious: true,
      ));
    } on EarningsException catch (e) {
      emit(EarningsError(
        message: e.message,
        year: _currentYear,
        month: _currentMonth,
      ));
    } catch (e) {
      emit(EarningsError(
        message: 'Failed to load earnings: $e',
        year: _currentYear,
        month: _currentMonth,
      ));
    }
  }
}
