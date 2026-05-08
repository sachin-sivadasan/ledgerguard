import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../providers/analytics_provider.dart';
import '../../providers/apps_provider.dart';
import '../../providers/organization_provider.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import 'revenue_tab.dart';
import 'forecasting_tab.dart';
import 'profit_tab.dart';
import 'cohort_tab.dart';
import 'multi_app_tab.dart';

class AnalyticsScreen extends StatefulWidget {
  const AnalyticsScreen({super.key});

  @override
  State<AnalyticsScreen> createState() => _AnalyticsScreenState();
}

class _AnalyticsScreenState extends State<AnalyticsScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<AnalyticsProvider>().setSelectedApp(appId);
  }

  @override
  Widget build(BuildContext context) {
    final appsProvider = context.watch<AppsProvider>();
    final hasApps = appsProvider.apps.isNotEmpty;

    if (!hasApps) {
      return LgPage(
        title: 'Analytics',
        child: LgEmptyState(
          icon: Icons.analytics_outlined,
          heading: 'No analytics yet',
          description:
              'Connect your Shopify app to see revenue analytics.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<AnalyticsProvider>();

    if (provider.error != null) {
      return LgPage(
        title: 'Analytics',
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.mrrSnapshots.isEmpty) {
      return LgPage(
        title: 'Analytics',
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final appsList = context.watch<AppsProvider>().apps;
    final showAppFilter = appsList.length > 1;

    return LgPage(
      title: 'Analytics',
      subtitle: 'Revenue analysis and forecasting',
      scrollable: false,
      child: DefaultTabController(
        length: 5,
        child: Column(
          children: [
            if (showAppFilter) ...[
              Align(
                alignment: Alignment.centerLeft,
                child: Builder(builder: (context) {
                  final canViewAllApps = context.watch<OrganizationProvider>().canViewAllApps;
                  return PopupMenuButton<String?>(
                    onSelected: provider.setSelectedApp,
                    itemBuilder: (_) => [
                      if (canViewAllApps)
                        const PopupMenuItem(
                            value: null, child: Text('All Apps')),
                      ...appsList.map((app) => PopupMenuItem(
                            value: app.id,
                            child: Text(app.name),
                          )),
                    ],
                    child: Chip(
                      label: Text(provider.selectedAppId != null
                          ? appsList
                              .firstWhere(
                                  (a) => a.id == provider.selectedAppId,
                                  orElse: () => appsList.first)
                              .name
                          : canViewAllApps
                              ? 'All Apps'
                              : appsList.first.name),
                      deleteIcon: canViewAllApps && provider.selectedAppId != null
                          ? const Icon(Icons.close, size: 14)
                          : null,
                      onDeleted: canViewAllApps && provider.selectedAppId != null
                          ? () => provider.setSelectedApp(null)
                          : null,
                    ),
                  );
                }),
              ),
              const SizedBox(height: LgSpacing.s300),
            ],
            const TabBar(
              isScrollable: true,
              tabAlignment: TabAlignment.start,
              padding: EdgeInsets.zero,
              tabs: [
                Tab(text: 'Revenue'),
                Tab(text: 'Forecasting'),
                Tab(text: 'Profit & Expense'),
                Tab(text: 'Cohorts'),
                Tab(text: 'Multi-App'),
              ],
            ),
            const SizedBox(height: LgSpacing.s400),
            Expanded(
              child: TabBarView(
                children: [
                  RevenueTab(),
                  ForecastingTab(),
                  ProfitTab(),
                  CohortTab(),
                  MultiAppTab(),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
