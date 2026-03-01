import '../../domain/entities/shopify_app.dart';
import '../../domain/entities/revenue_share_tier.dart';
import '../../domain/repositories/app_repository.dart';

/// Mock implementation of AppRepository for development
class MockAppRepository implements AppRepository {
  ShopifyApp? _selectedApp;

  /// Simulated delay for API calls
  final Duration delay;

  /// Mock apps to return with tier info
  static const List<ShopifyApp> _mockApps = [
    ShopifyApp(
      id: 'app-1',
      name: 'Product Reviews Pro',
      description: 'Collect and display product reviews',
      installCount: 1250,
      revenueShareTier: RevenueShareTier.smallDev0,
    ),
    ShopifyApp(
      id: 'app-2',
      name: 'Inventory Sync',
      description: 'Real-time inventory synchronization',
      installCount: 890,
      revenueShareTier: RevenueShareTier.default20,
    ),
    ShopifyApp(
      id: 'app-3',
      name: 'Email Marketing Suite',
      description: 'Automated email campaigns',
      installCount: 2100,
      revenueShareTier: RevenueShareTier.smallDev15,
    ),
    ShopifyApp(
      id: 'app-4',
      name: 'Analytics Dashboard',
      description: 'Advanced store analytics',
      installCount: 560,
      revenueShareTier: RevenueShareTier.largeDev15,
    ),
  ];

  MockAppRepository({
    this.delay = const Duration(milliseconds: 1000),
  });

  @override
  Future<List<ShopifyApp>> fetchAvailableApps() async {
    await Future.delayed(delay);
    return _mockApps;
  }

  @override
  Future<List<ShopifyApp>> fetchTrackedApps() async {
    await Future.delayed(delay);
    // Return only the selected app if one is selected
    if (_selectedApp != null) {
      return [_selectedApp!];
    }
    return [];
  }

  @override
  Future<ShopifyApp?> getSelectedApp() async {
    await Future.delayed(const Duration(milliseconds: 100));
    return _selectedApp;
  }

  @override
  Future<void> saveSelectedApp(ShopifyApp app) async {
    await Future.delayed(const Duration(milliseconds: 100));
    _selectedApp = app;
  }

  @override
  Future<void> clearSelectedApp() async {
    await Future.delayed(const Duration(milliseconds: 100));
    _selectedApp = null;
  }

  @override
  Future<ShopifyApp> updateAppTier(String appId, RevenueShareTier tier) async {
    await Future.delayed(delay);

    // Update selected app if it matches
    if (_selectedApp?.id == appId) {
      _selectedApp = _selectedApp!.copyWith(revenueShareTier: tier);
    }

    return ShopifyApp(
      id: appId,
      name: _selectedApp?.name ?? 'Mock App',
      revenueShareTier: tier,
    );
  }

  @override
  Future<FeeSummary> getFeeSummary(String appId, {DateTime? start, DateTime? end}) async {
    await Future.delayed(delay);

    // Return mock fee summary
    return const FeeSummary(
      transactionCount: 156,
      totalGrossCents: 784500,       // $7,845.00
      totalRevenueShareCents: 0,      // 0% tier
      totalProcessingFeeCents: 22750, // 2.9% of gross
      totalTaxOnFeesCents: 1820,      // 8% tax on fees
      totalFeesCents: 24570,
      totalNetCents: 759930,          // $7,599.30
      avgRevenueSharePct: 0.0,
      avgProcessingFeePct: 2.9,
      effectiveFeePct: 3.13,
      savings: TierSavings(
        defaultFeesCents: 179835,     // What 20% + 2.9% would be
        currentFeesCents: 24570,
        savingsCents: 155265,          // $1,552.65 saved
        savingsPct: 86.3,
      ),
    );
  }

  @override
  Future<FeeBreakdownResponse> getFeeBreakdown(String appId, {int amountCents = 4900}) async {
    await Future.delayed(delay);

    // Calculate mock breakdowns for all tiers
    final allTiers = RevenueShareTier.values.map((tier) {
      final breakdown = FeeBreakdown.calculate(
        grossAmountCents: amountCents,
        tier: tier,
      );
      return TierBreakdown(
        tier: tier,
        isCurrent: tier == (_selectedApp?.revenueShareTier ?? RevenueShareTier.default20),
        breakdown: breakdown,
      );
    }).toList();

    final currentTier = _selectedApp?.revenueShareTier ?? RevenueShareTier.default20;
    final currentBreakdown = FeeBreakdown.calculate(
      grossAmountCents: amountCents,
      tier: currentTier,
    );

    return FeeBreakdownResponse(
      amountCents: amountCents,
      currentTierBreakdown: currentBreakdown,
      allTiers: allTiers,
    );
  }

  @override
  Future<SyncResult> syncData() async {
    await Future.delayed(delay);

    // Return mock sync result
    return SyncResult(
      transactionCount: 42,
      syncedAt: DateTime.now(),
      appName: _selectedApp?.name ?? 'Mock App',
    );
  }
}
