import 'package:flutter/foundation.dart';
import '../mock_data/mock_stores.dart';
import '../mock_data/mock_subscriptions.dart';
import '../models/store_model.dart';
import '../models/subscription_model.dart';

class StoreProvider extends ChangeNotifier {
  String _searchQuery = '';
  String? _selectedAppId;

  String get searchQuery => _searchQuery;
  String? get selectedAppId => _selectedAppId;

  void setSelectedApp(String? appId) {
    _selectedAppId = appId;
    notifyListeners();
  }

  List<Store> get stores {
    var list = mockStores.toList();

    if (_selectedAppId != null) {
      list = list.where((s) => s.installedAppIds.contains(_selectedAppId)).toList();
    }

    if (_searchQuery.isNotEmpty) {
      final q = _searchQuery.toLowerCase();
      list = list.where((s) => s.shopDomain.toLowerCase().contains(q)).toList();
    }

    return list;
  }

  Store? getById(String id) {
    try {
      return mockStores.firstWhere((s) => s.id == id);
    } catch (_) {
      return null;
    }
  }

  List<Subscription> getSubscriptionsForStore(String shopDomain) {
    return mockSubscriptions.where((s) => s.shopDomain == shopDomain).toList();
  }

  void setSearch(String query) {
    _searchQuery = query;
    notifyListeners();
  }
}
