import 'package:shared_preferences/shared_preferences.dart';

import '../../core/network/api_client.dart';
import '../../domain/entities/shopify_app.dart';
import '../../domain/repositories/app_repository.dart';

/// API implementation of AppRepository
class ApiAppRepository implements AppRepository {
  final ApiClient _apiClient;
  static const _selectedAppKey = 'selected_app_id';
  static const _selectedAppNameKey = 'selected_app_name';

  ShopifyApp? _selectedApp;

  ApiAppRepository({required ApiClient apiClient}) : _apiClient = apiClient;

  @override
  Future<List<ShopifyApp>> fetchApps() async {
    try {
      final response = await _apiClient.get('/api/v1/apps/available');

      if (response.statusCode == 200) {
        final data = response.data as Map<String, dynamic>;
        final apps = data['apps'] as List<dynamic>? ?? [];

        return apps.map((app) {
          final appMap = app as Map<String, dynamic>;
          return ShopifyApp(
            id: appMap['id'] as String,
            name: appMap['name'] as String,
            description: '',
            installCount: 0,
          );
        }).toList();
      }

      // Return mock apps for development when API fails
      return _getMockApps();
    } catch (e) {
      // If API fails, return mock apps for development
      return _getMockApps();
    }
  }

  /// Mock apps for development/testing
  List<ShopifyApp> _getMockApps() {
    return const [
      ShopifyApp(
        id: 'gid://partners/App/demo-1',
        name: 'Demo App 1 (Development)',
        description: 'Mock app for testing',
        installCount: 100,
      ),
      ShopifyApp(
        id: 'gid://partners/App/demo-2',
        name: 'Demo App 2 (Development)',
        description: 'Mock app for testing',
        installCount: 50,
      ),
    ];
  }

  @override
  Future<ShopifyApp?> getSelectedApp() async {
    if (_selectedApp != null) {
      return _selectedApp;
    }

    // Try to load from SharedPreferences
    final prefs = await SharedPreferences.getInstance();
    final appId = prefs.getString(_selectedAppKey);
    final appName = prefs.getString(_selectedAppNameKey);

    if (appId != null && appName != null) {
      _selectedApp = ShopifyApp(
        id: appId,
        name: appName,
        description: '',
        installCount: 0,
      );
      return _selectedApp;
    }

    return null;
  }

  @override
  Future<void> saveSelectedApp(ShopifyApp app) async {
    _selectedApp = app;

    // Persist to SharedPreferences
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_selectedAppKey, app.id);
    await prefs.setString(_selectedAppNameKey, app.name);

    // Also notify backend
    try {
      await _apiClient.post(
        '/api/v1/apps/select',
        data: {
          'partner_app_id': app.id,
          'name': app.name,
        },
      );
    } catch (e) {
      // Continue even if backend call fails - we have local selection
    }
  }

  @override
  Future<void> clearSelectedApp() async {
    _selectedApp = null;

    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_selectedAppKey);
    await prefs.remove(_selectedAppNameKey);
  }
}
