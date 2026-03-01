import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/shopify_app.dart';
import 'package:ledgerguard/domain/repositories/app_repository.dart';
import 'package:ledgerguard/presentation/blocs/app_selection/app_selection.dart';

class MockAppRepository extends Mock implements AppRepository {}

class FakeShopifyApp extends Fake implements ShopifyApp {}

void main() {
  late MockAppRepository mockRepository;

  const testApps = [
    ShopifyApp(id: 'app-1', name: 'App One'),
    ShopifyApp(id: 'app-2', name: 'App Two'),
    ShopifyApp(id: 'app-3', name: 'App Three'),
  ];

  setUpAll(() {
    registerFallbackValue(FakeShopifyApp());
  });

  setUp(() {
    mockRepository = MockAppRepository();
  });

  group('AppSelectionBloc', () {
    test('initial state is AppSelectionInitial', () {
      final bloc = AppSelectionBloc(appRepository: mockRepository);
      expect(bloc.state, const AppSelectionInitial());
      bloc.close();
    });

    group('FetchAppsRequested', () {
      blocTest<AppSelectionBloc, AppSelectionState>(
        'emits [Loading, Loaded] when apps are fetched successfully',
        build: () {
          when(() => mockRepository.fetchAvailableApps())
              .thenAnswer((_) async => testApps);
          when(() => mockRepository.getSelectedApp())
              .thenAnswer((_) async => null);
          return AppSelectionBloc(appRepository: mockRepository);
        },
        act: (bloc) => bloc.add(const FetchAppsRequested()),
        expect: () => [
          isA<AppSelectionLoading>(),
          isA<AppSelectionLoaded>()
              .having((s) => s.apps, 'apps', testApps)
              .having((s) => s.selectedApp, 'selectedApp', isNull),
        ],
      );

      blocTest<AppSelectionBloc, AppSelectionState>(
        'emits [Loading, Loaded] with previously selected app',
        build: () {
          when(() => mockRepository.fetchAvailableApps())
              .thenAnswer((_) async => testApps);
          when(() => mockRepository.getSelectedApp())
              .thenAnswer((_) async => testApps[1]);
          return AppSelectionBloc(appRepository: mockRepository);
        },
        act: (bloc) => bloc.add(const FetchAppsRequested()),
        expect: () => [
          isA<AppSelectionLoading>(),
          isA<AppSelectionLoaded>()
              .having((s) => s.selectedApp, 'selectedApp', testApps[1]),
        ],
      );

      blocTest<AppSelectionBloc, AppSelectionState>(
        'emits [Loading, Error] when no apps found',
        build: () {
          when(() => mockRepository.fetchAvailableApps()).thenAnswer((_) async => []);
          return AppSelectionBloc(appRepository: mockRepository);
        },
        act: (bloc) => bloc.add(const FetchAppsRequested()),
        expect: () => [
          isA<AppSelectionLoading>(),
          isA<AppSelectionError>()
              .having((s) => s.message, 'message', contains('No apps')),
        ],
      );

      blocTest<AppSelectionBloc, AppSelectionState>(
        'emits [Loading, Error] when fetch fails',
        build: () {
          when(() => mockRepository.fetchAvailableApps())
              .thenThrow(const FetchAppsException('Network error'));
          return AppSelectionBloc(appRepository: mockRepository);
        },
        act: (bloc) => bloc.add(const FetchAppsRequested()),
        expect: () => [
          isA<AppSelectionLoading>(),
          isA<AppSelectionError>()
              .having((s) => s.message, 'message', 'Network error'),
        ],
      );
    });

    group('AppSelected', () {
      blocTest<AppSelectionBloc, AppSelectionState>(
        'updates selectedApp in Loaded state',
        build: () => AppSelectionBloc(appRepository: mockRepository),
        seed: () => const AppSelectionLoaded(apps: testApps),
        act: (bloc) => bloc.add(AppSelected(testApps[2])),
        expect: () => [
          isA<AppSelectionLoaded>()
              .having((s) => s.selectedApp, 'selectedApp', testApps[2]),
        ],
      );

      blocTest<AppSelectionBloc, AppSelectionState>(
        'does nothing when not in Loaded state',
        build: () => AppSelectionBloc(appRepository: mockRepository),
        seed: () => const AppSelectionLoading(),
        act: (bloc) => bloc.add(AppSelected(testApps[0])),
        expect: () => [],
      );
    });

    group('ConfirmSelectionRequested', () {
      blocTest<AppSelectionBloc, AppSelectionState>(
        'emits [Saving, Confirmed] on successful save',
        build: () {
          when(() => mockRepository.saveSelectedApp(any()))
              .thenAnswer((_) async {});
          return AppSelectionBloc(appRepository: mockRepository);
        },
        seed: () => AppSelectionLoaded(apps: testApps, selectedApp: testApps[0]),
        act: (bloc) => bloc.add(const ConfirmSelectionRequested()),
        expect: () => [
          isA<AppSelectionSaving>()
              .having((s) => s.selectedApp, 'selectedApp', testApps[0]),
          isA<AppSelectionConfirmed>()
              .having((s) => s.selectedApp, 'selectedApp', testApps[0]),
        ],
        verify: (_) {
          verify(() => mockRepository.saveSelectedApp(testApps[0])).called(1);
        },
      );

      blocTest<AppSelectionBloc, AppSelectionState>(
        'emits [Saving, Error] on save failure',
        build: () {
          when(() => mockRepository.saveSelectedApp(any()))
              .thenThrow(const AppException('Save failed'));
          return AppSelectionBloc(appRepository: mockRepository);
        },
        seed: () => AppSelectionLoaded(apps: testApps, selectedApp: testApps[0]),
        act: (bloc) => bloc.add(const ConfirmSelectionRequested()),
        expect: () => [
          isA<AppSelectionSaving>(),
          isA<AppSelectionError>()
              .having((s) => s.message, 'message', 'Save failed')
              .having((s) => s.apps, 'apps', testApps),
        ],
      );

      blocTest<AppSelectionBloc, AppSelectionState>(
        'does nothing when no app selected',
        build: () => AppSelectionBloc(appRepository: mockRepository),
        seed: () => const AppSelectionLoaded(apps: testApps),
        act: (bloc) => bloc.add(const ConfirmSelectionRequested()),
        expect: () => [],
      );
    });

    group('LoadSelectedAppRequested', () {
      blocTest<AppSelectionBloc, AppSelectionState>(
        'emits [Confirmed] when previously selected app exists',
        build: () {
          when(() => mockRepository.getSelectedApp())
              .thenAnswer((_) async => testApps[1]);
          return AppSelectionBloc(appRepository: mockRepository);
        },
        act: (bloc) => bloc.add(const LoadSelectedAppRequested()),
        expect: () => [
          isA<AppSelectionConfirmed>()
              .having((s) => s.selectedApp, 'selectedApp', testApps[1]),
        ],
      );

      blocTest<AppSelectionBloc, AppSelectionState>(
        'emits nothing when no previously selected app',
        build: () {
          when(() => mockRepository.getSelectedApp())
              .thenAnswer((_) async => null);
          return AppSelectionBloc(appRepository: mockRepository);
        },
        act: (bloc) => bloc.add(const LoadSelectedAppRequested()),
        expect: () => [],
      );
    });
  });
}
