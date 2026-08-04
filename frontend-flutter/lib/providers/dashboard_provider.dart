import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:intl/intl.dart';
import '../core/dashboard_registry.dart';
import '../mock_data/mock_analytics.dart';
import '../mock_data/mock_events.dart';
import '../mock_data/mock_insights.dart';
import '../mock_data/mock_subscriptions.dart';
import '../models/analytics_model.dart';
import '../models/event_model.dart';
import '../models/insight_model.dart';
import '../models/subscription_model.dart';
import '../services/metrics_service.dart';
import '../services/user_preferences_service.dart';
import '../widgets/lg_risk_badge.dart';

enum DashboardTimeRange { thisWeek, thisMonth, lastMonth, threeMonths }

class DashboardProvider extends ChangeNotifier {
  final MetricsService _metricsService;
  final UserPreferencesService? _prefsService;

  bool _demoMode = false;
  bool _isLoading = false;
  String? _error;
  bool _isServiceUnavailable = false;
  Timer? _retryTimer;
  int _retryCount = 0;
  static const int _maxRetries = 3;
  static const Duration _retryInterval = Duration(seconds: 15);
  String? _selectedAppId;
  DashboardTimeRange _timeRange = DashboardTimeRange.thisWeek;
  CancelToken? _cancelToken;

  DashboardMetrics? _metrics;
  List<MrrSnapshot> _liveMrrTrend = const [];
  ForecastResult? _liveForecast;
  String? _forecastError;

  List<String> _primaryKpis = List.of(kDefaultKpiIds);
  List<String> _secondaryWidgets = List.of(kDefaultWidgetIds);

  DashboardProvider(this._metricsService, [this._prefsService]);

  bool get demoMode => _demoMode;
  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isServiceUnavailable => _isServiceUnavailable;
  String? get selectedAppId => _selectedAppId;
  DashboardTimeRange get timeRange => _timeRange;
  List<String> get primaryKpis => _primaryKpis;
  List<String> get secondaryWidgets => _secondaryWidgets;

  Future<void> loadDashboardPreferences() async {
    final prefs = await _prefsService?.getDashboardPreferences();
    if (prefs != null) {
      _primaryKpis = prefs.primaryKpis;
      _secondaryWidgets = prefs.secondaryWidgets;
      notifyListeners();
    }
  }

  void saveDashboardPreferences(
      List<String> kpis, List<String> widgets) {
    _primaryKpis = kpis;
    _secondaryWidgets = widgets;
    notifyListeners();
    _prefsService?.saveDashboardPreferences(kpis, widgets);
  }

  void setDemoMode(bool value) {
    _demoMode = value;
    notifyListeners();
  }

  void setSelectedApp(String? appId) {
    _selectedAppId = appId;
    notifyListeners();
    if (!_demoMode && appId != null) {
      loadMetrics(appId);
    }
  }

  void setTimeRange(DashboardTimeRange range) {
    if (range == _timeRange) return;
    _timeRange = range;
    notifyListeners();
    // Live KPIs are period-scoped on the backend, so re-fetch for the new window.
    if (!_demoMode && _selectedAppId != null) {
      loadMetrics(_selectedAppId!);
    }
  }

  /// The [start, end] window sent to /metrics for the selected chip. The backend derives
  /// the previous period + deltas from it.
  ({DateTime from, DateTime to}) _periodRange() {
    final now = DateTime.now();
    return switch (_timeRange) {
      DashboardTimeRange.thisWeek => (
          from: now.subtract(const Duration(days: 6)),
          to: now,
        ),
      DashboardTimeRange.thisMonth => (
          from: DateTime(now.year, now.month, 1),
          to: now,
        ),
      DashboardTimeRange.lastMonth => (
          from: DateTime(now.year, now.month - 1, 1),
          to: DateTime(now.year, now.month, 0), // last day of the previous month
        ),
      DashboardTimeRange.threeMonths => (
          from: DateTime(now.year, now.month - 2, 1),
          to: now,
        ),
    };
  }

  Future<void> loadMetrics(String appId) async {
    debugPrint('[DashboardProvider] loadMetrics called – appId=$appId demoMode=$_demoMode isLoading=$_isLoading');
    if (_demoMode) return;
    _cancelToken?.cancel('Superseded');
    _cancelToken = CancelToken();
    _isLoading = true;
    _error = null;
    _isServiceUnavailable = false;
    notifyListeners();
    try {
      final range = _periodRange();
      _metrics = await _metricsService.fetchMetrics(appId,
          from: range.from, to: range.to, cancelToken: _cancelToken);
      // Trend + forecast come from their own endpoints (the /metrics payload has neither).
      // Fetch them alongside; a failure in either degrades that widget only, not the KPIs.
      final token = _cancelToken;
      final results = await Future.wait([
        _metricsService.fetchMrrTrend(appId, months: 12, cancelToken: token),
        _metricsService.fetchForecast(appId, cancelToken: token),
      ]);
      // Bail if a newer loadMetrics superseded this one mid-flight (the services swallow
      // their own cancel errors, so the catch below never fires for a cancel).
      if (token != _cancelToken) return;
      _liveMrrTrend = results[0] as List<MrrSnapshot>;
      final (forecast, forecastErr) = results[1] as (ForecastResult?, String?);
      _liveForecast = forecast;
      _forecastError = forecastErr;
      _resetRetry();
      debugPrint('[DashboardProvider] loadMetrics success');
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) return;
      if (e.response?.statusCode == 503) {
        _isServiceUnavailable = true;
        _error = 'Service temporarily unavailable. Retrying...';
        _scheduleRetry(appId);
      } else {
        _error = e.message;
      }
      debugPrint('[DashboardProvider] loadMetrics error – $e');
    } catch (e) {
      _error = e.toString();
      debugPrint('[DashboardProvider] loadMetrics error – $e');
    }
    _isLoading = false;
    notifyListeners();
  }

  void _scheduleRetry(String appId) {
    if (_retryCount >= _maxRetries) return;
    _retryTimer?.cancel();
    _retryTimer = Timer(_retryInterval, () {
      _retryCount++;
      loadMetrics(appId);
    });
  }

  void _resetRetry() {
    _retryTimer?.cancel();
    _retryTimer = null;
    _retryCount = 0;
  }

  @override
  void dispose() {
    _retryTimer?.cancel();
    _cancelToken?.cancel('disposed');
    super.dispose();
  }

  DateTime get _activityCutoff {
    final now = DateTime.now();
    return switch (_timeRange) {
      DashboardTimeRange.thisWeek => now.subtract(const Duration(days: 7)),
      DashboardTimeRange.thisMonth => DateTime(now.year, now.month, 1),
      DashboardTimeRange.lastMonth => DateTime(now.year, now.month - 1, 1),
      DashboardTimeRange.threeMonths => DateTime(now.year, now.month - 2, 1),
    };
  }

  List<Subscription> get _filteredSubscriptions {
    if (_selectedAppId == null) return mockSubscriptions;
    return mockSubscriptions.where((s) => s.appId == _selectedAppId).toList();
  }

  // KPIs
  int get mrrCents {
    if (!_demoMode) return _metrics?.mrrCents ?? 0;
    return _filteredSubscriptions
        .where((s) => s.riskState == RiskState.safe)
        .fold<int>(0, (sum, s) => sum + s.priceCents);
  }

  String get mrrFormatted => _money.format(mrrCents / 100);

  double get renewalRate {
    if (!_demoMode) return _metrics?.renewalRate ?? 0;
    final subs = _filteredSubscriptions;
    final total = subs.length;
    final safe = subs.where((s) => s.riskState == RiskState.safe).length;
    return total > 0 ? safe / total * 100 : 0;
  }

  int get revenueAtRiskCents {
    if (!_demoMode) return _metrics?.revenueAtRiskCents ?? 0;
    return _filteredSubscriptions
        .where((s) =>
            s.riskState != RiskState.safe && s.riskState != RiskState.churned)
        .fold<int>(0, (sum, s) => sum + s.priceCents);
  }

  String get revenueAtRiskFormatted => _money.format(revenueAtRiskCents / 100);

  int get usageRevenueCents {
    if (!_demoMode) return _metrics?.usageRevenueCents ?? 0;
    return mockRevenueMix.usageCents;
  }

  String get usageRevenueFormatted => _money.format(usageRevenueCents / 100);

  String get totalRevenueFormatted =>
      _money.format((mrrCents + usageRevenueCents) / 100);

  Map<String, int> get activity {
    if (!_demoMode) return {};
    final cutoff = _activityCutoff;
    final types = {
      'Installs': EventType.appInstall,
      'Uninstalls': EventType.appUninstall,
      'Churns': EventType.subscriptionCancelled,
      'Bill. Failures': EventType.billingFailure,
      'Upgrades': EventType.planUpgrade,
      'Downgrades': EventType.planDowngrade,
    };
    return {
      for (final entry in types.entries)
        entry.key: mockEvents
            .where((e) =>
                e.type == entry.value &&
                e.date.isAfter(cutoff) &&
                (_selectedAppId == null || e.appId == _selectedAppId))
            .length,
    };
  }

  String get activityTitle => switch (_timeRange) {
        DashboardTimeRange.thisWeek => 'This Week Activity',
        DashboardTimeRange.thisMonth => 'This Month Activity',
        DashboardTimeRange.lastMonth => 'Last Month Activity',
        DashboardTimeRange.threeMonths => 'Last 3 Months Activity',
      };

  List<MrrSnapshot> get mrrTrend {
    if (!_demoMode) return _liveMrrTrend;
    return mockMrrSnapshots;
  }

  RiskDistribution get riskDistribution {
    if (!_demoMode) {
      return _metrics?.riskDistribution ??
          const RiskDistribution(
              safe: 0, oneCycle: 0, twoCycle: 0, churned: 0);
    }
    return mockRiskDistribution;
  }

  RevenueMix get revenueMix {
    if (!_demoMode) {
      return _metrics?.revenueMix ??
          const RevenueMix(
              recurringCents: 0, usageCents: 0, oneTimeCents: 0);
    }
    return mockRevenueMix;
  }

  /// Next-period forecast: the first real forecast point when live (else the demo mock).
  /// Null in live mode when the backend can't forecast yet (too few snapshots).
  ForecastPoint? get nextMonthForecast {
    if (_demoMode) return mockForecast.first;
    return _liveForecast?.points.isNotEmpty == true
        ? _liveForecast!.points.first
        : null;
  }

  /// Explains why the forecast is unavailable in live mode (e.g. insufficient snapshots).
  String? get forecastError => _demoMode ? null : _forecastError;

  List<AiInsight> get recentAlerts => mockInsights.take(3).toList();

  /// Real period-over-period trend for a KPI card: a signed "%" label + whether it's a
  /// GOOD move (green). Null when there's no live delta (demo mode or a KPI without one) —
  /// the card then shows no badge rather than a stale hardcoded one. Risk is inverted: a
  /// drop is good.
  ({String label, bool positive})? kpiTrend(String kpiId) {
    if (_demoMode || _metrics == null) return null;
    final m = _metrics!;
    final (double? pct, bool upIsGood) = switch (kpiId) {
      'active_mrr' => (m.mrrDeltaPct, true),
      'renewal_success_rate' => (m.renewalDeltaPct, true),
      'usage_revenue' => (m.usageDeltaPct, true),
      'revenue_at_risk' => (m.riskDeltaPct, false),
      _ => (null, true),
    };
    if (pct == null) return null;
    final label = '${pct >= 0 ? '+' : ''}${pct.toStringAsFixed(1)}%';
    return (label: label, positive: upIsGood ? pct >= 0 : pct <= 0);
  }
}

/// Formats whole-dollar currency with thousands separators.
final _money = NumberFormat.currency(symbol: '\$', decimalDigits: 0);
