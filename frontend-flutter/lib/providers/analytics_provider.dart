import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../mock_data/mock_analytics.dart';
import '../mock_data/mock_apps.dart';
import '../mock_data/mock_subscriptions.dart';
import '../models/analytics_model.dart';
import '../models/app_model.dart';
import '../services/metrics_service.dart';

class AnalyticsProvider extends ChangeNotifier {
  final MetricsService _metricsService;

  bool _demoMode = false;
  bool _isLoading = false;
  String? _error;
  int _selectedTab = 0;
  String? _selectedAppId;
  CancelToken? _cancelToken;

  DashboardMetrics? _liveMetrics;
  List<MrrMovement>? _liveMrrMovements;
  List<ForecastPoint>? _liveForecast;
  ForecastResult? _forecastResult;
  String _forecastModel = 'linear';
  List<ExpenseBreakdown>? _liveExpenses;
  List<CohortData>? _liveCohorts;
  List<ShopifyApp>? _liveApps;

  AnalyticsProvider(this._metricsService);

  String get forecastModel => _forecastModel;
  ForecastResult? get forecastResult => _forecastResult;

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
    if (_demoMode) return;
    _cancelToken?.cancel('Superseded');
    _cancelToken = CancelToken();
    _isLoading = true;
    _error = null;
    notifyListeners();
    try {
      _liveMetrics = await _metricsService.fetchMetrics(appId,
          cancelToken: _cancelToken);
      // Load secondary data in parallel (non-blocking)
      _metricsService
          .fetchMrrMovements(appId, cancelToken: _cancelToken)
          .then((movements) {
        _liveMrrMovements = movements;
        notifyListeners();
      });
      _loadForecast(appId);
      _loadCohorts(appId);
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) return;
      _error = e.message;
    } catch (e) {
      _error = e.toString();
    }
    _isLoading = false;
    notifyListeners();
  }

  Future<void> _loadForecast(String appId) async {
    final result = await _metricsService.fetchForecast(appId,
        model: _forecastModel, cancelToken: _cancelToken);
    if (result != null) {
      _forecastResult = result;
      _liveForecast = result.points;
      notifyListeners();
    }
  }

  Future<void> _loadCohorts(String appId) async {
    final cohorts = await _metricsService.fetchCohorts(appId,
        cancelToken: _cancelToken);
    _liveCohorts = cohorts;
    notifyListeners();
  }

  void setForecastModel(String model) {
    _forecastModel = model;
    notifyListeners();
    if (!_demoMode && _selectedAppId != null) {
      _loadForecast(_selectedAppId!);
    }
  }

  // Revenue tab
  List<MrrSnapshot> get mrrSnapshots {
    if (!_demoMode) return _liveMetrics?.mrrTrend ?? [];
    return mockMrrSnapshots;
  }

  List<MrrMovement> get mrrMovements {
    if (!_demoMode) return _liveMrrMovements ?? [];
    return mockMrrMovements;
  }

  RevenueMix get revenueMix {
    if (!_demoMode) {
      return _liveMetrics?.revenueMix ??
          const RevenueMix(
              recurringCents: 0, usageCents: 0, oneTimeCents: 0);
    }
    return mockRevenueMix;
  }

  // Forecasting tab
  List<ForecastPoint> get forecast {
    if (!_demoMode) return _liveForecast ?? [];
    return mockForecast;
  }

  // Profit tab
  List<ExpenseBreakdown> get expenses {
    if (!_demoMode) return _liveExpenses ?? [];
    return mockExpenses;
  }

  double get avgProfitMargin {
    final data = expenses;
    if (data.isEmpty) return 0;
    return data.map((e) => e.profitMarginPct).reduce((a, b) => a + b) /
        data.length;
  }

  // Cohorts tab
  List<CohortData> get cohorts {
    if (!_demoMode) return _liveCohorts ?? [];
    return mockCohorts;
  }

  // Multi-app tab
  List<ShopifyApp> get apps {
    if (!_demoMode) return _liveApps ?? [];
    return mockApps;
  }

  Map<String, int> appMrrCents() {
    final appList = apps;
    if (!_demoMode) return {};
    final map = <String, int>{};
    for (final app in appList) {
      map[app.id] = mockSubscriptions
          .where((s) => s.appId == app.id)
          .fold<int>(0, (sum, s) => sum + s.priceCents);
    }
    return map;
  }

  Map<String, int> appSubCount() {
    final appList = apps;
    if (!_demoMode) return {};
    final map = <String, int>{};
    for (final app in appList) {
      map[app.id] =
          mockSubscriptions.where((s) => s.appId == app.id).length;
    }
    return map;
  }
}
