import 'package:equatable/equatable.dart';

import '../../../domain/entities/earnings_timeline.dart';
import '../../../domain/entities/time_range.dart';

/// Base class for earnings events
abstract class EarningsEvent extends Equatable {
  const EarningsEvent();

  @override
  List<Object?> get props => [];
}

/// Request to load earnings based on dashboard time range
/// This sets the initial month based on the time range preset
class LoadEarningsRequested extends EarningsEvent {
  final TimeRange timeRange;

  const LoadEarningsRequested(this.timeRange);

  @override
  List<Object?> get props => [timeRange];
}

/// Dashboard time range changed - sync to appropriate month
class EarningsTimeRangeChanged extends EarningsEvent {
  final TimeRange timeRange;

  const EarningsTimeRangeChanged(this.timeRange);

  @override
  List<Object?> get props => [timeRange];
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
