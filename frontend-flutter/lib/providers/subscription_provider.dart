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
    notifyListeners();
    try {
      final result = await _subscriptionService.fetchSubscriptions(
        appId,
        page: 1,
        pageSize: _pageSize,
        search: _searchQuery.isNotEmpty ? _searchQuery : null,
        cancelToken: _cancelToken,
      );
      _liveSubscriptions = result.items;
      _totalCount = result.total;
      _totalPages = result.totalPages;
      _currentPage = result.page;
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

    if (_statusFilter != null) {
      list = list.where((s) => s.status == _statusFilter).toList();
    }

    if (_riskFilter != null) {
      list = list.where((s) => s.riskState == _riskFilter).toList();
    }

    if (_planFilter != null) {
      list = list.where((s) => s.planName == _planFilter).toList();
    }

    return list;
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

  int get activeCount =>
      _appFilteredSubs
          .where((s) => s.status == SubscriptionStatus.active)
          .length;

  int get atRiskCount => _appFilteredSubs
      .where((s) =>
          s.riskState == RiskState.oneCycleMissed ||
          s.riskState == RiskState.twoCycleMissed)
      .length;

  int get churnedCount =>
      _appFilteredSubs.where((s) => s.riskState == RiskState.churned).length;

  String get avgPrice {
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
  }

  void setRiskFilter(RiskState? risk) {
    _riskFilter = risk;
    notifyListeners();
  }

  void setPlanFilter(String? plan) {
    _planFilter = plan;
    notifyListeners();
  }

  void clearFilters() {
    _searchQuery = '';
    _statusFilter = null;
    _riskFilter = null;
    _planFilter = null;
    notifyListeners();
  }
}
