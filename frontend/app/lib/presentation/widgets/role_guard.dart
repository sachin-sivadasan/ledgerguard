import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../domain/entities/user_profile.dart';
import '../blocs/role/role.dart';

/// Widget that shows/hides content based on user role
class RoleGuard extends StatelessWidget {
  /// The child to show if role requirement is met
  final Widget child;

  /// Required role to view content
  final UserRole requiredRole;

  /// Widget to show if role requirement is not met (defaults to nothing)
  final Widget? fallback;

  /// Whether to show loading indicator while role is loading
  final bool showLoading;

  const RoleGuard({
    super.key,
    required this.child,
    required this.requiredRole,
    this.fallback,
    this.showLoading = false,
  });

  /// Show content only for owners
  const RoleGuard.ownerOnly({
    super.key,
    required this.child,
    this.fallback,
    this.showLoading = false,
  }) : requiredRole = UserRole.owner;

  /// Show content only for admins (includes owners)
  const RoleGuard.adminOnly({
    super.key,
    required this.child,
    this.fallback,
    this.showLoading = false,
  }) : requiredRole = UserRole.admin;

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<RoleBloc, RoleState>(
      builder: (context, state) {
        if (state is RoleLoading && showLoading) {
          return const Center(
            child: SizedBox(
              height: 20,
              width: 20,
              child: CircularProgressIndicator(strokeWidth: 2),
            ),
          );
        }

        if (state is RoleLoaded && state.hasRole(requiredRole)) {
          return child;
        }

        return fallback ?? const SizedBox.shrink();
      },
    );
  }
}

/// Widget that shows content only for Pro tier users
class ProGuard extends StatelessWidget {
  final Widget child;
  final Widget? fallback;
  final bool showLoading;

  const ProGuard({
    super.key,
    required this.child,
    this.fallback,
    this.showLoading = false,
  });

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<RoleBloc, RoleState>(
      builder: (context, state) {
        if (state is RoleLoading && showLoading) {
          return const Center(
            child: SizedBox(
              height: 20,
              width: 20,
              child: CircularProgressIndicator(strokeWidth: 2),
            ),
          );
        }

        if (state is RoleLoaded && state.isPro) {
          return child;
        }

        return fallback ?? const SizedBox.shrink();
      },
    );
  }
}
