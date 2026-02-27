import 'package:equatable/equatable.dart';

/// Base class for insight events
abstract class InsightEvent extends Equatable {
  const InsightEvent();

  @override
  List<Object?> get props => [];
}

/// Request to load daily insight
class LoadInsightRequested extends InsightEvent {
  const LoadInsightRequested();
}

/// Request to refresh daily insight
class RefreshInsightRequested extends InsightEvent {
  const RefreshInsightRequested();
}
