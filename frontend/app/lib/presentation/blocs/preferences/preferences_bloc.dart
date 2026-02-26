import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../domain/entities/dashboard_preferences.dart';
import '../../../domain/repositories/dashboard_preferences_repository.dart';
import 'preferences_event.dart';
import 'preferences_state.dart';

/// Bloc for managing dashboard preferences
class PreferencesBloc extends Bloc<PreferencesEvent, PreferencesState> {
  final DashboardPreferencesRepository _repository;

  PreferencesBloc({
    required DashboardPreferencesRepository repository,
  })  : _repository = repository,
        super(const PreferencesInitial()) {
    on<LoadPreferencesRequested>(_onLoadPreferences);
    on<AddPrimaryKpiRequested>(_onAddPrimaryKpi);
    on<RemovePrimaryKpiRequested>(_onRemovePrimaryKpi);
    on<ReorderPrimaryKpiRequested>(_onReorderPrimaryKpi);
    on<ToggleSecondaryWidgetRequested>(_onToggleSecondaryWidget);
    on<SavePreferencesRequested>(_onSavePreferences);
    on<ResetPreferencesRequested>(_onResetPreferences);
  }

  Future<void> _onLoadPreferences(
    LoadPreferencesRequested event,
    Emitter<PreferencesState> emit,
  ) async {
    emit(const PreferencesLoading());

    try {
      final preferences = await _repository.fetchPreferences();
      emit(PreferencesLoaded(preferences: preferences));
    } on DashboardPreferencesException catch (e) {
      emit(PreferencesError(e.message));
    } catch (e) {
      emit(PreferencesError('Failed to load preferences: $e'));
    }
  }

  Future<void> _onAddPrimaryKpi(
    AddPrimaryKpiRequested event,
    Emitter<PreferencesState> emit,
  ) async {
    final currentState = state;
    if (currentState is PreferencesLoaded) {
      final newPreferences = currentState.preferences.addPrimaryKpi(event.kpi);
      emit(currentState.copyWith(
        preferences: newPreferences,
        hasUnsavedChanges: true,
      ));
    }
  }

  Future<void> _onRemovePrimaryKpi(
    RemovePrimaryKpiRequested event,
    Emitter<PreferencesState> emit,
  ) async {
    final currentState = state;
    if (currentState is PreferencesLoaded) {
      final newPreferences =
          currentState.preferences.removePrimaryKpi(event.kpi);
      emit(currentState.copyWith(
        preferences: newPreferences,
        hasUnsavedChanges: true,
      ));
    }
  }

  Future<void> _onReorderPrimaryKpi(
    ReorderPrimaryKpiRequested event,
    Emitter<PreferencesState> emit,
  ) async {
    final currentState = state;
    if (currentState is PreferencesLoaded) {
      final newPreferences = currentState.preferences.reorderPrimaryKpi(
        event.oldIndex,
        event.newIndex,
      );
      emit(currentState.copyWith(
        preferences: newPreferences,
        hasUnsavedChanges: true,
      ));
    }
  }

  Future<void> _onToggleSecondaryWidget(
    ToggleSecondaryWidgetRequested event,
    Emitter<PreferencesState> emit,
  ) async {
    final currentState = state;
    if (currentState is PreferencesLoaded) {
      final newPreferences =
          currentState.preferences.toggleSecondaryWidget(event.widget);
      emit(currentState.copyWith(
        preferences: newPreferences,
        hasUnsavedChanges: true,
      ));
    }
  }

  Future<void> _onSavePreferences(
    SavePreferencesRequested event,
    Emitter<PreferencesState> emit,
  ) async {
    final currentState = state;
    if (currentState is PreferencesLoaded) {
      emit(currentState.copyWith(isSaving: true));

      try {
        await _repository.savePreferences(currentState.preferences);
        emit(PreferencesSaved(preferences: currentState.preferences));
        // Transition back to loaded state without unsaved changes
        emit(PreferencesLoaded(
          preferences: currentState.preferences,
          hasUnsavedChanges: false,
        ));
      } on DashboardPreferencesException catch (e) {
        emit(PreferencesError(
          e.message,
          preferences: currentState.preferences,
        ));
        // Restore loaded state with unsaved changes
        emit(currentState.copyWith(isSaving: false));
      } catch (e) {
        emit(PreferencesError(
          'Failed to save preferences: $e',
          preferences: currentState.preferences,
        ));
        // Restore loaded state with unsaved changes
        emit(currentState.copyWith(isSaving: false));
      }
    }
  }

  Future<void> _onResetPreferences(
    ResetPreferencesRequested event,
    Emitter<PreferencesState> emit,
  ) async {
    final currentState = state;
    if (currentState is PreferencesLoaded) {
      emit(currentState.copyWith(
        preferences: DashboardPreferences.defaults(),
        hasUnsavedChanges: true,
      ));
    }
  }
}
