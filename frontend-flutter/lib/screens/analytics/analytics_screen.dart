import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import '../../mock_data/mock_apps.dart';
import '../../providers/analytics_provider.dart';
import '../../providers/apps_provider.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_page.dart';
import 'revenue_tab.dart';
import 'forecasting_tab.dart';
import 'profit_tab.dart';
import 'cohort_tab.dart';
import 'multi_app_tab.dart';

class AnalyticsScreen extends StatelessWidget {
  const AnalyticsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final hasApps = context.watch<AppsProvider>().apps.isNotEmpty;
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
    final showAppFilter = mockApps.length > 1;

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
                child: PopupMenuButton<String?>(
                  onSelected: provider.setSelectedApp,
                  itemBuilder: (_) => [
                    const PopupMenuItem(
                        value: null, child: Text('All Apps')),
                    ...mockApps.map((app) => PopupMenuItem(
                          value: app.id,
                          child: Text(app.name),
                        )),
                  ],
                  child: Chip(
                    label: Text(provider.selectedAppId != null
                        ? mockApps
                            .firstWhere(
                                (a) => a.id == provider.selectedAppId)
                            .name
                        : 'All Apps'),
                    deleteIcon: provider.selectedAppId != null
                        ? const Icon(Icons.close, size: 14)
                        : null,
                    onDeleted: provider.selectedAppId != null
                        ? () => provider.setSelectedApp(null)
                        : null,
                  ),
                ),
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
