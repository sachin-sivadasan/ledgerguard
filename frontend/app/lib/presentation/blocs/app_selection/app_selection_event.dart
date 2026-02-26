import 'package:equatable/equatable.dart';

import '../../../domain/entities/shopify_app.dart';

/// Base class for app selection events
abstract class AppSelectionEvent extends Equatable {
  const AppSelectionEvent();

  @override
  List<Object?> get props => [];
}

/// Request to fetch available apps
class FetchAppsRequested extends AppSelectionEvent {
  const FetchAppsRequested();
}

/// User selected an app from the list
class AppSelected extends AppSelectionEvent {
  final ShopifyApp app;

  const AppSelected(this.app);

  @override
  List<Object?> get props => [app];
}

/// User confirmed their app selection
class ConfirmSelectionRequested extends AppSelectionEvent {
  const ConfirmSelectionRequested();
}

/// Load previously selected app
class LoadSelectedAppRequested extends AppSelectionEvent {
  const LoadSelectedAppRequested();
}
