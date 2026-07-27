import '../providers/analytics_provider.dart';
import '../providers/apps_provider.dart';
import '../providers/churn_provider.dart';
import '../providers/cohorts_provider.dart';
import '../providers/dashboard_provider.dart';
import '../providers/earnings_provider.dart';
import '../providers/earnings_report_provider.dart';
import '../providers/events_provider.dart';
import '../providers/insights_provider.dart';
import '../providers/mrr_report_provider.dart';
import '../providers/retention_provider.dart';
import '../providers/usage_provider.dart';
import '../providers/usage_trends_provider.dart';
import '../providers/uninstall_context_provider.dart';
import '../providers/reviews_provider.dart';
import '../providers/revenue_at_risk_provider.dart';
import '../providers/revenue_mix_provider.dart';
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
  final RevenueAtRiskProvider _revenueAtRiskProvider;
  final RevenueMixProvider _revenueMixProvider;
  final ChurnProvider _churnProvider;
  final CohortsProvider _cohortsProvider;
  final RetentionProvider _retentionProvider;
  final UsageProvider _usageProvider;
  final UsageTrendsProvider _usageTrendsProvider;
  final MrrReportProvider _mrrReportProvider;
  final UninstallContextProvider _uninstallContextProvider;
  final ReviewsProvider _reviewsProvider;
  final AnalyticsProvider _analyticsProvider;
  final EarningsProvider _earningsProvider;
  final EarningsReportProvider _earningsReportProvider;
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
    required RevenueAtRiskProvider revenueAtRiskProvider,
    required RevenueMixProvider revenueMixProvider,
    required ChurnProvider churnProvider,
    required CohortsProvider cohortsProvider,
    required RetentionProvider retentionProvider,
    required UsageProvider usageProvider,
    required UsageTrendsProvider usageTrendsProvider,
    required MrrReportProvider mrrReportProvider,
    required UninstallContextProvider uninstallContextProvider,
    required ReviewsProvider reviewsProvider,
    required AnalyticsProvider analyticsProvider,
    required EarningsProvider earningsProvider,
    required EarningsReportProvider earningsReportProvider,
    required InsightsProvider insightsProvider,
    required WebhookProvider webhookProvider,
  })  : _appsProvider = appsProvider,
        _dashboardProvider = dashboardProvider,
        _subscriptionProvider = subscriptionProvider,
        _storeProvider = storeProvider,
        _transactionProvider = transactionProvider,
        _eventsProvider = eventsProvider,
        _riskProvider = riskProvider,
        _revenueAtRiskProvider = revenueAtRiskProvider,
        _revenueMixProvider = revenueMixProvider,
        _churnProvider = churnProvider,
        _cohortsProvider = cohortsProvider,
        _retentionProvider = retentionProvider,
        _usageProvider = usageProvider,
        _usageTrendsProvider = usageTrendsProvider,
        _mrrReportProvider = mrrReportProvider,
        _uninstallContextProvider = uninstallContextProvider,
        _reviewsProvider = reviewsProvider,
        _analyticsProvider = analyticsProvider,
        _earningsProvider = earningsProvider,
        _earningsReportProvider = earningsReportProvider,
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
    _revenueAtRiskProvider.setDemoMode(value);
    _revenueMixProvider.setDemoMode(value);
    _churnProvider.setDemoMode(value);
    _cohortsProvider.setDemoMode(value);
    _retentionProvider.setDemoMode(value);
    _usageProvider.setDemoMode(value);
    _usageTrendsProvider.setDemoMode(value);
    _mrrReportProvider.setDemoMode(value);
    _uninstallContextProvider.setDemoMode(value);
    _reviewsProvider.setDemoMode(value);
    _analyticsProvider.setDemoMode(value);
    _earningsProvider.setDemoMode(value);
    _earningsReportProvider.setDemoMode(value);
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
      _revenueAtRiskProvider.setSelectedApp(appId);
      _revenueMixProvider.setSelectedApp(appId);
      _churnProvider.setSelectedApp(appId);
      _cohortsProvider.setSelectedApp(appId);
      _retentionProvider.setSelectedApp(appId);
      _usageProvider.setSelectedApp(appId);
      _usageTrendsProvider.setSelectedApp(appId);
      _mrrReportProvider.setSelectedApp(appId);
      _uninstallContextProvider.setSelectedApp(appId);
      _reviewsProvider.setSelectedApp(appId);
      _analyticsProvider.setSelectedApp(appId);
      _earningsProvider.setSelectedApp(appId);
      _earningsReportProvider.setSelectedApp(appId);
      _insightsProvider.setSelectedApp(appId);
    }
  }
}
