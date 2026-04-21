import '../models/insight_model.dart';

final _now = DateTime.now();

final mockInsights = <AiInsight>[
  AiInsight(
    id: 'insight-1',
    date: _now.subtract(const Duration(hours: 2)),
    title: 'MRR Growth Accelerating',
    summary: 'Your MRR grew 4.2% this month, up from 3.1% last month. InventorySync Pro is driving 60% of new revenue. Consider increasing marketing spend on this app.',
    severity: InsightSeverity.info,
  ),
  AiInsight(
    id: 'insight-2',
    date: _now.subtract(const Duration(days: 1)),
    title: '3 Stores At Risk of Churning',
    summary: 'ink-press, jade-spa, and kite-kids have missed 2 billing cycles. Combined at-risk revenue: \$149.97/mo. Recommended action: send personalized re-engagement emails.',
    severity: InsightSeverity.warning,
  ),
  AiInsight(
    id: 'insight-3',
    date: _now.subtract(const Duration(days: 1, hours: 8)),
    title: 'ReviewBoost Churn Spike',
    summary: 'ReviewBoost had 2 cancellations this week, 3x the monthly average. Correlates with a recent 1-star review mentioning slow load times. Investigate performance.',
    severity: InsightSeverity.critical,
  ),
  AiInsight(
    id: 'insight-4',
    date: _now.subtract(const Duration(days: 2)),
    title: 'Usage Revenue Up 18%',
    summary: 'Usage-based charges increased 18% week-over-week. Top contributor: eco-shop.myshopify.com with \$47.50 in usage fees. This store may be ready for a Pro plan upsell.',
    severity: InsightSeverity.info,
  ),
  AiInsight(
    id: 'insight-5',
    date: _now.subtract(const Duration(days: 3)),
    title: 'January Cohort Outperforming',
    summary: 'Stores installed in January have 89% retention at month 2, vs 80% for the December cohort. Investigate what changed in onboarding.',
    severity: InsightSeverity.info,
  ),
];
