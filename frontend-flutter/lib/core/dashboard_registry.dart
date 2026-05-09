import 'package:flutter/material.dart';
import '../providers/dashboard_provider.dart';

class KpiDefinition {
  final String id;
  final String label;
  final IconData icon;
  final String Function(DashboardProvider dp) valueGetter;
  final String? trend;
  final bool? trendPositive;

  const KpiDefinition({
    required this.id,
    required this.label,
    required this.icon,
    required this.valueGetter,
    this.trend,
    this.trendPositive,
  });
}

class WidgetDefinition {
  final String id;
  final String label;

  const WidgetDefinition({required this.id, required this.label});
}

/// All available KPIs (superset — users choose up to 4)
final List<KpiDefinition> kAllKpis = [
  KpiDefinition(
    id: 'active_mrr',
    label: 'Monthly Recurring Revenue',
    icon: Icons.attach_money,
    valueGetter: (dp) => dp.mrrFormatted,
    trend: '+4.2%',
    trendPositive: true,
  ),
  KpiDefinition(
    id: 'renewal_success_rate',
    label: 'Renewal Rate',
    icon: Icons.autorenew,
    valueGetter: (dp) => '${dp.renewalRate.toStringAsFixed(1)}%',
    trend: '+1.3%',
    trendPositive: true,
  ),
  KpiDefinition(
    id: 'revenue_at_risk',
    label: 'Revenue at Risk',
    icon: Icons.warning_amber,
    valueGetter: (dp) => dp.revenueAtRiskFormatted,
    trend: '-\$120',
    trendPositive: true,
  ),
  KpiDefinition(
    id: 'usage_revenue',
    label: 'Usage Revenue',
    icon: Icons.trending_up,
    valueGetter: (dp) => dp.usageRevenueFormatted,
    trend: '+18%',
    trendPositive: true,
  ),
  KpiDefinition(
    id: 'churned',
    label: 'Churned Subscriptions',
    icon: Icons.cancel,
    valueGetter: (dp) => '${dp.riskDistribution.churned}',
  ),
  KpiDefinition(
    id: 'total_revenue',
    label: 'Total Revenue',
    icon: Icons.account_balance_wallet,
    valueGetter: (dp) => dp.totalRevenueFormatted,
  ),
];

/// All available secondary widgets
const List<WidgetDefinition> kAllWidgets = [
  WidgetDefinition(id: 'mrr_trend', label: 'MRR Trend (12 months)'),
  WidgetDefinition(id: 'risk_distribution_chart', label: 'Risk Distribution'),
  WidgetDefinition(id: 'forecast', label: 'Forecast (Next Month)'),
  WidgetDefinition(id: 'revenue_mix_chart', label: 'Revenue Mix'),
  WidgetDefinition(id: 'weekly_activity', label: 'Activity Summary'),
  WidgetDefinition(id: 'earnings_timeline', label: 'Earnings Overview'),
];

/// Frontend defaults (matches current hardcoded dashboard)
const kDefaultKpiIds = [
  'active_mrr',
  'renewal_success_rate',
  'revenue_at_risk',
  'usage_revenue',
];

const kDefaultWidgetIds = [
  'mrr_trend',
  'risk_distribution_chart',
  'forecast',
  'revenue_mix_chart',
  'weekly_activity',
];

/// Look up a KPI definition by ID
KpiDefinition? lookupKpi(String id) {
  for (final kpi in kAllKpis) {
    if (kpi.id == id) return kpi;
  }
  return null;
}
