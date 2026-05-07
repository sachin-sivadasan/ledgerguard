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
  String? _error;

  String _searchQuery = '';
  SubscriptionStatus? _statusFilter;
  RiskState? _riskFilter;
  String? _planFilter;
  String? _appFilter;

  List<Subscription> _liveSubscriptions = [];

  SubscriptionProvider(this._subscriptionService);

  bool get demoMode => _demoMode;
  bool get isLoading => _isLoading;
  String? get error => _error;
  String get searchQuery => _searchQuery;
  SubscriptionStatus? get statusFilter => _statusFilter;
  RiskState? get riskFilter => _riskFilter;
  String? get planFilter => _planFilter;
  String? get appFilter => _appFilter;

  void setDemoMode(bool value) {
    _demoMode = value;
    notifyListeners();
  }

  Future<void> loadSubscriptions(String appId) async {
    if (_demoMode || _isLoading) return;
    _isLoading = true;
    _error = null;
    notifyListeners();
    try {
      _liveSubscriptions =
          await _subscriptionService.fetchSubscriptions(appId);
    } catch (e) {
      _error = e.toString();
    }
    _isLoading = false;
    notifyListeners();
  }

  List<Subscription> get _allSubscriptions =>
      _demoMode ? mockSubscriptions : _liveSubscriptions;

  List<Subscription> get subscriptions {
    var list = _allSubscriptions.toList();

    if (_appFilter != null) {
      list = list.where((s) => s.appId == _appFilter).toList();
    }

    if (_searchQuery.isNotEmpty) {
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
    if (_appFilter == null) return _allSubscriptions;
    return _allSubscriptions.where((s) => s.appId == _appFilter).toList();
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

  void setAppFilter(String? appId) {
    _appFilter = appId;
    notifyListeners();
    if (!_demoMode && appId != null) {
      loadSubscriptions(appId);
    }
  }

  void setSearch(String query) {
    _searchQuery = query;
    notifyListeners();
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
    _appFilter = null;
    notifyListeners();
  }
}
