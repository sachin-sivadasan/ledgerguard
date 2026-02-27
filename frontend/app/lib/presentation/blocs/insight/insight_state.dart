import 'package:equatable/equatable.dart';

import '../../../domain/entities/daily_insight.dart';

/// Base class for insight states
abstract class InsightState extends Equatable {
  const InsightState();

  @override
  List<Object?> get props => [];
}

/// Initial state before loading
class InsightInitial extends InsightState {
  const InsightInitial();
}

/// Loading insight
class InsightLoading extends InsightState {
  const InsightLoading();
}

/// Insight loaded successfully
class InsightLoaded extends InsightState {
  final DailyInsight insight;
  final bool isRefreshing;

  const InsightLoaded({
    required this.insight,
    this.isRefreshing = false,
  });

  InsightLoaded copyWith({
    DailyInsight? insight,
    bool? isRefreshing,
  }) {
    return InsightLoaded(
      insight: insight ?? this.insight,
      isRefreshing: isRefreshing ?? this.isRefreshing,
    );
  }

  @override
  List<Object?> get props => [insight, isRefreshing];
}

/// No insight available
class InsightEmpty extends InsightState {
  const InsightEmpty();
}

/// Error loading insight
class InsightError extends InsightState {
  final String message;

  const InsightError(this.message);

  @override
  List<Object?> get props => [message];
}
