import 'package:equatable/equatable.dart';

import '../../../domain/entities/shopify_app.dart';

/// Base class for app selection states
abstract class AppSelectionState extends Equatable {
  const AppSelectionState();

  @override
  List<Object?> get props => [];
}

/// Initial state
class AppSelectionInitial extends AppSelectionState {
  const AppSelectionInitial();
}

/// Loading apps from backend
class AppSelectionLoading extends AppSelectionState {
  const AppSelectionLoading();
}

/// Apps loaded, ready for selection
class AppSelectionLoaded extends AppSelectionState {
  final List<ShopifyApp> apps;
  final ShopifyApp? selectedApp;

  const AppSelectionLoaded({
    required this.apps,
    this.selectedApp,
  });

  bool get hasSelection => selectedApp != null;

  AppSelectionLoaded copyWith({
    List<ShopifyApp>? apps,
    ShopifyApp? selectedApp,
  }) {
    return AppSelectionLoaded(
      apps: apps ?? this.apps,
      selectedApp: selectedApp ?? this.selectedApp,
    );
  }

  @override
  List<Object?> get props => [apps, selectedApp];
}

/// Saving selection
class AppSelectionSaving extends AppSelectionState {
  final List<ShopifyApp> apps;
  final ShopifyApp selectedApp;

  const AppSelectionSaving({
    required this.apps,
    required this.selectedApp,
  });

  @override
  List<Object?> get props => [apps, selectedApp];
}

/// Selection confirmed and saved
class AppSelectionConfirmed extends AppSelectionState {
  final ShopifyApp selectedApp;

  const AppSelectionConfirmed(this.selectedApp);

  @override
  List<Object?> get props => [selectedApp];
}

/// Error fetching or saving
class AppSelectionError extends AppSelectionState {
  final String message;
  final List<ShopifyApp>? apps;

  const AppSelectionError(this.message, {this.apps});

  @override
  List<Object?> get props => [message, apps];
}
