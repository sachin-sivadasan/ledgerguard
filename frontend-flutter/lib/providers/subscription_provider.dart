import 'package:flutter/foundation.dart';
import '../mock_data/mock_payment_history.dart';
import '../mock_data/mock_risk_timeline.dart';
import '../mock_data/mock_subscriptions.dart';
import '../models/payment_history_model.dart';
import '../models/risk_timeline_model.dart';
import '../models/subscription_model.dart';
import '../widgets/lg_risk_badge.dart';
import '../widgets/lg_status_badge.dart';

class SubscriptionProvider extends ChangeNotifier {
  String _searchQuery = '';
  SubscriptionStatus? _statusFilter;
  RiskState? _riskFilter;
  String? _planFilter;
  String? _appFilter;

  String get searchQuery => _searchQuery;
  SubscriptionStatus? get statusFilter => _statusFilter;
  RiskState? get riskFilter => _riskFilter;
  String? get planFilter => _planFilter;
  String? get appFilter => _appFilter;

  List<Subscription> get subscriptions {
    var list = mockSubscriptions.toList();

    if (_appFilter != null) {
      list = list.where((s) => s.appId == _appFilter).toList();
    }

    if (_searchQuery.isNotEmpty) {
      final q = _searchQuery.toLowerCase();
      list = list.where((s) =>
          s.shopDomain.toLowerCase().contains(q) ||
          s.planName.toLowerCase().contains(q)).toList();
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
      return mockSubscriptions.firstWhere((s) => s.id == id);
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
    if (_appFilter == null) return mockSubscriptions;
    return mockSubscriptions.where((s) => s.appId == _appFilter).toList();
  }

  // Summary stats (filtered by app)
  int get activeCount =>
      _appFilteredSubs.where((s) => s.status == SubscriptionStatus.active).length;

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
    final avg = subs.fold<int>(0, (sum, s) => sum + s.priceCents) / subs.length;
    return '\$${(avg / 100).toStringAsFixed(2)}';
  }

  void setAppFilter(String? appId) {
    _appFilter = appId;
    notifyListeners();
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
