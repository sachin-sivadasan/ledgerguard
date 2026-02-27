import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/dashboard_metrics.dart';
import '../../domain/entities/dashboard_preferences.dart';
import '../blocs/dashboard/dashboard.dart';
import '../blocs/preferences/preferences.dart';
import '../widgets/ai_insight_card.dart';
import '../widgets/dashboard_config_dialog.dart';
import '../widgets/kpi_card.dart';
import '../widgets/revenue_mix_chart.dart';
import '../widgets/risk_distribution_chart.dart';
import '../widgets/role_guard.dart';

/// Executive Dashboard page displaying key metrics
class DashboardPage extends StatelessWidget {
  const DashboardPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey[50],
      appBar: AppBar(
        title: const Text('Executive Dashboard'),
        actions: [
          IconButton(
            icon: const Icon(Icons.settings),
            tooltip: 'Configure Dashboard',
            onPressed: () {
              // Load preferences before showing dialog
              context
                  .read<PreferencesBloc>()
                  .add(const LoadPreferencesRequested());
              DashboardConfigDialog.show(context);
            },
          ),
          BlocBuilder<DashboardBloc, DashboardState>(
            builder: (context, state) {
              final isRefreshing =
                  state is DashboardLoaded && state.isRefreshing;
              return IconButton(
                icon: isRefreshing
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: Colors.white,
                        ),
                      )
                    : const Icon(Icons.refresh),
                onPressed: isRefreshing
                    ? null
                    : () {
                        context
                            .read<DashboardBloc>()
                            .add(const RefreshDashboardRequested());
                      },
              );
            },
          ),
          IconButton(
            icon: const Icon(Icons.person_outline),
            tooltip: 'Profile',
            onPressed: () => context.push('/profile'),
          ),
        ],
      ),
      body: BlocBuilder<DashboardBloc, DashboardState>(
        builder: (context, state) {
          if (state is DashboardInitial) {
            // Trigger load on first build
            context.read<DashboardBloc>().add(const LoadDashboardRequested());
            return const Center(child: CircularProgressIndicator());
          }

          if (state is DashboardLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state is DashboardEmpty) {
            return _buildEmptyState(context, state.message);
          }

          if (state is DashboardError) {
            return _buildErrorState(context, state.message);
          }

          if (state is DashboardLoaded) {
            return _buildDashboard(context, state.metrics);
          }

          return const SizedBox.shrink();
        },
      ),
    );
  }

  Widget _buildErrorState(BuildContext context, String message) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.error_outline,
              size: 64,
              color: Colors.red[300],
            ),
            const SizedBox(height: 16),
            Text(
              'Failed to load dashboard',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 8),
            Text(
              message,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Colors.grey[600],
                  ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: () {
                context
                    .read<DashboardBloc>()
                    .add(const LoadDashboardRequested());
              },
              icon: const Icon(Icons.refresh),
              label: const Text('Retry'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildEmptyState(BuildContext context, String message) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.analytics_outlined,
              size: 80,
              color: Colors.grey[400],
            ),
            const SizedBox(height: 24),
            Text(
              'No Metrics Yet',
              style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 12),
            Text(
              message,
              style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                    color: Colors.grey[600],
                  ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 32),
            ElevatedButton.icon(
              onPressed: () {
                context
                    .read<DashboardBloc>()
                    .add(const RefreshDashboardRequested());
              },
              icon: const Icon(Icons.sync),
              label: const Text('Sync Data'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildDashboard(BuildContext context, DashboardMetrics metrics) {
    return RefreshIndicator(
      onRefresh: () async {
        context.read<DashboardBloc>().add(const RefreshDashboardRequested());
        // Wait for refresh to complete
        await context.read<DashboardBloc>().stream.firstWhere(
              (state) => state is DashboardLoaded && !state.isRefreshing,
            );
      },
      child: SingleChildScrollView(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const ProGuard(child: AiInsightCard()),
            _buildSectionHeader(context, 'Primary KPIs'),
            const SizedBox(height: 16),
            _buildPrimaryKpis(context, metrics),
            const SizedBox(height: 32),
            _buildSectionHeader(context, 'Revenue & Risk'),
            const SizedBox(height: 16),
            _buildSecondarySection(context, metrics),
          ],
        ),
      ),
    );
  }

  Widget _buildSectionHeader(BuildContext context, String title) {
    return Text(
      title,
      style: Theme.of(context).textTheme.titleLarge?.copyWith(
            fontWeight: FontWeight.bold,
          ),
    );
  }

  Widget _buildPrimaryKpis(BuildContext context, DashboardMetrics metrics) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final isWide = constraints.maxWidth > 800;
        final isMedium = constraints.maxWidth > 500;

        if (isWide) {
          return Row(
            children: [
              Expanded(
                child: KpiCard(
                  title: 'Renewal Success Rate',
                  value: '${metrics.renewalSuccessRate.toStringAsFixed(1)}%',
                  subtitle: 'Last 30 days',
                  icon: Icons.trending_up,
                  color: AppTheme.success,
                  isLarge: true,
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: KpiCard(
                  title: 'Active MRR',
                  value: metrics.formattedMrr,
                  subtitle: 'Monthly recurring revenue',
                  icon: Icons.attach_money,
                  color: AppTheme.primary,
                  isLarge: true,
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: KpiCard(
                  title: 'Revenue at Risk',
                  value: metrics.formattedRevenueAtRisk,
                  subtitle: 'Needs attention',
                  icon: Icons.warning_amber,
                  color: AppTheme.warning,
                  isLarge: true,
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: KpiCard(
                  title: 'Churned',
                  value: metrics.formattedChurnedRevenue,
                  subtitle: '${metrics.churnedCount} subscriptions',
                  icon: Icons.trending_down,
                  color: AppTheme.danger,
                  isLarge: true,
                ),
              ),
            ],
          );
        } else if (isMedium) {
          return Column(
            children: [
              Row(
                children: [
                  Expanded(
                    child: KpiCard(
                      title: 'Renewal Success Rate',
                      value:
                          '${metrics.renewalSuccessRate.toStringAsFixed(1)}%',
                      subtitle: 'Last 30 days',
                      icon: Icons.trending_up,
                      color: AppTheme.success,
                    ),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: KpiCard(
                      title: 'Active MRR',
                      value: metrics.formattedMrr,
                      subtitle: 'Monthly recurring revenue',
                      icon: Icons.attach_money,
                      color: AppTheme.primary,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              Row(
                children: [
                  Expanded(
                    child: KpiCard(
                      title: 'Revenue at Risk',
                      value: metrics.formattedRevenueAtRisk,
                      subtitle: 'Needs attention',
                      icon: Icons.warning_amber,
                      color: AppTheme.warning,
                    ),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: KpiCard(
                      title: 'Churned',
                      value: metrics.formattedChurnedRevenue,
                      subtitle: '${metrics.churnedCount} subscriptions',
                      icon: Icons.trending_down,
                      color: AppTheme.danger,
                    ),
                  ),
                ],
              ),
            ],
          );
        } else {
          return Column(
            children: [
              KpiCard(
                title: 'Renewal Success Rate',
                value: '${metrics.renewalSuccessRate.toStringAsFixed(1)}%',
                subtitle: 'Last 30 days',
                icon: Icons.trending_up,
                color: AppTheme.success,
              ),
              const SizedBox(height: 16),
              KpiCard(
                title: 'Active MRR',
                value: metrics.formattedMrr,
                subtitle: 'Monthly recurring revenue',
                icon: Icons.attach_money,
                color: AppTheme.primary,
              ),
              const SizedBox(height: 16),
              KpiCard(
                title: 'Revenue at Risk',
                value: metrics.formattedRevenueAtRisk,
                subtitle: 'Needs attention',
                icon: Icons.warning_amber,
                color: AppTheme.warning,
              ),
              const SizedBox(height: 16),
              KpiCard(
                title: 'Churned',
                value: metrics.formattedChurnedRevenue,
                subtitle: '${metrics.churnedCount} subscriptions',
                icon: Icons.trending_down,
                color: AppTheme.danger,
              ),
            ],
          );
        }
      },
    );
  }

  Widget _buildSecondarySection(
      BuildContext context, DashboardMetrics metrics) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final isWide = constraints.maxWidth > 700;

        if (isWide) {
          return Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                child: Column(
                  children: [
                    Row(
                      children: [
                        Expanded(
                          child: KpiCardCompact(
                            title: 'Usage Revenue',
                            value: metrics.formattedUsageRevenue,
                            icon: Icons.data_usage,
                            color: AppTheme.secondary,
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: KpiCardCompact(
                            title: 'Total Revenue',
                            value: metrics.formattedTotalRevenue,
                            icon: Icons.account_balance_wallet,
                            color: AppTheme.primary,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    RevenueMixChart(revenueMix: metrics.revenueMix),
                  ],
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: RiskDistributionChart(
                    riskDistribution: metrics.riskDistribution),
              ),
            ],
          );
        } else {
          return Column(
            children: [
              KpiCardCompact(
                title: 'Usage Revenue',
                value: metrics.formattedUsageRevenue,
                icon: Icons.data_usage,
                color: AppTheme.secondary,
              ),
              const SizedBox(height: 12),
              KpiCardCompact(
                title: 'Total Revenue',
                value: metrics.formattedTotalRevenue,
                icon: Icons.account_balance_wallet,
                color: AppTheme.primary,
              ),
              const SizedBox(height: 16),
              RevenueMixChart(revenueMix: metrics.revenueMix),
              const SizedBox(height: 16),
              RiskDistributionChart(
                  riskDistribution: metrics.riskDistribution),
            ],
          );
        }
      },
    );
  }
}
