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

  factory ShopifyApp.fromJson(Map<String, dynamic> json) {
    return ShopifyApp(
      id: json['id'].toString(),
      name: json['name'] as String? ?? '',
      iconUrl: json['icon_url'] as String? ?? '',
      installCount: json['install_count'] as int? ?? 0,
      avgRating: (json['avg_rating'] as num?)?.toDouble() ?? 0.0,
      revenueShareTier: _parseRevenueShareTier(json['revenue_share_tier']),
      lastSyncAt: json['last_sync_at'] != null
          ? DateTime.tryParse(json['last_sync_at'].toString())
          : null,
      syncStatus: json['sync_status'] as String? ?? 'synced',
    );
  }

  static double _parseRevenueShareTier(dynamic value) {
    if (value == null) return 0.80;
    if (value is num) return value.toDouble();
    // Backend sends string like "STANDARD", "REDUCED_85", etc.
    final s = value.toString().toUpperCase();
    if (s.contains('85')) return 0.85;
    if (s.contains('80') || s == 'STANDARD') return 0.80;
    return 0.80;
  }
}
