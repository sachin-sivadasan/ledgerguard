import 'package:equatable/equatable.dart';

import '../../../domain/entities/earnings_timeline.dart';

/// Base class for earnings events
abstract class EarningsEvent extends Equatable {
  const EarningsEvent();

  @override
  List<Object?> get props => [];
}

/// Request to load earnings for current month
class LoadEarningsRequested extends EarningsEvent {
  const LoadEarningsRequested();
}

/// Request to navigate to previous month
class PreviousMonthRequested extends EarningsEvent {
  const PreviousMonthRequested();
}

/// Request to navigate to next month
class NextMonthRequested extends EarningsEvent {
  const NextMonthRequested();
}

/// Request to change earnings display mode
class EarningsModeChanged extends EarningsEvent {
  final EarningsMode mode;

  const EarningsModeChanged(this.mode);

  @override
  List<Object?> get props => [mode];
}

/// Request to refresh earnings data
class RefreshEarningsRequested extends EarningsEvent {
  const RefreshEarningsRequested();
}
