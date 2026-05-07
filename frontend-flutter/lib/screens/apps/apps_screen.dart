import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../providers/apps_provider.dart';
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
  }

  @override
  Widget build(BuildContext context) {
    final provider = context.watch<AppsProvider>();
    final theme = Theme.of(context);
    final dateFmt = DateFormat('MMM d, h:mm a');

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
                  // Apps list
                  ListView(
                    children: provider.apps.map((app) {
                      return Padding(
                        padding: const EdgeInsets.only(bottom: LgSpacing.s300),
                        child: LgCard(
                          child: Row(
                            children: [
                              Container(
                                width: 48, height: 48,
                                decoration: BoxDecoration(
                                  color: LgColors.primary.withValues(alpha: 0.1),
                                  borderRadius: BorderRadius.circular(8),
                                ),
                                child: Center(
                                  child: Text(
                                    app.name.substring(0, 2),
                                    style: const TextStyle(fontWeight: FontWeight.w600, color: LgColors.primary),
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
                                        const Icon(Icons.people, size: 14, color: LgColors.textSecondary),
                                        const SizedBox(width: 4),
                                        Text('${app.installCount} installs', style: theme.textTheme.bodySmall),
                                        const SizedBox(width: LgSpacing.s400),
                                        const Icon(Icons.star_rounded, size: 14, color: LgColors.starRating),
                                        const SizedBox(width: 4),
                                        Text('${app.avgRating}', style: theme.textTheme.bodySmall),
                                      ],
                                    ),
                                  ],
                                ),
                              ),
                              Column(
                                crossAxisAlignment: CrossAxisAlignment.end,
                                children: [
                                  LgBadge(
                                    label: app.syncStatus == 'synced' ? 'Synced' : 'Pending',
                                    tone: app.syncStatus == 'synced' ? BadgeTone.success : BadgeTone.warning,
                                  ),
                                  const SizedBox(height: LgSpacing.s100),
                                  if (app.lastSyncAt != null)
                                    Text(dateFmt.format(app.lastSyncAt!), style: theme.textTheme.bodySmall),
                                ],
                              ),
                            ],
                          ),
                        ),
                      );
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
