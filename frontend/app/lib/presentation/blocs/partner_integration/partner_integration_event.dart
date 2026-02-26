import 'package:equatable/equatable.dart';

/// Base class for partner integration events
abstract class PartnerIntegrationEvent extends Equatable {
  const PartnerIntegrationEvent();

  @override
  List<Object?> get props => [];
}

/// Request to check current integration status
class CheckIntegrationStatusRequested extends PartnerIntegrationEvent {
  const CheckIntegrationStatusRequested();
}

/// Request to connect via OAuth
class ConnectWithOAuthRequested extends PartnerIntegrationEvent {
  const ConnectWithOAuthRequested();
}

/// Request to save manual token
class SaveManualTokenRequested extends PartnerIntegrationEvent {
  final String partnerId;
  final String apiToken;

  const SaveManualTokenRequested({
    required this.partnerId,
    required this.apiToken,
  });

  @override
  List<Object?> get props => [partnerId, apiToken];
}

/// Request to disconnect integration
class DisconnectRequested extends PartnerIntegrationEvent {
  const DisconnectRequested();
}
