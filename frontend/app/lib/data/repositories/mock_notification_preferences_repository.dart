import 'package:flutter/material.dart';

import '../../domain/entities/notification_preferences.dart';
import '../../domain/repositories/notification_preferences_repository.dart';

/// Mock implementation of NotificationPreferencesRepository for development
class MockNotificationPreferencesRepository implements NotificationPreferencesRepository {
  NotificationPreferences _preferences = const NotificationPreferences(
    criticalAlertsEnabled: true,
    dailySummaryEnabled: true,
    dailySummaryTime: TimeOfDay(hour: 9, minute: 0),
  );

  /// Simulated delay for async operations
  final Duration delay;

  /// Whether to simulate errors
  final bool simulateError;

  MockNotificationPreferencesRepository({
    this.delay = const Duration(milliseconds: 500),
    this.simulateError = false,
  });

  @override
  Future<NotificationPreferences> fetchPreferences() async {
    await Future.delayed(delay);

    if (simulateError) {
      throw const LoadNotificationPreferencesException('Simulated error');
    }

    return _preferences;
  }

  @override
  Future<void> savePreferences(NotificationPreferences preferences) async {
    await Future.delayed(delay);

    if (simulateError) {
      throw const SaveNotificationPreferencesException('Simulated error');
    }

    _preferences = preferences;
  }
}
