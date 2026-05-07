import 'package:flutter/foundation.dart';
import '../mock_data/mock_analytics.dart';
import '../mock_data/mock_events.dart';
import '../mock_data/mock_insights.dart';
import '../mock_data/mock_subscriptions.dart';
import '../models/analytics_model.dart';
import '../models/event_model.dart';
import '../models/insight_model.dart';
import '../models/subscription_model.dart';
import '../services/metrics_service.dart';
import '../widgets/lg_risk_badge.dart';

enum DashboardTimeRange { thisWeek, thisMonth, lastMonth, threeMonths }

class DashboardProvider extends ChangeNotifier {
  final MetricsService _metricsService;

  bool _demoMode = true;
  bool _isLoading = false;
  String? _error;
  String? _selectedAppId;
  DashboardTimeRange _timeRange = DashboardTimeRange.thisWeek;

  DashboardMetrics? _metrics;

  DashboardProvider(this._metricsService);

  bool get demoMode => _demoMode;
  bool get isLoading => _isLoading;
  String? get error => _error;
  String? get selectedAppId => _selectedAppId;
  DashboardTimeRange get timeRange => _timeRange;

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
    _timeRange = range;
    notifyListeners();
  }

  Future<void> loadMetrics(String appId) async {
    debugPrint('[DashboardProvider] loadMetrics called – appId=$appId demoMode=$_demoMode isLoading=$_isLoading');
    if (_demoMode || _isLoading) return;
    _isLoading = true;
    _error = null;
    notifyListeners();
    try {
      _metrics = await _metricsService.fetchMetrics(appId);
      debugPrint('[DashboardProvider] loadMetrics success');
    } catch (e) {
      _error = e.toString();
      debugPrint('[DashboardProvider] loadMetrics error – $e');
    }
    _isLoading = false;
    notifyListeners();
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

  String get mrrFormatted => '\$${(mrrCents / 100).toStringAsFixed(0)}';

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

  String get revenueAtRiskFormatted =>
      '\$${(revenueAtRiskCents / 100).toStringAsFixed(0)}';

  int get usageRevenueCents {
    if (!_demoMode) return _metrics?.usageRevenueCents ?? 0;
    return mockRevenueMix.usageCents;
  }

  String get usageRevenueFormatted =>
      '\$${(usageRevenueCents / 100).toStringAsFixed(0)}';

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
    if (!_demoMode) return _metrics?.mrrTrend ?? [];
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

  ForecastPoint get nextMonthForecast => mockForecast.first;

  List<AiInsight> get recentAlerts => mockInsights.take(3).toList();
}
