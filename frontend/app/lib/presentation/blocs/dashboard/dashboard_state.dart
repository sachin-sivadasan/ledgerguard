import 'package:equatable/equatable.dart';

import '../../../domain/entities/dashboard_metrics.dart';

/// Base class for dashboard states
abstract class DashboardState extends Equatable {
  const DashboardState();

  @override
  List<Object?> get props => [];
}

/// Initial state before loading
class DashboardInitial extends DashboardState {
  const DashboardInitial();
}

/// Loading metrics
class DashboardLoading extends DashboardState {
  const DashboardLoading();
}

/// Metrics loaded successfully
class DashboardLoaded extends DashboardState {
  final DashboardMetrics metrics;
  final bool isRefreshing;

  const DashboardLoaded({
    required this.metrics,
    this.isRefreshing = false,
  });

  DashboardLoaded copyWith({
    DashboardMetrics? metrics,
    bool? isRefreshing,
  }) {
    return DashboardLoaded(
      metrics: metrics ?? this.metrics,
      isRefreshing: isRefreshing ?? this.isRefreshing,
    );
  }

  @override
  List<Object?> get props => [metrics, isRefreshing];
}

/// No metrics available (empty state)
class DashboardEmpty extends DashboardState {
  final String message;

  const DashboardEmpty({
    this.message = 'No metrics available. Sync your app data to see metrics.',
  });

  @override
  List<Object?> get props => [message];
}

/// Error loading metrics
class DashboardError extends DashboardState {
  final String message;

  const DashboardError(this.message);

  @override
  List<Object?> get props => [message];
}
