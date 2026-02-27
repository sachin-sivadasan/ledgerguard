import 'package:equatable/equatable.dart';

import '../../../domain/entities/notification_preferences.dart';

/// Base state for notification preferences
abstract class NotificationPreferencesState extends Equatable {
  const NotificationPreferencesState();

  @override
  List<Object?> get props => [];
}

/// Initial state
class NotificationPreferencesInitial extends NotificationPreferencesState {
  const NotificationPreferencesInitial();
}

/// Loading preferences
class NotificationPreferencesLoading extends NotificationPreferencesState {
  const NotificationPreferencesLoading();
}

/// Preferences loaded successfully
class NotificationPreferencesLoaded extends NotificationPreferencesState {
  final NotificationPreferences preferences;
  final bool isSaving;
  final bool hasUnsavedChanges;

  const NotificationPreferencesLoaded({
    required this.preferences,
    this.isSaving = false,
    this.hasUnsavedChanges = false,
  });

  NotificationPreferencesLoaded copyWith({
    NotificationPreferences? preferences,
    bool? isSaving,
    bool? hasUnsavedChanges,
  }) {
    return NotificationPreferencesLoaded(
      preferences: preferences ?? this.preferences,
      isSaving: isSaving ?? this.isSaving,
      hasUnsavedChanges: hasUnsavedChanges ?? this.hasUnsavedChanges,
    );
  }

  @override
  List<Object?> get props => [preferences, isSaving, hasUnsavedChanges];
}

/// Preferences saved successfully
class NotificationPreferencesSaved extends NotificationPreferencesState {
  final NotificationPreferences preferences;

  const NotificationPreferencesSaved({required this.preferences});

  @override
  List<Object?> get props => [preferences];
}

/// Error loading or saving preferences
class NotificationPreferencesError extends NotificationPreferencesState {
  final String message;
  final NotificationPreferences? previousPreferences;

  const NotificationPreferencesError(this.message, {this.previousPreferences});

  @override
  List<Object?> get props => [message, previousPreferences];
}
