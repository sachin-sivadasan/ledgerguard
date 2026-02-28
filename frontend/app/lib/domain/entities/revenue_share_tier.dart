import 'package:equatable/equatable.dart';

/// Shopify revenue share tier for an app
/// Based on Shopify's Reduced Revenue Share Plan
enum RevenueShareTier {
  /// Default 20% revenue share (not registered for reduced plan)
  default20('DEFAULT_20', 'Default (20%)', 20.0, 'Not registered for reduced revenue share plan'),

  /// 0% revenue share on first $1M lifetime (small developer)
  smallDev0('SMALL_DEV_0', 'Small Developer (0%)', 0.0, '0% on first \$1M lifetime (under \$1M earned)'),

  /// 15% revenue share after $1M lifetime (small developer)
  smallDev15('SMALL_DEV_15', 'Small Developer (15%)', 15.0, '15% after \$1M lifetime earnings'),

  /// 15% revenue share on all revenue (large developer)
  largeDev15('LARGE_DEV_15', 'Large Developer (15%)', 15.0, '15% on all revenue (large developer)');

  final String code;
  final String displayName;
  final double revenueSharePercent;
  final String description;

  const RevenueShareTier(this.code, this.displayName, this.revenueSharePercent, this.description);

  /// Processing fee is always 2.9% regardless of tier
  static const double processingFeePercent = 2.9;

  /// Whether this is a reduced plan tier
  bool get isReducedPlan => this != default20;

  /// Parse tier from API string
  static RevenueShareTier fromCode(String? code) {
    if (code == null) return default20;
    return RevenueShareTier.values.firstWhere(
      (t) => t.code == code,
      orElse: () => default20,
    );
  }

  /// Badge color for the tier
  int get badgeColor {
    switch (this) {
      case RevenueShareTier.default20:
        return 0xFFEF4444; // Red
      case RevenueShareTier.smallDev0:
        return 0xFF22C55E; // Green
      case RevenueShareTier.smallDev15:
        return 0xFFF59E0B; // Amber
      case RevenueShareTier.largeDev15:
        return 0xFFF59E0B; // Amber
    }
  }
}

/// Fee breakdown for a transaction or gross amount
class FeeBreakdown extends Equatable {
  final int grossAmountCents;
  final int revenueShareCents;
  final int processingFeeCents;
  final int taxOnFeesCents;
  final int totalFeesCents;
  final int netAmountCents;
  final double revenueSharePercent;
  final double processingFeePercent;

  const FeeBreakdown({
    required this.grossAmountCents,
    required this.revenueShareCents,
    required this.processingFeeCents,
    required this.taxOnFeesCents,
    required this.totalFeesCents,
    required this.netAmountCents,
    required this.revenueSharePercent,
    required this.processingFeePercent,
  });

  /// Calculate fee breakdown for a given tier and gross amount
  factory FeeBreakdown.calculate({
    required int grossAmountCents,
    required RevenueShareTier tier,
    double taxRate = 0.08, // Default 8% tax on fees
  }) {
    final revenueShareCents = (grossAmountCents * tier.revenueSharePercent / 100).round();
    final processingFeeCents = (grossAmountCents * RevenueShareTier.processingFeePercent / 100).round();
    final taxOnFeesCents = ((revenueShareCents + processingFeeCents) * taxRate).round();
    final totalFeesCents = revenueShareCents + processingFeeCents + taxOnFeesCents;
    final netAmountCents = grossAmountCents - totalFeesCents;

    return FeeBreakdown(
      grossAmountCents: grossAmountCents,
      revenueShareCents: revenueShareCents,
      processingFeeCents: processingFeeCents,
      taxOnFeesCents: taxOnFeesCents,
      totalFeesCents: totalFeesCents,
      netAmountCents: netAmountCents,
      revenueSharePercent: tier.revenueSharePercent,
      processingFeePercent: RevenueShareTier.processingFeePercent,
    );
  }

  @override
  List<Object?> get props => [
        grossAmountCents,
        revenueShareCents,
        processingFeeCents,
        taxOnFeesCents,
        totalFeesCents,
        netAmountCents,
      ];
}
