import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../domain/entities/subscription.dart';
import '../../domain/entities/subscription_filter.dart';
import '../blocs/subscription_list/subscription_list.dart';
import '../widgets/empty_state_widget.dart';
import '../widgets/error_state_widget.dart';
import '../widgets/pagination_controls.dart';
import '../widgets/subscription_filter_bar.dart';
import '../widgets/subscription_summary_bar.dart';
import '../widgets/subscription_tile.dart';

/// Page displaying list of subscriptions for an app with advanced filtering
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
                .add(LoadSubscriptionsRequested(appId: appId));
            return const Center(child: CircularProgressIndicator());
          }

          if (state is SubscriptionListLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state is SubscriptionListEmpty) {
            return _buildEmptyState(context, state);
          }

          if (state is SubscriptionListError) {
            return ErrorStateWidget(
              title: 'Failed to load subscriptions',
              message: state.message,
              onRetry: () => context
                  .read<SubscriptionListBloc>()
                  .add(LoadSubscriptionsRequested(appId: appId)),
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

  Widget _buildEmptyState(BuildContext context, SubscriptionListEmpty state) {
    return Column(
      children: [
        // Show summary bar if we have summary data
        if (state.summary != null)
          SubscriptionSummaryBar(summary: state.summary!),
        // Show filter bar if filters are active
        if (state.filters.hasActiveFilters)
          SubscriptionFilterBar(
            filters: state.filters,
            priceStats: state.priceStats,
            onFiltersChanged: (filters) => context
                .read<SubscriptionListBloc>()
                .add(ApplyFiltersRequested(filters)),
            onClearFilters: () => context
                .read<SubscriptionListBloc>()
                .add(const ClearFiltersRequested()),
          ),
        Expanded(
          child: EmptyStateWidget(
            title: state.filters.hasActiveFilters
                ? 'No Matching Subscriptions'
                : 'No Subscriptions',
            message: state.filters.hasActiveFilters
                ? 'Try adjusting your filters'
                : 'No subscriptions found for this app',
            icon: Icons.subscriptions_outlined,
            actionLabel: state.filters.hasActiveFilters ? 'Clear Filters' : null,
            actionIcon: state.filters.hasActiveFilters ? Icons.clear_all : null,
            onAction: state.filters.hasActiveFilters
                ? () => context
                    .read<SubscriptionListBloc>()
                    .add(const ClearFiltersRequested())
                : null,
          ),
        ),
      ],
    );
  }

  Widget _buildContent(BuildContext context, SubscriptionListLoaded state) {
    return Column(
      children: [
        // Summary bar
        SubscriptionSummaryBar(
          summary: state.summary,
          isLoading: state.isLoading,
        ),
        // Filter bar
        SubscriptionFilterBar(
          filters: state.filters,
          priceStats: state.priceStats,
          onFiltersChanged: (filters) => context
              .read<SubscriptionListBloc>()
              .add(ApplyFiltersRequested(filters)),
          onClearFilters: () => context
              .read<SubscriptionListBloc>()
              .add(const ClearFiltersRequested()),
          isLoading: state.isLoading,
        ),
        // Table header
        _SubscriptionTableHeader(
          sort: state.filters.sort,
          sortAscending: state.filters.sortAscending,
          onSortChanged: (sort, ascending) => context
              .read<SubscriptionListBloc>()
              .add(ChangeSortRequested(sort, ascending: ascending)),
          isLoading: state.isLoading,
        ),
        // Subscription list
        Expanded(
          child: Stack(
            children: [
              RefreshIndicator(
                onRefresh: () async {
                  context
                      .read<SubscriptionListBloc>()
                      .add(const RefreshSubscriptionsRequested());
                  await context
                      .read<SubscriptionListBloc>()
                      .stream
                      .firstWhere((s) =>
                          s is SubscriptionListLoaded && !s.isRefreshing ||
                          s is SubscriptionListError);
                },
                child: ListView.builder(
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                  itemCount: state.subscriptions.length,
                  itemBuilder: (context, index) {
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
              // Loading overlay
              if (state.isLoading && !state.isRefreshing)
                Container(
                  color: Colors.white.withOpacity(0.7),
                  child: const Center(
                    child: CircularProgressIndicator(),
                  ),
                ),
            ],
          ),
        ),
        // Pagination controls
        PaginationControls(
          page: state.page,
          pageSize: state.pageSize,
          totalPages: state.totalPages,
          total: state.total,
          onPageChanged: (page) => context
              .read<SubscriptionListBloc>()
              .add(ChangePageRequested(page)),
          onPageSizeChanged: (pageSize) => context
              .read<SubscriptionListBloc>()
              .add(ChangePageSizeRequested(pageSize)),
          isLoading: state.isLoading,
        ),
      ],
    );
  }

  void _navigateToDetail(BuildContext context, Subscription subscription) {
    context.push('/apps/$appId/subscriptions/${subscription.id}');
  }
}

/// Sortable table header for subscription list
class _SubscriptionTableHeader extends StatelessWidget {
  final SubscriptionSort sort;
  final bool sortAscending;
  final void Function(SubscriptionSort sort, bool ascending) onSortChanged;
  final bool isLoading;

  const _SubscriptionTableHeader({
    required this.sort,
    required this.sortAscending,
    required this.onSortChanged,
    this.isLoading = false,
  });

  void _onTap(SubscriptionSort column) {
    if (isLoading) return;
    if (sort == column) {
      onSortChanged(column, !sortAscending);
    } else {
      onSortChanged(column, true);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      decoration: BoxDecoration(
        color: Colors.grey[100],
        border: Border(
          bottom: BorderSide(color: Colors.grey[300]!),
        ),
      ),
      child: Row(
        children: [
          Expanded(
            flex: 3,
            child: _SortableHeader(
              label: 'Shop',
              sort: SubscriptionSort.shopName,
              currentSort: sort,
              sortAscending: sortAscending,
              onTap: () => _onTap(SubscriptionSort.shopName),
            ),
          ),
          Expanded(
            flex: 2,
            child: _SortableHeader(
              label: 'Price',
              sort: SubscriptionSort.price,
              currentSort: sort,
              sortAscending: sortAscending,
              onTap: () => _onTap(SubscriptionSort.price),
            ),
          ),
          Expanded(
            flex: 2,
            child: _SortableHeader(
              label: 'Risk',
              sort: SubscriptionSort.riskState,
              currentSort: sort,
              sortAscending: sortAscending,
              onTap: () => _onTap(SubscriptionSort.riskState),
            ),
          ),
        ],
      ),
    );
  }
}

class _SortableHeader extends StatelessWidget {
  final String label;
  final SubscriptionSort sort;
  final SubscriptionSort currentSort;
  final bool sortAscending;
  final VoidCallback onTap;

  const _SortableHeader({
    required this.label,
    required this.sort,
    required this.currentSort,
    required this.sortAscending,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final isActive = sort == currentSort;

    return InkWell(
      onTap: onTap,
      child: Row(
        children: [
          Text(
            label,
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  fontWeight: isActive ? FontWeight.bold : FontWeight.w500,
                  color: isActive ? Colors.blue : Colors.grey[700],
                ),
          ),
          if (isActive) ...[
            const SizedBox(width: 4),
            Icon(
              sortAscending ? Icons.arrow_upward : Icons.arrow_downward,
              size: 14,
              color: Colors.blue,
            ),
          ],
        ],
      ),
    );
  }
}
