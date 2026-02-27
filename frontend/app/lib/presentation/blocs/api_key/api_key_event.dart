import 'package:equatable/equatable.dart';

abstract class ApiKeyEvent extends Equatable {
  const ApiKeyEvent();

  @override
  List<Object?> get props => [];
}

/// Load all API keys
class LoadApiKeysRequested extends ApiKeyEvent {
  const LoadApiKeysRequested();
}

/// Create a new API key
class CreateApiKeyRequested extends ApiKeyEvent {
  final String name;

  const CreateApiKeyRequested(this.name);

  @override
  List<Object?> get props => [name];
}

/// Revoke an API key
class RevokeApiKeyRequested extends ApiKeyEvent {
  final String keyId;

  const RevokeApiKeyRequested(this.keyId);

  @override
  List<Object?> get props => [keyId];
}

/// Dismiss the key created dialog
class DismissKeyCreatedRequested extends ApiKeyEvent {
  const DismissKeyCreatedRequested();
}
