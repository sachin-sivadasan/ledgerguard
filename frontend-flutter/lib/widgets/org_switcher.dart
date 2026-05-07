import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/organization_provider.dart';

/// Org switcher widget for the app shell. Shows a dropdown when user
/// belongs to 2+ organizations. Hidden when user has exactly 1 org.
class OrgSwitcher extends StatelessWidget {
  const OrgSwitcher({super.key});

  @override
  Widget build(BuildContext context) {
    final orgProvider = context.watch<OrganizationProvider>();

    // Don't show if single org or no orgs
    if (!orgProvider.hasMultipleOrgs) return const SizedBox.shrink();

    final currentOrg = orgProvider.currentOrg;
    final theme = Theme.of(context);

    return PopupMenuButton<String>(
      tooltip: 'Switch organization',
      offset: const Offset(0, 40),
      onSelected: (orgId) {
        orgProvider.selectOrganization(orgId);
      },
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
        decoration: BoxDecoration(
          border: Border.all(color: theme.colorScheme.outline.withValues(alpha: 0.3)),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.business,
                size: 16, color: theme.colorScheme.onSurfaceVariant),
            const SizedBox(width: 8),
            Text(
              currentOrg?.name ?? 'Select org',
              style: theme.textTheme.bodySmall
                  ?.copyWith(fontWeight: FontWeight.w600),
            ),
            const SizedBox(width: 4),
            Icon(Icons.arrow_drop_down,
                size: 18, color: theme.colorScheme.onSurfaceVariant),
          ],
        ),
      ),
      itemBuilder: (_) => orgProvider.memberships
          .map((m) => PopupMenuItem<String>(
                value: m.orgId,
                child: Row(
                  children: [
                    Icon(
                      m.orgId == currentOrg?.id
                          ? Icons.check_circle
                          : Icons.circle_outlined,
                      size: 16,
                      color: m.orgId == currentOrg?.id
                          ? theme.colorScheme.primary
                          : theme.colorScheme.outline,
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Text(m.orgId,
                          overflow: TextOverflow.ellipsis),
                    ),
                    const SizedBox(width: 8),
                    Text(m.role,
                        style: theme.textTheme.labelSmall
                            ?.copyWith(color: theme.colorScheme.outline)),
                  ],
                ),
              ))
          .toList(),
    );
  }
}
