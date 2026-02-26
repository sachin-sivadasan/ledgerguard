import 'package:equatable/equatable.dart';

import '../../../domain/entities/risk_summary.dart';

/// Base class for risk states
abstract class RiskState extends Equatable {
  const RiskState();

  @override
  List<Object?> get props => [];
}

/// Initial state before loading
class RiskInitial extends RiskState {
  const RiskInitial();
}

/// Loading risk summary
class RiskLoading extends RiskState {
  const RiskLoading();
}

/// Risk summary loaded successfully
class RiskLoaded extends RiskState {
  final RiskSummary summary;
  final bool isRefreshing;

  const RiskLoaded({
    required this.summary,
    this.isRefreshing = false,
  });

  RiskLoaded copyWith({
    RiskSummary? summary,
    bool? isRefreshing,
  }) {
    return RiskLoaded(
      summary: summary ?? this.summary,
      isRefreshing: isRefreshing ?? this.isRefreshing,
    );
  }

  @override
  List<Object?> get props => [summary, isRefreshing];
}

/// No risk data available
class RiskEmpty extends RiskState {
  final String message;

  const RiskEmpty({
    this.message = 'No risk data available. Sync your app data first.',
  });

  @override
  List<Object?> get props => [message];
}

/// Error loading risk summary
class RiskError extends RiskState {
  final String message;

  const RiskError(this.message);

  @override
  List<Object?> get props => [message];
}
