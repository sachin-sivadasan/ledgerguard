import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../domain/entities/subscription_filter.dart';
import '../../../domain/repositories/subscription_repository.dart';
import 'subscription_list_event.dart';
import 'subscription_list_state.dart';

/// Bloc for managing subscription list state
class SubscriptionListBloc
    extends Bloc<SubscriptionListEvent, SubscriptionListState> {
  final SubscriptionRepository _repository;

  String? _currentAppId;
  SubscriptionFilters _currentFilters = const SubscriptionFilters();
  int _currentPage = 1;
  int _currentPageSize = 25;

  // Cache summary and price stats to avoid refetching
  SubscriptionSummary? _cachedSummary;
  PriceStats? _cachedPriceStats;

  SubscriptionListBloc({
    required SubscriptionRepository repository,
  })  : _repository = repository,
        super(const SubscriptionListInitial()) {
    on<LoadSubscriptionsRequested>(_onLoadSubscriptions);
    on<FetchSubscriptionsRequested>(_onFetchSubscriptions);
    on<RefreshSubscriptionsRequested>(_onRefreshSubscriptions);
    on<ApplyFiltersRequested>(_onApplyFilters);
    on<ChangePageRequested>(_onChangePage);
    on<ChangePageSizeRequested>(_onChangePageSize);
    on<ChangeSortRequested>(_onChangeSort);
    on<SearchRequested>(_onSearch);
    on<ClearFiltersRequested>(_onClearFilters);
    on<FilterByRiskStateRequested>(_onFilterByRiskState);
    on<LoadMoreSubscriptionsRequested>(_onLoadMore);
  }

  Future<void> _onLoadSubscriptions(
    LoadSubscriptionsRequested event,
    Emitter<SubscriptionListState> emit,
  ) async {
    _currentAppId = event.appId;
    _currentPage = 1;
    _currentFilters = const SubscriptionFilters();
    emit(const SubscriptionListLoading());

    try {
      // Fetch summary, price stats, and subscriptions in parallel
      final results = await Future.wait([
        _repository.getSummary(event.appId),
        _repository.getPriceStats(event.appId),
        _repository.getSubscriptionsFiltered(
          event.appId,
          filters: _currentFilters,
          page: _currentPage,
          pageSize: _currentPageSize,
        ),
      ]);

      final summary = results[0] as SubscriptionSummary;
      final priceStats = results[1] as PriceStats;
      final response = results[2] as PaginatedSubscriptionResponse;

      _cachedSummary = summary;
      _cachedPriceStats = priceStats;

      if (response.subscriptions.isEmpty) {
        emit(SubscriptionListEmpty(
          appId: event.appId,
          summary: summary,
          priceStats: priceStats,
          filters: _currentFilters,
        ));
      } else {
        emit(SubscriptionListLoaded(
          subscriptions: response.subscriptions,
          summary: summary,
          priceStats: priceStats,
          filters: _currentFilters,
          page: response.page,
          pageSize: response.pageSize,
          total: response.total,
          totalPages: response.totalPages,
          appId: event.appId,
        ));
      }
    } on SubscriptionException catch (e) {
      emit(SubscriptionListError(e.message));
    } catch (e) {
      emit(SubscriptionListError('Failed to load subscriptions: $e'));
    }
  }

  /// Legacy support for FetchSubscriptionsRequested
  Future<void> _onFetchSubscriptions(
    FetchSubscriptionsRequested event,
    Emitter<SubscriptionListState> emit,
  ) async {
    add(LoadSubscriptionsRequested(appId: event.appId));
  }

  Future<void> _onRefreshSubscriptions(
    RefreshSubscriptionsRequested event,
    Emitter<SubscriptionListState> emit,
  ) async {
    final currentState = state;
    if (_currentAppId == null) return;

    if (currentState is SubscriptionListLoaded) {
      emit(currentState.copyWith(isRefreshing: true));
    }

    try {
      // Refetch summary and price stats
      final results = await Future.wait([
        _repository.getSummary(_currentAppId!),
        _repository.getPriceStats(_currentAppId!),
        _repository.getSubscriptionsFiltered(
          _currentAppId!,
          filters: _currentFilters,
          page: _currentPage,
          pageSize: _currentPageSize,
        ),
      ]);

      final summary = results[0] as SubscriptionSummary;
      final priceStats = results[1] as PriceStats;
      final response = results[2] as PaginatedSubscriptionResponse;

      _cachedSummary = summary;
      _cachedPriceStats = priceStats;

      if (response.subscriptions.isEmpty) {
        emit(SubscriptionListEmpty(
          appId: _currentAppId!,
          summary: summary,
          priceStats: priceStats,
          filters: _currentFilters,
        ));
      } else {
        emit(SubscriptionListLoaded(
          subscriptions: response.subscriptions,
          summary: summary,
          priceStats: priceStats,
          filters: _currentFilters,
          page: response.page,
          pageSize: response.pageSize,
          total: response.total,
          totalPages: response.totalPages,
          appId: _currentAppId!,
        ));
      }
    } catch (e) {
      if (currentState is SubscriptionListLoaded) {
        emit(currentState.copyWith(isRefreshing: false));
      }
    }
  }

  Future<void> _onApplyFilters(
    ApplyFiltersRequested event,
    Emitter<SubscriptionListState> emit,
  ) async {
    if (_currentAppId == null) return;

    _currentFilters = event.filters;
    _currentPage = 1; // Reset to first page when filters change

    await _fetchSubscriptions(emit);
  }

  Future<void> _onChangePage(
    ChangePageRequested event,
    Emitter<SubscriptionListState> emit,
  ) async {
    if (_currentAppId == null) return;

    _currentPage = event.page;
    await _fetchSubscriptions(emit);
  }

  Future<void> _onChangePageSize(
    ChangePageSizeRequested event,
    Emitter<SubscriptionListState> emit,
  ) async {
    if (_currentAppId == null) return;

    _currentPageSize = event.pageSize;
    _currentPage = 1; // Reset to first page when page size changes
    await _fetchSubscriptions(emit);
  }

  Future<void> _onChangeSort(
    ChangeSortRequested event,
    Emitter<SubscriptionListState> emit,
  ) async {
    if (_currentAppId == null) return;

    _currentFilters = _currentFilters.copyWith(
      sort: event.sort,
      sortAscending: event.ascending,
    );
    _currentPage = 1; // Reset to first page when sort changes
    await _fetchSubscriptions(emit);
  }

  Future<void> _onSearch(
    SearchRequested event,
    Emitter<SubscriptionListState> emit,
  ) async {
    if (_currentAppId == null) return;

    _currentFilters = _currentFilters.copyWith(
      searchQuery: event.query.isEmpty ? null : event.query,
      clearSearchQuery: event.query.isEmpty,
    );
    _currentPage = 1; // Reset to first page when search changes
    await _fetchSubscriptions(emit);
  }

  Future<void> _onClearFilters(
    ClearFiltersRequested event,
    Emitter<SubscriptionListState> emit,
  ) async {
    if (_currentAppId == null) return;

    _currentFilters = const SubscriptionFilters();
    _currentPage = 1;
    await _fetchSubscriptions(emit);
  }

  /// Legacy support for FilterByRiskStateRequested
  Future<void> _onFilterByRiskState(
    FilterByRiskStateRequested event,
    Emitter<SubscriptionListState> emit,
  ) async {
    if (_currentAppId == null) return;

    if (event.riskState == null) {
      _currentFilters = _currentFilters.copyWith(riskStates: {});
    } else {
      _currentFilters = _currentFilters.copyWith(
        riskStates: {event.riskState!},
      );
    }
    _currentPage = 1;
    await _fetchSubscriptions(emit);
  }

  /// Legacy support for LoadMoreSubscriptionsRequested
  Future<void> _onLoadMore(
    LoadMoreSubscriptionsRequested event,
    Emitter<SubscriptionListState> emit,
  ) async {
    final currentState = state;
    if (currentState is SubscriptionListLoaded &&
        currentState.hasMore &&
        !currentState.isLoading) {
      _currentPage = currentState.page + 1;
      await _fetchSubscriptions(emit);
    }
  }

  /// Helper to fetch subscriptions with current filters and page
  Future<void> _fetchSubscriptions(
    Emitter<SubscriptionListState> emit,
  ) async {
    final currentState = state;
    if (currentState is SubscriptionListLoaded) {
      emit(currentState.copyWith(isLoading: true));
    } else {
      emit(const SubscriptionListLoading());
    }

    try {
      final response = await _repository.getSubscriptionsFiltered(
        _currentAppId!,
        filters: _currentFilters,
        page: _currentPage,
        pageSize: _currentPageSize,
      );

      if (response.subscriptions.isEmpty && _currentPage == 1) {
        emit(SubscriptionListEmpty(
          appId: _currentAppId!,
          summary: _cachedSummary,
          priceStats: _cachedPriceStats,
          filters: _currentFilters,
        ));
      } else {
        emit(SubscriptionListLoaded(
          subscriptions: response.subscriptions,
          summary: _cachedSummary ??
              const SubscriptionSummary(
                activeCount: 0,
                atRiskCount: 0,
                churnedCount: 0,
                avgPriceCents: 0,
                totalCount: 0,
              ),
          priceStats: _cachedPriceStats,
          filters: _currentFilters,
          page: response.page,
          pageSize: response.pageSize,
          total: response.total,
          totalPages: response.totalPages,
          appId: _currentAppId!,
        ));
      }
    } on SubscriptionException catch (e) {
      emit(SubscriptionListError(e.message));
    } catch (e) {
      emit(SubscriptionListError('Failed to load subscriptions: $e'));
    }
  }
}
