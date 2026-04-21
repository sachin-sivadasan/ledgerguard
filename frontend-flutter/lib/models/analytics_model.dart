class MrrSnapshot {
  final DateTime date;
  final int mrrCents;

  const MrrSnapshot({required this.date, required this.mrrCents});

  double get mrrDollars => mrrCents / 100;
}

class RevenueMix {
  final int recurringCents;
  final int usageCents;
  final int oneTimeCents;

  const RevenueMix({
    required this.recurringCents,
    required this.usageCents,
    required this.oneTimeCents,
  });

  int get totalCents => recurringCents + usageCents + oneTimeCents;
  double get recurringPct =>
      totalCents > 0 ? recurringCents / totalCents * 100 : 0;
  double get usagePct =>
      totalCents > 0 ? usageCents / totalCents * 100 : 0;
  double get oneTimePct =>
      totalCents > 0 ? oneTimeCents / totalCents * 100 : 0;
}

class RiskDistribution {
  final int safe;
  final int oneCycle;
  final int twoCycle;
  final int churned;

  const RiskDistribution({
    required this.safe,
    required this.oneCycle,
    required this.twoCycle,
    required this.churned,
  });

  int get total => safe + oneCycle + twoCycle + churned;
}

class MrrMovement {
  final String month;
  final int newCents;
  final int expansionCents;
  final int contractionCents;
  final int churnedCents;

  const MrrMovement({
    required this.month,
    required this.newCents,
    required this.expansionCents,
    required this.contractionCents,
    required this.churnedCents,
  });

  int get netCents =>
      newCents + expansionCents - contractionCents - churnedCents;
}

class ForecastPoint {
  final DateTime date;
  final double optimistic;
  final double expected;
  final double pessimistic;

  const ForecastPoint({
    required this.date,
    required this.optimistic,
    required this.expected,
    required this.pessimistic,
  });
}

class CohortData {
  final String cohortMonth;
  final int initialStores;
  final List<double> retentionPcts;

  const CohortData({
    required this.cohortMonth,
    required this.initialStores,
    required this.retentionPcts,
  });
}

class ExpenseBreakdown {
  final String month;
  final int grossRevenueCents;
  final int shopifyCutCents;
  final int infraCostCents;
  final int paymentFeesCents;

  const ExpenseBreakdown({
    required this.month,
    required this.grossRevenueCents,
    required this.shopifyCutCents,
    required this.infraCostCents,
    required this.paymentFeesCents,
  });

  int get netProfitCents =>
      grossRevenueCents - shopifyCutCents - infraCostCents - paymentFeesCents;
  double get profitMarginPct =>
      grossRevenueCents > 0 ? netProfitCents / grossRevenueCents * 100 : 0;
}
