import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:get_it/get_it.dart';
import 'package:go_router/go_router.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/dashboard_metrics.dart';
import '../../domain/entities/time_range.dart';
import '../../domain/repositories/app_repository.dart';
import '../blocs/dashboard/dashboard.dart';
import '../blocs/earnings/earnings.dart';
import '../blocs/preferences/preferences.dart';
import '../widgets/ai_insight_card.dart';
import '../widgets/dashboard_config_dialog.dart';
import '../widgets/earnings_timeline_chart.dart';
import '../widgets/kpi_card.dart';
import '../widgets/revenue_mix_chart.dart';
import '../widgets/risk_distribution_chart.dart';
import '../widgets/role_guard.dart';
import '../widgets/shared.dart';
import '../widgets/time_range_selector.dart';

/// Executive Dashboard page displaying key metrics
class DashboardPage extends StatelessWidget {
  const DashboardPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey[50],
      appBar: AppBar(
        title: const Text('Dashboard'),
        actions: [
          // Time Range Selector - responsive
          BlocBuilder<DashboardBloc, DashboardState>(
            builder: (context, state) {
              final timeRange = state is DashboardLoaded
                  ? state.timeRange
                  : TimeRange.thisMonth();
              return Padding(
                padding: const EdgeInsets.symmetric(vertical: 8),
                child: TimeRangeSelector(
                  currentRange: timeRange,
                  onRangeChanged: (range) {
                    context.read<DashboardBloc>().add(TimeRangeChanged(range));
                  },
                ),
              );
            },
          ),
          const SizedBox(width: 4),
          // Refresh button - always visible
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
          // Overflow menu for secondary actions
          PopupMenuButton<String>(
            icon: const Icon(Icons.more_vert),
            tooltip: 'More options',
            onSelected: (value) {
              switch (value) {
                case 'subscriptions':
                  _navigateToSubscriptions(context);
                  break;
                case 'settings':
                  context
                      .read<PreferencesBloc>()
                      .add(const LoadPreferencesRequested());
                  DashboardConfigDialog.show(context);
                  break;
                case 'profile':
                  context.push('/profile');
                  break;
              }
            },
            itemBuilder: (context) => [
              const PopupMenuItem(
                value: 'subscriptions',
                child: ListTile(
                  leading: Icon(Icons.subscriptions_outlined),
                  title: Text('Subscriptions'),
                  contentPadding: EdgeInsets.zero,
                  visualDensity: VisualDensity.compact,
                ),
              ),
              const PopupMenuItem(
                value: 'settings',
                child: ListTile(
                  leading: Icon(Icons.settings),
                  title: Text('Configure Dashboard'),
                  contentPadding: EdgeInsets.zero,
                  visualDensity: VisualDensity.compact,
                ),
              ),
              const PopupMenuItem(
                value: 'profile',
                child: ListTile(
                  leading: Icon(Icons.person_outline),
                  title: Text('Profile'),
                  contentPadding: EdgeInsets.zero,
                  visualDensity: VisualDensity.compact,
                ),
              ),
            ],
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
            return _buildDashboard(context, state.metrics, state.timeRange);
          }

          return const SizedBox.shrink();
        },
      ),
    );
  }

  Future<void> _navigateToSubscriptions(BuildContext context) async {
    final appRepository = GetIt.instance<AppRepository>();
    final selectedApp = await appRepository.getSelectedApp();
    if (selectedApp != null && context.mounted) {
      // Extract numeric ID from GID (e.g., "gid://partners/App/4599915" -> "4599915")
      final parts = selectedApp.id.split('/');
      final numericAppId = parts.isNotEmpty ? parts.last : selectedApp.id;
      context.push('/apps/$numericAppId/subscriptions');
    } else if (context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please select an app first')),
      );
    }
  }

  Widget _buildErrorState(BuildContext context, String message) {
    // Check if this is a "no app selected" error
    if (message.contains('No app selected') ||
        message.contains('select an app')) {
      return _buildOnboardingState(context);
    }

    return ErrorStateWidget(
      title: 'Failed to load dashboard',
      message: message,
      onRetry: () =>
          context.read<DashboardBloc>().add(const LoadDashboardRequested()),
    );
  }

  Widget _buildOnboardingState(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.rocket_launch_outlined,
              size: 80,
              color: AppTheme.primary.withOpacity(0.5),
            ),
            const SizedBox(height: 24),
            Text(
              'Welcome to LedgerGuard!',
              style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 12),
            Text(
              'Connect your Shopify Partner account to start tracking your app revenue.',
              style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                    color: Colors.grey[600],
                  ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 32),
            SizedBox(
              width: 280,
              height: 48,
              child: ElevatedButton.icon(
                onPressed: () => context.go('/partner-integration'),
                icon: const Icon(Icons.link),
                label: const Text('Connect Partner Account'),
              ),
            ),
            const SizedBox(height: 16),
            TextButton(
              onPressed: () => context.go('/app-selection'),
              child: const Text('I already connected, select an app'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildEmptyState(BuildContext context, String message) {
    return EmptyStateWidget(
      title: 'No Metrics Yet',
      message: message,
      icon: Icons.analytics_outlined,
      actionLabel: 'Sync Data',
      actionIcon: Icons.sync,
      onAction: () =>
          context.read<DashboardBloc>().add(const RefreshDashboardRequested()),
    );
  }

  Widget _buildDashboard(
    BuildContext context,
    DashboardMetrics metrics,
    TimeRange timeRange,
  ) {
    return RefreshIndicator(
      onRefresh: () async {
        context.read<DashboardBloc>().add(const RefreshDashboardRequested());
        // Wait for refresh to complete
        await context.read<DashboardBloc>().stream.firstWhere(
              (state) => state is DashboardLoaded && !state.isRefreshing,
            );
      },
      child: LayoutBuilder(
        builder: (context, constraints) {
          // Responsive padding: smaller on mobile
          final padding = constraints.maxWidth < 600 ? 12.0 : 20.0;

          return SingleChildScrollView(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: EdgeInsets.all(padding),
            child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const ProGuard(child: AiInsightCard()),
            _buildSectionHeader(context, 'Primary KPIs'),
            const SizedBox(height: 16),
            _buildPrimaryKpis(context, metrics, timeRange),
            const SizedBox(height: 32),
            _buildSectionHeader(context, 'Revenue & Risk'),
            const SizedBox(height: 16),
            _buildSecondarySection(context, metrics),
            const SizedBox(height: 32),
            _buildSectionHeader(context, 'Earnings Timeline'),
            const SizedBox(height: 16),
            BlocProvider(
              create: (_) => GetIt.instance<EarningsBloc>(),
              child: EarningsTimelineChart(timeRange: timeRange),
            ),
          ],
        ),
          );
        },
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

  Widget _buildPrimaryKpis(
    BuildContext context,
    DashboardMetrics metrics,
    TimeRange timeRange,
  ) {
    final periodSubtitle = timeRange.preset.displayName;
    final delta = metrics.delta;

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
                  subtitle: periodSubtitle,
                  icon: Icons.trending_up,
                  color: AppTheme.success,
                  isLarge: true,
                  delta: delta?.renewalSuccessIndicator,
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
                  delta: delta?.activeMrrIndicator,
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
                  delta: delta?.revenueAtRiskIndicator,
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
                  delta: delta?.churnCountIndicator,
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
                      subtitle: periodSubtitle,
                      icon: Icons.trending_up,
                      color: AppTheme.success,
                      delta: delta?.renewalSuccessIndicator,
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
                      delta: delta?.activeMrrIndicator,
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
                      delta: delta?.revenueAtRiskIndicator,
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
                      delta: delta?.churnCountIndicator,
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
                subtitle: periodSubtitle,
                icon: Icons.trending_up,
                color: AppTheme.success,
                delta: delta?.renewalSuccessIndicator,
              ),
              const SizedBox(height: 16),
              KpiCard(
                title: 'Active MRR',
                value: metrics.formattedMrr,
                subtitle: 'Monthly recurring revenue',
                icon: Icons.attach_money,
                color: AppTheme.primary,
                delta: delta?.activeMrrIndicator,
              ),
              const SizedBox(height: 16),
              KpiCard(
                title: 'Revenue at Risk',
                value: metrics.formattedRevenueAtRisk,
                subtitle: 'Needs attention',
                icon: Icons.warning_amber,
                color: AppTheme.warning,
                delta: delta?.revenueAtRiskIndicator,
              ),
              const SizedBox(height: 16),
              KpiCard(
                title: 'Churned',
                value: metrics.formattedChurnedRevenue,
                subtitle: '${metrics.churnedCount} subscriptions',
                icon: Icons.trending_down,
                color: AppTheme.danger,
                delta: delta?.churnCountIndicator,
              ),
            ],
          );
        }
      },
    );
  }

  Widget _buildSecondarySection(
      BuildContext context, DashboardMetrics metrics) {
    final delta = metrics.delta;

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
                            delta: delta?.usageRevenueIndicator,
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: KpiCardCompact(
                            title: 'Total Revenue',
                            value: metrics.formattedTotalRevenue,
                            icon: Icons.account_balance_wallet,
                            color: AppTheme.primary,
                            delta: delta?.totalRevenueIndicator,
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
                delta: delta?.usageRevenueIndicator,
              ),
              const SizedBox(height: 12),
              KpiCardCompact(
                title: 'Total Revenue',
                value: metrics.formattedTotalRevenue,
                icon: Icons.account_balance_wallet,
                color: AppTheme.primary,
                delta: delta?.totalRevenueIndicator,
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
