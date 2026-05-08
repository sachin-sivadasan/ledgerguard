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
  String? _error;
  String _searchQuery = '';
  String? _selectedAppId;
  CancelToken? _cancelToken;

  List<Store> _liveStores = [];
  final List<Subscription> _liveStoreSubscriptions = [];

  StoreProvider(this._storeService, this._subscriptionService);

  bool get demoMode => _demoMode;
  bool get isLoading => _isLoading;
  String? get error => _error;
  String get searchQuery => _searchQuery;
  String? get selectedAppId => _selectedAppId;

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
    notifyListeners();
    try {
      _liveStores = await _storeService.fetchStores(appId,
          cancelToken: _cancelToken);
      final subs = await _subscriptionService.fetchSubscriptions(appId,
          cancelToken: _cancelToken);
      _liveStoreSubscriptions
        ..clear()
        ..addAll(subs);
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) return;
      _error = e.message;
    } catch (e) {
      _error = e.toString();
    }
    _isLoading = false;
    notifyListeners();
  }

  List<Store> get stores {
    var list = _demoMode ? mockStores.toList() : _liveStores.toList();

    if (_selectedAppId != null) {
      list = list
          .where((s) => s.installedAppIds.contains(_selectedAppId))
          .toList();
    }

    if (_searchQuery.isNotEmpty) {
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
  }
}
