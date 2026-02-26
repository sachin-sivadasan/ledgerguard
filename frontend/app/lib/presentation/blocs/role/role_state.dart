import 'package:equatable/equatable.dart';

import '../../../domain/entities/user_profile.dart';

/// Base class for role states
abstract class RoleState extends Equatable {
  const RoleState();

  @override
  List<Object?> get props => [];
}

/// Initial state before role is fetched
class RoleInitial extends RoleState {
  const RoleInitial();
}

/// Loading state while fetching role
class RoleLoading extends RoleState {
  const RoleLoading();
}

/// Role loaded successfully
class RoleLoaded extends RoleState {
  final UserProfile profile;

  const RoleLoaded(this.profile);

  /// Convenience getters
  bool get isOwner => profile.isOwner;
  bool get isAdmin => profile.isAdmin;
  bool get isPro => profile.isPro;
  UserRole get role => profile.role;
  PlanTier get planTier => profile.planTier;

  /// Check if user has required role
  bool hasRole(UserRole required) => profile.hasRole(required);

  @override
  List<Object?> get props => [profile];
}

/// Error fetching role
class RoleError extends RoleState {
  final String message;

  const RoleError(this.message);

  @override
  List<Object?> get props => [message];
}
