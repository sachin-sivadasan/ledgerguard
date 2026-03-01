import 'package:shared_preferences/shared_preferences.dart';

import '../../core/network/api_client.dart';
import '../../domain/entities/shopify_app.dart';
import '../../domain/entities/revenue_share_tier.dart';
import '../../domain/repositories/app_repository.dart';

/// API implementation of AppRepository
class ApiAppRepository implements AppRepository {
  final ApiClient _apiClient;
  static const _selectedAppKey = 'selected_app_id';
  static const _selectedAppNameKey = 'selected_app_name';
  static const _selectedAppTierKey = 'selected_app_tier';

  ShopifyApp? _selectedApp;

  ApiAppRepository({required ApiClient apiClient}) : _apiClient = apiClient;

  @override
  Future<List<ShopifyApp>> fetchAvailableApps() async {
    try {
      final response = await _apiClient.get('/api/v1/apps/available');

      if (response.statusCode == 200) {
        final data = response.data as Map<String, dynamic>;
        final apps = data['apps'] as List<dynamic>? ?? [];

        return apps.map((app) => _parseAvailableApp(app as Map<String, dynamic>)).toList();
      }

      throw const AppException('Failed to fetch available apps');
    } catch (e) {
      if (e is AppException) rethrow;
      throw AppException('Failed to fetch available apps: $e');
    }
  }

  @override
  Future<List<ShopifyApp>> fetchTrackedApps() async {
    try {
      final response = await _apiClient.get('/api/v1/apps');

      if (response.statusCode == 200) {
        final data = response.data as Map<String, dynamic>;
        final apps = data['apps'] as List<dynamic>? ?? [];

        return apps.map((app) => _parseApp(app as Map<String, dynamic>)).toList();
      }

      // Return mock apps for development when API fails
      return _getMockApps();
    } catch (e) {
      // If API fails, return mock apps for development
      return _getMockApps();
    }
  }

  /// Parse app from /api/v1/apps/available response (Shopify Partner API)
  ShopifyApp _parseAvailableApp(Map<String, dynamic> appMap) {
    return ShopifyApp(
      id: appMap['id'] as String,  // Full GID: gid://partners/App/4599915
      name: appMap['name'] as String,
    );
  }

  /// Parse app from /api/v1/apps response (tracked apps)
  ShopifyApp _parseApp(Map<String, dynamic> appMap) {
    return ShopifyApp(
      id: appMap['id'] as String,
      name: appMap['name'] as String,
      description: appMap['description'] as String? ?? '',
      installCount: appMap['install_count'] as int? ?? 0,
      revenueShareTier: RevenueShareTier.fromCode(
        appMap['revenue_share_tier'] as String?,
      ),
      createdAt: appMap['created_at'] != null
          ? DateTime.tryParse(appMap['created_at'] as String)
          : null,
      updatedAt: appMap['updated_at'] != null
          ? DateTime.tryParse(appMap['updated_at'] as String)
          : null,
    );
  }

  /// Mock apps for development/testing
  List<ShopifyApp> _getMockApps() {
    return const [
      ShopifyApp(
        id: 'gid://partners/App/demo-1',
        name: 'Demo App 1 (Development)',
        description: 'Mock app for testing',
        installCount: 100,
        revenueShareTier: RevenueShareTier.default20,
      ),
      ShopifyApp(
        id: 'gid://partners/App/demo-2',
        name: 'Demo App 2 (Development)',
        description: 'Mock app for testing',
        installCount: 50,
        revenueShareTier: RevenueShareTier.smallDev0,
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
    await prefs.remove(_selectedAppTierKey);
  }

  @override
  Future<ShopifyApp> updateAppTier(String appId, RevenueShareTier tier) async {
    // Extract numeric ID to avoid URL routing issues with slashes in GID
    final numericAppId = _extractNumericId(appId);
    final response = await _apiClient.patch(
      '/api/v1/apps/$numericAppId/tier',
      data: {'revenue_share_tier': tier.code},
    );

    if (response.statusCode == 200) {
      final data = response.data as Map<String, dynamic>;

      // Update cached app if it's the selected one
      if (_selectedApp?.id == appId) {
        _selectedApp = _selectedApp!.copyWith(revenueShareTier: tier);

        // Update SharedPreferences
        final prefs = await SharedPreferences.getInstance();
        await prefs.setString(_selectedAppTierKey, tier.code);
      }

      // Return updated app info from response
      return ShopifyApp(
        id: appId,
        name: _selectedApp?.name ?? '',
        revenueShareTier: RevenueShareTier.fromCode(
          data['revenue_share_tier'] as String?,
        ),
      );
    }

    throw const AppException('Failed to update tier');
  }

  @override
  Future<FeeSummary> getFeeSummary(String appId, {DateTime? start, DateTime? end}) async {
    final numericAppId = _extractNumericId(appId);
    final queryParams = <String, String>{};
    if (start != null) {
      queryParams['start'] = start.toIso8601String().split('T').first;
    }
    if (end != null) {
      queryParams['end'] = end.toIso8601String().split('T').first;
    }

    final response = await _apiClient.get(
      '/api/v1/apps/$numericAppId/fees/summary',
      queryParameters: queryParams,
    );

    if (response.statusCode == 200) {
      return FeeSummary.fromJson(response.data as Map<String, dynamic>);
    }

    throw const AppException('Failed to fetch fee summary');
  }

  @override
  Future<FeeBreakdownResponse> getFeeBreakdown(String appId, {int amountCents = 4900}) async {
    final numericAppId = _extractNumericId(appId);
    final response = await _apiClient.get(
      '/api/v1/apps/$numericAppId/fees/breakdown',
      queryParameters: {'amount_cents': amountCents.toString()},
    );

    if (response.statusCode == 200) {
      final data = response.data as Map<String, dynamic>;
      final currentTier = data['current_tier'] as Map<String, dynamic>;
      final allTiers = data['all_tiers'] as List<dynamic>;

      return FeeBreakdownResponse(
        amountCents: data['amount_cents'] as int,
        currentTierBreakdown: FeeBreakdown(
          grossAmountCents: currentTier['gross_cents'] as int,
          revenueShareCents: currentTier['revenue_share_cents'] as int,
          processingFeeCents: currentTier['processing_fee_cents'] as int,
          taxOnFeesCents: currentTier['tax_on_fees_cents'] as int,
          totalFeesCents: currentTier['total_fees_cents'] as int,
          netAmountCents: currentTier['net_cents'] as int,
          revenueSharePercent: (currentTier['revenue_share_pct'] as num?)?.toDouble() ?? 0,
          processingFeePercent: (currentTier['processing_fee_pct'] as num?)?.toDouble() ?? 2.9,
        ),
        allTiers: allTiers.map((tierData) {
          final td = tierData as Map<String, dynamic>;
          return TierBreakdown(
            tier: RevenueShareTier.fromCode(td['tier'] as String?),
            isCurrent: td['is_current'] as bool? ?? false,
            breakdown: FeeBreakdown(
              grossAmountCents: td['gross_cents'] as int,
              revenueShareCents: td['revenue_share_cents'] as int,
              processingFeeCents: td['processing_fee_cents'] as int,
              taxOnFeesCents: td['tax_on_fees_cents'] as int,
              totalFeesCents: td['total_fees_cents'] as int,
              netAmountCents: td['net_cents'] as int,
              revenueSharePercent: (td['revenue_share_pct'] as num?)?.toDouble() ?? 0,
              processingFeePercent: (td['processing_fee_pct'] as num?)?.toDouble() ?? 2.9,
            ),
          );
        }).toList(),
      );
    }

    throw const AppException('Failed to fetch fee breakdown');
  }

  @override
  Future<SyncResult> syncData() async {
    try {
      final response = await _apiClient.post('/api/v1/sync');

      if (response.statusCode == 200) {
        final data = response.data as Map<String, dynamic>;
        final results = data['results'] as List<dynamic>? ?? [];

        if (results.isEmpty) {
          return SyncResult(
            transactionCount: 0,
            syncedAt: DateTime.now(),
          );
        }

        // Aggregate results from all apps
        int totalTransactions = 0;
        String? lastError;
        String? appName;

        for (final result in results) {
          final r = result as Map<String, dynamic>;
          totalTransactions += (r['transaction_count'] as int?) ?? 0;
          appName = r['app_name'] as String?;
          if (r['error'] != null) {
            lastError = r['error'] as String;
          }
        }

        return SyncResult(
          transactionCount: totalTransactions,
          syncedAt: DateTime.now(),
          appName: appName,
          error: lastError,
        );
      }

      throw const AppException('Failed to sync data');
    } catch (e) {
      return SyncResult(
        transactionCount: 0,
        syncedAt: DateTime.now(),
        error: e.toString(),
      );
    }
  }

  /// Extracts numeric ID from Shopify GID
  /// e.g., "gid://partners/App/4599915" -> "4599915"
  String _extractNumericId(String gid) {
    final parts = gid.split('/');
    return parts.isNotEmpty ? parts.last : gid;
  }
}
