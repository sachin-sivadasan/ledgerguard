class ShopifyApp {
  final String id;
  final String name;
  final String iconUrl;
  final int installCount;
  final double avgRating;
  final double revenueShareTier;
  final DateTime? lastSyncAt;
  final String syncStatus;

  const ShopifyApp({
    required this.id,
    required this.name,
    required this.iconUrl,
    required this.installCount,
    required this.avgRating,
    this.revenueShareTier = 0.80,
    this.lastSyncAt,
    this.syncStatus = 'synced',
  });
}
