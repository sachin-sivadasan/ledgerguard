import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../domain/entities/partner_integration.dart';
import '../../../domain/repositories/partner_integration_repository.dart';
import 'partner_integration_event.dart';
import 'partner_integration_state.dart';

/// Bloc for managing partner integration state
class PartnerIntegrationBloc
    extends Bloc<PartnerIntegrationEvent, PartnerIntegrationState> {
  final PartnerIntegrationRepository _repository;

  PartnerIntegrationBloc({
    required PartnerIntegrationRepository repository,
  })  : _repository = repository,
        super(const PartnerIntegrationInitial()) {
    on<CheckIntegrationStatusRequested>(_onCheckStatus);
    on<ConnectWithOAuthRequested>(_onConnectWithOAuth);
    on<SaveManualTokenRequested>(_onSaveManualToken);
    on<DisconnectRequested>(_onDisconnect);
  }

  Future<void> _onCheckStatus(
    CheckIntegrationStatusRequested event,
    Emitter<PartnerIntegrationState> emit,
  ) async {
    emit(const PartnerIntegrationLoading(message: 'Checking status...'));

    try {
      final integration = await _repository.getIntegrationStatus();

      if (integration.isConnected) {
        emit(PartnerIntegrationConnected(integration));
      } else {
        emit(const PartnerIntegrationNotConnected());
      }
    } on PartnerIntegrationException catch (e) {
      emit(PartnerIntegrationError(e.message));
    } catch (e) {
      emit(PartnerIntegrationError('Failed to check status: $e'));
    }
  }

  Future<void> _onConnectWithOAuth(
    ConnectWithOAuthRequested event,
    Emitter<PartnerIntegrationState> emit,
  ) async {
    emit(const PartnerIntegrationLoading(message: 'Connecting...'));

    try {
      final integration = await _repository.connectWithOAuth();

      emit(PartnerIntegrationSuccess(
        integration: integration,
        message: 'Successfully connected to Shopify Partner!',
      ));
    } on PartnerIntegrationException catch (e) {
      emit(PartnerIntegrationError(e.message));
    } catch (e) {
      emit(PartnerIntegrationError('Failed to connect: $e'));
    }
  }

  Future<void> _onSaveManualToken(
    SaveManualTokenRequested event,
    Emitter<PartnerIntegrationState> emit,
  ) async {
    emit(const PartnerIntegrationLoading(message: 'Saving token...'));

    try {
      final integration = await _repository.saveManualToken(
        partnerId: event.partnerId,
        apiToken: event.apiToken,
      );

      emit(PartnerIntegrationSuccess(
        integration: integration,
        message: 'Token saved successfully!',
      ));
    } on PartnerIntegrationException catch (e) {
      emit(PartnerIntegrationError(e.message));
    } catch (e) {
      emit(PartnerIntegrationError('Failed to save token: $e'));
    }
  }

  Future<void> _onDisconnect(
    DisconnectRequested event,
    Emitter<PartnerIntegrationState> emit,
  ) async {
    emit(const PartnerIntegrationLoading(message: 'Disconnecting...'));

    try {
      await _repository.disconnect();
      emit(const PartnerIntegrationNotConnected());
    } on PartnerIntegrationException catch (e) {
      emit(PartnerIntegrationError(e.message));
    } catch (e) {
      emit(PartnerIntegrationError('Failed to disconnect: $e'));
    }
  }
}
