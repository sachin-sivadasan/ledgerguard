import 'package:equatable/equatable.dart';

import '../../../domain/entities/store_health.dart';

/// States for StoreHealthBloc
abstract class StoreHealthState extends Equatable {
  const StoreHealthState();

  @override
  List<Object?> get props => [];
}

/// Initial state
class StoreHealthInitial extends StoreHealthState {
  const StoreHealthInitial();
}

/// Loading state
class StoreHealthLoading extends StoreHealthState {
  const StoreHealthLoading();
}

/// Loaded state with data
class StoreHealthLoaded extends StoreHealthState {
  final StoreHealth storeHealth;
  final bool isRefreshing;

  const StoreHealthLoaded({
    required this.storeHealth,
    this.isRefreshing = false,
  });

  StoreHealthLoaded copyWith({
    StoreHealth? storeHealth,
    bool? isRefreshing,
  }) {
    return StoreHealthLoaded(
      storeHealth: storeHealth ?? this.storeHealth,
      isRefreshing: isRefreshing ?? this.isRefreshing,
    );
  }

  @override
  List<Object?> get props => [storeHealth, isRefreshing];
}

/// Error state
class StoreHealthError extends StoreHealthState {
  final String message;

  const StoreHealthError(this.message);

  @override
  List<Object?> get props => [message];
}

/// Store not found state
class StoreHealthNotFound extends StoreHealthState {
  final String domain;

  const StoreHealthNotFound(this.domain);

  @override
  List<Object?> get props => [domain];
}
