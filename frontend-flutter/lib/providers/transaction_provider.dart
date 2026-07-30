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
  // Server-side aggregate over the FULL filtered set (live mode). Null until loaded.
  TransactionSummary? _summary;
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
    _summary = null;
    notifyListeners();
    final chargeType =
        _typeFilter != null ? chargeTypeToApi(_typeFilter!) : null;
    try {
      final result = await _transactionService.fetchTransactions(
        appId,
        page: 1,
        pageSize: _pageSize,
        chargeType: chargeType,
        cancelToken: _cancelToken,
      );
      _liveTransactions = result.items;
      _totalCount = result.total;
      _totalPages = result.totalPages;
      _currentPage = result.page;
      // Totals must reflect the whole dataset, not just the first page —
      // fetch the server-side aggregate alongside the first page.
      _summary = await _transactionService.fetchSummary(
        appId,
        chargeType: chargeType,
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
      final result = await _transactionService.fetchTransactions(
        _selectedAppId!,
        page: _currentPage + 1,
        pageSize: _pageSize,
        chargeType:
            _typeFilter != null ? chargeTypeToApi(_typeFilter!) : null,
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

    // Live data is already type-filtered server-side; only filter client-side in demo mode.
    if (_typeFilter != null && _demoMode) {
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

  // Totals come from the server-side aggregate over the entire filtered dataset
  // in live mode. Fall back to folding over the visible rows when there is no
  // server summary yet, in demo mode, or when a client-only store filter is
  // active (store isn't sent to the server, so the aggregate would ignore it and
  // disagree with the narrowed list).
  bool get _useServerTotals =>
      !_demoMode && _summary != null && _storeFilter == null;
  int get totalGrossCents => _useServerTotals
      ? _summary!.grossCents
      : transactions.fold<int>(0, (sum, t) => sum + t.grossAmountCents);
  int get totalNetCents => _useServerTotals
      ? _summary!.netCents
      : transactions.fold<int>(0, (sum, t) => sum + t.netAmountCents);
  int get shopifyCutCents => _useServerTotals
      ? _summary!.shopifyCutCents
      : totalGrossCents - totalNetCents;

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
