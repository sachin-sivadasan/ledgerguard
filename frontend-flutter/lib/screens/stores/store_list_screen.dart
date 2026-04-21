import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import '../../mock_data/mock_apps.dart';
import '../../providers/apps_provider.dart';
import '../../providers/store_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_risk_badge.dart';
import '../../widgets/lg_search_field.dart';

class StoreListScreen extends StatelessWidget {
  const StoreListScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final hasApps = context.watch<AppsProvider>().apps.isNotEmpty;
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
    final stores = provider.stores;
    final theme = Theme.of(context);
    final showAppFilter = mockApps.length > 1;

    return LgPage(
      title: 'Stores',
      subtitle: '${stores.length} stores',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Wrap(
            spacing: LgSpacing.s300,
            crossAxisAlignment: WrapCrossAlignment.center,
            children: [
              if (showAppFilter)
                PopupMenuButton<String?>(
                  onSelected: provider.setSelectedApp,
                  itemBuilder: (_) => [
                    const PopupMenuItem(value: null, child: Text('All Apps')),
                    ...mockApps.map((app) => PopupMenuItem(
                          value: app.id,
                          child: Text(app.name),
                        )),
                  ],
                  child: Chip(
                    label: Text(provider.selectedAppId != null
                        ? mockApps.firstWhere((a) => a.id == provider.selectedAppId).name
                        : 'All Apps'),
                    deleteIcon: provider.selectedAppId != null
                        ? const Icon(Icons.close, size: 14)
                        : null,
                    onDeleted: provider.selectedAppId != null
                        ? () => provider.setSelectedApp(null)
                        : null,
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
          Wrap(
            spacing: LgSpacing.s400,
            runSpacing: LgSpacing.s400,
            children: stores.map((store) {
              return SizedBox(
                width: 350,
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
