import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'app.dart';
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
import 'providers/webhook_provider.dart';
import 'services/mixpanel_service.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  final mixpanel = MixpanelService();
  await mixpanel.init();

  runApp(
    MultiProvider(
      providers: [
        Provider<MixpanelService>.value(value: mixpanel),
        ChangeNotifierProvider(create: (_) => AuthProvider()),
        ChangeNotifierProvider(create: (_) => DashboardProvider()),
        ChangeNotifierProvider(create: (_) => SubscriptionProvider()),
        ChangeNotifierProvider(create: (_) => StoreProvider()),
        ChangeNotifierProvider(create: (_) => TransactionProvider()),
        ChangeNotifierProvider(create: (_) => EventsProvider()),
        ChangeNotifierProvider(create: (_) => WebhookProvider()),
        ChangeNotifierProvider(create: (_) => RiskProvider()),
        ChangeNotifierProvider(create: (_) => AnalyticsProvider()),
        ChangeNotifierProvider(create: (_) => EarningsProvider()),
        ChangeNotifierProvider(create: (_) => AppsProvider()),
        ChangeNotifierProvider(create: (_) => ApiKeyProvider()),
        ChangeNotifierProvider(create: (_) => InsightsProvider()),
        ChangeNotifierProvider(create: (_) => SettingsProvider()),
      ],
      child: const App(),
    ),
  );
}
