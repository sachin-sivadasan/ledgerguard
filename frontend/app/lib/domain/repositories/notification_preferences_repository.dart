import '../entities/notification_preferences.dart';

/// Repository interface for notification preferences
abstract class NotificationPreferencesRepository {
  /// Fetch user's notification preferences
  Future<NotificationPreferences> fetchPreferences();

  /// Save user's notification preferences
  Future<void> savePreferences(NotificationPreferences preferences);
}

/// Base exception for notification preferences repository errors
class NotificationPreferencesException implements Exception {
  final String message;
  final String code;

  const NotificationPreferencesException(this.message, {this.code = 'unknown'});

  @override
  String toString() => 'NotificationPreferencesException: $message (code: $code)';
}

/// Exception when user is not authenticated
class UnauthorizedNotificationPreferencesException extends NotificationPreferencesException {
  const UnauthorizedNotificationPreferencesException()
      : super('User must be authenticated to access notification preferences',
            code: 'unauthorized');
}

/// Exception when preferences fail to save
class SaveNotificationPreferencesException extends NotificationPreferencesException {
  const SaveNotificationPreferencesException([String message = 'Failed to save notification preferences'])
      : super(message, code: 'save-failed');
}

/// Exception when preferences fail to load
class LoadNotificationPreferencesException extends NotificationPreferencesException {
  const LoadNotificationPreferencesException([String message = 'Failed to load notification preferences'])
      : super(message, code: 'load-failed');
}
