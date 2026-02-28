import 'package:equatable/equatable.dart';

/// Events for StoreHealthBloc
abstract class StoreHealthEvent extends Equatable {
  const StoreHealthEvent();

  @override
  List<Object?> get props => [];
}

/// Load store health data
class LoadStoreHealthRequested extends StoreHealthEvent {
  final String appId;
  final String domain;

  const LoadStoreHealthRequested({
    required this.appId,
    required this.domain,
  });

  @override
  List<Object?> get props => [appId, domain];
}

/// Refresh store health data
class RefreshStoreHealthRequested extends StoreHealthEvent {
  const RefreshStoreHealthRequested();
}
