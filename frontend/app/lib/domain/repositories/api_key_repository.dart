import '../entities/api_key.dart';

/// Repository interface for API key management
abstract class ApiKeyRepository {
  /// Get all API keys for the current user
  Future<List<ApiKey>> getApiKeys();

  /// Create a new API key with the given name
  /// Returns the full key (shown only once)
  Future<ApiKeyCreationResult> createApiKey(String name);

  /// Revoke (delete) an API key
  Future<void> revokeApiKey(String keyId);
}

/// Base exception for API key operations
class ApiKeyException implements Exception {
  final String message;
  const ApiKeyException(this.message);

  @override
  String toString() => message;
}

/// Thrown when API key limit is reached
class ApiKeyLimitException extends ApiKeyException {
  const ApiKeyLimitException() : super('API key limit reached. Please revoke an existing key.');
}

/// Thrown when API key is not found
class ApiKeyNotFoundException extends ApiKeyException {
  const ApiKeyNotFoundException() : super('API key not found.');
}

/// Thrown when user is not authorized
class ApiKeyUnauthorizedException extends ApiKeyException {
  const ApiKeyUnauthorizedException() : super('Not authorized to manage API keys.');
}
