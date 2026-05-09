import '../providers/analytics_provider.dart';
import '../providers/apps_provider.dart';
import '../providers/dashboard_provider.dart';
import '../providers/earnings_provider.dart';
import '../providers/events_provider.dart';
import '../providers/insights_provider.dart';
import '../providers/risk_provider.dart';
import '../providers/store_provider.dart';
import '../providers/subscription_provider.dart';
import '../providers/transaction_provider.dart';
import '../providers/webhook_provider.dart';

/// Centralizes the demo mode toggle so settings_screen doesn't need to
/// enumerate every provider individually.
class DemoModeCoordinator {
  final AppsProvider _appsProvider;
  final DashboardProvider _dashboardProvider;
  final SubscriptionProvider _subscriptionProvider;
  final StoreProvider _storeProvider;
  final TransactionProvider _transactionProvider;
  final EventsProvider _eventsProvider;
  final RiskProvider _riskProvider;
  final AnalyticsProvider _analyticsProvider;
  final EarningsProvider _earningsProvider;
  final InsightsProvider _insightsProvider;
  final WebhookProvider _webhookProvider;

  DemoModeCoordinator({
    required AppsProvider appsProvider,
    required DashboardProvider dashboardProvider,
    required SubscriptionProvider subscriptionProvider,
    required StoreProvider storeProvider,
    required TransactionProvider transactionProvider,
    required EventsProvider eventsProvider,
    required RiskProvider riskProvider,
    required AnalyticsProvider analyticsProvider,
    required EarningsProvider earningsProvider,
    required InsightsProvider insightsProvider,
    required WebhookProvider webhookProvider,
  })  : _appsProvider = appsProvider,
        _dashboardProvider = dashboardProvider,
        _subscriptionProvider = subscriptionProvider,
        _storeProvider = storeProvider,
        _transactionProvider = transactionProvider,
        _eventsProvider = eventsProvider,
        _riskProvider = riskProvider,
        _analyticsProvider = analyticsProvider,
        _earningsProvider = earningsProvider,
        _insightsProvider = insightsProvider,
        _webhookProvider = webhookProvider;

  /// Set demo mode on all providers at once.
  void setDemoMode(bool value) {
    _appsProvider.setDemoMode(value);
    _dashboardProvider.setDemoMode(value);
    _subscriptionProvider.setDemoMode(value);
    _storeProvider.setDemoMode(value);
    _transactionProvider.setDemoMode(value);
    _eventsProvider.setDemoMode(value);
    _riskProvider.setDemoMode(value);
    _analyticsProvider.setDemoMode(value);
    _earningsProvider.setDemoMode(value);
    _insightsProvider.setDemoMode(value);
    _webhookProvider.setDemoMode(value);
  }

  /// Toggle to live mode: set demo off, load apps, then load all screen data.
  Future<void> switchToLiveMode() async {
    setDemoMode(false);
    await _appsProvider.loadApps();
    final apps = _appsProvider.apps;
    if (apps.isNotEmpty) {
      final appId = apps.first.id;
      _dashboardProvider.setSelectedApp(appId);
      _subscriptionProvider.setSelectedApp(appId);
      _storeProvider.setSelectedApp(appId);
      _transactionProvider.setSelectedApp(appId);
      _eventsProvider.setSelectedApp(appId);
      _riskProvider.setSelectedApp(appId);
      _analyticsProvider.setSelectedApp(appId);
      _earningsProvider.setSelectedApp(appId);
      _insightsProvider.setSelectedApp(appId);
    }
  }
}
