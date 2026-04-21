import 'package:flutter/foundation.dart';
import '../mock_data/mock_analytics.dart';
import '../mock_data/mock_events.dart';
import '../mock_data/mock_insights.dart';
import '../mock_data/mock_subscriptions.dart';
import '../models/analytics_model.dart';
import '../models/event_model.dart';
import '../models/insight_model.dart';
import '../models/subscription_model.dart';
import '../widgets/lg_risk_badge.dart';

enum DashboardTimeRange { thisWeek, thisMonth, lastMonth, threeMonths }

class DashboardProvider extends ChangeNotifier {
  String? _selectedAppId;
  DashboardTimeRange _timeRange = DashboardTimeRange.thisWeek;

  String? get selectedAppId => _selectedAppId;
  DashboardTimeRange get timeRange => _timeRange;

  void setSelectedApp(String? appId) {
    _selectedAppId = appId;
    notifyListeners();
  }

  void setTimeRange(DashboardTimeRange range) {
    _timeRange = range;
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
    return _filteredSubscriptions
        .where((s) => s.riskState == RiskState.safe)
        .fold<int>(0, (sum, s) => sum + s.priceCents);
  }

  String get mrrFormatted => '\$${(mrrCents / 100).toStringAsFixed(0)}';

  double get renewalRate {
    final subs = _filteredSubscriptions;
    final total = subs.length;
    final safe = subs.where((s) => s.riskState == RiskState.safe).length;
    return total > 0 ? safe / total * 100 : 0;
  }

  int get revenueAtRiskCents {
    return _filteredSubscriptions
        .where((s) => s.riskState != RiskState.safe && s.riskState != RiskState.churned)
        .fold<int>(0, (sum, s) => sum + s.priceCents);
  }

  String get revenueAtRiskFormatted =>
      '\$${(revenueAtRiskCents / 100).toStringAsFixed(0)}';

  int get usageRevenueCents => mockRevenueMix.usageCents;
  String get usageRevenueFormatted =>
      '\$${(usageRevenueCents / 100).toStringAsFixed(0)}';

  /// Event totals for the selected time range, filtered by selected app.
  Map<String, int> get activity {
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

  // Chart data
  List<MrrSnapshot> get mrrTrend => mockMrrSnapshots;
  RiskDistribution get riskDistribution => mockRiskDistribution;
  RevenueMix get revenueMix => mockRevenueMix;

  // Forecast summary
  ForecastPoint get nextMonthForecast => mockForecast.first;

  // Recent alerts
  List<AiInsight> get recentAlerts => mockInsights.take(3).toList();
}
