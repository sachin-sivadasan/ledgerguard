import 'package:flutter/foundation.dart';
import '../mock_data/mock_analytics.dart';
import '../mock_data/mock_apps.dart';
import '../mock_data/mock_subscriptions.dart';
import '../models/analytics_model.dart';
import '../models/app_model.dart';

class AnalyticsProvider extends ChangeNotifier {
  int _selectedTab = 0;
  String? _selectedAppId;

  int get selectedTab => _selectedTab;
  String? get selectedAppId => _selectedAppId;

  void setTab(int tab) {
    _selectedTab = tab;
    notifyListeners();
  }

  void setSelectedApp(String? appId) {
    _selectedAppId = appId;
    notifyListeners();
  }

  // Revenue tab
  List<MrrSnapshot> get mrrSnapshots => mockMrrSnapshots;
  List<MrrMovement> get mrrMovements => mockMrrMovements;
  RevenueMix get revenueMix => mockRevenueMix;

  // Forecasting tab
  List<ForecastPoint> get forecast => mockForecast;

  // Profit tab
  List<ExpenseBreakdown> get expenses => mockExpenses;
  double get avgProfitMargin {
    if (mockExpenses.isEmpty) return 0;
    return mockExpenses.map((e) => e.profitMarginPct).reduce((a, b) => a + b) /
        mockExpenses.length;
  }

  // Cohorts tab
  List<CohortData> get cohorts => mockCohorts;

  // Multi-app tab
  List<ShopifyApp> get apps => mockApps;

  Map<String, int> appMrrCents() {
    final map = <String, int>{};
    for (final app in mockApps) {
      map[app.id] = mockSubscriptions
          .where((s) => s.appId == app.id)
          .fold<int>(0, (sum, s) => sum + s.priceCents);
    }
    return map;
  }

  Map<String, int> appSubCount() {
    final map = <String, int>{};
    for (final app in mockApps) {
      map[app.id] = mockSubscriptions.where((s) => s.appId == app.id).length;
    }
    return map;
  }
}
