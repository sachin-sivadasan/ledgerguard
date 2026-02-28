import 'package:equatable/equatable.dart';

import '../../../domain/entities/earnings_timeline.dart';

/// Base class for earnings states
abstract class EarningsState extends Equatable {
  const EarningsState();

  @override
  List<Object?> get props => [];
}

/// Initial state before any data is loaded
class EarningsInitial extends EarningsState {
  const EarningsInitial();
}

/// Loading state while fetching earnings data
class EarningsLoading extends EarningsState {
  final int year;
  final int month;

  const EarningsLoading({
    required this.year,
    required this.month,
  });

  @override
  List<Object?> get props => [year, month];
}

/// Loaded state with earnings data
class EarningsLoaded extends EarningsState {
  /// The earnings timeline data
  final EarningsTimeline timeline;

  /// Current display mode
  final EarningsMode mode;

  /// Current year being displayed
  final int year;

  /// Current month being displayed (1-12)
  final int month;

  /// Whether we can navigate to next month
  final bool canGoNext;

  /// Whether we can navigate to previous month
  final bool canGoPrevious;

  const EarningsLoaded({
    required this.timeline,
    required this.mode,
    required this.year,
    required this.month,
    this.canGoNext = false,
    this.canGoPrevious = true,
  });

  /// Copy with updated values
  EarningsLoaded copyWith({
    EarningsTimeline? timeline,
    EarningsMode? mode,
    int? year,
    int? month,
    bool? canGoNext,
    bool? canGoPrevious,
  }) {
    return EarningsLoaded(
      timeline: timeline ?? this.timeline,
      mode: mode ?? this.mode,
      year: year ?? this.year,
      month: month ?? this.month,
      canGoNext: canGoNext ?? this.canGoNext,
      canGoPrevious: canGoPrevious ?? this.canGoPrevious,
    );
  }

  @override
  List<Object?> get props => [timeline, mode, year, month, canGoNext, canGoPrevious];
}

/// Empty state when no earnings data exists
class EarningsEmpty extends EarningsState {
  final String message;
  final int year;
  final int month;
  final bool canGoNext;
  final bool canGoPrevious;

  const EarningsEmpty({
    required this.message,
    required this.year,
    required this.month,
    this.canGoNext = false,
    this.canGoPrevious = true,
  });

  @override
  List<Object?> get props => [message, year, month, canGoNext, canGoPrevious];
}

/// Error state when loading fails
class EarningsError extends EarningsState {
  final String message;
  final int year;
  final int month;

  const EarningsError({
    required this.message,
    required this.year,
    required this.month,
  });

  @override
  List<Object?> get props => [message, year, month];
}
