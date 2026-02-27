import 'package:equatable/equatable.dart';
import 'package:flutter/material.dart';

/// Base event for notification preferences
abstract class NotificationPreferencesEvent extends Equatable {
  const NotificationPreferencesEvent();

  @override
  List<Object?> get props => [];
}

/// Load notification preferences
class LoadNotificationPreferencesRequested extends NotificationPreferencesEvent {
  const LoadNotificationPreferencesRequested();
}

/// Toggle critical alerts
class ToggleCriticalAlertsRequested extends NotificationPreferencesEvent {
  final bool enabled;

  const ToggleCriticalAlertsRequested({required this.enabled});

  @override
  List<Object?> get props => [enabled];
}

/// Toggle daily summary
class ToggleDailySummaryRequested extends NotificationPreferencesEvent {
  final bool enabled;

  const ToggleDailySummaryRequested({required this.enabled});

  @override
  List<Object?> get props => [enabled];
}

/// Update daily summary time
class UpdateDailySummaryTimeRequested extends NotificationPreferencesEvent {
  final TimeOfDay time;

  const UpdateDailySummaryTimeRequested({required this.time});

  @override
  List<Object?> get props => [time.hour, time.minute];
}

/// Save notification preferences
class SaveNotificationPreferencesRequested extends NotificationPreferencesEvent {
  const SaveNotificationPreferencesRequested();
}
