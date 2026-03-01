import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:get_it/get_it.dart';
import 'package:go_router/go_router.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/dashboard_metrics.dart';
import '../../domain/entities/dashboard_preferences.dart';
import '../../domain/entities/time_range.dart';
import '../../domain/repositories/app_repository.dart';
import '../blocs/app_selection/app_selection.dart';
import '../blocs/dashboard/dashboard.dart';
import '../blocs/earnings/earnings.dart';
import '../blocs/preferences/preferences.dart';
import '../widgets/ai_insight_card.dart';
import '../widgets/app_selector.dart';
import '../widgets/dashboard_config_dialog.dart';
import '../widgets/earnings_timeline_chart.dart';
import '../widgets/fee_insights_card.dart';
import '../widgets/kpi_card.dart';
import '../widgets/revenue_mix_chart.dart';
import '../widgets/risk_distribution_chart.dart';
import '../widgets/role_guard.dart';
import '../widgets/shared.dart';
import '../widgets/time_range_selector.dart';

/// Executive Dashboard page displaying key metrics
class DashboardPage extends StatefulWidget {
  const DashboardPage({super.key});

  @override
  State<DashboardPage> createState() => _DashboardPageState();
}

class _DashboardPageState extends State<DashboardPage> {
  @override
  void initState() {
    super.initState();
    // Load preferences when dashboard initializes
    context.read<PreferencesBloc>().add(const LoadPreferencesRequested());
    // Load tracked apps for multi-app selector
    context.read<AppSelectionBloc>().add(const FetchAppsRequested());
  }

  void _onAppChanged() {
    // Refresh dashboard when app selection changes
    context.read<DashboardBloc>().add(const LoadDashboardRequested());
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey[50],
      appBar: AppBar(
        title: const Text('Dashboard'),
        actions: [
          // App Selector for multi-app support
          AppSelector(onAppChanged: _onAppChanged),
          const SizedBox(width: 8),
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
                case 'sync':
                  _triggerSync(context);
                  break;
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
                value: 'sync',
                child: ListTile(
                  leading: Icon(Icons.sync),
                  title: Text('Sync Data'),
                  contentPadding: EdgeInsets.zero,
                  visualDensity: VisualDensity.compact,
                ),
              ),
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
      body: BlocBuilder<PreferencesBloc, PreferencesState>(
        builder: (context, prefsState) {
          // Get preferences (use defaults if not loaded yet)
          final preferences = prefsState is PreferencesLoaded
              ? prefsState.preferences
              : DashboardPreferences.defaults();

          return BlocBuilder<DashboardBloc, DashboardState>(
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
                return _buildDashboard(
                  context,
                  state.metrics,
                  state.timeRange,
                  preferences,
                );
              }

              return const SizedBox.shrink();
            },
          );
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

  Future<void> _triggerSync(BuildContext context) async {
    final scaffoldMessenger = ScaffoldMessenger.of(context);

    // Show syncing indicator
    scaffoldMessenger.showSnackBar(
      const SnackBar(
        content: Row(
          children: [
            SizedBox(
              width: 16,
              height: 16,
              child: CircularProgressIndicator(
                strokeWidth: 2,
                color: Colors.white,
              ),
            ),
            SizedBox(width: 12),
            Text('Syncing data from Shopify...'),
          ],
        ),
        duration: Duration(seconds: 30),
      ),
    );

    try {
      final appRepository = GetIt.instance<AppRepository>();
      final result = await appRepository.syncData();

      scaffoldMessenger.hideCurrentSnackBar();

      if (result.isSuccess) {
        scaffoldMessenger.showSnackBar(
          SnackBar(
            content: Text(
              'Sync complete! ${result.transactionCount} transactions synced.',
            ),
            backgroundColor: Colors.green,
          ),
        );

        // Refresh dashboard after sync
        if (context.mounted) {
          context
              .read<DashboardBloc>()
              .add(const RefreshDashboardRequested());
        }
      } else {
        scaffoldMessenger.showSnackBar(
          SnackBar(
            content: Text('Sync failed: ${result.error}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } catch (e) {
      scaffoldMessenger.hideCurrentSnackBar();
      scaffoldMessenger.showSnackBar(
        SnackBar(
          content: Text('Sync failed: $e'),
          backgroundColor: Colors.red,
        ),
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
    DashboardPreferences preferences,
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
                if (preferences.primaryKpis.isNotEmpty) ...[
                  _buildSectionHeader(context, 'Primary KPIs'),
                  const SizedBox(height: 16),
                  _buildPrimaryKpis(context, metrics, timeRange, preferences),
                  const SizedBox(height: 32),
                ],
                if (_hasSecondaryWidgets(preferences)) ...[
                  _buildSectionHeader(context, 'Revenue & Risk'),
                  const SizedBox(height: 16),
                  _buildSecondarySection(context, metrics, preferences),
                  const SizedBox(height: 32),
                ],
                _buildSectionHeader(context, 'Fee Insights'),
                const SizedBox(height: 16),
                FeeInsightsCard(totalGrossCents: metrics.totalRevenue),
                if (preferences.isSecondaryWidgetEnabled(
                    SecondaryWidget.earningsTimeline)) ...[
                  const SizedBox(height: 32),
                  _buildSectionHeader(context, 'Earnings Timeline'),
                  const SizedBox(height: 16),
                  BlocProvider(
                    create: (_) => GetIt.instance<EarningsBloc>(),
                    child: EarningsTimelineChart(timeRange: timeRange),
                  ),
                ],
              ],
            ),
          );
        },
      ),
    );
  }

  bool _hasSecondaryWidgets(DashboardPreferences preferences) {
    return preferences
            .isSecondaryWidgetEnabled(SecondaryWidget.usageRevenue) ||
        preferences.isSecondaryWidgetEnabled(SecondaryWidget.totalRevenue) ||
        preferences.isSecondaryWidgetEnabled(SecondaryWidget.revenueMixChart) ||
        preferences
            .isSecondaryWidgetEnabled(SecondaryWidget.riskDistributionChart);
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
    DashboardPreferences preferences,
  ) {
    final kpis = preferences.primaryKpis;
    if (kpis.isEmpty) return const SizedBox.shrink();

    return LayoutBuilder(
      builder: (context, constraints) {
        final isWide = constraints.maxWidth > 800;
        final isMedium = constraints.maxWidth > 500;

        final kpiWidgets = kpis.map((kpi) {
          return _buildKpiCard(
            context,
            kpi,
            metrics,
            timeRange,
            isLarge: isWide,
          );
        }).toList();

        if (isWide) {
          return Row(
            children: kpiWidgets
                .expand((widget) => [Expanded(child: widget), const SizedBox(width: 16)])
                .toList()
              ..removeLast(),
          );
        } else if (isMedium) {
          // Grid layout for medium screens
          final rows = <Widget>[];
          for (var i = 0; i < kpiWidgets.length; i += 2) {
            final rowWidgets = <Widget>[Expanded(child: kpiWidgets[i])];
            if (i + 1 < kpiWidgets.length) {
              rowWidgets.add(const SizedBox(width: 16));
              rowWidgets.add(Expanded(child: kpiWidgets[i + 1]));
            }
            rows.add(Row(children: rowWidgets));
            if (i + 2 < kpiWidgets.length) {
              rows.add(const SizedBox(height: 16));
            }
          }
          return Column(children: rows);
        } else {
          // Stacked layout for small screens
          return Column(
            children: kpiWidgets
                .expand((widget) => [widget, const SizedBox(height: 16)])
                .toList()
              ..removeLast(),
          );
        }
      },
    );
  }

  Widget _buildKpiCard(
    BuildContext context,
    KpiType kpiType,
    DashboardMetrics metrics,
    TimeRange timeRange, {
    bool isLarge = false,
  }) {
    final periodSubtitle = timeRange.preset.displayName;
    final delta = metrics.delta;

    switch (kpiType) {
      case KpiType.renewalSuccessRate:
        return KpiCard(
          title: 'Renewal Success Rate',
          value: '${metrics.renewalSuccessRate.toStringAsFixed(1)}%',
          subtitle: periodSubtitle,
          icon: Icons.trending_up,
          color: AppTheme.success,
          isLarge: isLarge,
          delta: delta?.renewalSuccessIndicator,
        );
      case KpiType.activeMrr:
        return KpiCard(
          title: 'Active MRR',
          value: metrics.formattedMrr,
          subtitle: 'Monthly recurring revenue',
          icon: Icons.attach_money,
          color: AppTheme.primary,
          isLarge: isLarge,
          delta: delta?.activeMrrIndicator,
        );
      case KpiType.revenueAtRisk:
        return KpiCard(
          title: 'Revenue at Risk',
          value: metrics.formattedRevenueAtRisk,
          subtitle: 'Needs attention',
          icon: Icons.warning_amber,
          color: AppTheme.warning,
          isLarge: isLarge,
          delta: delta?.revenueAtRiskIndicator,
        );
      case KpiType.churned:
        return KpiCard(
          title: 'Churned',
          value: metrics.formattedChurnedRevenue,
          subtitle: '${metrics.churnedCount} subscriptions',
          icon: Icons.trending_down,
          color: AppTheme.danger,
          isLarge: isLarge,
          delta: delta?.churnCountIndicator,
        );
      case KpiType.usageRevenue:
        return KpiCard(
          title: 'Usage Revenue',
          value: metrics.formattedUsageRevenue,
          subtitle: periodSubtitle,
          icon: Icons.data_usage,
          color: AppTheme.secondary,
          isLarge: isLarge,
          delta: delta?.usageRevenueIndicator,
        );
      case KpiType.totalRevenue:
        return KpiCard(
          title: 'Total Revenue',
          value: metrics.formattedTotalRevenue,
          subtitle: periodSubtitle,
          icon: Icons.account_balance_wallet,
          color: AppTheme.primary,
          isLarge: isLarge,
          delta: delta?.totalRevenueIndicator,
        );
    }
  }

  Widget _buildSecondarySection(
    BuildContext context,
    DashboardMetrics metrics,
    DashboardPreferences preferences,
  ) {
    final delta = metrics.delta;

    final showUsageRevenue =
        preferences.isSecondaryWidgetEnabled(SecondaryWidget.usageRevenue);
    final showTotalRevenue =
        preferences.isSecondaryWidgetEnabled(SecondaryWidget.totalRevenue);
    final showRevenueMix =
        preferences.isSecondaryWidgetEnabled(SecondaryWidget.revenueMixChart);
    final showRiskDistribution = preferences
        .isSecondaryWidgetEnabled(SecondaryWidget.riskDistributionChart);

    // If nothing to show, return empty
    if (!showUsageRevenue &&
        !showTotalRevenue &&
        !showRevenueMix &&
        !showRiskDistribution) {
      return const SizedBox.shrink();
    }

    return LayoutBuilder(
      builder: (context, constraints) {
        final isWide = constraints.maxWidth > 700;

        final widgets = <Widget>[];

        // Build KPI cards
        final kpiCards = <Widget>[];
        if (showUsageRevenue) {
          kpiCards.add(
            KpiCardCompact(
              title: 'Usage Revenue',
              value: metrics.formattedUsageRevenue,
              icon: Icons.data_usage,
              color: AppTheme.secondary,
              delta: delta?.usageRevenueIndicator,
            ),
          );
        }
        if (showTotalRevenue) {
          kpiCards.add(
            KpiCardCompact(
              title: 'Total Revenue',
              value: metrics.formattedTotalRevenue,
              icon: Icons.account_balance_wallet,
              color: AppTheme.primary,
              delta: delta?.totalRevenueIndicator,
            ),
          );
        }

        if (isWide) {
          // Wide layout: side-by-side columns
          final leftColumn = <Widget>[];
          if (kpiCards.isNotEmpty) {
            if (kpiCards.length == 2) {
              leftColumn.add(
                Row(
                  children: [
                    Expanded(child: kpiCards[0]),
                    const SizedBox(width: 12),
                    Expanded(child: kpiCards[1]),
                  ],
                ),
              );
            } else if (kpiCards.length == 1) {
              leftColumn.add(kpiCards[0]);
            }
          }
          if (showRevenueMix) {
            if (leftColumn.isNotEmpty) {
              leftColumn.add(const SizedBox(height: 16));
            }
            leftColumn.add(RevenueMixChart(revenueMix: metrics.revenueMix));
          }

          if (leftColumn.isNotEmpty && showRiskDistribution) {
            return Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(child: Column(children: leftColumn)),
                const SizedBox(width: 16),
                Expanded(
                  child: RiskDistributionChart(
                    riskDistribution: metrics.riskDistribution,
                  ),
                ),
              ],
            );
          } else if (leftColumn.isNotEmpty) {
            return Column(children: leftColumn);
          } else if (showRiskDistribution) {
            return RiskDistributionChart(
              riskDistribution: metrics.riskDistribution,
            );
          }
        } else {
          // Narrow layout: stacked
          for (final card in kpiCards) {
            widgets.add(card);
            widgets.add(const SizedBox(height: 12));
          }
          if (showRevenueMix) {
            widgets.add(RevenueMixChart(revenueMix: metrics.revenueMix));
            widgets.add(const SizedBox(height: 16));
          }
          if (showRiskDistribution) {
            widgets.add(
              RiskDistributionChart(
                riskDistribution: metrics.riskDistribution,
              ),
            );
          }

          // Remove trailing spacer if present
          if (widgets.isNotEmpty && widgets.last is SizedBox) {
            widgets.removeLast();
          }

          return Column(children: widgets);
        }

        return const SizedBox.shrink();
      },
    );
  }
}
