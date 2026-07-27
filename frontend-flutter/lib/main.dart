import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import 'app.dart';
import 'core/config/app_config.dart';
import 'core/demo_mode_coordinator.dart';
import 'core/navigation/navigation_refresh_notifier.dart';
import 'core/network/api_client.dart';
import 'firebase_options.dart';
import 'providers/analytics_provider.dart';
import 'providers/auth_provider.dart';
import 'providers/api_key_provider.dart';
import 'providers/apps_provider.dart';
import 'providers/churn_provider.dart';
import 'providers/cohorts_provider.dart';
import 'providers/dashboard_provider.dart';
import 'providers/earnings_provider.dart';
import 'providers/earnings_report_provider.dart';
import 'providers/events_provider.dart';
import 'providers/insights_provider.dart';
import 'providers/mrr_report_provider.dart';
import 'providers/retention_provider.dart';
import 'providers/usage_provider.dart';
import 'providers/usage_trends_provider.dart';
import 'providers/subscriptions_provider.dart';
import 'providers/payout_schedule_provider.dart';
import 'providers/payout_history_provider.dart';
import 'providers/reviews_provider.dart';
import 'providers/revenue_at_risk_provider.dart';
import 'providers/revenue_mix_provider.dart';
import 'providers/risk_provider.dart';
import 'providers/settings_provider.dart';
import 'providers/store_provider.dart';
import 'providers/subscription_provider.dart';
import 'providers/sync_status_provider.dart';
import 'providers/transaction_provider.dart';
import 'providers/uninstall_context_provider.dart';
import 'providers/organization_provider.dart';
import 'providers/webhook_provider.dart';
import 'services/app_service.dart';
import 'services/churn_service.dart';
import 'services/cohorts_service.dart';
import 'services/earnings_report_service.dart';
import 'services/earnings_service.dart';
import 'services/events_service.dart';
import 'services/insights_service.dart';
import 'services/metrics_service.dart';
import 'services/mrr_report_service.dart';
import 'services/retention_service.dart';
import 'services/usage_service.dart';
import 'services/usage_trends_service.dart';
import 'services/subscriptions_service.dart';
import 'services/payout_schedule_service.dart';
import 'services/payout_history_service.dart';
import 'services/reviews_service.dart';
import 'services/mixpanel_service.dart';
import 'services/revenue_at_risk_service.dart';
import 'services/revenue_mix_service.dart';
import 'services/risk_service.dart';
import 'services/store_service.dart';
import 'services/subscription_service.dart';
import 'services/organization_service.dart';
import 'services/sync_status_service.dart';
import 'services/transaction_service.dart';
import 'services/uninstall_context_service.dart';
import 'services/user_preferences_service.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await Firebase.initializeApp(options: DefaultFirebaseOptions.currentPlatform);

  final mixpanel = MixpanelService();
  await mixpanel.init();

  final apiClient = ApiClient(baseUrl: AppConfig.apiBaseUrl);

  // Create services
  final appService = AppService(apiClient);
  final metricsService = MetricsService(apiClient);
  final subscriptionService = SubscriptionService(apiClient);
  final transactionService = TransactionService(apiClient);
  final earningsService = EarningsService(apiClient);
  final earningsReportService = EarningsReportService(apiClient);
  final riskService = RiskService(apiClient);
  final revenueAtRiskService = RevenueAtRiskService(apiClient);
  final revenueMixService = RevenueMixService(apiClient);
  final churnService = ChurnService(apiClient);
  final cohortsService = CohortsService(apiClient);
  final retentionService = RetentionService(apiClient);
  final usageService = UsageService(apiClient);
  final usageTrendsService = UsageTrendsService(apiClient);
  final subscriptionsService = SubscriptionsService(apiClient);
  final payoutScheduleService = PayoutScheduleService(apiClient);
  final payoutHistoryService = PayoutHistoryService(apiClient);
  final mrrReportService = MrrReportService(apiClient);
  final uninstallContextService = UninstallContextService(apiClient);
  final reviewsService = ReviewsService(apiClient);
  final eventsService = EventsService(apiClient);
  final insightsService = InsightsService(apiClient);
  final storeService = StoreService(apiClient);
  final organizationService = OrganizationService(apiClient);
  final syncStatusService = SyncStatusService(apiClient);
  final userPreferencesService = UserPreferencesService(apiClient);

  // Create providers (kept as local vars for coordinator wiring)
  final appsProvider = AppsProvider(appService, userPreferencesService);
  final dashboardProvider = DashboardProvider(metricsService, userPreferencesService);
  final subscriptionProvider = SubscriptionProvider(subscriptionService);
  final storeProvider = StoreProvider(storeService, subscriptionService);
  final transactionProvider = TransactionProvider(transactionService);
  final eventsProvider = EventsProvider(eventsService);
  final riskProvider = RiskProvider(riskService);
  final revenueAtRiskProvider = RevenueAtRiskProvider(revenueAtRiskService);
  final revenueMixProvider = RevenueMixProvider(revenueMixService);
  final churnProvider = ChurnProvider(churnService);
  final cohortsProvider = CohortsProvider(cohortsService);
  final retentionProvider = RetentionProvider(retentionService);
  final usageProvider = UsageProvider(usageService);
  final usageTrendsProvider = UsageTrendsProvider(usageTrendsService);
  final subscriptionsProvider = SubscriptionsProvider(subscriptionsService);
  final payoutScheduleProvider = PayoutScheduleProvider(payoutScheduleService);
  final payoutHistoryProvider = PayoutHistoryProvider(payoutHistoryService);
  final mrrReportProvider = MrrReportProvider(mrrReportService);
  final uninstallContextProvider =
      UninstallContextProvider(uninstallContextService);
  final reviewsProvider = ReviewsProvider(reviewsService);
  final analyticsProvider = AnalyticsProvider(metricsService, earningsService);
  final earningsProvider = EarningsProvider(earningsService);
  final earningsReportProvider = EarningsReportProvider(earningsReportService);
  final insightsProvider = InsightsProvider(insightsService);
  final webhookProvider = WebhookProvider();

  final demoCoordinator = DemoModeCoordinator(
    appsProvider: appsProvider,
    dashboardProvider: dashboardProvider,
    subscriptionProvider: subscriptionProvider,
    storeProvider: storeProvider,
    transactionProvider: transactionProvider,
    eventsProvider: eventsProvider,
    riskProvider: riskProvider,
    revenueAtRiskProvider: revenueAtRiskProvider,
    revenueMixProvider: revenueMixProvider,
    churnProvider: churnProvider,
    cohortsProvider: cohortsProvider,
    retentionProvider: retentionProvider,
    usageProvider: usageProvider,
    usageTrendsProvider: usageTrendsProvider,
    subscriptionsProvider: subscriptionsProvider,
    payoutScheduleProvider: payoutScheduleProvider,
    payoutHistoryProvider: payoutHistoryProvider,
    mrrReportProvider: mrrReportProvider,
    uninstallContextProvider: uninstallContextProvider,
    reviewsProvider: reviewsProvider,
    analyticsProvider: analyticsProvider,
    earningsProvider: earningsProvider,
    earningsReportProvider: earningsReportProvider,
    insightsProvider: insightsProvider,
    webhookProvider: webhookProvider,
  );

  runApp(
    MultiProvider(
      providers: [
        Provider<MixpanelService>.value(value: mixpanel),
        Provider<ApiClient>.value(value: apiClient),
        Provider<DemoModeCoordinator>.value(value: demoCoordinator),
        ChangeNotifierProvider(create: (_) => NavigationRefreshNotifier()),
        ChangeNotifierProvider(create: (_) => AuthProvider()),
        ChangeNotifierProvider.value(value: dashboardProvider),
        ChangeNotifierProvider.value(value: subscriptionProvider),
        ChangeNotifierProvider.value(value: storeProvider),
        ChangeNotifierProvider.value(value: transactionProvider),
        ChangeNotifierProvider.value(value: eventsProvider),
        ChangeNotifierProvider.value(value: webhookProvider),
        ChangeNotifierProvider.value(value: riskProvider),
        ChangeNotifierProvider.value(value: revenueAtRiskProvider),
        ChangeNotifierProvider.value(value: revenueMixProvider),
        ChangeNotifierProvider.value(value: churnProvider),
        ChangeNotifierProvider.value(value: cohortsProvider),
        ChangeNotifierProvider.value(value: retentionProvider),
        ChangeNotifierProvider.value(value: usageProvider),
        ChangeNotifierProvider.value(value: usageTrendsProvider),
        ChangeNotifierProvider.value(value: subscriptionsProvider),
        ChangeNotifierProvider.value(value: payoutScheduleProvider),
        ChangeNotifierProvider.value(value: payoutHistoryProvider),
        ChangeNotifierProvider.value(value: mrrReportProvider),
        ChangeNotifierProvider.value(value: uninstallContextProvider),
        ChangeNotifierProvider.value(value: reviewsProvider),
        ChangeNotifierProvider.value(value: analyticsProvider),
        ChangeNotifierProvider.value(value: earningsProvider),
        ChangeNotifierProvider.value(value: earningsReportProvider),
        ChangeNotifierProvider.value(value: appsProvider),
        ChangeNotifierProvider(create: (_) => ApiKeyProvider()),
        ChangeNotifierProvider.value(value: insightsProvider),
        ChangeNotifierProvider(create: (_) => SyncStatusProvider(syncStatusService)),
        ChangeNotifierProvider(create: (_) => SettingsProvider(userPreferencesService)),
        ChangeNotifierProvider(
            create: (_) => OrganizationProvider(organizationService,
                apiClient: apiClient,
                prefsService: userPreferencesService)),
      ],
      child: const App(),
    ),
  );
}
