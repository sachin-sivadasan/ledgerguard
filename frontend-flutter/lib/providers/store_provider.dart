import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../mock_data/mock_stores.dart';
import '../mock_data/mock_subscriptions.dart';
import '../models/store_model.dart';
import '../models/subscription_model.dart';
import '../services/store_service.dart';
import '../services/subscription_service.dart';

class StoreProvider extends ChangeNotifier {
  final StoreService _storeService;
  final SubscriptionService _subscriptionService;

  bool _demoMode = false;
  bool _isLoading = false;
  bool _isLoadingMore = false;
  String? _error;
  String _searchQuery = '';
  String? _selectedAppId;
  CancelToken? _cancelToken;
  Timer? _searchDebounce;

  List<Store> _liveStores = [];
  final List<Subscription> _liveStoreSubscriptions = [];
  int _currentPage = 1;
  int _totalPages = 1;
  int _totalCount = 0;
  static const int _pageSize = 20;

  // Store-detail state — fetched by domain so it works on a cold deep-link and for
  // any store beyond the loaded list page (SD-1/SD-2), not just page 1.
  Store? _detailStore;
  List<Subscription> _detailSubs = [];
  bool _detailLoading = false;
  String? _detailError;

  Store? get detailStore => _detailStore;
  List<Subscription> get detailSubscriptions => _detailSubs;
  bool get detailLoading => _detailLoading;
  String? get detailError => _detailError;

  StoreProvider(this._storeService, this._subscriptionService);

  bool get demoMode => _demoMode;
  bool get isLoading => _isLoading;
  bool get isLoadingMore => _isLoadingMore;
  String? get error => _error;
  String get searchQuery => _searchQuery;
  String? get selectedAppId => _selectedAppId;
  bool get hasMore => _currentPage < _totalPages;
  int get totalCount => _totalCount;

  void setDemoMode(bool value) {
    _demoMode = value;
    notifyListeners();
  }

  void setSelectedApp(String? appId) {
    _selectedAppId = appId;
    notifyListeners();
    if (!_demoMode && appId != null) {
      loadStores(appId);
    }
  }

  Future<void> loadStores(String appId) async {
    if (_demoMode) return;
    _cancelToken?.cancel('Superseded');
    _cancelToken = CancelToken();
    _isLoading = true;
    _error = null;
    _currentPage = 1;
    _liveStores = [];
    notifyListeners();
    try {
      final result = await _storeService.fetchStores(
        appId,
        page: 1,
        pageSize: _pageSize,
        search: _searchQuery.isNotEmpty ? _searchQuery : null,
        cancelToken: _cancelToken,
      );
      _liveStores = result.items;
      _totalCount = result.total;
      _totalPages = result.totalPages;
      _currentPage = result.page;

      final subsResult = await _subscriptionService.fetchSubscriptions(
        appId,
        page: 1,
        pageSize: 100,
        cancelToken: _cancelToken,
      );
      _liveStoreSubscriptions
        ..clear()
        ..addAll(subsResult.items);
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
      final result = await _storeService.fetchStores(
        _selectedAppId!,
        page: _currentPage + 1,
        pageSize: _pageSize,
        search: _searchQuery.isNotEmpty ? _searchQuery : null,
      );
      _liveStores.addAll(result.items);
      _currentPage = result.page;
      _totalPages = result.totalPages;
      _totalCount = result.total;
    } catch (e) {
      debugPrint('[StoreProvider] loadMore error: $e');
    }
    _isLoadingMore = false;
    notifyListeners();
  }

  List<Store> get stores {
    var list = _demoMode ? mockStores.toList() : _liveStores.toList();

    // Only filter by app in demo mode — live data is already fetched per-app
    if (_selectedAppId != null && _demoMode) {
      list = list
          .where((s) => s.installedAppIds.contains(_selectedAppId))
          .toList();
    }

    // Client-side search only in demo mode — live mode uses server-side search
    if (_searchQuery.isNotEmpty && _demoMode) {
      final q = _searchQuery.toLowerCase();
      list =
          list.where((s) => s.shopDomain.toLowerCase().contains(q)).toList();
    }

    return list;
  }

  Store? getById(String id) {
    try {
      final source = _demoMode ? mockStores : _liveStores;
      return source.firstWhere((s) => s.id == id);
    } catch (_) {
      return null;
    }
  }

  /// Loads a single store (and its subscriptions) BY DOMAIN from the server, so
  /// the detail page resolves on a deep-link / for stores past the loaded list
  /// page. The route param (storeId) is the shop domain. Reuses the server-side
  /// `search` on both endpoints, then keeps the exact-domain match.
  Future<void> loadStoreDetail(String appId, String domain) async {
    if (_demoMode) {
      _detailStore = getById(domain);
      _detailSubs = _detailStore != null
          ? getSubscriptionsForStore(_detailStore!.shopDomain)
          : const [];
      _detailLoading = false;
      _detailError = null;
      notifyListeners();
      return;
    }

    _detailLoading = true;
    _detailError = null;
    _detailStore = null;
    _detailSubs = const [];
    notifyListeners();
    try {
      final storeRes =
          await _storeService.fetchStores(appId, search: domain, pageSize: 5);
      final match = matchStoreByDomain(storeRes.items, domain);
      _detailStore = match;
      if (match != null) {
        final subRes = await _subscriptionService.fetchSubscriptions(appId,
            search: domain, pageSize: 25);
        _detailSubs = subRes.items
            .where((s) => s.shopDomain == match.shopDomain)
            .toList();
      }
    } on DioException catch (e) {
      _detailError = e.message;
    } catch (e) {
      _detailError = e.toString();
    }
    _detailLoading = false;
    notifyListeners();
  }

  /// Picks the exact-domain store from a search result set, else the first hit
  /// (server `search` is a substring match), else null (not found).
  static Store? matchStoreByDomain(List<Store> results, String domain) {
    for (final s in results) {
      if (s.shopDomain == domain) return s;
    }
    return results.isNotEmpty ? results.first : null;
  }

  List<Subscription> getSubscriptionsForStore(String shopDomain) {
    if (_demoMode) {
      return mockSubscriptions
          .where((s) => s.shopDomain == shopDomain)
          .toList();
    }
    return _liveStoreSubscriptions
        .where((s) => s.shopDomain == shopDomain)
        .toList();
  }

  void setSearch(String query) {
    _searchQuery = query;
    notifyListeners();
    if (!_demoMode && _selectedAppId != null) {
      _searchDebounce?.cancel();
      _searchDebounce = Timer(const Duration(milliseconds: 300), () {
        loadStores(_selectedAppId!);
      });
    }
  }
}
