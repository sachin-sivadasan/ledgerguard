import 'package:equatable/equatable.dart';

/// User roles matching backend
enum UserRole {
  owner,
  admin;

  static UserRole fromString(String value) {
    switch (value.toUpperCase()) {
      case 'OWNER':
        return UserRole.owner;
      case 'ADMIN':
        return UserRole.admin;
      default:
        return UserRole.admin;
    }
  }

  String toJson() => name.toUpperCase();

  /// Check if this role has at least the given role's permissions
  bool hasPermission(UserRole required) {
    // OWNER has all permissions, ADMIN has only ADMIN permissions
    if (this == UserRole.owner) return true;
    return this == required;
  }
}

/// Plan tiers matching backend
enum PlanTier {
  starter,
  pro;

  static PlanTier fromString(String value) {
    switch (value.toUpperCase()) {
      case 'PRO':
        return PlanTier.pro;
      case 'STARTER':
      default:
        return PlanTier.starter;
    }
  }

  String toJson() => name.toUpperCase();

  bool get isPro => this == PlanTier.pro;
}

/// User profile with role and plan information
class UserProfile extends Equatable {
  final String id;
  final String email;
  final UserRole role;
  final PlanTier planTier;
  final String? displayName;

  const UserProfile({
    required this.id,
    required this.email,
    required this.role,
    required this.planTier,
    this.displayName,
  });

  bool get isOwner => role == UserRole.owner;
  bool get isAdmin => role == UserRole.admin || role == UserRole.owner;
  bool get isPro => planTier.isPro;

  /// Check if user has required role permission
  bool hasRole(UserRole required) => role.hasPermission(required);

  @override
  List<Object?> get props => [id, email, role, planTier, displayName];
}
