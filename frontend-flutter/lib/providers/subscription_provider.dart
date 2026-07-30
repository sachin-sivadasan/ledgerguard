import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../mock_data/mock_payment_history.dart';
import '../mock_data/mock_risk_timeline.dart';
import '../mock_data/mock_subscriptions.dart';
import '../models/payment_history_model.dart';
import '../models/risk_timeline_model.dart';
import '../models/subscription_model.dart';
import '../services/subscription_service.dart';
import '../widgets/lg_risk_badge.dart';
import '../widgets/lg_status_badge.dart';

class SubscriptionProvider extends ChangeNotifier {
  final SubscriptionService _subscriptionService;

  bool _demoMode = false;
  bool _isLoading = false;
  bool _isLoadingMore = false;
  String? _error;
  CancelToken? _cancelToken;
  Timer? _searchDebounce;

  String _searchQuery = '';
  SubscriptionStatus? _statusFilter;
  RiskState? _riskFilter;
  String? _planFilter;
  String? _selectedAppId;

  List<Subscription> _liveSubscriptions = [];
  int _currentPage = 1;
  int _totalPages = 1;
  int _totalCount = 0;
  // Server-side KPI totals over ALL subscriptions (live mode). Null until loaded / in
  // demo mode, where the counts are computed from the local list instead.
  SubscriptionSummary? _summary;
  static const int _pageSize = 25;

  SubscriptionProvider(this._subscriptionService);

  bool get demoMode => _demoMode;
  bool get isLoading => _isLoading;
  bool get isLoadingMore => _isLoadingMore;
  String? get error => _error;
  String get searchQuery => _searchQuery;
  SubscriptionStatus? get statusFilter => _statusFilter;
  RiskState? get riskFilter => _riskFilter;
  String? get planFilter => _planFilter;
  String? get selectedAppId => _selectedAppId;
  bool get hasMore => _currentPage < _totalPages;
  int get totalCount => _totalCount;

  void setDemoMode(bool value) {
    _demoMode = value;
    notifyListeners();
  }

  Future<void> loadSubscriptions(String appId) async {
    if (_demoMode) return;
    _cancelToken?.cancel('Superseded');
    _cancelToken = CancelToken();
    _isLoading = true;
    _error = null;
    _currentPage = 1;
    _liveSubscriptions = [];
    _summary = null;
    notifyListeners();
    try {
      final result = await _subscriptionService.fetchSubscriptions(
        appId,
        page: 1,
        pageSize: _pageSize,
        search: _searchQuery.isNotEmpty ? _searchQuery : null,
        riskState: _riskFilter != null
            ? Subscription.riskStateToApi(_riskFilter!)
            : null,
        status: _statusFilter != null
            ? Subscription.statusToApi(_statusFilter!)
            : null,
        plan: _planFilter,
        cancelToken: _cancelToken,
      );
      _liveSubscriptions = result.items;
      _totalCount = result.total;
      _totalPages = result.totalPages;
      _currentPage = result.page;
      // KPI counts come from a server-side aggregate over ALL subscriptions, not the
      // loaded page (returns null on failure → KPIs fall back to 0 rather than a
      // page-scoped undercount).
      _summary = await _subscriptionService.fetchSummary(
        appId,
        cancelToken: _cancelToken,
      );
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) return;
      _error = e.message;
    } catch (e) {
      _error = e.toString();
    }
    _isLoading = false;
    notifyListeners();
  }

  Future<void> loadMore() async {
    if (_demoMode || _isLoadingMore || !hasMore || _selectedAppId == null) {
      return;
    }
    _isLoadingMore = true;
    notifyListeners();
    try {
      final result = await _subscriptionService.fetchSubscriptions(
        _selectedAppId!,
        page: _currentPage + 1,
        pageSize: _pageSize,
        search: _searchQuery.isNotEmpty ? _searchQuery : null,
        riskState: _riskFilter != null
            ? Subscription.riskStateToApi(_riskFilter!)
            : null,
        status: _statusFilter != null
            ? Subscription.statusToApi(_statusFilter!)
            : null,
        plan: _planFilter,
      );
      _liveSubscriptions.addAll(result.items);
      _currentPage = result.page;
      _totalPages = result.totalPages;
      _totalCount = result.total;
    } catch (e) {
      debugPrint('[SubscriptionProvider] loadMore error: $e');
    }
    _isLoadingMore = false;
    notifyListeners();
  }

  List<Subscription> get _allSubscriptions =>
      _demoMode ? mockSubscriptions : _liveSubscriptions;

  List<Subscription> get subscriptions {
    var list = _allSubscriptions.toList();

    // Only filter by app in demo mode — live data is already fetched per-app
    if (_selectedAppId != null && _demoMode) {
      list = list.where((s) => s.appId == _selectedAppId).toList();
    }

    // Client-side search only in demo mode — live mode uses server-side search
    if (_searchQuery.isNotEmpty && _demoMode) {
      final q = _searchQuery.toLowerCase();
      list = list
          .where((s) =>
              s.shopDomain.toLowerCase().contains(q) ||
              s.planName.toLowerCase().contains(q))
          .toList();
    }

    // Live mode filters server-side (see loadSubscriptions); only demo mode filters
    // the local mock list here.
    if (_demoMode) {
      if (_statusFilter != null) {
        list = list.where((s) => s.status == _statusFilter).toList();
      }
      if (_riskFilter != null) {
        list = list.where((s) => s.riskState == _riskFilter).toList();
      }
      if (_planFilter != null) {
        list = list.where((s) => s.planName == _planFilter).toList();
      }
    }

    return list;
  }

  /// Distinct non-empty plan names for the Plan filter dropdown (from loaded rows;
  /// empty for apps whose subscriptions carry no plan name).
  List<String> get availablePlans {
    final names = _allSubscriptions
        .map((s) => s.planName)
        .where((p) => p.isNotEmpty)
        .toSet()
        .toList()
      ..sort();
    return names;
  }

  Subscription? getById(String id) {
    try {
      return _allSubscriptions.firstWhere((s) => s.id == id);
    } catch (_) {
      return null;
    }
  }

  List<PaymentHistoryEntry> getPaymentHistory(String subscriptionId) {
    return generatePaymentHistory(subscriptionId);
  }

  List<RiskTimelineEntry> getRiskTimeline(String subscriptionId) {
    return generateRiskTimeline(subscriptionId);
  }

  List<Subscription> get _appFilteredSubs {
    if (_selectedAppId == null || !_demoMode) return _allSubscriptions;
    return _allSubscriptions.where((s) => s.appId == _selectedAppId).toList();
  }

  // KPI counts: live mode uses the server-side summary (all subscriptions); demo mode
  // computes from the local mock list. This keeps the cards accurate regardless of how
  // many pages are loaded. (Live "active" = SAFE risk_state, matching the server's
  // SAFE/at-risk/churned partition of the total.)
  int get activeCount => _demoMode
      ? _appFilteredSubs
          .where((s) => s.status == SubscriptionStatus.active)
          .length
      : (_summary?.activeCount ?? 0);

  int get atRiskCount => _demoMode
      ? _appFilteredSubs
          .where((s) =>
              s.riskState == RiskState.oneCycleMissed ||
              s.riskState == RiskState.twoCycleMissed)
          .length
      : (_summary?.atRiskCount ?? 0);

  int get churnedCount => _demoMode
      ? _appFilteredSubs.where((s) => s.riskState == RiskState.churned).length
      : (_summary?.churnedCount ?? 0);

  String get avgPrice {
    if (!_demoMode) {
      return '\$${((_summary?.avgPriceCents ?? 0) / 100).toStringAsFixed(2)}';
    }
    final subs = _appFilteredSubs;
    if (subs.isEmpty) return '\$0.00';
    final avg =
        subs.fold<int>(0, (sum, s) => sum + s.priceCents) / subs.length;
    return '\$${(avg / 100).toStringAsFixed(2)}';
  }

  void setSelectedApp(String? appId) {
    _selectedAppId = appId;
    notifyListeners();
    if (!_demoMode && appId != null) {
      loadSubscriptions(appId);
    }
  }

  void setSearch(String query) {
    _searchQuery = query;
    notifyListeners();
    if (!_demoMode && _selectedAppId != null) {
      _searchDebounce?.cancel();
      _searchDebounce = Timer(const Duration(milliseconds: 300), () {
        loadSubscriptions(_selectedAppId!);
      });
    }
  }

  void setStatusFilter(SubscriptionStatus? status) {
    _statusFilter = status;
    notifyListeners();
    _reloadForFilterChange();
  }

  void setRiskFilter(RiskState? risk) {
    _riskFilter = risk;
    notifyListeners();
    _reloadForFilterChange();
  }

  void setPlanFilter(String? plan) {
    _planFilter = plan;
    notifyListeners();
    _reloadForFilterChange();
  }

  void clearFilters() {
    _searchQuery = '';
    _statusFilter = null;
    _riskFilter = null;
    _planFilter = null;
    notifyListeners();
    _reloadForFilterChange();
  }

  // Live mode re-queries the server with the active filters (server-side filtering over
  // ALL subscriptions, not just the loaded page). Demo mode filters the local list.
  void _reloadForFilterChange() {
    if (!_demoMode && _selectedAppId != null) {
      loadSubscriptions(_selectedAppId!);
    }
  }
}
