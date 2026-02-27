import 'package:flutter_test/flutter_test.dart';
import 'package:bloc_test/bloc_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/api_key.dart';
import 'package:ledgerguard/domain/repositories/api_key_repository.dart';
import 'package:ledgerguard/presentation/blocs/api_key/api_key.dart';

class MockApiKeyRepository extends Mock implements ApiKeyRepository {}

void main() {
  late ApiKeyRepository repository;

  final testApiKey1 = ApiKey(
    id: 'key-1',
    name: 'Test Key 1',
    keyPrefix: 'lg_test_abc1234',
    createdAt: DateTime(2024, 1, 15),
    lastUsedAt: DateTime(2024, 2, 1),
  );

  final testApiKey2 = ApiKey(
    id: 'key-2',
    name: 'Test Key 2',
    keyPrefix: 'lg_test_def5678',
    createdAt: DateTime(2024, 1, 20),
    lastUsedAt: null,
  );

  final testCreationResult = ApiKeyCreationResult(
    apiKey: testApiKey1,
    fullKey: 'lg_test_abc1234_full_secret_key',
  );

  setUp(() {
    repository = MockApiKeyRepository();
  });

  group('ApiKeyBloc', () {
    test('initial state is ApiKeyInitial', () {
      final bloc = ApiKeyBloc(repository: repository);
      expect(bloc.state, equals(const ApiKeyInitial()));
    });

    group('LoadApiKeysRequested', () {
      blocTest<ApiKeyBloc, ApiKeyState>(
        'emits [Loading, Loaded] when keys are loaded successfully',
        setUp: () {
          when(() => repository.getApiKeys())
              .thenAnswer((_) async => [testApiKey1, testApiKey2]);
        },
        build: () => ApiKeyBloc(repository: repository),
        act: (bloc) => bloc.add(const LoadApiKeysRequested()),
        expect: () => [
          const ApiKeyLoading(),
          ApiKeyLoaded(apiKeys: [testApiKey1, testApiKey2]),
        ],
        verify: (_) {
          verify(() => repository.getApiKeys()).called(1);
        },
      );

      blocTest<ApiKeyBloc, ApiKeyState>(
        'emits [Loading, Empty] when no keys exist',
        setUp: () {
          when(() => repository.getApiKeys()).thenAnswer((_) async => []);
        },
        build: () => ApiKeyBloc(repository: repository),
        act: (bloc) => bloc.add(const LoadApiKeysRequested()),
        expect: () => [
          const ApiKeyLoading(),
          const ApiKeyEmpty(),
        ],
      );

      blocTest<ApiKeyBloc, ApiKeyState>(
        'emits [Loading, Error] when loading fails',
        setUp: () {
          when(() => repository.getApiKeys())
              .thenThrow(const ApiKeyException('Network error'));
        },
        build: () => ApiKeyBloc(repository: repository),
        act: (bloc) => bloc.add(const LoadApiKeysRequested()),
        expect: () => [
          const ApiKeyLoading(),
          const ApiKeyError('Network error'),
        ],
      );
    });

    group('CreateApiKeyRequested', () {
      blocTest<ApiKeyBloc, ApiKeyState>(
        'emits [Loaded(creating), Created] when key is created successfully from loaded state',
        setUp: () {
          when(() => repository.getApiKeys())
              .thenAnswer((_) async => [testApiKey2]);
          when(() => repository.createApiKey('New Key'))
              .thenAnswer((_) async => testCreationResult);
        },
        build: () => ApiKeyBloc(repository: repository),
        seed: () => ApiKeyLoaded(apiKeys: [testApiKey2]),
        act: (bloc) => bloc.add(const CreateApiKeyRequested('New Key')),
        expect: () => [
          ApiKeyLoaded(apiKeys: [testApiKey2], isCreating: true),
          ApiKeyCreated(
            apiKeys: [testApiKey1, testApiKey2],
            fullKey: 'lg_test_abc1234_full_secret_key',
            keyName: 'Test Key 1',
          ),
        ],
      );

      blocTest<ApiKeyBloc, ApiKeyState>(
        'emits [Loaded(creating), Created] when key is created from empty state',
        setUp: () {
          when(() => repository.createApiKey('First Key'))
              .thenAnswer((_) async => testCreationResult);
        },
        build: () => ApiKeyBloc(repository: repository),
        seed: () => const ApiKeyEmpty(),
        act: (bloc) => bloc.add(const CreateApiKeyRequested('First Key')),
        expect: () => [
          const ApiKeyLoaded(apiKeys: [], isCreating: true),
          ApiKeyCreated(
            apiKeys: [testApiKey1],
            fullKey: 'lg_test_abc1234_full_secret_key',
            keyName: 'Test Key 1',
          ),
        ],
      );

      blocTest<ApiKeyBloc, ApiKeyState>(
        'emits [Loaded(creating), Error] when creation fails with limit exceeded',
        setUp: () {
          when(() => repository.createApiKey(any()))
              .thenThrow(const ApiKeyLimitException());
        },
        build: () => ApiKeyBloc(repository: repository),
        seed: () => ApiKeyLoaded(apiKeys: [testApiKey1]),
        act: (bloc) => bloc.add(const CreateApiKeyRequested('Another Key')),
        expect: () => [
          ApiKeyLoaded(apiKeys: [testApiKey1], isCreating: true),
          ApiKeyError(
            'API key limit reached. Please revoke an existing key.',
            previousKeys: [testApiKey1],
          ),
        ],
      );
    });

    group('RevokeApiKeyRequested', () {
      blocTest<ApiKeyBloc, ApiKeyState>(
        'emits [Loaded(revoking), Loaded] when key is revoked successfully',
        setUp: () {
          when(() => repository.revokeApiKey('key-1'))
              .thenAnswer((_) async {});
        },
        build: () => ApiKeyBloc(repository: repository),
        seed: () => ApiKeyLoaded(apiKeys: [testApiKey1, testApiKey2]),
        act: (bloc) => bloc.add(const RevokeApiKeyRequested('key-1')),
        expect: () => [
          ApiKeyLoaded(
            apiKeys: [testApiKey1, testApiKey2],
            isRevoking: true,
            revokingKeyId: 'key-1',
          ),
          ApiKeyLoaded(apiKeys: [testApiKey2]),
        ],
      );

      blocTest<ApiKeyBloc, ApiKeyState>(
        'emits [Loaded(revoking), Empty] when last key is revoked',
        setUp: () {
          when(() => repository.revokeApiKey('key-1'))
              .thenAnswer((_) async {});
        },
        build: () => ApiKeyBloc(repository: repository),
        seed: () => ApiKeyLoaded(apiKeys: [testApiKey1]),
        act: (bloc) => bloc.add(const RevokeApiKeyRequested('key-1')),
        expect: () => [
          ApiKeyLoaded(
            apiKeys: [testApiKey1],
            isRevoking: true,
            revokingKeyId: 'key-1',
          ),
          const ApiKeyEmpty(),
        ],
      );

      blocTest<ApiKeyBloc, ApiKeyState>(
        'emits [Loaded(revoking), Error] when revoke fails',
        setUp: () {
          when(() => repository.revokeApiKey(any()))
              .thenThrow(const ApiKeyNotFoundException());
        },
        build: () => ApiKeyBloc(repository: repository),
        seed: () => ApiKeyLoaded(apiKeys: [testApiKey1]),
        act: (bloc) => bloc.add(const RevokeApiKeyRequested('key-1')),
        expect: () => [
          ApiKeyLoaded(
            apiKeys: [testApiKey1],
            isRevoking: true,
            revokingKeyId: 'key-1',
          ),
          ApiKeyError(
            'API key not found.',
            previousKeys: [testApiKey1],
          ),
        ],
      );

      blocTest<ApiKeyBloc, ApiKeyState>(
        'does nothing when not in Loaded state',
        build: () => ApiKeyBloc(repository: repository),
        seed: () => const ApiKeyEmpty(),
        act: (bloc) => bloc.add(const RevokeApiKeyRequested('key-1')),
        expect: () => [],
      );
    });

    group('DismissKeyCreatedRequested', () {
      blocTest<ApiKeyBloc, ApiKeyState>(
        'emits [Loaded] when dismissing from Created state with keys',
        build: () => ApiKeyBloc(repository: repository),
        seed: () => ApiKeyCreated(
          apiKeys: [testApiKey1],
          fullKey: 'secret',
          keyName: 'Test',
        ),
        act: (bloc) => bloc.add(const DismissKeyCreatedRequested()),
        expect: () => [
          ApiKeyLoaded(apiKeys: [testApiKey1]),
        ],
      );

      blocTest<ApiKeyBloc, ApiKeyState>(
        'emits [Empty] when dismissing from Created state with no keys',
        build: () => ApiKeyBloc(repository: repository),
        seed: () => const ApiKeyCreated(
          apiKeys: [],
          fullKey: 'secret',
          keyName: 'Test',
        ),
        act: (bloc) => bloc.add(const DismissKeyCreatedRequested()),
        expect: () => [
          const ApiKeyEmpty(),
        ],
      );

      blocTest<ApiKeyBloc, ApiKeyState>(
        'does nothing when not in Created state',
        build: () => ApiKeyBloc(repository: repository),
        seed: () => ApiKeyLoaded(apiKeys: [testApiKey1]),
        act: (bloc) => bloc.add(const DismissKeyCreatedRequested()),
        expect: () => [],
      );
    });
  });
}
