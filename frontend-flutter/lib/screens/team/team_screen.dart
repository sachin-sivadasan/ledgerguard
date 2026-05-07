import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../models/organization_model.dart';
import '../../providers/organization_provider.dart';
import '../../widgets/lg_page.dart';

class TeamScreen extends StatefulWidget {
  const TeamScreen({super.key});

  @override
  State<TeamScreen> createState() => _TeamScreenState();
}

class _TeamScreenState extends State<TeamScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<OrganizationProvider>().loadMembers();
    });
  }

  @override
  Widget build(BuildContext context) {
    final orgProvider = context.watch<OrganizationProvider>();
    final org = orgProvider.currentOrg;
    final members = orgProvider.members;
    final isAdmin = orgProvider.isAdmin;
    final theme = Theme.of(context);

    return LgPage(
      title: 'Team',
      subtitle: org != null
          ? '${members.length}/${org.maxMembers} members · ${org.planTier} plan'
          : null,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Invite button (admin only)
          if (isAdmin)
            Padding(
              padding: const EdgeInsets.only(bottom: 24),
              child: FilledButton.icon(
                onPressed: () => _showInviteDialog(context),
                icon: const Icon(Icons.person_add, size: 18),
                label: const Text('Invite Member'),
              ),
            ),

          if (orgProvider.isMembersLoading)
            const Center(child: CircularProgressIndicator())
          else if (members.isEmpty)
            Center(
              child: Padding(
                padding: const EdgeInsets.all(48),
                child: Column(
                  children: [
                    Icon(Icons.group_outlined,
                        size: 48, color: theme.colorScheme.outline),
                    const SizedBox(height: 16),
                    Text('No team members yet',
                        style: theme.textTheme.titleMedium),
                  ],
                ),
              ),
            )
          else
            ...members.map((member) => _MemberTile(
                  member: member,
                  isAdmin: isAdmin,
                  isOwner: orgProvider.isOwner,
                  onSuspend: () => orgProvider.suspendMember(member.id),
                  onUnsuspend: () => orgProvider.unsuspendMember(member.id),
                  onRemove: () => _confirmRemove(context, member),
                  onChangeRole: (role) =>
                      orgProvider.changeRole(member.id, role),
                )),
        ],
      ),
    );
  }

  void _showInviteDialog(BuildContext context) {
    final emailController = TextEditingController();
    String selectedRole = 'VIEWER';

    showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) => AlertDialog(
          title: const Text('Invite Team Member'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(
                controller: emailController,
                decoration: const InputDecoration(
                  labelText: 'Email address',
                  hintText: 'teammate@example.com',
                ),
                keyboardType: TextInputType.emailAddress,
              ),
              const SizedBox(height: 16),
              DropdownButtonFormField<String>(
                initialValue: selectedRole,
                decoration: const InputDecoration(labelText: 'Role'),
                items: const [
                  DropdownMenuItem(value: 'VIEWER', child: Text('Viewer')),
                  DropdownMenuItem(value: 'ADMIN', child: Text('Admin')),
                ],
                onChanged: (v) {
                  if (v != null) setDialogState(() => selectedRole = v);
                },
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () async {
                final email = emailController.text.trim();
                if (email.isEmpty) return;
                final orgProvider = context.read<OrganizationProvider>();
                final messenger = ScaffoldMessenger.of(context);
                Navigator.pop(ctx);
                final inv = await orgProvider.inviteMember(email, selectedRole);
                if (inv != null && mounted) {
                  messenger.showSnackBar(
                    SnackBar(content: Text('Invitation sent to $email')),
                  );
                }
              },
              child: const Text('Send Invite'),
            ),
          ],
        ),
      ),
    );
  }

  void _confirmRemove(BuildContext context, OrgMember member) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Remove Member'),
        content: Text('Remove ${member.userId} from the organization?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Cancel'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(
                backgroundColor: Theme.of(context).colorScheme.error),
            onPressed: () {
              Navigator.pop(ctx);
              context.read<OrganizationProvider>().removeMember(member.id);
            },
            child: const Text('Remove'),
          ),
        ],
      ),
    );
  }
}

class _MemberTile extends StatelessWidget {
  final OrgMember member;
  final bool isAdmin;
  final bool isOwner;
  final VoidCallback onSuspend;
  final VoidCallback onUnsuspend;
  final VoidCallback onRemove;
  final ValueChanged<String> onChangeRole;

  const _MemberTile({
    required this.member,
    required this.isAdmin,
    required this.isOwner,
    required this.onSuspend,
    required this.onUnsuspend,
    required this.onRemove,
    required this.onChangeRole,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: _roleColor(member.role).withValues(alpha:0.15),
          child: Icon(_roleIcon(member.role), color: _roleColor(member.role)),
        ),
        title: Text(member.userId,
            style: theme.textTheme.bodyMedium
                ?.copyWith(fontWeight: FontWeight.w600)),
        subtitle: Row(
          children: [
            _RoleBadge(role: member.role),
            const SizedBox(width: 8),
            if (member.isSuspended)
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(
                  color: theme.colorScheme.error.withValues(alpha:0.1),
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text('Suspended',
                    style: theme.textTheme.labelSmall
                        ?.copyWith(color: theme.colorScheme.error)),
              ),
          ],
        ),
        trailing: _buildActions(context),
      ),
    );
  }

  Widget? _buildActions(BuildContext context) {
    // Don't show actions on owner or if not admin
    if (member.isOwner || !isAdmin) return null;

    return PopupMenuButton<String>(
      onSelected: (action) {
        switch (action) {
          case 'suspend':
            onSuspend();
          case 'unsuspend':
            onUnsuspend();
          case 'remove':
            onRemove();
          case 'make_admin':
            onChangeRole('ADMIN');
          case 'make_viewer':
            onChangeRole('VIEWER');
        }
      },
      itemBuilder: (_) => [
        if (member.isActive && !member.isAdmin)
          const PopupMenuItem(value: 'make_admin', child: Text('Make Admin')),
        if (member.isActive && member.isAdmin && isOwner)
          const PopupMenuItem(
              value: 'make_viewer', child: Text('Make Viewer')),
        if (member.isActive)
          const PopupMenuItem(value: 'suspend', child: Text('Suspend')),
        if (member.isSuspended)
          const PopupMenuItem(value: 'unsuspend', child: Text('Unsuspend')),
        const PopupMenuItem(
          value: 'remove',
          child: Text('Remove', style: TextStyle(color: Colors.red)),
        ),
      ],
    );
  }

  Color _roleColor(String role) {
    switch (role) {
      case 'OWNER':
        return Colors.amber.shade700;
      case 'ADMIN':
        return Colors.blue;
      case 'VIEWER':
        return Colors.grey;
      default:
        return Colors.grey;
    }
  }

  IconData _roleIcon(String role) {
    switch (role) {
      case 'OWNER':
        return Icons.star;
      case 'ADMIN':
        return Icons.admin_panel_settings;
      case 'VIEWER':
        return Icons.visibility;
      default:
        return Icons.person;
    }
  }
}

class _RoleBadge extends StatelessWidget {
  final String role;
  const _RoleBadge({required this.role});

  @override
  Widget build(BuildContext context) {
    final color = switch (role) {
      'OWNER' => Colors.amber.shade700,
      'ADMIN' => Colors.blue,
      _ => Colors.grey,
    };

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      decoration: BoxDecoration(
        color: color.withValues(alpha:0.1),
        borderRadius: BorderRadius.circular(4),
        border: Border.all(color: color.withValues(alpha:0.3)),
      ),
      child: Text(
        switch (role) {
          'OWNER' => 'Owner',
          'ADMIN' => 'Admin',
          'VIEWER' => 'Viewer',
          _ => role,
        },
        style: Theme.of(context)
            .textTheme
            .labelSmall
            ?.copyWith(color: color, fontWeight: FontWeight.w600),
      ),
    );
  }
}
