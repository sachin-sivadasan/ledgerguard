import 'package:equatable/equatable.dart';

import '../../../domain/entities/dashboard_metrics.dart';
import '../../../domain/entities/time_range.dart';

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
  final TimeRange? timeRange;

  const DashboardLoading({this.timeRange});

  @override
  List<Object?> get props => [timeRange];
}

/// Metrics loaded successfully
class DashboardLoaded extends DashboardState {
  final DashboardMetrics metrics;
  final TimeRange timeRange;
  final bool isRefreshing;

  const DashboardLoaded({
    required this.metrics,
    required this.timeRange,
    this.isRefreshing = false,
  });

  DashboardLoaded copyWith({
    DashboardMetrics? metrics,
    TimeRange? timeRange,
    bool? isRefreshing,
  }) {
    return DashboardLoaded(
      metrics: metrics ?? this.metrics,
      timeRange: timeRange ?? this.timeRange,
      isRefreshing: isRefreshing ?? this.isRefreshing,
    );
  }

  @override
  List<Object?> get props => [metrics, timeRange, isRefreshing];
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
