import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/user_profile.dart';
import '../blocs/auth/auth.dart';
import '../blocs/role/role.dart';

/// Settings page providing access to all application settings
class SettingsPage extends StatelessWidget {
  const SettingsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Settings'),
        automaticallyImplyLeading: false,
      ),
      body: BlocBuilder<RoleBloc, RoleState>(
        builder: (context, roleState) {
          final isAdmin = roleState is RoleLoaded &&
              roleState.hasRole(UserRole.admin);

          return SingleChildScrollView(
            padding: const EdgeInsets.all(20),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // App Settings Section
                _buildSectionHeader(context, 'App Settings', Icons.tune_outlined),
                const SizedBox(height: 12),
                _buildCard(
                  context,
                  children: [
                    _buildNavigationTile(
                      context,
                      icon: Icons.dashboard_outlined,
                      iconColor: Colors.blue,
                      title: 'Preferences',
                      subtitle: 'Dashboard and display settings',
                      onTap: () => context.push('/settings/preferences'),
                    ),
                    _buildDivider(),
                    _buildNavigationTile(
                      context,
                      icon: Icons.tune_outlined,
                      iconColor: Colors.purple,
                      title: 'App Settings',
                      subtitle: 'Revenue share tier and fee settings',
                      onTap: () => context.push('/settings/app'),
                    ),
                  ],
                ),

                const SizedBox(height: 24),

                // Notifications Section
                _buildSectionHeader(context, 'Notifications', Icons.notifications_outlined),
                const SizedBox(height: 12),
                _buildCard(
                  context,
                  children: [
                    _buildNavigationTile(
                      context,
                      icon: Icons.notifications_outlined,
                      iconColor: Colors.orange,
                      title: 'Notification Preferences',
                      subtitle: 'Manage alert and summary settings',
                      onTap: () => context.push('/settings/notifications'),
                    ),
                  ],
                ),

                const SizedBox(height: 24),

                // Integrations Section
                _buildSectionHeader(context, 'Integrations', Icons.hub_outlined),
                const SizedBox(height: 12),
                _buildCard(
                  context,
                  children: [
                    _buildNavigationTile(
                      context,
                      icon: Icons.storefront_outlined,
                      iconColor: Colors.green,
                      title: 'Shopify Partner',
                      subtitle: 'Manage Partner API connection',
                      onTap: () => context.push('/partner-integration'),
                    ),
                    if (isAdmin) ...[
                      _buildDivider(),
                      _buildNavigationTile(
                        context,
                        icon: Icons.code_outlined,
                        iconColor: Colors.deepOrange,
                        title: 'Manual Integration',
                        subtitle: 'Admin: Configure token manually',
                        onTap: () => context.push('/admin/manual-integration'),
                      ),
                    ],
                  ],
                ),

                const SizedBox(height: 24),

                // Developer Section
                _buildSectionHeader(context, 'Developer', Icons.code_outlined),
                const SizedBox(height: 12),
                _buildCard(
                  context,
                  children: [
                    _buildNavigationTile(
                      context,
                      icon: Icons.key_outlined,
                      iconColor: Colors.indigo,
                      title: 'API Keys',
                      subtitle: 'Manage Revenue API access',
                      onTap: () => context.push('/settings/api-keys'),
                    ),
                  ],
                ),

                const SizedBox(height: 24),

                // Account Section
                _buildSectionHeader(context, 'Account', Icons.person_outline),
                const SizedBox(height: 12),
                _buildCard(
                  context,
                  children: [
                    _buildNavigationTile(
                      context,
                      icon: Icons.account_circle_outlined,
                      iconColor: Colors.blue,
                      title: 'Profile',
                      subtitle: 'View and edit your profile',
                      onTap: () => context.push('/profile'),
                    ),
                    _buildDivider(),
                    _buildLogoutTile(context),
                  ],
                ),

                const SizedBox(height: 32),

                // Version info
                Center(
                  child: Text(
                    'LedgerGuard v1.0.0',
                    style: Theme.of(context).textTheme.labelMedium?.copyWith(
                          color: Colors.grey[400],
                        ),
                  ),
                ),

                const SizedBox(height: 16),
              ],
            ),
          );
        },
      ),
    );
  }

  Widget _buildSectionHeader(BuildContext context, String title, IconData icon) {
    return Row(
      children: [
        Icon(icon, size: 20, color: Colors.grey[600]),
        const SizedBox(width: 8),
        Text(
          title,
          style: Theme.of(context).textTheme.titleSmall?.copyWith(
                color: Colors.grey[700],
                letterSpacing: 0.5,
              ),
        ),
      ],
    );
  }

  Widget _buildCard(BuildContext context, {required List<Widget> children}) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(16),
        child: Column(children: children),
      ),
    );
  }

  Widget _buildNavigationTile(
    BuildContext context, {
    required IconData icon,
    required Color iconColor,
    required String title,
    required String subtitle,
    required VoidCallback onTap,
  }) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          child: Row(
            children: [
              Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  color: iconColor.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Icon(icon, color: iconColor, size: 20),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      title,
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            fontWeight: FontWeight.w500,
                          ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      subtitle,
                      style: Theme.of(context).textTheme.labelMedium?.copyWith(
                            color: Colors.grey[500],
                          ),
                    ),
                  ],
                ),
              ),
              Icon(Icons.chevron_right, color: Colors.grey[400], size: 20),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildLogoutTile(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: () => _showLogoutConfirmation(context),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          child: Row(
            children: [
              Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  color: AppTheme.danger.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Icon(Icons.logout, color: AppTheme.danger, size: 20),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Log Out',
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            fontWeight: FontWeight.w500,
                            color: AppTheme.danger,
                          ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      'Sign out of your account',
                      style: Theme.of(context).textTheme.labelMedium?.copyWith(
                            color: Colors.grey[500],
                          ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildDivider() {
    return Divider(
      height: 1,
      thickness: 1,
      color: Colors.grey[100],
      indent: 62,
    );
  }

  void _showLogoutConfirmation(BuildContext context) {
    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        title: const Text('Log Out'),
        content: const Text('Are you sure you want to log out?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(dialogContext).pop(),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.of(dialogContext).pop();
              context.read<AuthBloc>().add(const SignOutRequested());
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: AppTheme.danger,
              foregroundColor: Colors.white,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
            ),
            child: const Text('Log Out'),
          ),
        ],
      ),
    );
  }
}
