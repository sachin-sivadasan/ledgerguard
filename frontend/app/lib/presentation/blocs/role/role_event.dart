import 'package:equatable/equatable.dart';

/// Base class for role events
abstract class RoleEvent extends Equatable {
  const RoleEvent();

  @override
  List<Object?> get props => [];
}

/// Fetch user role from backend
class FetchRoleRequested extends RoleEvent {
  final String authToken;

  const FetchRoleRequested({required this.authToken});

  @override
  List<Object?> get props => [authToken];
}

/// Clear role state (on logout)
class ClearRoleRequested extends RoleEvent {
  const ClearRoleRequested();
}
