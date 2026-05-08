import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import 'app.dart';
import 'core/config/app_config.dart';
import 'core/demo_mode_coordinator.dart';
import 'core/network/api_client.dart';
import 'firebase_options.dart';
import 'providers/analytics_provider.dart';
import 'providers/auth_provider.dart';
import 'providers/api_key_provider.dart';
import 'providers/apps_provider.dart';
import 'providers/dashboard_provider.dart';
import 'providers/earnings_provider.dart';
import 'providers/events_provider.dart';
import 'providers/insights_provider.dart';
import 'providers/risk_provider.dart';
import 'providers/settings_provider.dart';
import 'providers/store_provider.dart';
import 'providers/subscription_provider.dart';
import 'providers/transaction_provider.dart';
import 'providers/organization_provider.dart';
import 'providers/webhook_provider.dart';
import 'services/app_service.dart';
import 'services/earnings_service.dart';
import 'services/events_service.dart';
import 'services/insights_service.dart';
import 'services/metrics_service.dart';
import 'services/mixpanel_service.dart';
import 'services/risk_service.dart';
import 'services/store_service.dart';
import 'services/subscription_service.dart';
import 'services/organization_service.dart';
import 'services/transaction_service.dart';

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
  final riskService = RiskService(apiClient);
  final eventsService = EventsService(apiClient);
  final insightsService = InsightsService(apiClient);
  final storeService = StoreService(apiClient);
  final organizationService = OrganizationService(apiClient);

  // Create providers (kept as local vars for coordinator wiring)
  final appsProvider = AppsProvider(appService);
  final dashboardProvider = DashboardProvider(metricsService);
  final subscriptionProvider = SubscriptionProvider(subscriptionService);
  final storeProvider = StoreProvider(storeService, subscriptionService);
  final transactionProvider = TransactionProvider(transactionService);
  final eventsProvider = EventsProvider(eventsService);
  final riskProvider = RiskProvider(riskService);
  final analyticsProvider = AnalyticsProvider(metricsService);
  final earningsProvider = EarningsProvider(earningsService);
  final insightsProvider = InsightsProvider(insightsService);

  final demoCoordinator = DemoModeCoordinator(
    appsProvider: appsProvider,
    dashboardProvider: dashboardProvider,
    subscriptionProvider: subscriptionProvider,
    storeProvider: storeProvider,
    transactionProvider: transactionProvider,
    eventsProvider: eventsProvider,
    riskProvider: riskProvider,
    analyticsProvider: analyticsProvider,
    earningsProvider: earningsProvider,
    insightsProvider: insightsProvider,
  );

  runApp(
    MultiProvider(
      providers: [
        Provider<MixpanelService>.value(value: mixpanel),
        Provider<ApiClient>.value(value: apiClient),
        Provider<DemoModeCoordinator>.value(value: demoCoordinator),
        ChangeNotifierProvider(create: (_) => AuthProvider()),
        ChangeNotifierProvider.value(value: dashboardProvider),
        ChangeNotifierProvider.value(value: subscriptionProvider),
        ChangeNotifierProvider.value(value: storeProvider),
        ChangeNotifierProvider.value(value: transactionProvider),
        ChangeNotifierProvider.value(value: eventsProvider),
        ChangeNotifierProvider(create: (_) => WebhookProvider()),
        ChangeNotifierProvider.value(value: riskProvider),
        ChangeNotifierProvider.value(value: analyticsProvider),
        ChangeNotifierProvider.value(value: earningsProvider),
        ChangeNotifierProvider.value(value: appsProvider),
        ChangeNotifierProvider(create: (_) => ApiKeyProvider()),
        ChangeNotifierProvider.value(value: insightsProvider),
        ChangeNotifierProvider(create: (_) => SettingsProvider()),
        ChangeNotifierProvider(
            create: (_) => OrganizationProvider(organizationService,
                apiClient: apiClient)),
      ],
      child: const App(),
    ),
  );
}
