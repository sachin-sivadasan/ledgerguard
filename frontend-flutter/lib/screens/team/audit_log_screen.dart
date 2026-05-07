import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../providers/organization_provider.dart';
import '../../widgets/lg_page.dart';

class AuditLogScreen extends StatefulWidget {
  const AuditLogScreen({super.key});

  @override
  State<AuditLogScreen> createState() => _AuditLogScreenState();
}

class _AuditLogScreenState extends State<AuditLogScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<OrganizationProvider>().loadAuditLog();
    });
  }

  @override
  Widget build(BuildContext context) {
    final orgProvider = context.watch<OrganizationProvider>();
    final entries = orgProvider.auditEntries;
    final theme = Theme.of(context);

    return LgPage(
      title: 'Audit Log',
      subtitle: orgProvider.currentOrg?.name,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (orgProvider.isAuditLoading)
            const Center(child: CircularProgressIndicator())
          else if (entries.isEmpty)
            Center(
              child: Padding(
                padding: const EdgeInsets.all(48),
                child: Column(
                  children: [
                    Icon(Icons.history,
                        size: 48, color: theme.colorScheme.outline),
                    const SizedBox(height: 16),
                    Text('No activity yet',
                        style: theme.textTheme.titleMedium),
                  ],
                ),
              ),
            )
          else
            ...entries.map((entry) => Card(
                  margin: const EdgeInsets.only(bottom: 8),
                  child: ListTile(
                    leading: CircleAvatar(
                      radius: 18,
                      backgroundColor:
                          _actionColor(entry.action).withValues(alpha:0.1),
                      child: Icon(_actionIcon(entry.action),
                          size: 18, color: _actionColor(entry.action)),
                    ),
                    title: Text(entry.actionLabel,
                        style: theme.textTheme.bodyMedium
                            ?.copyWith(fontWeight: FontWeight.w500)),
                    subtitle: Text(
                      _formatTime(entry.createdAt),
                      style: theme.textTheme.bodySmall
                          ?.copyWith(color: theme.colorScheme.outline),
                    ),
                    trailing: entry.metadata != null
                        ? Tooltip(
                            message: entry.metadata.toString(),
                            child: Icon(Icons.info_outline,
                                size: 18,
                                color: theme.colorScheme.outline),
                          )
                        : null,
                  ),
                )),
        ],
      ),
    );
  }

  Color _actionColor(String action) {
    if (action.contains('created') || action.contains('joined')) {
      return Colors.green;
    }
    if (action.contains('removed') || action.contains('deleted')) {
      return Colors.red;
    }
    if (action.contains('suspended')) return Colors.orange;
    return Colors.blue;
  }

  IconData _actionIcon(String action) {
    if (action.contains('invited')) return Icons.mail_outline;
    if (action.contains('joined')) return Icons.person_add;
    if (action.contains('removed') || action.contains('deleted')) {
      return Icons.person_remove;
    }
    if (action.contains('suspended')) return Icons.pause_circle_outline;
    if (action.contains('unsuspended')) return Icons.play_circle_outline;
    if (action.contains('role')) return Icons.swap_horiz;
    if (action.contains('webhook')) return Icons.webhook;
    if (action.contains('org')) return Icons.business;
    return Icons.history;
  }

  String _formatTime(DateTime dt) {
    final now = DateTime.now();
    final diff = now.difference(dt);
    if (diff.inMinutes < 1) return 'Just now';
    if (diff.inHours < 1) return '${diff.inMinutes}m ago';
    if (diff.inDays < 1) return '${diff.inHours}h ago';
    if (diff.inDays < 7) return '${diff.inDays}d ago';
    return '${dt.month}/${dt.day}/${dt.year}';
  }
}
