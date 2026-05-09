class MrrSnapshot {
  final DateTime date;
  final int mrrCents;

  const MrrSnapshot({required this.date, required this.mrrCents});

  double get mrrDollars => mrrCents / 100;

  factory MrrSnapshot.fromJson(Map<String, dynamic> json) {
    return MrrSnapshot(
      date: DateTime.parse(
          json['date'] as String? ?? DateTime.now().toIso8601String()),
      mrrCents: json['mrr_cents'] as int? ?? 0,
    );
  }
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

  factory RevenueMix.fromJson(Map<String, dynamic> json) {
    return RevenueMix(
      recurringCents: json['recurring_cents'] as int? ?? 0,
      usageCents: json['usage_cents'] as int? ?? 0,
      oneTimeCents: json['one_time_cents'] as int? ?? 0,
    );
  }
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

  factory RiskDistribution.fromJson(Map<String, dynamic> json) {
    return RiskDistribution(
      safe: json['safe'] as int? ?? 0,
      oneCycle: json['one_cycle'] as int? ?? 0,
      twoCycle: json['two_cycle'] as int? ?? 0,
      churned: json['churned'] as int? ?? 0,
    );
  }
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

  factory ExpenseBreakdown.fromJson(Map<String, dynamic> json) {
    return ExpenseBreakdown(
      month: json['month'] as String? ?? '',
      grossRevenueCents: json['gross_cents'] as int? ?? 0,
      shopifyCutCents: json['shopify_cut_cents'] as int? ?? 0,
      infraCostCents: json['infrastructure_cents'] as int? ?? 0,
      paymentFeesCents: json['processing_fee_cents'] as int? ?? 0,
    );
  }

  int get netProfitCents =>
      grossRevenueCents - shopifyCutCents - infraCostCents - paymentFeesCents;
  double get profitMarginPct =>
      grossRevenueCents > 0 ? netProfitCents / grossRevenueCents * 100 : 0;
}
