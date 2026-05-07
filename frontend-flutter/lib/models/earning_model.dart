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
    return EarningPeriod(
      id: json['id'].toString(),
      month: json['month'] as String? ?? '',
      startDate: DateTime.parse(json['start_date'] as String? ??
          DateTime.now().toIso8601String()),
      endDate: DateTime.parse(
          json['end_date'] as String? ?? DateTime.now().toIso8601String()),
      grossCents: json['gross_cents'] as int? ?? 0,
      shopifyCutCents: json['shopify_cut_cents'] as int? ?? 0,
      netEarningsCents: json['net_earnings_cents'] as int? ?? 0,
      status: _parseStatus(json['status'] as String? ?? 'PENDING'),
      paidOutDate: json['paid_out_date'] != null
          ? DateTime.parse(json['paid_out_date'] as String)
          : null,
    );
  }

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

  String get rateLabel => '${ratePct.toStringAsFixed(0)}%';
  String? get thresholdFormatted => thresholdCents != null
      ? '\$${(thresholdCents! / 100).toStringAsFixed(0)}'
      : null;
}
