import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../domain/entities/api_key.dart';
import '../../../domain/repositories/api_key_repository.dart';
import 'api_key_event.dart';
import 'api_key_state.dart';

class ApiKeyBloc extends Bloc<ApiKeyEvent, ApiKeyState> {
  final ApiKeyRepository _repository;

  ApiKeyBloc({required ApiKeyRepository repository})
      : _repository = repository,
        super(const ApiKeyInitial()) {
    on<LoadApiKeysRequested>(_onLoadApiKeys);
    on<CreateApiKeyRequested>(_onCreateApiKey);
    on<RevokeApiKeyRequested>(_onRevokeApiKey);
    on<DismissKeyCreatedRequested>(_onDismissKeyCreated);
  }

  Future<void> _onLoadApiKeys(
    LoadApiKeysRequested event,
    Emitter<ApiKeyState> emit,
  ) async {
    emit(const ApiKeyLoading());

    try {
      final apiKeys = await _repository.getApiKeys();

      if (apiKeys.isEmpty) {
        emit(const ApiKeyEmpty());
      } else {
        emit(ApiKeyLoaded(apiKeys: apiKeys));
      }
    } on ApiKeyException catch (e) {
      emit(ApiKeyError(e.message));
    } catch (e) {
      emit(ApiKeyError('Failed to load API keys: $e'));
    }
  }

  Future<void> _onCreateApiKey(
    CreateApiKeyRequested event,
    Emitter<ApiKeyState> emit,
  ) async {
    final currentState = state;
    List<ApiKey> currentKeys = [];

    if (currentState is ApiKeyLoaded) {
      currentKeys = currentState.apiKeys;
      emit(currentState.copyWith(isCreating: true));
    } else if (currentState is ApiKeyEmpty) {
      emit(const ApiKeyLoaded(apiKeys: [], isCreating: true));
    }

    try {
      final result = await _repository.createApiKey(event.name);

      final updatedKeys = [result.apiKey, ...currentKeys];

      emit(ApiKeyCreated(
        apiKeys: updatedKeys,
        fullKey: result.fullKey,
        keyName: result.apiKey.name,
      ));
    } on ApiKeyException catch (e) {
      emit(ApiKeyError(e.message, previousKeys: currentKeys));
    } catch (e) {
      emit(ApiKeyError('Failed to create API key: $e', previousKeys: currentKeys));
    }
  }

  Future<void> _onRevokeApiKey(
    RevokeApiKeyRequested event,
    Emitter<ApiKeyState> emit,
  ) async {
    final currentState = state;

    if (currentState is! ApiKeyLoaded) return;

    emit(currentState.copyWith(isRevoking: true, revokingKeyId: event.keyId));

    try {
      await _repository.revokeApiKey(event.keyId);

      final updatedKeys = currentState.apiKeys
          .where((key) => key.id != event.keyId)
          .toList();

      if (updatedKeys.isEmpty) {
        emit(const ApiKeyEmpty());
      } else {
        emit(ApiKeyLoaded(apiKeys: updatedKeys));
      }
    } on ApiKeyException catch (e) {
      emit(ApiKeyError(e.message, previousKeys: currentState.apiKeys));
    } catch (e) {
      emit(ApiKeyError('Failed to revoke API key: $e', previousKeys: currentState.apiKeys));
    }
  }

  void _onDismissKeyCreated(
    DismissKeyCreatedRequested event,
    Emitter<ApiKeyState> emit,
  ) {
    final currentState = state;

    if (currentState is ApiKeyCreated) {
      if (currentState.apiKeys.isEmpty) {
        emit(const ApiKeyEmpty());
      } else {
        emit(ApiKeyLoaded(apiKeys: currentState.apiKeys));
      }
    }
  }
}
