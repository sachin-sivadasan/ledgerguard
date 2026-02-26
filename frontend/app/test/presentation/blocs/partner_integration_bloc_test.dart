import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/partner_integration.dart';
import 'package:ledgerguard/domain/repositories/partner_integration_repository.dart';
import 'package:ledgerguard/presentation/blocs/partner_integration/partner_integration.dart';

class MockPartnerIntegrationRepository extends Mock
    implements PartnerIntegrationRepository {}

void main() {
  late MockPartnerIntegrationRepository mockRepository;

  final connectedIntegration = PartnerIntegration(
    partnerId: 'partner-123',
    status: IntegrationStatus.connected,
    connectedAt: DateTime(2024, 1, 1),
  );

  setUp(() {
    mockRepository = MockPartnerIntegrationRepository();
  });

  group('PartnerIntegrationBloc', () {
    test('initial state is PartnerIntegrationInitial', () {
      final bloc = PartnerIntegrationBloc(repository: mockRepository);
      expect(bloc.state, const PartnerIntegrationInitial());
      bloc.close();
    });

    group('CheckIntegrationStatusRequested', () {
      blocTest<PartnerIntegrationBloc, PartnerIntegrationState>(
        'emits [Loading, Connected] when connected',
        build: () {
          when(() => mockRepository.getIntegrationStatus())
              .thenAnswer((_) async => connectedIntegration);
          return PartnerIntegrationBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const CheckIntegrationStatusRequested()),
        expect: () => [
          isA<PartnerIntegrationLoading>(),
          isA<PartnerIntegrationConnected>()
              .having((s) => s.integration, 'integration', connectedIntegration),
        ],
      );

      blocTest<PartnerIntegrationBloc, PartnerIntegrationState>(
        'emits [Loading, NotConnected] when not connected',
        build: () {
          when(() => mockRepository.getIntegrationStatus())
              .thenAnswer((_) async => const PartnerIntegration());
          return PartnerIntegrationBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const CheckIntegrationStatusRequested()),
        expect: () => [
          isA<PartnerIntegrationLoading>(),
          isA<PartnerIntegrationNotConnected>(),
        ],
      );

      blocTest<PartnerIntegrationBloc, PartnerIntegrationState>(
        'emits [Loading, Error] when repository throws',
        build: () {
          when(() => mockRepository.getIntegrationStatus()).thenThrow(
            const PartnerIntegrationException('Network error'),
          );
          return PartnerIntegrationBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const CheckIntegrationStatusRequested()),
        expect: () => [
          isA<PartnerIntegrationLoading>(),
          isA<PartnerIntegrationError>()
              .having((s) => s.message, 'message', 'Network error'),
        ],
      );
    });

    group('ConnectWithOAuthRequested', () {
      blocTest<PartnerIntegrationBloc, PartnerIntegrationState>(
        'emits [Loading, Success] on successful OAuth connection',
        build: () {
          when(() => mockRepository.connectWithOAuth())
              .thenAnswer((_) async => connectedIntegration);
          return PartnerIntegrationBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const ConnectWithOAuthRequested()),
        expect: () => [
          isA<PartnerIntegrationLoading>(),
          isA<PartnerIntegrationSuccess>()
              .having((s) => s.integration, 'integration', connectedIntegration)
              .having((s) => s.message, 'message', contains('Successfully')),
        ],
      );

      blocTest<PartnerIntegrationBloc, PartnerIntegrationState>(
        'emits [Loading, Error] on OAuth failure',
        build: () {
          when(() => mockRepository.connectWithOAuth()).thenThrow(
            const PartnerIntegrationException('OAuth failed'),
          );
          return PartnerIntegrationBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const ConnectWithOAuthRequested()),
        expect: () => [
          isA<PartnerIntegrationLoading>(),
          isA<PartnerIntegrationError>()
              .having((s) => s.message, 'message', 'OAuth failed'),
        ],
      );
    });

    group('SaveManualTokenRequested', () {
      blocTest<PartnerIntegrationBloc, PartnerIntegrationState>(
        'emits [Loading, Success] on successful token save',
        build: () {
          when(() => mockRepository.saveManualToken(
                partnerId: any(named: 'partnerId'),
                apiToken: any(named: 'apiToken'),
              )).thenAnswer((_) async => connectedIntegration);
          return PartnerIntegrationBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const SaveManualTokenRequested(
          partnerId: 'partner-123',
          apiToken: 'token-abc',
        )),
        expect: () => [
          isA<PartnerIntegrationLoading>(),
          isA<PartnerIntegrationSuccess>()
              .having((s) => s.message, 'message', contains('saved')),
        ],
        verify: (_) {
          verify(() => mockRepository.saveManualToken(
                partnerId: 'partner-123',
                apiToken: 'token-abc',
              )).called(1);
        },
      );

      blocTest<PartnerIntegrationBloc, PartnerIntegrationState>(
        'emits [Loading, Error] when token is invalid',
        build: () {
          when(() => mockRepository.saveManualToken(
                partnerId: any(named: 'partnerId'),
                apiToken: any(named: 'apiToken'),
              )).thenThrow(
            const PartnerIntegrationException('Invalid token'),
          );
          return PartnerIntegrationBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const SaveManualTokenRequested(
          partnerId: 'partner-123',
          apiToken: '',
        )),
        expect: () => [
          isA<PartnerIntegrationLoading>(),
          isA<PartnerIntegrationError>()
              .having((s) => s.message, 'message', 'Invalid token'),
        ],
      );
    });

    group('DisconnectRequested', () {
      blocTest<PartnerIntegrationBloc, PartnerIntegrationState>(
        'emits [Loading, NotConnected] on successful disconnect',
        build: () {
          when(() => mockRepository.disconnect()).thenAnswer((_) async {});
          return PartnerIntegrationBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const DisconnectRequested()),
        expect: () => [
          isA<PartnerIntegrationLoading>(),
          isA<PartnerIntegrationNotConnected>(),
        ],
      );

      blocTest<PartnerIntegrationBloc, PartnerIntegrationState>(
        'emits [Loading, Error] on disconnect failure',
        build: () {
          when(() => mockRepository.disconnect()).thenThrow(
            const PartnerIntegrationException('Disconnect failed'),
          );
          return PartnerIntegrationBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const DisconnectRequested()),
        expect: () => [
          isA<PartnerIntegrationLoading>(),
          isA<PartnerIntegrationError>()
              .having((s) => s.message, 'message', 'Disconnect failed'),
        ],
      );
    });
  });
}
