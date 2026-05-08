import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../providers/apps_provider.dart';
import '../../providers/sync_status_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_badge.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_page.dart';
import 'reviews_tab.dart';

class AppsScreen extends StatefulWidget {
  const AppsScreen({super.key});

  @override
  State<AppsScreen> createState() => _AppsScreenState();
}

class _AppsScreenState extends State<AppsScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _maybeLoadData());
  }

  void _maybeLoadData() {
    final apps = context.read<AppsProvider>();
    if (!apps.demoMode && !apps.isLoading) {
      apps.loadApps();
    }
    // Start watching sync status for all apps
    if (apps.apps.isNotEmpty && !apps.demoMode) {
      final appIds = apps.apps.map((a) => a.id).toList();
      context.read<SyncStatusProvider>().startWatching(appIds);
    }
  }

  @override
  Widget build(BuildContext context) {
    final provider = context.watch<AppsProvider>();

    if (provider.apps.isEmpty) {
      return LgPage(
        title: 'Apps',
        subtitle: '0 connected apps',
        child: LgEmptyState(
          icon: Icons.apps_outlined,
          heading: 'Connect Your First App',
          description:
              'Link your Shopify Partner account to start tracking app revenue.',
          actionLabel: 'Connect Partner Account',
          onAction: () => context.go('/settings/connect-shopify'),
        ),
      );
    }

    // Ensure sync watching is active when apps list changes
    if (!provider.demoMode) {
      final appIds = provider.apps.map((a) => a.id).toList();
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) {
          context.read<SyncStatusProvider>().startWatching(appIds);
        }
      });
    }

    return LgPage(
      title: 'Apps',
      subtitle: '${provider.apps.length} connected apps',
      scrollable: false,
      child: DefaultTabController(
        length: 2,
        child: Column(
          children: [
            const TabBar(
              isScrollable: true,
              tabAlignment: TabAlignment.start,
              padding: EdgeInsets.zero,
              tabs: [
                Tab(text: 'Connected Apps'),
                Tab(text: 'Reviews'),
              ],
            ),
            const SizedBox(height: LgSpacing.s400),
            Expanded(
              child: TabBarView(
                children: [
                  // Apps list with sync status
                  ListView(
                    children: provider.apps.map((app) {
                      return _AppCardWithSync(app: app);
                    }).toList(),
                  ),
                  // Reviews tab
                  const ReviewsTab(),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _AppCardWithSync extends StatelessWidget {
  final dynamic app;
  const _AppCardWithSync({required this.app});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final syncProvider = context.watch<SyncStatusProvider>();
    final syncState = syncProvider.getState(app.id);

    return Padding(
      padding: const EdgeInsets.only(bottom: LgSpacing.s300),
      child: LgCard(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Main row: icon + info + status
            Row(
              children: [
                Container(
                  width: 48,
                  height: 48,
                  decoration: BoxDecoration(
                    color: LgColors.primary.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Center(
                    child: Text(
                      app.name.substring(0, 2),
                      style: const TextStyle(
                          fontWeight: FontWeight.w600,
                          color: LgColors.primary),
                    ),
                  ),
                ),
                const SizedBox(width: LgSpacing.s400),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(app.name, style: theme.textTheme.titleSmall),
                      const SizedBox(height: LgSpacing.s100),
                      Row(
                        children: [
                          const Icon(Icons.people,
                              size: 14, color: LgColors.textSecondary),
                          const SizedBox(width: 4),
                          Text('${app.installCount} installs',
                              style: theme.textTheme.bodySmall),
                          const SizedBox(width: LgSpacing.s400),
                          const Icon(Icons.star_rounded,
                              size: 14, color: LgColors.starRating),
                          const SizedBox(width: 4),
                          Text('${app.avgRating}',
                              style: theme.textTheme.bodySmall),
                        ],
                      ),
                    ],
                  ),
                ),
                if (syncState.isSyncing)
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const SizedBox(
                        width: 14,
                        height: 14,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      ),
                      const SizedBox(width: 6),
                      Text('Syncing',
                          style: TextStyle(
                              fontSize: 12, color: LgColors.textSecondary)),
                    ],
                  )
                else
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      LgBadge(
                        label: app.syncStatus == 'synced' ? 'Synced' : 'Pending',
                        tone: app.syncStatus == 'synced'
                            ? BadgeTone.success
                            : BadgeTone.warning,
                      ),
                      if (app.lastSyncAt != null) ...[
                        const SizedBox(height: LgSpacing.s100),
                        Text(
                          _formatSyncTime(app.lastSyncAt!),
                          style: theme.textTheme.bodySmall,
                        ),
                      ],
                    ],
                  ),
              ],
            ),

            // Sync progress or action row
            if (syncState.isSyncing) ...[
              const SizedBox(height: LgSpacing.s300),
              // Progress bar
              ClipRRect(
                borderRadius: BorderRadius.circular(4),
                child: LinearProgressIndicator(
                  value: syncState.progress,
                  minHeight: 6,
                  backgroundColor: LgColors.surfaceSecondary,
                  valueColor:
                      const AlwaysStoppedAnimation<Color>(LgColors.primary),
                ),
              ),
              const SizedBox(height: LgSpacing.s200),
              Row(
                children: [
                  Expanded(
                    child: Text(
                      syncState.message ?? 'Syncing...',
                      style: TextStyle(
                          fontSize: 12, color: LgColors.textSecondary),
                    ),
                  ),
                  if (syncState.progress != null)
                    Text(
                      '${(syncState.progress! * 100).toInt()}%',
                      style: TextStyle(
                          fontSize: 12,
                          fontWeight: FontWeight.w600,
                          color: LgColors.textSecondary),
                    ),
                  const SizedBox(width: LgSpacing.s300),
                  SizedBox(
                    height: 28,
                    child: OutlinedButton(
                      onPressed: () =>
                          context.read<SyncStatusProvider>().cancelSync(app.id),
                      style: OutlinedButton.styleFrom(
                        padding:
                            const EdgeInsets.symmetric(horizontal: 12),
                        textStyle: const TextStyle(fontSize: 12),
                      ),
                      child: const Text('Cancel'),
                    ),
                  ),
                ],
              ),
            ] else if (syncState.error != null) ...[
              const SizedBox(height: LgSpacing.s200),
              Row(
                children: [
                  const Icon(Icons.error_outline,
                      size: 14, color: LgColors.critical),
                  const SizedBox(width: 4),
                  Expanded(
                    child: Text(
                      'Sync failed',
                      style:
                          TextStyle(fontSize: 12, color: LgColors.critical),
                    ),
                  ),
                  SizedBox(
                    height: 28,
                    child: OutlinedButton(
                      onPressed: () => context
                          .read<SyncStatusProvider>()
                          .triggerSync(app.id),
                      style: OutlinedButton.styleFrom(
                        padding:
                            const EdgeInsets.symmetric(horizontal: 12),
                        textStyle: const TextStyle(fontSize: 12),
                      ),
                      child: const Text('Retry'),
                    ),
                  ),
                ],
              ),
            ] else ...[
              const SizedBox(height: LgSpacing.s200),
              Align(
                alignment: Alignment.centerLeft,
                child: SizedBox(
                  height: 28,
                  child: OutlinedButton.icon(
                    onPressed: () => context
                        .read<SyncStatusProvider>()
                        .triggerSync(app.id),
                    icon: const Icon(Icons.sync, size: 14),
                    label: const Text('Sync Now'),
                    style: OutlinedButton.styleFrom(
                      padding:
                          const EdgeInsets.symmetric(horizontal: 12),
                      textStyle: const TextStyle(fontSize: 12),
                    ),
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }

  String _formatSyncTime(DateTime syncAt) {
    final diff = DateTime.now().difference(syncAt);
    if (diff.inMinutes < 1) return 'Synced just now';
    if (diff.inMinutes < 60) return 'Synced ${diff.inMinutes} min ago';
    if (diff.inHours < 24) return 'Synced ${diff.inHours} hr ago';
    return 'Synced ${DateFormat('MMM d').format(syncAt)}';
  }
}
