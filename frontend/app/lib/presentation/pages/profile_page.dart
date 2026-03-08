import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../core/theme/app_theme.dart';
import '../../domain/entities/user_profile.dart';
import '../blocs/auth/auth.dart';
import '../blocs/role/role.dart';
import '../widgets/shared.dart';

/// Profile page displaying user information and settings with premium design
class ProfilePage extends StatelessWidget {
  const ProfilePage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey[50],
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
                  displayName: authState.user.displayName,
                  profile: roleState.profile,
                );
              }

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
    String? displayName,
    required UserProfile profile,
  }) {
    return CustomScrollView(
      slivers: [
        // Pinned app bar with gradient — stays visible on scroll
        SliverAppBar(
          pinned: true,
          backgroundColor: AppTheme.primary,
          foregroundColor: Colors.white,
          title: const Text('Profile'),
          actions: [
            IconButton(
              icon: const Icon(Icons.edit_outlined, color: Colors.white70),
              onPressed: () => _showComingSoon(context, 'Edit Profile'),
            ),
          ],
          expandedHeight: 300,
          flexibleSpace: FlexibleSpaceBar(
            background: _buildProfileBanner(context, email, displayName, profile),
          ),
        ),

        // Main content
        SliverPadding(
          padding: const EdgeInsets.all(20),
          sliver: SliverList(
            delegate: SliverChildListDelegate([
              // Quick Stats Row
              _buildQuickStats(context, profile),

              const SizedBox(height: 24),

              // Account Section
              _buildSectionHeader(context, 'Account', Icons.person_outline),
              const SizedBox(height: 12),
              _buildPremiumCard(
                context,
                children: [
                  _buildPremiumTile(
                    context,
                    icon: Icons.email_outlined,
                    iconColor: Colors.blue,
                    title: 'Email Address',
                    subtitle: email,
                  ),
                  _buildDivider(),
                  _buildPremiumTile(
                    context,
                    icon: Icons.security_outlined,
                    iconColor: Colors.green,
                    title: 'Account Security',
                    subtitle: 'Password protected',
                    trailing: const Icon(Icons.check_circle, color: Colors.green, size: 20),
                  ),
                ],
              ),

              const SizedBox(height: 24),

              // Subscription Section
              _buildSectionHeader(context, 'Subscription', Icons.workspace_premium_outlined),
              const SizedBox(height: 12),
              _buildSubscriptionCard(context, profile),

              const SizedBox(height: 24),

              // Support Section
              _buildSectionHeader(context, 'Support', Icons.help_outline),
              const SizedBox(height: 12),
              _buildPremiumCard(
                context,
                children: [
                  _buildNavigationTile(
                    context,
                    icon: Icons.menu_book_outlined,
                    iconColor: Colors.indigo,
                    title: 'Documentation',
                    subtitle: 'API docs and guides',
                    onTap: () => _showComingSoon(context, 'Documentation'),
                  ),
                  _buildDivider(),
                  _buildNavigationTile(
                    context,
                    icon: Icons.chat_bubble_outline,
                    iconColor: Colors.teal,
                    title: 'Contact Support',
                    subtitle: 'Get help from our team',
                    onTap: () => _showComingSoon(context, 'Support'),
                  ),
                ],
              ),

              const SizedBox(height: 32),

              // Logout Button
              _buildLogoutButton(context),

              const SizedBox(height: 24),

              // Version info
              Center(
                child: Text(
                  'LedgerGuard v1.0.0',
                  style: TextStyle(
                    color: Colors.grey[400],
                    fontSize: 12,
                  ),
                ),
              ),

              const SizedBox(height: 32),
            ]),
          ),
        ),
      ],
    );
  }

  Widget _buildProfileBanner(
    BuildContext context,
    String email,
    String? displayName,
    UserProfile profile,
  ) {
    return ClipRect(
      child: Container(
        decoration: BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [
              AppTheme.primary,
              AppTheme.primary.withBlue(180),
            ],
          ),
        ),
        child: LayoutBuilder(
          builder: (context, constraints) {
            // SliverAppBar toolbar height (~kToolbarHeight + status bar)
            final topPadding = MediaQuery.of(context).padding.top + kToolbarHeight;
            final availableHeight = constraints.maxHeight - topPadding;

            return Padding(
              padding: EdgeInsets.only(top: topPadding),
              child: SizedBox(
                height: availableHeight,
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(24, 8, 24, 16),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      // Avatar with border
                      Container(
                        padding: const EdgeInsets.all(3),
                        decoration: BoxDecoration(
                          shape: BoxShape.circle,
                          border: Border.all(color: Colors.white30, width: 2),
                        ),
                        child: CircleAvatar(
                          radius: 32,
                          backgroundColor: Colors.white.withOpacity(0.2),
                          child: Text(
                            _getInitials(displayName ?? email),
                            style: const TextStyle(
                              fontSize: 24,
                              fontWeight: FontWeight.bold,
                              color: Colors.white,
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(height: 10),

                      // Name
                      Text(
                        displayName ?? email.split('@').first,
                        style: const TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                          color: Colors.white,
                        ),
                        overflow: TextOverflow.ellipsis,
                      ),
                      const SizedBox(height: 2),

                      // Email
                      Text(
                        email,
                        style: TextStyle(
                          fontSize: 13,
                          color: Colors.white.withOpacity(0.8),
                        ),
                        overflow: TextOverflow.ellipsis,
                      ),
                      const SizedBox(height: 10),

                      // Badges Row
                      Wrap(
                        alignment: WrapAlignment.center,
                        spacing: 8,
                        runSpacing: 6,
                        children: [
                          _buildHeaderBadge(
                            context,
                            icon: Icons.verified_user,
                            label: profile.role.displayName,
                            color: Colors.white.withOpacity(0.2),
                          ),
                          _buildHeaderBadge(
                            context,
                            icon: profile.planTier.isPro ? Icons.star : Icons.star_border,
                            label: profile.planTier.displayName,
                            color: profile.planTier.isPro
                                ? Colors.amber.withOpacity(0.3)
                                : Colors.white.withOpacity(0.2),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
            );
          },
        ),
      ),
    );
  }

  Widget _buildHeaderBadge(
    BuildContext context, {
    required IconData icon,
    required String label,
    required Color color,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(
        color: color,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 16, color: Colors.white),
          const SizedBox(width: 6),
          Text(
            label,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 12,
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildQuickStats(BuildContext context, UserProfile profile) {
    return Row(
      children: [
        Expanded(
          child: _buildStatCard(
            context,
            icon: Icons.apps,
            value: '1',
            label: 'Connected Apps',
            color: Colors.blue,
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: _buildStatCard(
            context,
            icon: Icons.key,
            value: '0',
            label: 'API Keys',
            color: Colors.purple,
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: _buildStatCard(
            context,
            icon: Icons.calendar_today,
            value: _getMemberSince(),
            label: 'Member Since',
            color: Colors.green,
          ),
        ),
      ],
    );
  }

  Widget _buildStatCard(
    BuildContext context, {
    required IconData icon,
    required String value,
    required String label,
    required Color color,
  }) {
    return Container(
      padding: const EdgeInsets.all(16),
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
      child: Column(
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: color.withOpacity(0.1),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Icon(icon, color: color, size: 20),
          ),
          const SizedBox(height: 12),
          Text(
            value,
            style: Theme.of(context).textTheme.titleLarge?.copyWith(
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 2),
          Text(
            label,
            style: TextStyle(
              color: Colors.grey[600],
              fontSize: 11,
            ),
            textAlign: TextAlign.center,
          ),
        ],
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
          style: TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.w600,
            color: Colors.grey[700],
            letterSpacing: 0.5,
          ),
        ),
      ],
    );
  }

  Widget _buildPremiumCard(BuildContext context, {required List<Widget> children}) {
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

  Widget _buildSubscriptionCard(BuildContext context, UserProfile profile) {
    final isPro = profile.planTier.isPro;

    return Container(
      decoration: BoxDecoration(
        gradient: isPro
            ? LinearGradient(
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
                colors: [
                  Colors.amber[700]!,
                  Colors.orange[600]!,
                ],
              )
            : null,
        color: isPro ? null : Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: isPro
                ? Colors.amber.withOpacity(0.3)
                : Colors.black.withOpacity(0.04),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: isPro ? Colors.white.withOpacity(0.2) : Colors.grey[100],
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Icon(
                    isPro ? Icons.workspace_premium : Icons.rocket_launch,
                    color: isPro ? Colors.white : AppTheme.primary,
                    size: 24,
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        isPro ? 'Pro Plan' : 'Free Plan',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                          color: isPro ? Colors.white : Colors.grey[800],
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        isPro ? 'Full access to all features' : 'Basic features included',
                        style: TextStyle(
                          fontSize: 13,
                          color: isPro ? Colors.white70 : Colors.grey[600],
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            if (!isPro) ...[
              const SizedBox(height: 20),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  _buildFeatureChip('AI Insights', false),
                  _buildFeatureChip('Priority Support', false),
                  _buildFeatureChip('Advanced Analytics', false),
                ],
              ),
              const SizedBox(height: 16),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: () => _showComingSoon(context, 'Upgrade'),
                  style: ElevatedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  child: const Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.bolt, size: 18),
                      SizedBox(width: 8),
                      Text('Upgrade to Pro'),
                    ],
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildFeatureChip(String label, bool included) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: Colors.grey[100],
        borderRadius: BorderRadius.circular(12),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            included ? Icons.check : Icons.lock_outline,
            size: 12,
            color: Colors.grey[500],
          ),
          const SizedBox(width: 4),
          Text(
            label,
            style: TextStyle(
              fontSize: 11,
              color: Colors.grey[600],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPremiumTile(
    BuildContext context, {
    required IconData icon,
    required Color iconColor,
    required String title,
    required String subtitle,
    Widget? trailing,
  }) {
    return Padding(
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
                  style: TextStyle(
                    fontSize: 12,
                    color: Colors.grey[500],
                  ),
                ),
                const SizedBox(height: 2),
                Text(
                  subtitle,
                  style: const TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ],
            ),
          ),
          if (trailing != null) trailing,
        ],
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
                      style: const TextStyle(
                        fontSize: 15,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      subtitle,
                      style: TextStyle(
                        fontSize: 12,
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

  Widget _buildDivider() {
    return Divider(
      height: 1,
      thickness: 1,
      color: Colors.grey[100],
      indent: 62,
    );
  }

  Widget _buildLogoutButton(BuildContext context) {
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
      child: Material(
        color: Colors.transparent,
        borderRadius: BorderRadius.circular(16),
        child: InkWell(
          onTap: () => _showLogoutConfirmation(context),
          borderRadius: BorderRadius.circular(16),
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: 16),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.logout, color: AppTheme.danger, size: 20),
                const SizedBox(width: 8),
                Text(
                  'Log Out',
                  style: TextStyle(
                    color: AppTheme.danger,
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  void _showComingSoon(BuildContext context, String feature) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(Icons.info_outline, color: Colors.white),
            const SizedBox(width: 8),
            Expanded(
              child: Text('$feature coming soon!'),
            ),
          ],
        ),
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
      ),
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

  String _getInitials(String name) {
    if (name.isEmpty) return '?';
    final parts = name.contains('@') ? name.split('@')[0].split('.') : name.split(' ');
    if (parts.isEmpty) return '?';
    if (parts.length == 1) {
      return parts[0].isNotEmpty ? parts[0].substring(0, 1).toUpperCase() : '?';
    }
    return '${parts[0].isNotEmpty ? parts[0][0] : ''}${parts.length > 1 && parts[1].isNotEmpty ? parts[1][0] : ''}'.toUpperCase();
  }

  String _getMemberSince() {
    // For now, return a placeholder. In real app, get from user data
    return '2024';
  }
}
