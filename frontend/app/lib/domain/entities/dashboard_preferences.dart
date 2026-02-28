import 'package:equatable/equatable.dart';
import 'package:flutter/material.dart';

/// Available KPI types for the dashboard
enum KpiType {
  renewalSuccessRate('renewal_success_rate', 'Renewal Success Rate'),
  activeMrr('active_mrr', 'Active MRR'),
  revenueAtRisk('revenue_at_risk', 'Revenue at Risk'),
  churned('churned', 'Churned'),
  usageRevenue('usage_revenue', 'Usage Revenue'),
  totalRevenue('total_revenue', 'Total Revenue');

  final String id;
  final String displayName;

  const KpiType(this.id, this.displayName);

  static KpiType fromId(String id) {
    return KpiType.values.firstWhere(
      (k) => k.id == id,
      orElse: () => KpiType.activeMrr,
    );
  }
}

/// Available secondary widgets for the dashboard
enum SecondaryWidget {
  usageRevenue('usage_revenue', 'Usage Revenue', Icons.data_usage),
  totalRevenue('total_revenue', 'Total Revenue', Icons.account_balance_wallet),
  revenueMixChart('revenue_mix_chart', 'Revenue Mix', Icons.pie_chart_outline),
  riskDistributionChart('risk_distribution_chart', 'Risk Distribution', Icons.donut_small),
  earningsTimeline('earnings_timeline', 'Earnings Timeline', Icons.bar_chart);

  final String id;
  final String displayName;
  final IconData icon;

  const SecondaryWidget(this.id, this.displayName, this.icon);

  static SecondaryWidget fromId(String id) {
    return SecondaryWidget.values.firstWhere(
      (w) => w.id == id,
      orElse: () => SecondaryWidget.revenueMixChart,
    );
  }
}

/// User preferences for dashboard configuration
class DashboardPreferences extends Equatable {
  /// Ordered list of primary KPIs to display (max 4)
  final List<KpiType> primaryKpis;

  /// Set of enabled secondary widgets
  final Set<SecondaryWidget> enabledSecondaryWidgets;

  const DashboardPreferences({
    required this.primaryKpis,
    required this.enabledSecondaryWidgets,
  });

  /// Default dashboard preferences
  factory DashboardPreferences.defaults() {
    return const DashboardPreferences(
      primaryKpis: [
        KpiType.renewalSuccessRate,
        KpiType.activeMrr,
        KpiType.revenueAtRisk,
        KpiType.churned,
      ],
      enabledSecondaryWidgets: {
        SecondaryWidget.usageRevenue,
        SecondaryWidget.totalRevenue,
        SecondaryWidget.revenueMixChart,
        SecondaryWidget.riskDistributionChart,
        SecondaryWidget.earningsTimeline,
      },
    );
  }

  /// Create from JSON map
  factory DashboardPreferences.fromJson(Map<String, dynamic> json) {
    final primaryKpiIds = (json['primary_kpis'] as List<dynamic>?)
            ?.map((e) => e as String)
            .toList() ??
        [];
    final secondaryWidgetIds = (json['secondary_widgets'] as List<dynamic>?)
            ?.map((e) => e as String)
            .toSet() ??
        {};

    return DashboardPreferences(
      primaryKpis: primaryKpiIds.map(KpiType.fromId).toList(),
      enabledSecondaryWidgets:
          secondaryWidgetIds.map(SecondaryWidget.fromId).toSet(),
    );
  }

  /// Convert to JSON map
  Map<String, dynamic> toJson() {
    return {
      'primary_kpis': primaryKpis.map((k) => k.id).toList(),
      'secondary_widgets':
          enabledSecondaryWidgets.map((w) => w.id).toList(),
    };
  }

  /// Create a copy with updated primary KPIs
  DashboardPreferences copyWithPrimaryKpis(List<KpiType> kpis) {
    // Ensure max 4 KPIs
    final limitedKpis = kpis.take(4).toList();
    return DashboardPreferences(
      primaryKpis: limitedKpis,
      enabledSecondaryWidgets: enabledSecondaryWidgets,
    );
  }

  /// Create a copy with a toggled secondary widget
  DashboardPreferences toggleSecondaryWidget(SecondaryWidget widget) {
    final newSet = Set<SecondaryWidget>.from(enabledSecondaryWidgets);
    if (newSet.contains(widget)) {
      newSet.remove(widget);
    } else {
      newSet.add(widget);
    }
    return DashboardPreferences(
      primaryKpis: primaryKpis,
      enabledSecondaryWidgets: newSet,
    );
  }

  /// Check if a secondary widget is enabled
  bool isSecondaryWidgetEnabled(SecondaryWidget widget) {
    return enabledSecondaryWidgets.contains(widget);
  }

  /// Reorder a primary KPI
  DashboardPreferences reorderPrimaryKpi(int oldIndex, int newIndex) {
    final kpis = List<KpiType>.from(primaryKpis);
    if (oldIndex < 0 ||
        oldIndex >= kpis.length ||
        newIndex < 0 ||
        newIndex >= kpis.length) {
      return this;
    }
    final item = kpis.removeAt(oldIndex);
    kpis.insert(newIndex, item);
    return DashboardPreferences(
      primaryKpis: kpis,
      enabledSecondaryWidgets: enabledSecondaryWidgets,
    );
  }

  /// Add a primary KPI (if under max limit)
  DashboardPreferences addPrimaryKpi(KpiType kpi) {
    if (primaryKpis.length >= 4 || primaryKpis.contains(kpi)) {
      return this;
    }
    return DashboardPreferences(
      primaryKpis: [...primaryKpis, kpi],
      enabledSecondaryWidgets: enabledSecondaryWidgets,
    );
  }

  /// Remove a primary KPI
  DashboardPreferences removePrimaryKpi(KpiType kpi) {
    if (!primaryKpis.contains(kpi)) {
      return this;
    }
    return DashboardPreferences(
      primaryKpis: primaryKpis.where((k) => k != kpi).toList(),
      enabledSecondaryWidgets: enabledSecondaryWidgets,
    );
  }

  @override
  List<Object?> get props => [primaryKpis, enabledSecondaryWidgets];
}
