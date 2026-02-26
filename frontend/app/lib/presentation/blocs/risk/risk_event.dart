import 'package:equatable/equatable.dart';

/// Base class for risk events
abstract class RiskEvent extends Equatable {
  const RiskEvent();

  @override
  List<Object?> get props => [];
}

/// Request to load risk summary
class LoadRiskSummaryRequested extends RiskEvent {
  const LoadRiskSummaryRequested();
}

/// Request to refresh risk summary
class RefreshRiskSummaryRequested extends RiskEvent {
  const RefreshRiskSummaryRequested();
}
