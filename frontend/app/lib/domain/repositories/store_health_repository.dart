import '../entities/store_health.dart';

/// Repository interface for store health data
abstract class StoreHealthRepository {
  /// Fetch store health details by domain
  ///
  /// [appId] - The numeric app ID
  /// [domain] - The myshopify domain (e.g., "store-name.myshopify.com")
  Future<StoreHealth> getStoreHealth(String appId, String domain);
}

/// Exception thrown when store health data cannot be fetched
class StoreHealthException implements Exception {
  final String message;
  const StoreHealthException(this.message);

  @override
  String toString() => message;
}

/// Exception thrown when store is not found
class StoreNotFoundException implements Exception {
  final String domain;
  const StoreNotFoundException(this.domain);

  @override
  String toString() => 'Store not found: $domain';
}
