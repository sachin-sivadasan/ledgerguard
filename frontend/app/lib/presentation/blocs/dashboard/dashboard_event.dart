import 'package:equatable/equatable.dart';

/// Base class for dashboard events
abstract class DashboardEvent extends Equatable {
  const DashboardEvent();

  @override
  List<Object?> get props => [];
}

/// Request to load dashboard metrics
class LoadDashboardRequested extends DashboardEvent {
  const LoadDashboardRequested();
}

/// Request to refresh dashboard metrics
class RefreshDashboardRequested extends DashboardEvent {
  const RefreshDashboardRequested();
}
