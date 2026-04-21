import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:ledgerguard_flutter/app.dart';
import 'package:ledgerguard_flutter/providers/analytics_provider.dart';
import 'package:ledgerguard_flutter/providers/apps_provider.dart';
import 'package:ledgerguard_flutter/providers/dashboard_provider.dart';
import 'package:ledgerguard_flutter/providers/insights_provider.dart';
import 'package:ledgerguard_flutter/providers/risk_provider.dart';
import 'package:ledgerguard_flutter/providers/settings_provider.dart';
import 'package:ledgerguard_flutter/providers/store_provider.dart';
import 'package:ledgerguard_flutter/providers/subscription_provider.dart';
import 'package:ledgerguard_flutter/providers/transaction_provider.dart';

void main() {
  testWidgets('App renders with navigation', (WidgetTester tester) async {
    await tester.pumpWidget(
      MultiProvider(
        providers: [
          ChangeNotifierProvider(create: (_) => DashboardProvider()),
          ChangeNotifierProvider(create: (_) => SubscriptionProvider()),
          ChangeNotifierProvider(create: (_) => StoreProvider()),
          ChangeNotifierProvider(create: (_) => TransactionProvider()),
          ChangeNotifierProvider(create: (_) => RiskProvider()),
          ChangeNotifierProvider(create: (_) => AnalyticsProvider()),
          ChangeNotifierProvider(create: (_) => AppsProvider()),
          ChangeNotifierProvider(create: (_) => InsightsProvider()),
          ChangeNotifierProvider(create: (_) => SettingsProvider()),
        ],
        child: const App(),
      ),
    );
    expect(find.text('Dashboard'), findsWidgets);
  });
}
