import 'package:flutter/foundation.dart';
import '../mock_data/mock_analytics.dart';
import '../mock_data/mock_apps.dart';
import '../mock_data/mock_subscriptions.dart';
import '../models/analytics_model.dart';
import '../models/app_model.dart';
import '../services/metrics_service.dart';

class AnalyticsProvider extends ChangeNotifier {
  final MetricsService _metricsService;

  bool _demoMode = true;
  bool _isLoading = false;
  String? _error;
  int _selectedTab = 0;
  String? _selectedAppId;

  DashboardMetrics? _liveMetrics;

  AnalyticsProvider(this._metricsService);

  bool get demoMode => _demoMode;
  bool get isLoading => _isLoading;
  String? get error => _error;
  int get selectedTab => _selectedTab;
  String? get selectedAppId => _selectedAppId;

  void setDemoMode(bool value) {
    _demoMode = value;
    notifyListeners();
  }

  void setTab(int tab) {
    _selectedTab = tab;
    notifyListeners();
  }

  void setSelectedApp(String? appId) {
    _selectedAppId = appId;
    notifyListeners();
    if (!_demoMode && appId != null) {
      loadAnalytics(appId);
    }
  }

  Future<void> loadAnalytics(String appId) async {
    if (_demoMode || _isLoading) return;
    _isLoading = true;
    _error = null;
    notifyListeners();
    try {
      _liveMetrics = await _metricsService.fetchMetrics(appId);
    } catch (e) {
      _error = e.toString();
    }
    _isLoading = false;
    notifyListeners();
  }

  // Revenue tab
  List<MrrSnapshot> get mrrSnapshots {
    if (!_demoMode) return _liveMetrics?.mrrTrend ?? [];
    return mockMrrSnapshots;
  }

  List<MrrMovement> get mrrMovements => mockMrrMovements;

  RevenueMix get revenueMix {
    if (!_demoMode) {
      return _liveMetrics?.revenueMix ??
          const RevenueMix(
              recurringCents: 0, usageCents: 0, oneTimeCents: 0);
    }
    return mockRevenueMix;
  }

  // Forecasting tab
  List<ForecastPoint> get forecast => mockForecast;

  // Profit tab
  List<ExpenseBreakdown> get expenses => mockExpenses;
  double get avgProfitMargin {
    if (mockExpenses.isEmpty) return 0;
    return mockExpenses
            .map((e) => e.profitMarginPct)
            .reduce((a, b) => a + b) /
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
      map[app.id] =
          mockSubscriptions.where((s) => s.appId == app.id).length;
    }
    return map;
  }
}
