import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/user_profile.dart';
import '../blocs/auth/auth.dart';
import '../blocs/role/role.dart';
import '../widgets/shared.dart';

/// Profile page displaying user information and settings
class ProfilePage extends StatelessWidget {
  const ProfilePage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey[50],
      appBar: AppBar(
        title: const Text('Profile'),
      ),
      body: BlocBuilder<AuthBloc, AuthState>(
        builder: (context, authState) {
          if (authState is! Authenticated) {
            return const Center(child: CircularProgressIndicator());
          }

          return BlocBuilder<RoleBloc, RoleState>(
            builder: (context, roleState) {
              if (roleState is RoleLoading) {
                return const Center(child: CircularProgressIndicator());
              }

              if (roleState is RoleError) {
                return _buildErrorState(context, roleState.message);
              }

              if (roleState is RoleLoaded) {
                return _buildContent(
                  context,
                  email: authState.user.email ?? 'No email',
                  profile: roleState.profile,
                );
              }

              // Initial state - trigger load
              return const Center(child: CircularProgressIndicator());
            },
          );
        },
      ),
    );
  }

  Widget _buildErrorState(BuildContext context, String message) {
    return ErrorStateWidget(
      title: 'Failed to load profile',
      message: message,
    );
  }

  Widget _buildContent(
    BuildContext context, {
    required String email,
    required UserProfile profile,
  }) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Profile Header
          _buildProfileHeader(context, email, profile),

          const SizedBox(height: 24),

          // Account Section
          _buildSection(
            context,
            title: 'Account',
            children: [
              _buildInfoTile(
                context,
                icon: Icons.email_outlined,
                label: 'Email',
                value: email,
              ),
              const Divider(height: 1),
              _buildInfoTile(
                context,
                icon: Icons.badge_outlined,
                label: 'Role',
                value: profile.role.displayName,
                trailing: _buildRoleBadge(context, profile.role),
              ),
              const Divider(height: 1),
              _buildInfoTile(
                context,
                icon: Icons.workspace_premium_outlined,
                label: 'Plan',
                value: profile.planTier.displayName,
                trailing: _buildPlanBadge(context, profile.planTier),
              ),
            ],
          ),

          // Upgrade Button (only for FREE tier)
          if (profile.planTier.isFree) ...[
            const SizedBox(height: 16),
            _buildUpgradeCard(context),
          ],

          const SizedBox(height: 24),

          // Integrations Section
          _buildSection(
            context,
            title: 'Integrations',
            children: [
              _buildNavigationTile(
                context,
                icon: Icons.storefront_outlined,
                label: 'Shopify Partner Account',
                onTap: () => context.push('/partner-integration'),
              ),
            ],
          ),

          const SizedBox(height: 24),

          // Settings Section
          _buildSection(
            context,
            title: 'Settings',
            children: [
              _buildNavigationTile(
                context,
                icon: Icons.notifications_outlined,
                label: 'Notification Settings',
                onTap: () => context.push('/settings/notifications'),
              ),
              const Divider(height: 1),
              _buildNavigationTile(
                context,
                icon: Icons.key_outlined,
                label: 'API Keys',
                onTap: () => context.push('/settings/api-keys'),
              ),
            ],
          ),

          const SizedBox(height: 24),

          // Logout Button
          SizedBox(
            width: double.infinity,
            child: OutlinedButton.icon(
              onPressed: () => _showLogoutConfirmation(context),
              icon: const Icon(Icons.logout, color: AppTheme.danger),
              label: const Text(
                'Log Out',
                style: TextStyle(color: AppTheme.danger),
              ),
              style: OutlinedButton.styleFrom(
                padding: const EdgeInsets.symmetric(vertical: 16),
                side: const BorderSide(color: AppTheme.danger),
              ),
            ),
          ),

          const SizedBox(height: 32),
        ],
      ),
    );
  }

  Widget _buildProfileHeader(
    BuildContext context,
    String email,
    UserProfile profile,
  ) {
    return Center(
      child: Column(
        children: [
          CircleAvatar(
            radius: 48,
            backgroundColor: AppTheme.primary.withOpacity(0.1),
            child: Text(
              _getInitials(email),
              style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                    color: AppTheme.primary,
                    fontWeight: FontWeight.bold,
                  ),
            ),
          ),
          const SizedBox(height: 16),
          Text(
            email,
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w500,
                ),
          ),
          const SizedBox(height: 8),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              _buildRoleBadge(context, profile.role),
              const SizedBox(width: 8),
              _buildPlanBadge(context, profile.planTier),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildSection(
    BuildContext context, {
    required String title,
    required List<Widget> children,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.only(left: 4, bottom: 8),
          child: Text(
            title,
            style: Theme.of(context).textTheme.titleSmall?.copyWith(
                  color: Colors.grey[600],
                  fontWeight: FontWeight.w600,
                ),
          ),
        ),
        Card(
          elevation: 0,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
            side: BorderSide(color: Colors.grey[200]!),
          ),
          child: Column(children: children),
        ),
      ],
    );
  }

  Widget _buildInfoTile(
    BuildContext context, {
    required IconData icon,
    required String label,
    required String value,
    Widget? trailing,
  }) {
    return ListTile(
      leading: Icon(icon, color: Colors.grey[600]),
      title: Text(
        label,
        style: Theme.of(context).textTheme.bodySmall?.copyWith(
              color: Colors.grey[600],
            ),
      ),
      subtitle: Text(
        value,
        style: Theme.of(context).textTheme.bodyLarge?.copyWith(
              fontWeight: FontWeight.w500,
            ),
      ),
      trailing: trailing,
    );
  }

  Widget _buildNavigationTile(
    BuildContext context, {
    required IconData icon,
    required String label,
    required VoidCallback onTap,
  }) {
    return ListTile(
      leading: Icon(icon, color: Colors.grey[600]),
      title: Text(label),
      trailing: Icon(Icons.chevron_right, color: Colors.grey[400]),
      onTap: onTap,
    );
  }

  Widget _buildRoleBadge(BuildContext context, UserRole role) {
    return RoleBadge(role: role);
  }

  Widget _buildPlanBadge(BuildContext context, PlanTier tier) {
    return PlanBadge(tier: tier);
  }

  Widget _buildUpgradeCard(BuildContext context) {
    return Card(
      elevation: 0,
      color: AppTheme.primary.withOpacity(0.05),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: AppTheme.primary.withOpacity(0.2)),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: AppTheme.primary.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: const Icon(
                    Icons.rocket_launch,
                    color: AppTheme.primary,
                    size: 20,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Upgrade to Pro',
                        style: Theme.of(context).textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.bold,
                            ),
                      ),
                      Text(
                        'Unlock AI insights and advanced features',
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: Colors.grey[600],
                            ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: () => _showUpgradeComingSoon(context),
                child: const Text('Upgrade Now'),
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _showUpgradeComingSoon(BuildContext context) {
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Row(
          children: [
            Icon(Icons.info_outline, color: Colors.white),
            SizedBox(width: 8),
            Text('Upgrade functionality coming soon!'),
          ],
        ),
        behavior: SnackBarBehavior.floating,
      ),
    );
  }

  void _showLogoutConfirmation(BuildContext context) {
    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Log Out'),
        content: const Text('Are you sure you want to log out?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(dialogContext).pop(),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () {
              Navigator.of(dialogContext).pop();
              context.read<AuthBloc>().add(const SignOutRequested());
            },
            style: TextButton.styleFrom(foregroundColor: AppTheme.danger),
            child: const Text('Log Out'),
          ),
        ],
      ),
    );
  }

  String _getInitials(String email) {
    if (email.isEmpty) return '?';
    final parts = email.split('@');
    if (parts.isEmpty) return '?';
    final name = parts[0];
    if (name.isEmpty) return '?';
    if (name.length == 1) return name.toUpperCase();
    return name.substring(0, 2).toUpperCase();
  }
}
