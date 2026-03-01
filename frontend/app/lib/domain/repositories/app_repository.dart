import '../entities/shopify_app.dart';
import '../entities/revenue_share_tier.dart';

/// Repository interface for Shopify app operations
abstract class AppRepository {
  /// Fetch available apps from Shopify Partner API (for app selection)
  Future<List<ShopifyApp>> fetchAvailableApps();

  /// Fetch tracked apps from backend (already selected apps)
  Future<List<ShopifyApp>> fetchTrackedApps();

  /// Get the currently selected app (from local storage)
  Future<ShopifyApp?> getSelectedApp();

  /// Save the selected app locally
  Future<void> saveSelectedApp(ShopifyApp app);

  /// Clear the selected app
  Future<void> clearSelectedApp();

  /// Update the revenue share tier for an app
  Future<ShopifyApp> updateAppTier(String appId, RevenueShareTier tier);

  /// Get fee summary for an app
  Future<FeeSummary> getFeeSummary(String appId, {DateTime? start, DateTime? end});

  /// Get fee breakdown for a hypothetical amount
  Future<FeeBreakdownResponse> getFeeBreakdown(String appId, {int amountCents = 4900});

  /// Trigger a data sync from Shopify Partner API
  Future<SyncResult> syncData();
}

/// Result of a sync operation
class SyncResult {
  final int transactionCount;
  final DateTime syncedAt;
  final String? appName;
  final String? error;

  const SyncResult({
    required this.transactionCount,
    required this.syncedAt,
    this.appName,
    this.error,
  });

  bool get isSuccess => error == null;
}

/// Fee summary response from API
class FeeSummary {
  final int transactionCount;
  final int totalGrossCents;
  final int totalRevenueShareCents;
  final int totalProcessingFeeCents;
  final int totalTaxOnFeesCents;
  final int totalFeesCents;
  final int totalNetCents;
  final double avgRevenueSharePct;
  final double avgProcessingFeePct;
  final double effectiveFeePct;
  final TierSavings savings;

  const FeeSummary({
    required this.transactionCount,
    required this.totalGrossCents,
    required this.totalRevenueShareCents,
    required this.totalProcessingFeeCents,
    required this.totalTaxOnFeesCents,
    required this.totalFeesCents,
    required this.totalNetCents,
    required this.avgRevenueSharePct,
    required this.avgProcessingFeePct,
    required this.effectiveFeePct,
    required this.savings,
  });

  factory FeeSummary.fromJson(Map<String, dynamic> json) {
    final summary = json['summary'] as Map<String, dynamic>;
    final savings = json['savings'] as Map<String, dynamic>;
    return FeeSummary(
      transactionCount: summary['transaction_count'] ?? 0,
      totalGrossCents: summary['total_gross_cents'] ?? 0,
      totalRevenueShareCents: summary['total_revenue_share_cents'] ?? 0,
      totalProcessingFeeCents: summary['total_processing_fee_cents'] ?? 0,
      totalTaxOnFeesCents: summary['total_tax_on_fees_cents'] ?? 0,
      totalFeesCents: summary['total_fees_cents'] ?? 0,
      totalNetCents: summary['total_net_cents'] ?? 0,
      avgRevenueSharePct: (summary['avg_revenue_share_pct'] ?? 0).toDouble(),
      avgProcessingFeePct: (summary['avg_processing_fee_pct'] ?? 0).toDouble(),
      effectiveFeePct: (summary['effective_fee_pct'] ?? 0).toDouble(),
      savings: TierSavings.fromJson(savings),
    );
  }
}

/// Tier savings compared to default 20%
class TierSavings {
  final int defaultFeesCents;
  final int currentFeesCents;
  final int savingsCents;
  final double savingsPct;

  const TierSavings({
    required this.defaultFeesCents,
    required this.currentFeesCents,
    required this.savingsCents,
    required this.savingsPct,
  });

  factory TierSavings.fromJson(Map<String, dynamic> json) {
    return TierSavings(
      defaultFeesCents: json['default_fees_cents'] ?? 0,
      currentFeesCents: json['current_fees_cents'] ?? 0,
      savingsCents: json['savings_cents'] ?? 0,
      savingsPct: (json['savings_pct'] ?? 0).toDouble(),
    );
  }
}

/// Fee breakdown response with all tiers
class FeeBreakdownResponse {
  final int amountCents;
  final FeeBreakdown currentTierBreakdown;
  final List<TierBreakdown> allTiers;

  const FeeBreakdownResponse({
    required this.amountCents,
    required this.currentTierBreakdown,
    required this.allTiers,
  });
}

/// Fee breakdown for a specific tier
class TierBreakdown {
  final RevenueShareTier tier;
  final bool isCurrent;
  final FeeBreakdown breakdown;

  const TierBreakdown({
    required this.tier,
    required this.isCurrent,
    required this.breakdown,
  });
}

/// Exception for app-related errors
class AppException implements Exception {
  final String message;
  final String? code;

  const AppException(this.message, {this.code});

  @override
  String toString() => message;
}

/// No apps found in Partner account
class NoAppsFoundException extends AppException {
  const NoAppsFoundException() : super('No apps found in your Partner account');
}

/// Failed to fetch apps
class FetchAppsException extends AppException {
  const FetchAppsException([String message = 'Failed to fetch apps'])
      : super(message);
}
