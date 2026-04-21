import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../models/api_key_model.dart';
import '../../providers/api_key_provider.dart';
import '../../providers/apps_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_badge.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_confirmation_dialog.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_page.dart';

class ApiKeysScreen extends StatelessWidget {
  const ApiKeysScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final hasApps = context.watch<AppsProvider>().apps.isNotEmpty;
    if (!hasApps) {
      return LgPage(
        title: 'API Keys',
        child: LgEmptyState(
          icon: Icons.key,
          heading: 'No API keys yet',
          description:
              'Connect your Shopify app to manage API keys.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<ApiKeyProvider>();
    final keys = provider.keys;
    final dateFmt = DateFormat('MMM d, y');
    final theme = Theme.of(context);

    return LgPage(
      title: 'API Keys',
      subtitle: '${provider.activeKeys.length} active keys',
      primaryAction: LgPageAction(
        label: 'Create API Key',
        onPressed: () => _showCreateDialog(context, provider),
        isPrimary: true,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          ...keys.map((key) {
            final isActive = key.status == ApiKeyStatus.active;
            return Padding(
              padding: const EdgeInsets.only(bottom: LgSpacing.s300),
              child: LgCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Expanded(
                          child: Text(key.name,
                              style: theme.textTheme.titleSmall),
                        ),
                        LgBadge(
                          label: isActive ? 'ACTIVE' : 'REVOKED',
                          tone: isActive
                              ? BadgeTone.success
                              : BadgeTone.critical,
                        ),
                      ],
                    ),
                    const SizedBox(height: LgSpacing.s200),
                    Text(
                      key.keyPrefix,
                      style: TextStyle(
                        fontSize: 13,
                        fontFamily: 'monospace',
                        color: LgColors.textSecondary,
                      ),
                    ),
                    const SizedBox(height: LgSpacing.s200),
                    Row(
                      children: [
                        Text(
                          'Created ${dateFmt.format(key.createdAt)}',
                          style: TextStyle(
                              fontSize: 12, color: LgColors.textSecondary),
                        ),
                        const SizedBox(width: LgSpacing.s400),
                        Text(
                          key.lastUsedAt != null
                              ? 'Last used ${dateFmt.format(key.lastUsedAt!)}'
                              : 'Never used',
                          style: TextStyle(
                              fontSize: 12, color: LgColors.textSecondary),
                        ),
                      ],
                    ),
                    const SizedBox(height: LgSpacing.s300),
                    Wrap(
                      spacing: LgSpacing.s200,
                      runSpacing: LgSpacing.s200,
                      children: [
                        ...key.permissions.map((p) => LgBadge(
                              label: p.toUpperCase(),
                              tone: _permissionTone(p),
                            )),
                        if (isActive) ...[
                          const SizedBox(width: LgSpacing.s300),
                          TextButton.icon(
                            onPressed: () => _confirmRevoke(
                                context, provider, key),
                            icon: const Icon(Icons.block, size: 16),
                            label: const Text('Revoke'),
                            style: TextButton.styleFrom(
                                foregroundColor: LgColors.critical),
                          ),
                        ],
                      ],
                    ),
                  ],
                ),
              ),
            );
          }),
        ],
      ),
    );
  }

  void _confirmRevoke(
      BuildContext context, ApiKeyProvider provider, ApiKey key) {
    showDialog(
      context: context,
      builder: (ctx) => LgConfirmationDialog(
        title: 'Revoke API Key',
        message:
            'Are you sure you want to revoke "${key.name}"? This action cannot be undone.',
        confirmLabel: 'Revoke',
        destructive: true,
        onConfirm: () => provider.revokeKey(key.id),
      ),
    );
  }

  void _showCreateDialog(BuildContext context, ApiKeyProvider provider) {
    final nameController = TextEditingController();
    final permissions = <String>{'read'};

    showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) => AlertDialog(
          title: const Text('Create API Key'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              TextField(
                controller: nameController,
                decoration: const InputDecoration(
                  labelText: 'Key Name',
                  hintText: 'e.g. Production Key',
                ),
              ),
              const SizedBox(height: LgSpacing.s400),
              const Text('Permissions'),
              const SizedBox(height: LgSpacing.s200),
              ..._allPermissions.map((p) => CheckboxListTile(
                    title: Text(p.toUpperCase()),
                    value: permissions.contains(p),
                    dense: true,
                    onChanged: (v) {
                      setDialogState(() {
                        if (v == true) {
                          permissions.add(p);
                        } else {
                          permissions.remove(p);
                        }
                      });
                    },
                  )),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(ctx).pop(),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () {
                if (nameController.text.trim().isNotEmpty &&
                    permissions.isNotEmpty) {
                  provider.createKey(
                      nameController.text.trim(), permissions.toList());
                  Navigator.of(ctx).pop();
                }
              },
              child: const Text('Create'),
            ),
          ],
        ),
      ),
    );
  }
}

const _allPermissions = ['read', 'write', 'sync'];

BadgeTone _permissionTone(String permission) => switch (permission) {
      'read' => BadgeTone.info,
      'write' => BadgeTone.warning,
      'sync' => BadgeTone.attention,
      _ => BadgeTone.defaultTone,
    };
