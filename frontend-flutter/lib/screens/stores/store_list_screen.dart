import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../providers/apps_provider.dart';
import '../../providers/store_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_risk_badge.dart';
import '../../widgets/lg_search_field.dart';

class StoreListScreen extends StatefulWidget {
  const StoreListScreen({super.key});

  @override
  State<StoreListScreen> createState() => _StoreListScreenState();
}

class _StoreListScreenState extends State<StoreListScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<StoreProvider>().setSelectedApp(appId);
  }

  @override
  Widget build(BuildContext context) {
    final appsProvider = context.watch<AppsProvider>();
    final hasApps = appsProvider.apps.isNotEmpty;

    if (!hasApps) {
      return LgPage(
        title: 'Stores',
        child: LgEmptyState(
          icon: Icons.storefront,
          heading: 'No stores yet',
          description: 'Connect your Shopify app to see store data.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<StoreProvider>();

    if (provider.error != null) {
      return LgPage(
        title: 'Stores',
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.stores.isEmpty) {
      return LgPage(
        title: 'Stores',
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final stores = provider.stores;
    final theme = Theme.of(context);
    final appsList = context.watch<AppsProvider>().apps;
    final showAppFilter = appsList.length > 1;
    final totalLabel = provider.demoMode
        ? '${stores.length} stores'
        : '${provider.totalCount} stores';

    return LgPage(
      title: 'Stores',
      subtitle: totalLabel,
      onRefresh: refreshData,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Wrap(
            spacing: LgSpacing.s300,
            crossAxisAlignment: WrapCrossAlignment.center,
            children: [
              if (showAppFilter)
                PopupMenuButton<String>(
                  onSelected: provider.setSelectedApp,
                  itemBuilder: (_) => [
                    ...appsList.map((app) => PopupMenuItem(
                          value: app.id,
                          child: Text(app.name),
                        )),
                  ],
                  child: Chip(
                    label: Text(appsList.firstWhere((a) => a.id == provider.selectedAppId, orElse: () => appsList.first).name),
                  ),
                ),
              LgSearchField(
                value: provider.searchQuery,
                onChanged: provider.setSearch,
                hintText: 'Search stores...',
              ),
            ],
          ),
          const SizedBox(height: LgSpacing.s400),
          LayoutBuilder(
            builder: (context, constraints) {
              final width = constraints.maxWidth;
              // Mobile: 1 col, Tablet: 2 col, Desktop: 3 col
              final cols = width < 600 ? 1 : (width < 900 ? 2 : 3);
              final spacing = LgSpacing.s400;
              final totalSpacing = spacing * (cols - 1);
              final cardWidth = cols == 1 ? width : (width - totalSpacing) / cols;

              return Wrap(
                spacing: spacing,
                runSpacing: spacing,
                children: stores.map((store) {
                  return SizedBox(
                    width: cardWidth,
                    child: MouseRegion(
                      cursor: SystemMouseCursors.click,
                      child: InkWell(
                        borderRadius: BorderRadius.circular(12),
                        onTap: () => context.go('/stores/${store.id}'),
                        child: LgCard(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  Expanded(
                                    child: Text(
                                      store.shopDomain.replaceAll('.myshopify.com', ''),
                                      style: theme.textTheme.titleSmall,
                                    ),
                                  ),
                                  LgRiskBadge(riskState: store.riskState),
                                ],
                              ),
                              const SizedBox(height: LgSpacing.s300),
                              Row(
                                children: [
                                  _Stat('Health', '${store.healthScore}%', _healthColor(store.healthScore)),
                                  const SizedBox(width: LgSpacing.s600),
                                  _Stat('LTV', store.ltvFormatted, LgColors.textPrimary),
                                  const SizedBox(width: LgSpacing.s600),
                                  _Stat('Apps', '${store.installedAppIds.length}', LgColors.textPrimary),
                                ],
                              ),
                            ],
                          ),
                        ),
                      ),
                    ),
                  );
                }).toList(),
              );
            },
          ),

          // Load More
          if (provider.hasMore && !provider.demoMode)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: LgSpacing.s400),
              child: Center(
                child: provider.isLoadingMore
                    ? const CircularProgressIndicator()
                    : OutlinedButton(
                        onPressed: provider.loadMore,
                        child: const Text('Load More'),
                      ),
              ),
            ),
        ],
      ),
    );
  }

  Color _healthColor(int score) {
    if (score >= 80) return LgColors.success;
    if (score >= 50) return LgColors.warning;
    return LgColors.critical;
  }
}

class _Stat extends StatelessWidget {
  final String label;
  final String value;
  final Color valueColor;
  const _Stat(this.label, this.value, this.valueColor);

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: Theme.of(context).textTheme.bodySmall),
        Text(value, style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: valueColor)),
      ],
    );
  }
}
