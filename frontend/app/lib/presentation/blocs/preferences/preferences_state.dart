import 'package:equatable/equatable.dart';

import '../../../domain/entities/dashboard_preferences.dart';

/// Base class for preferences states
abstract class PreferencesState extends Equatable {
  const PreferencesState();

  @override
  List<Object?> get props => [];
}

/// Initial state before loading
class PreferencesInitial extends PreferencesState {
  const PreferencesInitial();
}

/// Loading preferences
class PreferencesLoading extends PreferencesState {
  const PreferencesLoading();
}

/// Preferences loaded successfully
class PreferencesLoaded extends PreferencesState {
  final DashboardPreferences preferences;
  final bool isSaving;
  final bool hasUnsavedChanges;

  const PreferencesLoaded({
    required this.preferences,
    this.isSaving = false,
    this.hasUnsavedChanges = false,
  });

  PreferencesLoaded copyWith({
    DashboardPreferences? preferences,
    bool? isSaving,
    bool? hasUnsavedChanges,
  }) {
    return PreferencesLoaded(
      preferences: preferences ?? this.preferences,
      isSaving: isSaving ?? this.isSaving,
      hasUnsavedChanges: hasUnsavedChanges ?? this.hasUnsavedChanges,
    );
  }

  @override
  List<Object?> get props => [preferences, isSaving, hasUnsavedChanges];
}

/// Preferences saved successfully
class PreferencesSaved extends PreferencesState {
  final DashboardPreferences preferences;

  const PreferencesSaved({required this.preferences});

  @override
  List<Object?> get props => [preferences];
}

/// Error loading or saving preferences
class PreferencesError extends PreferencesState {
  final String message;
  final DashboardPreferences? preferences;

  const PreferencesError(this.message, {this.preferences});

  @override
  List<Object?> get props => [message, preferences];
}
