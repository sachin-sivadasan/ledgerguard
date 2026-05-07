import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:ledgerguard_flutter/app.dart';
import 'package:ledgerguard_flutter/core/network/api_client.dart';
import 'package:ledgerguard_flutter/providers/analytics_provider.dart';
import 'package:ledgerguard_flutter/providers/apps_provider.dart';
import 'package:ledgerguard_flutter/providers/auth_provider.dart';
import 'package:ledgerguard_flutter/providers/dashboard_provider.dart';
import 'package:ledgerguard_flutter/providers/earnings_provider.dart';
import 'package:ledgerguard_flutter/providers/events_provider.dart';
import 'package:ledgerguard_flutter/providers/insights_provider.dart';
import 'package:ledgerguard_flutter/providers/risk_provider.dart';
import 'package:ledgerguard_flutter/providers/settings_provider.dart';
import 'package:ledgerguard_flutter/providers/store_provider.dart';
import 'package:ledgerguard_flutter/providers/subscription_provider.dart';
import 'package:ledgerguard_flutter/providers/transaction_provider.dart';
import 'package:ledgerguard_flutter/services/app_service.dart';
import 'package:ledgerguard_flutter/services/earnings_service.dart';
import 'package:ledgerguard_flutter/services/events_service.dart';
import 'package:ledgerguard_flutter/services/insights_service.dart';
import 'package:ledgerguard_flutter/services/metrics_service.dart';
import 'package:ledgerguard_flutter/services/risk_service.dart';
import 'package:ledgerguard_flutter/services/store_service.dart';
import 'package:ledgerguard_flutter/services/subscription_service.dart';
import 'package:ledgerguard_flutter/services/transaction_service.dart';
import 'package:ledgerguard_flutter/services/mixpanel_service.dart';

void main() {
  // Note: Tests requiring Firebase Auth need firebase_core mock setup.
  // This basic smoke test just verifies the widget tree builds with demo mode.
  testWidgets('App renders with navigation', (WidgetTester tester) async {
    final apiClient = ApiClient(baseUrl: 'http://localhost:8080');
    final appService = AppService(apiClient);
    final metricsService = MetricsService(apiClient);
    final subscriptionService = SubscriptionService(apiClient);
    final transactionService = TransactionService(apiClient);
    final earningsService = EarningsService(apiClient);
    final riskService = RiskService(apiClient);
    final eventsService = EventsService(apiClient);
    final insightsService = InsightsService(apiClient);
    final storeService = StoreService(apiClient);

    await tester.pumpWidget(
      MultiProvider(
        providers: [
          Provider<MixpanelService>.value(value: MixpanelService()),
          ChangeNotifierProvider(create: (_) => AuthProvider()),
          ChangeNotifierProvider(
              create: (_) => DashboardProvider(metricsService)),
          ChangeNotifierProvider(
              create: (_) => SubscriptionProvider(subscriptionService)),
          ChangeNotifierProvider(
              create: (_) =>
                  StoreProvider(storeService, subscriptionService)),
          ChangeNotifierProvider(
              create: (_) => TransactionProvider(transactionService)),
          ChangeNotifierProvider(
              create: (_) => RiskProvider(riskService)),
          ChangeNotifierProvider(
              create: (_) => AnalyticsProvider(metricsService)),
          ChangeNotifierProvider(
              create: (_) => EarningsProvider(earningsService)),
          ChangeNotifierProvider(
              create: (_) => AppsProvider(appService)),
          ChangeNotifierProvider(
              create: (_) => EventsProvider(eventsService)),
          ChangeNotifierProvider(
              create: (_) => InsightsProvider(insightsService)),
          ChangeNotifierProvider(create: (_) => SettingsProvider()),
        ],
        child: const App(),
      ),
    );
    // Auth guard will redirect to login since no Firebase is initialized
    await tester.pumpAndSettle();
  });
}
