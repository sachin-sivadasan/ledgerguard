import '../entities/shopify_app.dart';

/// Repository interface for Shopify app operations
abstract class AppRepository {
  /// Fetch list of apps from the connected Partner account
  Future<List<ShopifyApp>> fetchApps();

  /// Get the currently selected app (from local storage)
  Future<ShopifyApp?> getSelectedApp();

  /// Save the selected app locally
  Future<void> saveSelectedApp(ShopifyApp app);

  /// Clear the selected app
  Future<void> clearSelectedApp();
}

/// Exception for app-related errors
class AppException implements Exception {
  final String message;
  final String? code;

  const AppException(this.message, {this.code});

  @override
  String toString() => message;
}

/// No apps found in Partner account
class NoAppsFoundException extends AppException {
  const NoAppsFoundException() : super('No apps found in your Partner account');
}

/// Failed to fetch apps
class FetchAppsException extends AppException {
  const FetchAppsException([String message = 'Failed to fetch apps'])
      : super(message);
}
