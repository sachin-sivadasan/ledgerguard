import 'package:flutter/foundation.dart';
import '../mock_data/mock_transactions.dart';
import '../models/transaction_model.dart';

class TransactionProvider extends ChangeNotifier {
  ChargeType? _typeFilter;
  String? _appFilter;
  String? _storeFilter;

  ChargeType? get typeFilter => _typeFilter;
  String? get appFilter => _appFilter;
  String? get storeFilter => _storeFilter;

  List<Transaction> get transactions {
    var list = mockTransactions.toList();

    if (_typeFilter != null) {
      list = list.where((t) => t.chargeType == _typeFilter).toList();
    }
    if (_appFilter != null) {
      list = list.where((t) => t.appId == _appFilter).toList();
    }
    if (_storeFilter != null) {
      list = list.where((t) => t.shopDomain.contains(_storeFilter!)).toList();
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
