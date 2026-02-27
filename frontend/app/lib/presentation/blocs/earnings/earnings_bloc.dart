import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../domain/entities/earnings_timeline.dart';
import '../../../domain/repositories/earnings_repository.dart';
import 'earnings_event.dart';
import 'earnings_state.dart';

/// BLoC for managing earnings timeline state
class EarningsBloc extends Bloc<EarningsEvent, EarningsState> {
  final EarningsRepository _earningsRepository;

  int _currentYear;
  int _currentMonth;
  EarningsMode _currentMode;

  EarningsBloc({
    required EarningsRepository earningsRepository,
  })  : _earningsRepository = earningsRepository,
        _currentYear = DateTime.now().year,
        _currentMonth = DateTime.now().month,
        _currentMode = EarningsMode.combined,
        super(const EarningsInitial()) {
    on<LoadEarningsRequested>(_onLoadEarningsRequested);
    on<PreviousMonthRequested>(_onPreviousMonthRequested);
    on<NextMonthRequested>(_onNextMonthRequested);
    on<EarningsModeChanged>(_onEarningsModeChanged);
    on<RefreshEarningsRequested>(_onRefreshEarningsRequested);
  }

  Future<void> _onLoadEarningsRequested(
    LoadEarningsRequested event,
    Emitter<EarningsState> emit,
  ) async {
    // Reset to current month on initial load
    _currentYear = DateTime.now().year;
    _currentMonth = DateTime.now().month;

    emit(EarningsLoading(year: _currentYear, month: _currentMonth));

    try {
      final timeline = await _earningsRepository.fetchMonthlyEarnings(
        year: _currentYear,
        month: _currentMonth,
        mode: _currentMode,
      );

      if (timeline.earnings.isEmpty) {
        emit(EarningsEmpty(
          message: 'No earnings data for this month.',
          year: _currentYear,
          month: _currentMonth,
        ));
        return;
      }

      emit(EarningsLoaded(
        timeline: timeline,
        mode: _currentMode,
        canGoNext: false, // Current month, can't go forward
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

  Future<void> _onPreviousMonthRequested(
    PreviousMonthRequested event,
    Emitter<EarningsState> emit,
  ) async {
    // Navigate to previous month
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

    // Navigate to next month
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

    // Reload with new mode
    await _loadMonth(emit);
  }

  Future<void> _onRefreshEarningsRequested(
    RefreshEarningsRequested event,
    Emitter<EarningsState> emit,
  ) async {
    if (state is EarningsLoaded) {
      emit((state as EarningsLoaded).copyWith(isRefreshing: true));
    }

    try {
      final timeline = await _earningsRepository.fetchMonthlyEarnings(
        year: _currentYear,
        month: _currentMonth,
        mode: _currentMode,
      );

      final now = DateTime.now();
      final canGoNext =
          _currentYear < now.year || _currentMonth < now.month;

      if (timeline.earnings.isEmpty) {
        emit(EarningsEmpty(
          message: 'No earnings data for this month.',
          year: _currentYear,
          month: _currentMonth,
        ));
        return;
      }

      emit(EarningsLoaded(
        timeline: timeline,
        mode: _currentMode,
        canGoNext: canGoNext,
        canGoPrevious: true,
      ));
    } on EarningsException catch (e) {
      if (state is EarningsLoaded) {
        // Keep current data but show error
        emit((state as EarningsLoaded).copyWith(isRefreshing: false));
      } else {
        emit(EarningsError(
          message: e.message,
          year: _currentYear,
          month: _currentMonth,
        ));
      }
    }
  }

  Future<void> _loadMonth(Emitter<EarningsState> emit) async {
    emit(EarningsLoading(year: _currentYear, month: _currentMonth));

    try {
      final timeline = await _earningsRepository.fetchMonthlyEarnings(
        year: _currentYear,
        month: _currentMonth,
        mode: _currentMode,
      );

      final now = DateTime.now();
      final canGoNext =
          _currentYear < now.year || _currentMonth < now.month;

      if (timeline.earnings.isEmpty) {
        emit(EarningsEmpty(
          message: 'No earnings data for this month.',
          year: _currentYear,
          month: _currentMonth,
        ));
        return;
      }

      emit(EarningsLoaded(
        timeline: timeline,
        mode: _currentMode,
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
