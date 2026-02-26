import 'package:equatable/equatable.dart';

/// Status of partner integration
enum IntegrationStatus {
  notConnected,
  connecting,
  connected,
  error,
}

/// Partner integration entity
class PartnerIntegration extends Equatable {
  final String? partnerId;
  final IntegrationStatus status;
  final String? errorMessage;
  final DateTime? connectedAt;

  const PartnerIntegration({
    this.partnerId,
    this.status = IntegrationStatus.notConnected,
    this.errorMessage,
    this.connectedAt,
  });

  bool get isConnected => status == IntegrationStatus.connected;
  bool get isConnecting => status == IntegrationStatus.connecting;
  bool get hasError => status == IntegrationStatus.error;

  PartnerIntegration copyWith({
    String? partnerId,
    IntegrationStatus? status,
    String? errorMessage,
    DateTime? connectedAt,
  }) {
    return PartnerIntegration(
      partnerId: partnerId ?? this.partnerId,
      status: status ?? this.status,
      errorMessage: errorMessage ?? this.errorMessage,
      connectedAt: connectedAt ?? this.connectedAt,
    );
  }

  @override
  List<Object?> get props => [partnerId, status, errorMessage, connectedAt];
}
