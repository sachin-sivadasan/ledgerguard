import '../../domain/entities/shopify_app.dart';
import '../../domain/repositories/app_repository.dart';

/// Mock implementation of AppRepository for development
class MockAppRepository implements AppRepository {
  ShopifyApp? _selectedApp;

  /// Simulated delay for API calls
  final Duration delay;

  /// Mock apps to return
  static const List<ShopifyApp> _mockApps = [
    ShopifyApp(
      id: 'app-1',
      name: 'Product Reviews Pro',
      description: 'Collect and display product reviews',
      installCount: 1250,
    ),
    ShopifyApp(
      id: 'app-2',
      name: 'Inventory Sync',
      description: 'Real-time inventory synchronization',
      installCount: 890,
    ),
    ShopifyApp(
      id: 'app-3',
      name: 'Email Marketing Suite',
      description: 'Automated email campaigns',
      installCount: 2100,
    ),
    ShopifyApp(
      id: 'app-4',
      name: 'Analytics Dashboard',
      description: 'Advanced store analytics',
      installCount: 560,
    ),
  ];

  MockAppRepository({
    this.delay = const Duration(milliseconds: 1000),
  });

  @override
  Future<List<ShopifyApp>> fetchApps() async {
    await Future.delayed(delay);
    return _mockApps;
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
}
