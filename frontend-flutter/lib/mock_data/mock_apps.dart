import '../models/app_model.dart';

final mockApps = [
  ShopifyApp(
    id: '00000000-0000-0000-0000-000000000001',
    shopifyId: 'app-1',
    name: 'InventorySync Pro',
    iconUrl: 'https://placehold.co/64x64/5C6AC4/white?text=IS',
    installCount: 1247,
    avgRating: 4.6,
    revenueShareTier: 0.80,
    lastSyncAt: DateTime.now().subtract(const Duration(hours: 2)),
    syncStatus: 'synced',
  ),
  ShopifyApp(
    id: '00000000-0000-0000-0000-000000000002',
    shopifyId: 'app-2',
    name: 'ReviewBoost',
    iconUrl: 'https://placehold.co/64x64/008060/white?text=RB',
    installCount: 832,
    avgRating: 4.3,
    revenueShareTier: 0.80,
    lastSyncAt: DateTime.now().subtract(const Duration(hours: 1)),
    syncStatus: 'synced',
  ),
  ShopifyApp(
    id: '00000000-0000-0000-0000-000000000003',
    shopifyId: 'app-3',
    name: 'ShipTracker',
    iconUrl: 'https://placehold.co/64x64/2C6ECB/white?text=ST',
    installCount: 2089,
    avgRating: 4.8,
    revenueShareTier: 0.80,
    lastSyncAt: DateTime.now().subtract(const Duration(minutes: 30)),
    syncStatus: 'synced',
  ),
];
