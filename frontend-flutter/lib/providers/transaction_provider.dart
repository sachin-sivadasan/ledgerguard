import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../mock_data/mock_transactions.dart';
import '../models/transaction_model.dart';
import '../services/transaction_service.dart';

class TransactionProvider extends ChangeNotifier {
  final TransactionService _transactionService;

  bool _demoMode = false;
  bool _isLoading = false;
  bool _isLoadingMore = false;
  String? _error;
  CancelToken? _cancelToken;

  ChargeType? _typeFilter;
  String? _selectedAppId;
  String? _storeFilter;

  List<Transaction> _liveTransactions = [];
  int _currentPage = 1;
  int _totalPages = 1;
  int _totalCount = 0;
  static const int _pageSize = 20;

  TransactionProvider(this._transactionService);

  bool get demoMode => _demoMode;
  bool get isLoading => _isLoading;
  bool get isLoadingMore => _isLoadingMore;
  String? get error => _error;
  ChargeType? get typeFilter => _typeFilter;
  String? get selectedAppId => _selectedAppId;
  String? get storeFilter => _storeFilter;
  bool get hasMore => _currentPage < _totalPages;
  int get totalCount => _totalCount;
  int get currentPage => _currentPage;

  void setDemoMode(bool value) {
    _demoMode = value;
    notifyListeners();
  }

  Future<void> loadTransactions(String appId) async {
    if (_demoMode) return;
    _cancelToken?.cancel('Superseded');
    _cancelToken = CancelToken();
    _isLoading = true;
    _error = null;
    _currentPage = 1;
    _liveTransactions = [];
    notifyListeners();
    try {
      final result = await _transactionService.fetchTransactions(
        appId,
        page: 1,
        pageSize: _pageSize,
        cancelToken: _cancelToken,
      );
      _liveTransactions = result.items;
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
      final result = await _transactionService.fetchTransactions(
        _selectedAppId!,
        page: _currentPage + 1,
        pageSize: _pageSize,
      );
      _liveTransactions.addAll(result.items);
      _currentPage = result.page;
      _totalPages = result.totalPages;
      _totalCount = result.total;
    } catch (e) {
      debugPrint('[TransactionProvider] loadMore error: $e');
    }
    _isLoadingMore = false;
    notifyListeners();
  }

  List<Transaction> get transactions {
    var list =
        _demoMode ? mockTransactions.toList() : _liveTransactions.toList();

    if (_typeFilter != null) {
      list = list.where((t) => t.chargeType == _typeFilter).toList();
    }
    // Only filter by app in demo mode — live data is already fetched per-app
    if (_selectedAppId != null && _demoMode) {
      list = list.where((t) => t.appId == _selectedAppId).toList();
    }
    if (_storeFilter != null) {
      list = list
          .where((t) => t.shopDomain.contains(_storeFilter!))
          .toList();
    }

    list.sort((a, b) => b.date.compareTo(a.date));
    return list;
  }

  int get totalGrossCents =>
      transactions.fold<int>(0, (sum, t) => sum + t.grossAmountCents);
  int get totalNetCents =>
      transactions.fold<int>(0, (sum, t) => sum + t.netAmountCents);

  void setTypeFilter(ChargeType? type) {
    _typeFilter = type;
    if (!_demoMode && _selectedAppId != null) {
      loadTransactions(_selectedAppId!);
    }
    notifyListeners();
  }

  void setSelectedApp(String? appId) {
    _selectedAppId = appId;
    notifyListeners();
    if (!_demoMode && appId != null) {
      loadTransactions(appId);
    }
  }

  void setStoreFilter(String? store) {
    _storeFilter = store;
    notifyListeners();
  }

  void clearFilters() {
    _typeFilter = null;
    _storeFilter = null;
    if (!_demoMode && _selectedAppId != null) {
      loadTransactions(_selectedAppId!);
    }
    notifyListeners();
  }
}
