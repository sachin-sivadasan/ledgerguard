import 'package:equatable/equatable.dart';

import '../../../domain/entities/api_key.dart';

abstract class ApiKeyState extends Equatable {
  const ApiKeyState();

  @override
  List<Object?> get props => [];
}

/// Initial state before loading
class ApiKeyInitial extends ApiKeyState {
  const ApiKeyInitial();
}

/// Loading API keys
class ApiKeyLoading extends ApiKeyState {
  const ApiKeyLoading();
}

/// API keys loaded successfully
class ApiKeyLoaded extends ApiKeyState {
  final List<ApiKey> apiKeys;
  final bool isCreating;
  final bool isRevoking;
  final String? revokingKeyId;

  const ApiKeyLoaded({
    required this.apiKeys,
    this.isCreating = false,
    this.isRevoking = false,
    this.revokingKeyId,
  });

  ApiKeyLoaded copyWith({
    List<ApiKey>? apiKeys,
    bool? isCreating,
    bool? isRevoking,
    String? revokingKeyId,
  }) {
    return ApiKeyLoaded(
      apiKeys: apiKeys ?? this.apiKeys,
      isCreating: isCreating ?? this.isCreating,
      isRevoking: isRevoking ?? this.isRevoking,
      revokingKeyId: revokingKeyId,
    );
  }

  @override
  List<Object?> get props => [apiKeys, isCreating, isRevoking, revokingKeyId];
}

/// API key created successfully - shows full key once
class ApiKeyCreated extends ApiKeyState {
  final List<ApiKey> apiKeys;
  final String fullKey;
  final String keyName;

  const ApiKeyCreated({
    required this.apiKeys,
    required this.fullKey,
    required this.keyName,
  });

  @override
  List<Object?> get props => [apiKeys, fullKey, keyName];
}

/// No API keys exist
class ApiKeyEmpty extends ApiKeyState {
  const ApiKeyEmpty();
}

/// Error loading or managing API keys
class ApiKeyError extends ApiKeyState {
  final String message;
  final List<ApiKey>? previousKeys;

  const ApiKeyError(this.message, {this.previousKeys});

  @override
  List<Object?> get props => [message, previousKeys];
}
