import '../../core/network/api_client.dart';
import '../../domain/entities/store_health.dart';
import '../../domain/repositories/store_health_repository.dart';

/// API implementation of StoreHealthRepository
class ApiStoreHealthRepository implements StoreHealthRepository {
  final ApiClient _apiClient;

  ApiStoreHealthRepository(this._apiClient);

  @override
  Future<StoreHealth> getStoreHealth(String appId, String domain) async {
    try {
      final response = await _apiClient.get(
        '/api/v1/apps/$appId/stores/$domain/health',
      );

      if (response.statusCode == 404) {
        throw StoreNotFoundException(domain);
      }

      if (response.statusCode != 200) {
        throw StoreHealthException(
          'Failed to fetch store health: ${response.statusCode}',
        );
      }

      return StoreHealth.fromJson(response.data as Map<String, dynamic>);
    } catch (e) {
      if (e is StoreNotFoundException || e is StoreHealthException) {
        rethrow;
      }
      throw StoreHealthException('Failed to fetch store health: $e');
    }
  }
}
