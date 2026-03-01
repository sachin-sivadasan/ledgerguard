import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../domain/entities/shopify_app.dart';
import '../../../domain/repositories/app_repository.dart';
import 'app_selection_event.dart';
import 'app_selection_state.dart';

/// Bloc for managing app selection state
class AppSelectionBloc extends Bloc<AppSelectionEvent, AppSelectionState> {
  final AppRepository _appRepository;

  AppSelectionBloc({
    required AppRepository appRepository,
  })  : _appRepository = appRepository,
        super(const AppSelectionInitial()) {
    on<FetchAppsRequested>(_onFetchApps);
    on<AppSelected>(_onAppSelected);
    on<ConfirmSelectionRequested>(_onConfirmSelection);
    on<LoadSelectedAppRequested>(_onLoadSelectedApp);
  }

  Future<void> _onFetchApps(
    FetchAppsRequested event,
    Emitter<AppSelectionState> emit,
  ) async {
    emit(const AppSelectionLoading());

    try {
      final apps = await _appRepository.fetchAvailableApps();

      if (apps.isEmpty) {
        emit(const AppSelectionError('No apps found in your Partner account'));
        return;
      }

      // Check if there's a previously selected app
      final selectedApp = await _appRepository.getSelectedApp();
      final matchingApp = selectedApp != null
          ? apps.where((a) => a.id == selectedApp.id).firstOrNull
          : null;

      emit(AppSelectionLoaded(
        apps: apps,
        selectedApp: matchingApp,
      ));
    } on AppException catch (e) {
      emit(AppSelectionError(e.message));
    } catch (e) {
      emit(AppSelectionError('Failed to fetch apps: $e'));
    }
  }

  void _onAppSelected(
    AppSelected event,
    Emitter<AppSelectionState> emit,
  ) {
    final currentState = state;
    if (currentState is AppSelectionLoaded) {
      emit(currentState.copyWith(selectedApp: event.app));
    }
  }

  Future<void> _onConfirmSelection(
    ConfirmSelectionRequested event,
    Emitter<AppSelectionState> emit,
  ) async {
    final currentState = state;
    if (currentState is AppSelectionLoaded && currentState.hasSelection) {
      final selectedApp = currentState.selectedApp!;

      emit(AppSelectionSaving(
        apps: currentState.apps,
        selectedApp: selectedApp,
      ));

      try {
        await _appRepository.saveSelectedApp(selectedApp);
        emit(AppSelectionConfirmed(selectedApp));
      } on AppException catch (e) {
        emit(AppSelectionError(e.message, apps: currentState.apps));
      } catch (e) {
        emit(AppSelectionError('Failed to save selection: $e', apps: currentState.apps));
      }
    }
  }

  Future<void> _onLoadSelectedApp(
    LoadSelectedAppRequested event,
    Emitter<AppSelectionState> emit,
  ) async {
    try {
      final selectedApp = await _appRepository.getSelectedApp();
      if (selectedApp != null) {
        emit(AppSelectionConfirmed(selectedApp));
      }
    } catch (e) {
      // Silently fail - no previously selected app
    }
  }
}
