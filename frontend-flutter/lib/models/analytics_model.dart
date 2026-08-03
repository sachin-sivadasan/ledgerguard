class MrrSnapshot {
  final DateTime date;
  final int mrrCents;

  const MrrSnapshot({required this.date, required this.mrrCents});

  double get mrrDollars => mrrCents / 100;

  factory MrrSnapshot.fromJson(Map<String, dynamic> json) {
    return MrrSnapshot(
      date: DateTime.parse(
          json['date'] as String? ?? DateTime.now().toIso8601String()),
      mrrCents: (json['mrr_cents'] as int?) ??
          (json['active_mrr_cents'] as int?) ??
          0,
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

  factory MrrMovement.fromJson(Map<String, dynamic> json) {
    return MrrMovement(
      month: json['month'] as String? ?? '',
      newCents: json['new_cents'] as int? ?? 0,
      expansionCents: json['expansion_cents'] as int? ?? 0,
      contractionCents: json['contraction_cents'] as int? ?? 0,
      churnedCents: json['churned_cents'] as int? ?? 0,
    );
  }

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

  factory ForecastPoint.fromJson(Map<String, dynamic> json) {
    return ForecastPoint(
      date: DateTime.parse(
          json['date'] as String? ?? DateTime.now().toIso8601String()),
      expected: (json['expected_cents'] as num?)?.toDouble() ?? 0,
      optimistic: (json['optimistic_cents'] as num?)?.toDouble() ?? 0,
      pessimistic: (json['pessimistic_cents'] as num?)?.toDouble() ?? 0,
    );
  }

  // Convert from cents to dollars for display
  double get expectedDollars => expected / 100;
  double get optimisticDollars => optimistic / 100;
  double get pessimisticDollars => pessimistic / 100;
}

class ForecastResult {
  final String model;
  final int dataPointsUsed;
  final List<ForecastPoint> points;

  const ForecastResult({
    required this.model,
    required this.dataPointsUsed,
    required this.points,
  });

  factory ForecastResult.fromJson(Map<String, dynamic> json) {
    return ForecastResult(
      model: json['model'] as String? ?? 'linear',
      dataPointsUsed: json['data_points_used'] as int? ?? 0,
      points: (json['forecast'] as List<dynamic>?)
              ?.map(
                  (e) => ForecastPoint.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
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

  factory CohortData.fromJson(Map<String, dynamic> json) {
    return CohortData(
      cohortMonth: json['cohort_month'] as String? ?? '',
      initialStores: json['initial_stores'] as int? ?? 0,
      retentionPcts: (json['retention_pcts'] as List<dynamic>?)
              ?.map((e) => (e as num).toDouble())
              .toList() ??
          [],
    );
  }
}

class StoreRevenue {
  final String domain;
  final String shopName;
  final int revenueCents;
  final int transactionCount;
  final double pctOfTotal;

  const StoreRevenue({
    required this.domain,
    required this.shopName,
    required this.revenueCents,
    required this.transactionCount,
    required this.pctOfTotal,
  });

  double get revenueDollars => revenueCents / 100;

  factory StoreRevenue.fromJson(Map<String, dynamic> json) {
    return StoreRevenue(
      domain: json['domain'] as String? ?? '',
      shopName: json['shop_name'] as String? ?? '',
      revenueCents: (json['revenue_cents'] as num?)?.toInt() ?? 0,
      transactionCount: (json['transaction_count'] as num?)?.toInt() ?? 0,
      pctOfTotal: (json['pct_of_total'] as num?)?.toDouble() ?? 0,
    );
  }
}

class RevenueConcentration {
  final int totalRevenueCents;
  final List<StoreRevenue> stores;

  const RevenueConcentration({
    required this.totalRevenueCents,
    required this.stores,
  });

  factory RevenueConcentration.fromJson(Map<String, dynamic> json) {
    return RevenueConcentration(
      totalRevenueCents:
          (json['total_revenue_cents'] as num?)?.toInt() ?? 0,
      stores: (json['stores'] as List<dynamic>?)
              ?.map(
                  (e) => StoreRevenue.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

class AppComparison {
  final String id;
  final String name;
  final int mrrCents;
  final int atRiskCents;
  final int subscriptionCount;
  final double renewalRate;

  const AppComparison({
    required this.id,
    required this.name,
    required this.mrrCents,
    this.atRiskCents = 0,
    this.subscriptionCount = 0,
    this.renewalRate = 0,
  });

  double get mrrDollars => mrrCents / 100;

  factory AppComparison.fromJson(Map<String, dynamic> json) {
    return AppComparison(
      id: json['id'] as String? ?? '',
      name: json['name'] as String? ?? '',
      mrrCents: (json['mrr_cents'] as num?)?.toInt() ?? 0,
      atRiskCents: (json['at_risk_cents'] as num?)?.toInt() ?? 0,
      subscriptionCount: (json['subscription_count'] as num?)?.toInt() ?? 0,
      renewalRate: (json['renewal_rate'] as num?)?.toDouble() ?? 0,
    );
  }
}

class ExpenseBreakdown {
  final String month;
  final int grossRevenueCents;
  final int shopifyCutCents;
  final int infraCostCents;
  final int paymentFeesCents;

  /// Fee Guard: expected Shopify cut (gross × the app's revenue-share tier), the
  /// actual−expected variance, and whether they agree within tolerance. `feeGuardOk`
  /// defaults to true so pre-Guard/empty payloads don't render a false alarm.
  final int expectedCutCents;
  final int feeVarianceCents;
  final bool feeGuardOk;
  final double effectiveFeePct;

  const ExpenseBreakdown({
    required this.month,
    required this.grossRevenueCents,
    required this.shopifyCutCents,
    required this.infraCostCents,
    required this.paymentFeesCents,
    this.expectedCutCents = 0,
    this.feeVarianceCents = 0,
    this.feeGuardOk = true,
    this.effectiveFeePct = 0,
  });

  factory ExpenseBreakdown.fromJson(Map<String, dynamic> json) {
    return ExpenseBreakdown(
      month: json['month'] as String? ?? '',
      grossRevenueCents: json['gross_cents'] as int? ?? 0,
      shopifyCutCents: json['shopify_cut_cents'] as int? ?? 0,
      infraCostCents: json['infrastructure_cents'] as int? ?? 0,
      paymentFeesCents: json['processing_fee_cents'] as int? ?? 0,
      expectedCutCents: json['expected_cut_cents'] as int? ?? 0,
      feeVarianceCents: json['fee_variance_cents'] as int? ?? 0,
      feeGuardOk: json['fee_guard_ok'] as bool? ?? true,
      effectiveFeePct: (json['effective_fee_pct'] as num?)?.toDouble() ?? 0,
    );
  }

  int get netProfitCents =>
      grossRevenueCents - shopifyCutCents - infraCostCents - paymentFeesCents;
  double get profitMarginPct =>
      grossRevenueCents > 0 ? netProfitCents / grossRevenueCents * 100 : 0;
}
