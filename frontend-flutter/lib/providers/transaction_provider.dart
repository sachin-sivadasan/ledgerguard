import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../mock_data/mock_transactions.dart';
import '../models/transaction_model.dart';
import '../services/transaction_service.dart';

class TransactionProvider extends ChangeNotifier {
  final TransactionService _transactionService;

  bool _demoMode = false;
  bool _isLoading = false;
  String? _error;
  CancelToken? _cancelToken;

  ChargeType? _typeFilter;
  String? _appFilter;
  String? _storeFilter;

  List<Transaction> _liveTransactions = [];

  TransactionProvider(this._transactionService);

  bool get demoMode => _demoMode;
  bool get isLoading => _isLoading;
  String? get error => _error;
  ChargeType? get typeFilter => _typeFilter;
  String? get appFilter => _appFilter;
  String? get storeFilter => _storeFilter;

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
    notifyListeners();
    try {
      _liveTransactions = await _transactionService.fetchTransactions(appId,
          cancelToken: _cancelToken);
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) return;
      _error = e.message;
    } catch (e) {
      _error = e.toString();
    }
    _isLoading = false;
    notifyListeners();
  }

  List<Transaction> get transactions {
    var list =
        _demoMode ? mockTransactions.toList() : _liveTransactions.toList();

    if (_typeFilter != null) {
      list = list.where((t) => t.chargeType == _typeFilter).toList();
    }
    if (_appFilter != null) {
      list = list.where((t) => t.appId == _appFilter).toList();
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
    notifyListeners();
  }

  void setAppFilter(String? appId) {
    _appFilter = appId;
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
    _appFilter = null;
    _storeFilter = null;
    notifyListeners();
  }
}
