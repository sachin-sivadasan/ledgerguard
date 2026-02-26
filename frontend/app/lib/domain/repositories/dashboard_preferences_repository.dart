import '../entities/dashboard_preferences.dart';

/// Repository interface for dashboard preferences
abstract class DashboardPreferencesRepository {
  /// Fetch user's dashboard preferences from the backend
  Future<DashboardPreferences> fetchPreferences();

  /// Save user's dashboard preferences to the backend
  Future<void> savePreferences(DashboardPreferences preferences);
}

/// Base exception for dashboard preferences errors
class DashboardPreferencesException implements Exception {
  final String message;
  final String code;

  const DashboardPreferencesException(this.message, {this.code = 'unknown'});

  @override
  String toString() => 'DashboardPreferencesException: $message (code: $code)';
}

/// Exception when fetching preferences fails
class FetchPreferencesException extends DashboardPreferencesException {
  const FetchPreferencesException([String message = 'Failed to fetch preferences'])
      : super(message, code: 'fetch-failed');
}

/// Exception when saving preferences fails
class SavePreferencesException extends DashboardPreferencesException {
  const SavePreferencesException([String message = 'Failed to save preferences'])
      : super(message, code: 'save-failed');
}

/// Exception when user is not authenticated
class UnauthorizedPreferencesException extends DashboardPreferencesException {
  const UnauthorizedPreferencesException()
      : super('User must be authenticated to access preferences',
            code: 'unauthorized');
}
