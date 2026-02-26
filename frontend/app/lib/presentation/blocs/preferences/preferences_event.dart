import 'package:equatable/equatable.dart';

import '../../../domain/entities/dashboard_preferences.dart';

/// Base class for preferences events
abstract class PreferencesEvent extends Equatable {
  const PreferencesEvent();

  @override
  List<Object?> get props => [];
}

/// Request to load preferences
class LoadPreferencesRequested extends PreferencesEvent {
  const LoadPreferencesRequested();
}

/// Request to add a primary KPI
class AddPrimaryKpiRequested extends PreferencesEvent {
  final KpiType kpi;

  const AddPrimaryKpiRequested(this.kpi);

  @override
  List<Object?> get props => [kpi];
}

/// Request to remove a primary KPI
class RemovePrimaryKpiRequested extends PreferencesEvent {
  final KpiType kpi;

  const RemovePrimaryKpiRequested(this.kpi);

  @override
  List<Object?> get props => [kpi];
}

/// Request to reorder a primary KPI
class ReorderPrimaryKpiRequested extends PreferencesEvent {
  final int oldIndex;
  final int newIndex;

  const ReorderPrimaryKpiRequested({
    required this.oldIndex,
    required this.newIndex,
  });

  @override
  List<Object?> get props => [oldIndex, newIndex];
}

/// Request to toggle a secondary widget
class ToggleSecondaryWidgetRequested extends PreferencesEvent {
  final SecondaryWidget widget;

  const ToggleSecondaryWidgetRequested(this.widget);

  @override
  List<Object?> get props => [widget];
}

/// Request to save current preferences
class SavePreferencesRequested extends PreferencesEvent {
  const SavePreferencesRequested();
}

/// Request to reset preferences to defaults
class ResetPreferencesRequested extends PreferencesEvent {
  const ResetPreferencesRequested();
}
