import 'package:equatable/equatable.dart';

import '../../../domain/entities/partner_integration.dart';

/// Base class for partner integration states
abstract class PartnerIntegrationState extends Equatable {
  const PartnerIntegrationState();

  @override
  List<Object?> get props => [];
}

/// Initial state
class PartnerIntegrationInitial extends PartnerIntegrationState {
  const PartnerIntegrationInitial();
}

/// Loading state (checking status, connecting, saving)
class PartnerIntegrationLoading extends PartnerIntegrationState {
  final String? message;

  const PartnerIntegrationLoading({this.message});

  @override
  List<Object?> get props => [message];
}

/// Not connected state
class PartnerIntegrationNotConnected extends PartnerIntegrationState {
  const PartnerIntegrationNotConnected();
}

/// Connected state
class PartnerIntegrationConnected extends PartnerIntegrationState {
  final PartnerIntegration integration;

  const PartnerIntegrationConnected(this.integration);

  @override
  List<Object?> get props => [integration];
}

/// Success state (after action completed)
class PartnerIntegrationSuccess extends PartnerIntegrationState {
  final PartnerIntegration integration;
  final String message;

  const PartnerIntegrationSuccess({
    required this.integration,
    required this.message,
  });

  @override
  List<Object?> get props => [integration, message];
}

/// Error state
class PartnerIntegrationError extends PartnerIntegrationState {
  final String message;

  const PartnerIntegrationError(this.message);

  @override
  List<Object?> get props => [message];
}
