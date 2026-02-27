import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../domain/entities/notification_preferences.dart';
import '../../../domain/repositories/notification_preferences_repository.dart';
import 'notification_preferences_event.dart';
import 'notification_preferences_state.dart';

/// Bloc for managing notification preferences
class NotificationPreferencesBloc
    extends Bloc<NotificationPreferencesEvent, NotificationPreferencesState> {
  final NotificationPreferencesRepository _repository;
  NotificationPreferences? _originalPreferences;

  NotificationPreferencesBloc({
    required NotificationPreferencesRepository repository,
  })  : _repository = repository,
        super(const NotificationPreferencesInitial()) {
    on<LoadNotificationPreferencesRequested>(_onLoad);
    on<ToggleCriticalAlertsRequested>(_onToggleCriticalAlerts);
    on<ToggleDailySummaryRequested>(_onToggleDailySummary);
    on<UpdateDailySummaryTimeRequested>(_onUpdateDailySummaryTime);
    on<SaveNotificationPreferencesRequested>(_onSave);
  }

  Future<void> _onLoad(
    LoadNotificationPreferencesRequested event,
    Emitter<NotificationPreferencesState> emit,
  ) async {
    emit(const NotificationPreferencesLoading());

    try {
      final preferences = await _repository.fetchPreferences();
      _originalPreferences = preferences;
      emit(NotificationPreferencesLoaded(preferences: preferences));
    } on NotificationPreferencesException catch (e) {
      emit(NotificationPreferencesError(e.message));
    } catch (e) {
      emit(NotificationPreferencesError('Failed to load preferences: $e'));
    }
  }

  void _onToggleCriticalAlerts(
    ToggleCriticalAlertsRequested event,
    Emitter<NotificationPreferencesState> emit,
  ) {
    final currentState = state;
    if (currentState is NotificationPreferencesLoaded) {
      final newPreferences = currentState.preferences.copyWith(
        criticalAlertsEnabled: event.enabled,
      );
      emit(currentState.copyWith(
        preferences: newPreferences,
        hasUnsavedChanges: _hasChanges(newPreferences),
      ));
    }
  }

  void _onToggleDailySummary(
    ToggleDailySummaryRequested event,
    Emitter<NotificationPreferencesState> emit,
  ) {
    final currentState = state;
    if (currentState is NotificationPreferencesLoaded) {
      final newPreferences = currentState.preferences.copyWith(
        dailySummaryEnabled: event.enabled,
      );
      emit(currentState.copyWith(
        preferences: newPreferences,
        hasUnsavedChanges: _hasChanges(newPreferences),
      ));
    }
  }

  void _onUpdateDailySummaryTime(
    UpdateDailySummaryTimeRequested event,
    Emitter<NotificationPreferencesState> emit,
  ) {
    final currentState = state;
    if (currentState is NotificationPreferencesLoaded) {
      final newPreferences = currentState.preferences.copyWith(
        dailySummaryTime: event.time,
      );
      emit(currentState.copyWith(
        preferences: newPreferences,
        hasUnsavedChanges: _hasChanges(newPreferences),
      ));
    }
  }

  Future<void> _onSave(
    SaveNotificationPreferencesRequested event,
    Emitter<NotificationPreferencesState> emit,
  ) async {
    final currentState = state;
    if (currentState is NotificationPreferencesLoaded) {
      emit(currentState.copyWith(isSaving: true));

      try {
        await _repository.savePreferences(currentState.preferences);
        _originalPreferences = currentState.preferences;
        emit(NotificationPreferencesSaved(preferences: currentState.preferences));
        // Return to loaded state after save
        emit(NotificationPreferencesLoaded(
          preferences: currentState.preferences,
          hasUnsavedChanges: false,
        ));
      } on NotificationPreferencesException catch (e) {
        emit(NotificationPreferencesError(
          e.message,
          previousPreferences: currentState.preferences,
        ));
        // Return to loaded state with unsaved changes
        emit(currentState.copyWith(isSaving: false));
      } catch (e) {
        emit(NotificationPreferencesError(
          'Failed to save preferences: $e',
          previousPreferences: currentState.preferences,
        ));
        emit(currentState.copyWith(isSaving: false));
      }
    }
  }

  bool _hasChanges(NotificationPreferences newPreferences) {
    // If no original preferences set, consider any change as unsaved
    if (_originalPreferences == null) return true;
    return newPreferences != _originalPreferences;
  }
}
