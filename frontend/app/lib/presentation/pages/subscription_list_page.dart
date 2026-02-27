import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../domain/entities/subscription.dart';
import '../blocs/subscription_list/subscription_list.dart';
import '../widgets/empty_state_widget.dart';
import '../widgets/error_state_widget.dart';
import '../widgets/risk_badge.dart';
import '../widgets/subscription_tile.dart';

/// Page displaying list of subscriptions for an app
class SubscriptionListPage extends StatelessWidget {
  final String appId;

  const SubscriptionListPage({
    super.key,
    required this.appId,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey[50],
      appBar: AppBar(
        title: const Text('Subscriptions'),
        actions: [
          BlocBuilder<SubscriptionListBloc, SubscriptionListState>(
            builder: (context, state) {
              final isRefreshing =
                  state is SubscriptionListLoaded && state.isRefreshing;
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
                    : () => context
                        .read<SubscriptionListBloc>()
                        .add(const RefreshSubscriptionsRequested()),
              );
            },
          ),
        ],
      ),
      body: BlocBuilder<SubscriptionListBloc, SubscriptionListState>(
        builder: (context, state) {
          if (state is SubscriptionListInitial) {
            context
                .read<SubscriptionListBloc>()
                .add(FetchSubscriptionsRequested(appId: appId));
            return const Center(child: CircularProgressIndicator());
          }

          if (state is SubscriptionListLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state is SubscriptionListEmpty) {
            return EmptyStateWidget(
              title: 'No Subscriptions',
              message: 'No subscriptions found for this app',
              icon: Icons.subscriptions_outlined,
            );
          }

          if (state is SubscriptionListError) {
            return ErrorStateWidget(
              title: 'Failed to load subscriptions',
              message: state.message,
              onRetry: () => context
                  .read<SubscriptionListBloc>()
                  .add(FetchSubscriptionsRequested(appId: appId)),
            );
          }

          if (state is SubscriptionListLoaded) {
            return _buildContent(context, state);
          }

          return const SizedBox.shrink();
        },
      ),
    );
  }

  Widget _buildContent(BuildContext context, SubscriptionListLoaded state) {
    return LayoutBuilder(
      builder: (context, constraints) {
        // Responsive padding
        final listPadding = constraints.maxWidth < 600 ? 12.0 : 16.0;

        return Column(
          children: [
            // Filter bar
            _buildFilterBar(context, state),
            // Subscription list
            Expanded(
              child: RefreshIndicator(
                onRefresh: () async {
                  context
                      .read<SubscriptionListBloc>()
                      .add(const RefreshSubscriptionsRequested());
                  // Wait for the state to change
                  await context
                      .read<SubscriptionListBloc>()
                      .stream
                      .firstWhere((s) =>
                          s is SubscriptionListLoaded && !s.isRefreshing ||
                          s is SubscriptionListError);
                },
                child: ListView.builder(
                  padding: EdgeInsets.all(listPadding),
                  itemCount: state.subscriptions.length + (state.hasMore ? 1 : 0),
                  itemBuilder: (context, index) {
                    if (index == state.subscriptions.length) {
                      // Load more indicator
                      _loadMore(context, state);
                      return const Padding(
                        padding: EdgeInsets.all(16),
                        child: Center(child: CircularProgressIndicator()),
                      );
                    }

                    final subscription = state.subscriptions[index];
                    return Padding(
                      padding: const EdgeInsets.only(bottom: 10),
                      child: SubscriptionTile(
                        subscription: subscription,
                        onTap: () => _navigateToDetail(context, subscription),
                      ),
                    );
                  },
                ),
              ),
            ),
          ],
        );
      },
    );
  }

  Widget _buildFilterBar(BuildContext context, SubscriptionListLoaded state) {
    final screenWidth = MediaQuery.of(context).size.width;
    final isCompact = screenWidth < 400;

    return Container(
      padding: EdgeInsets.symmetric(
        horizontal: isCompact ? 12 : 16,
        vertical: isCompact ? 8 : 12,
      ),
      color: Colors.white,
      child: Row(
        children: [
          Text(
            isCompact ? '${state.total}' : '${state.total} subscriptions',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: Colors.grey[600],
                  fontSize: isCompact ? 13 : null,
                ),
          ),
          const Spacer(),
          // Filter dropdown
          PopupMenuButton<RiskState?>(
            initialValue: state.filterRiskState,
            onSelected: (riskState) {
              context
                  .read<SubscriptionListBloc>()
                  .add(FilterByRiskStateRequested(riskState: riskState));
            },
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              decoration: BoxDecoration(
                color: state.filterRiskState != null
                    ? Colors.blue.withOpacity(0.1)
                    : Colors.grey[100],
                borderRadius: BorderRadius.circular(8),
                border: Border.all(
                  color: state.filterRiskState != null
                      ? Colors.blue.withOpacity(0.3)
                      : Colors.grey[300]!,
                ),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(
                    Icons.filter_list,
                    size: 18,
                    color: state.filterRiskState != null
                        ? Colors.blue
                        : Colors.grey[600],
                  ),
                  const SizedBox(width: 6),
                  Text(
                    state.filterRiskState?.displayName ?? 'All',
                    style: TextStyle(
                      color: state.filterRiskState != null
                          ? Colors.blue
                          : Colors.grey[700],
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                  const SizedBox(width: 4),
                  Icon(
                    Icons.arrow_drop_down,
                    size: 20,
                    color: state.filterRiskState != null
                        ? Colors.blue
                        : Colors.grey[600],
                  ),
                ],
              ),
            ),
            itemBuilder: (context) => [
              const PopupMenuItem<RiskState?>(
                value: null,
                child: Text('All'),
              ),
              PopupMenuItem<RiskState>(
                value: RiskState.safe,
                child: Row(
                  children: [
                    RiskBadge(riskState: RiskState.safe, isCompact: true),
                    const SizedBox(width: 8),
                    const Text('Safe'),
                  ],
                ),
              ),
              PopupMenuItem<RiskState>(
                value: RiskState.oneCycleMissed,
                child: Row(
                  children: [
                    RiskBadge(riskState: RiskState.oneCycleMissed, isCompact: true),
                    const SizedBox(width: 8),
                    const Text('At Risk'),
                  ],
                ),
              ),
              PopupMenuItem<RiskState>(
                value: RiskState.twoCyclesMissed,
                child: Row(
                  children: [
                    RiskBadge(riskState: RiskState.twoCyclesMissed, isCompact: true),
                    const SizedBox(width: 8),
                    const Text('High Risk'),
                  ],
                ),
              ),
              PopupMenuItem<RiskState>(
                value: RiskState.churned,
                child: Row(
                  children: [
                    RiskBadge(riskState: RiskState.churned, isCompact: true),
                    const SizedBox(width: 8),
                    const Text('Churned'),
                  ],
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  void _loadMore(BuildContext context, SubscriptionListLoaded state) {
    if (!state.isLoadingMore) {
      context
          .read<SubscriptionListBloc>()
          .add(const LoadMoreSubscriptionsRequested());
    }
  }

  void _navigateToDetail(BuildContext context, Subscription subscription) {
    context.push('/apps/$appId/subscriptions/${subscription.id}');
  }
}
