enum EarningStatus { pending, available, paidOut }

class EarningPeriod {
  final String id;
  final String month;
  final DateTime startDate;
  final DateTime endDate;
  final int grossCents;
  final int shopifyCutCents;
  final int netEarningsCents;
  final EarningStatus status;
  final DateTime? paidOutDate;

  const EarningPeriod({
    required this.id,
    required this.month,
    required this.startDate,
    required this.endDate,
    required this.grossCents,
    required this.shopifyCutCents,
    required this.netEarningsCents,
    required this.status,
    this.paidOutDate,
  });

  factory EarningPeriod.fromJson(Map<String, dynamic> json) {
    // The live /earnings endpoint returns per-date rows shaped as
    // {date, total_amount_cents} (net earning for that day), while demo/mock
    // and any richer source use {month, start_date, gross_cents, ...}. Accept both.
    final dateStr = json['date'] as String?;
    final fallbackDate = dateStr ?? DateTime.now().toIso8601String();
    return EarningPeriod(
      id: json['id']?.toString() ?? dateStr ?? '',
      month: json['month'] as String? ?? '',
      startDate: DateTime.parse(json['start_date'] as String? ?? fallbackDate),
      endDate: DateTime.parse(json['end_date'] as String? ?? fallbackDate),
      grossCents: json['gross_cents'] as int? ?? 0,
      shopifyCutCents: json['shopify_cut_cents'] as int? ?? 0,
      netEarningsCents:
          json['net_earnings_cents'] as int? ?? json['total_amount_cents'] as int? ?? 0,
      status: _parseStatus(json['status'] as String? ?? 'PENDING'),
      paidOutDate: json['paid_out_date'] != null
          ? DateTime.parse(json['paid_out_date'] as String)
          : null,
    );
  }

  /// True when this row carries a real gross/fee breakdown (demo/rich source).
  /// The live per-date earnings feed only has a net amount, so callers should
  /// render just the net rather than a misleading "Gross: \$0.00 / Shopify: \$0.00".
  bool get hasFeeBreakdown => grossCents > 0 || shopifyCutCents > 0;

  static EarningStatus _parseStatus(String s) {
    switch (s.toUpperCase()) {
      case 'AVAILABLE':
        return EarningStatus.available;
      case 'PAID_OUT':
        return EarningStatus.paidOut;
      default:
        return EarningStatus.pending;
    }
  }

  String get grossFormatted =>
      '\$${(grossCents / 100).toStringAsFixed(2)}';
  String get netFormatted =>
      '\$${(netEarningsCents / 100).toStringAsFixed(2)}';
  String get shopifyCutFormatted =>
      '\$${(shopifyCutCents / 100).toStringAsFixed(2)}';

  String get statusLabel => switch (status) {
        EarningStatus.pending => 'Pending',
        EarningStatus.available => 'Available',
        EarningStatus.paidOut => 'Paid Out',
      };
}

class FeeBreakdown {
  final int grossCents;
  final double shopifyFeePct;
  final int shopifyFeeCents;
  final double processingFeePct;
  final int processingFeeCents;
  final int netCents;

  const FeeBreakdown({
    required this.grossCents,
    required this.shopifyFeePct,
    required this.shopifyFeeCents,
    required this.processingFeePct,
    required this.processingFeeCents,
    required this.netCents,
  });

  String get grossFormatted => '\$${(grossCents / 100).toStringAsFixed(2)}';
  String get shopifyFeeFormatted =>
      '\$${(shopifyFeeCents / 100).toStringAsFixed(2)}';
  String get processingFeeFormatted =>
      '\$${(processingFeeCents / 100).toStringAsFixed(2)}';
  String get netFormatted => '\$${(netCents / 100).toStringAsFixed(2)}';
}

class FeeBreakdownResponse {
  final int grossCents;
  final List<TierFeeBreakdown> tiers;

  const FeeBreakdownResponse({
    required this.grossCents,
    required this.tiers,
  });

  factory FeeBreakdownResponse.fromJson(Map<String, dynamic> json) {
    return FeeBreakdownResponse(
      grossCents: json['gross_cents'] as int? ?? 0,
      tiers: (json['tiers'] as List<dynamic>?)
              ?.map((t) =>
                  TierFeeBreakdown.fromJson(t as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

class TierFeeBreakdown {
  final String tierName;
  final double ratePct;
  final int shopifyFeeCents;
  final int processingFeeCents;
  final int netCents;

  const TierFeeBreakdown({
    required this.tierName,
    required this.ratePct,
    required this.shopifyFeeCents,
    required this.processingFeeCents,
    required this.netCents,
  });

  factory TierFeeBreakdown.fromJson(Map<String, dynamic> json) {
    return TierFeeBreakdown(
      tierName: json['tier_name'] as String? ?? '',
      ratePct: (json['rate_pct'] as num?)?.toDouble() ?? 0,
      shopifyFeeCents: json['shopify_fee_cents'] as int? ?? 0,
      processingFeeCents: json['processing_fee_cents'] as int? ?? 0,
      netCents: json['net_cents'] as int? ?? 0,
    );
  }
}

class FeeSummary {
  final int transactionCount;
  final int grossCents;
  final int shopifyFeeCents;
  final int processingFeeCents;
  final int netCents;
  final int savingsCents;
  final String currentTier;

  const FeeSummary({
    required this.transactionCount,
    required this.grossCents,
    required this.shopifyFeeCents,
    required this.processingFeeCents,
    required this.netCents,
    required this.savingsCents,
    required this.currentTier,
  });

  factory FeeSummary.fromJson(Map<String, dynamic> json) {
    return FeeSummary(
      transactionCount: json['transaction_count'] as int? ?? 0,
      grossCents: json['gross_cents'] as int? ?? 0,
      shopifyFeeCents: json['shopify_fee_cents'] as int? ?? 0,
      processingFeeCents: json['processing_fee_cents'] as int? ?? 0,
      netCents: json['net_cents'] as int? ?? 0,
      savingsCents: json['savings_cents'] as int? ?? 0,
      currentTier: json['current_tier'] as String? ?? '',
    );
  }

  String get savingsFormatted =>
      '\$${(savingsCents / 100).toStringAsFixed(2)}';
  String get netFormatted => '\$${(netCents / 100).toStringAsFixed(2)}';
}

class EarningsStatus {
  final int pendingCents;
  final int availableCents;
  final int paidOutCents;
  final List<UpcomingAvailability> upcoming;

  const EarningsStatus({
    required this.pendingCents,
    required this.availableCents,
    required this.paidOutCents,
    required this.upcoming,
  });

  factory EarningsStatus.fromJson(Map<String, dynamic> json) {
    // Live API keys: total_pending_cents / total_available_cents /
    // total_paid_out_cents / upcoming_availability. Fall back to the shorter
    // names for backward/demo compatibility.
    return EarningsStatus(
      pendingCents:
          json['total_pending_cents'] as int? ?? json['pending_cents'] as int? ?? 0,
      availableCents:
          json['total_available_cents'] as int? ?? json['available_cents'] as int? ?? 0,
      paidOutCents:
          json['total_paid_out_cents'] as int? ?? json['paid_out_cents'] as int? ?? 0,
      upcoming: ((json['upcoming_availability'] ?? json['upcoming'])
                  as List<dynamic>?)
              ?.map((u) =>
                  UpcomingAvailability.fromJson(u as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  String get pendingFormatted =>
      '\$${(pendingCents / 100).toStringAsFixed(2)}';
  String get availableFormatted =>
      '\$${(availableCents / 100).toStringAsFixed(2)}';
  String get paidOutFormatted =>
      '\$${(paidOutCents / 100).toStringAsFixed(2)}';
}

class UpcomingAvailability {
  final DateTime date;
  final int amountCents;

  const UpcomingAvailability({
    required this.date,
    required this.amountCents,
  });

  factory UpcomingAvailability.fromJson(Map<String, dynamic> json) {
    return UpcomingAvailability(
      date: DateTime.parse(
          json['date'] as String? ?? DateTime.now().toIso8601String()),
      amountCents: json['amount_cents'] as int? ?? 0,
    );
  }

  String get amountFormatted =>
      '\$${(amountCents / 100).toStringAsFixed(2)}';
}

class RevenueShareTier {
  final String name;
  final String description;
  final int? thresholdCents;
  final double ratePct;
  final bool isCurrentTier;

  const RevenueShareTier({
    required this.name,
    required this.description,
    this.thresholdCents,
    required this.ratePct,
    required this.isCurrentTier,
  });

  factory RevenueShareTier.fromJson(Map<String, dynamic> json,
      {String? currentTierName}) {
    final name = json['name'] as String? ?? '';
    return RevenueShareTier(
      name: name,
      description: json['description'] as String? ?? '',
      thresholdCents: json['threshold_cents'] as int?,
      ratePct: (json['rate_pct'] as num?)?.toDouble() ?? 0,
      isCurrentTier: currentTierName != null
          ? name == currentTierName
          : json['is_current_tier'] as bool? ?? false,
    );
  }

  String get rateLabel => '${ratePct.toStringAsFixed(0)}%';
  String? get thresholdFormatted => thresholdCents != null
      ? '\$${(thresholdCents! / 100).toStringAsFixed(0)}'
      : null;
}
