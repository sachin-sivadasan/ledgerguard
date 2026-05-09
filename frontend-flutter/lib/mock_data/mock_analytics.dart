import '../models/analytics_model.dart';

final _now = DateTime.now();

// 12 months of daily MRR snapshots (trending upward with realistic dips)
final mockMrrSnapshots = _generateMrrSnapshots();

List<MrrSnapshot> _generateMrrSnapshots() {
  final snapshots = <MrrSnapshot>[];
  int baseMrr = 380000; // Start at $3,800

  for (int day = 365; day >= 0; day--) {
    final date = _now.subtract(Duration(days: day));
    // Upward trend with noise
    final trend = ((365 - day) * 15).round(); // ~$0.15/day growth
    final noise = (day % 7 == 0) ? -5000 : (day % 13 == 0 ? -8000 : 2000);
    final mrr = baseMrr + trend + noise;
    snapshots.add(MrrSnapshot(date: date, mrrCents: mrr));
  }
  return snapshots;
}

final mockRevenueMix = const RevenueMix(
  recurringCents: 548000, // $5,480
  usageCents: 102750,     // $1,027.50
  oneTimeCents: 34250,    // $342.50
);

final mockRiskDistribution = const RiskDistribution(
  safe: 25,
  oneCycle: 8,
  twoCycle: 4,
  churned: 3,
);

final mockMrrMovements = <MrrMovement>[
  const MrrMovement(month: 'Oct', newCents: 15000, expansionCents: 8000, contractionCents: 3000, churnedCents: 6000),
  const MrrMovement(month: 'Nov', newCents: 18000, expansionCents: 5000, contractionCents: 2000, churnedCents: 4000),
  const MrrMovement(month: 'Dec', newCents: 12000, expansionCents: 10000, contractionCents: 4000, churnedCents: 8000),
  const MrrMovement(month: 'Jan', newCents: 22000, expansionCents: 7000, contractionCents: 1000, churnedCents: 5000),
  const MrrMovement(month: 'Feb', newCents: 19000, expansionCents: 9000, contractionCents: 3000, churnedCents: 3000),
  const MrrMovement(month: 'Mar', newCents: 25000, expansionCents: 11000, contractionCents: 2000, churnedCents: 4000),
];

// 12-month forecast (values in cents, consistent with API)
final mockForecast = List.generate(12, (i) {
  final date = DateTime(_now.year, _now.month + i + 1);
  final baseCents = 580000.0 + i * 18000; // $5,800 + $180/mo growth
  return ForecastPoint(
    date: date,
    optimistic: baseCents * 1.15,
    expected: baseCents,
    pessimistic: baseCents * 0.85,
  );
});

// 6 monthly cohorts with retention
final mockCohorts = <CohortData>[
  const CohortData(cohortMonth: 'Oct 2025', initialStores: 12, retentionPcts: [100, 92, 85, 80, 77, 75]),
  const CohortData(cohortMonth: 'Nov 2025', initialStores: 15, retentionPcts: [100, 87, 80, 75, 72]),
  const CohortData(cohortMonth: 'Dec 2025', initialStores: 8,  retentionPcts: [100, 88, 82, 78]),
  const CohortData(cohortMonth: 'Jan 2026', initialStores: 18, retentionPcts: [100, 94, 89]),
  const CohortData(cohortMonth: 'Feb 2026', initialStores: 14, retentionPcts: [100, 86]),
  const CohortData(cohortMonth: 'Mar 2026', initialStores: 11, retentionPcts: [100]),
];

// Monthly expense breakdowns
final mockExpenses = <ExpenseBreakdown>[
  const ExpenseBreakdown(month: 'Oct', grossRevenueCents: 620000, shopifyCutCents: 124000, infraCostCents: 4900, paymentFeesCents: 17980),
  const ExpenseBreakdown(month: 'Nov', grossRevenueCents: 645000, shopifyCutCents: 129000, infraCostCents: 4900, paymentFeesCents: 18705),
  const ExpenseBreakdown(month: 'Dec', grossRevenueCents: 598000, shopifyCutCents: 119600, infraCostCents: 4900, paymentFeesCents: 17342),
  const ExpenseBreakdown(month: 'Jan', grossRevenueCents: 682000, shopifyCutCents: 136400, infraCostCents: 4900, paymentFeesCents: 19778),
  const ExpenseBreakdown(month: 'Feb', grossRevenueCents: 710000, shopifyCutCents: 142000, infraCostCents: 4900, paymentFeesCents: 20590),
  const ExpenseBreakdown(month: 'Mar', grossRevenueCents: 685000, shopifyCutCents: 137000, infraCostCents: 4900, paymentFeesCents: 19865),
];
