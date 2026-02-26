import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/dashboard_preferences.dart';
import 'package:ledgerguard/domain/repositories/dashboard_preferences_repository.dart';
import 'package:ledgerguard/presentation/blocs/preferences/preferences.dart';

class MockDashboardPreferencesRepository extends Mock
    implements DashboardPreferencesRepository {}

void main() {
  late MockDashboardPreferencesRepository mockRepository;

  final defaultPreferences = DashboardPreferences.defaults();
  final customPreferences = DashboardPreferences(
    primaryKpis: [KpiType.activeMrr, KpiType.totalRevenue],
    enabledSecondaryWidgets: {SecondaryWidget.revenueMixChart},
  );

  setUpAll(() {
    registerFallbackValue(DashboardPreferences.defaults());
  });

  setUp(() {
    mockRepository = MockDashboardPreferencesRepository();
  });

  group('PreferencesBloc', () {
    test('initial state is PreferencesInitial', () {
      final bloc = PreferencesBloc(repository: mockRepository);
      expect(bloc.state, const PreferencesInitial());
      bloc.close();
    });

    group('LoadPreferencesRequested', () {
      blocTest<PreferencesBloc, PreferencesState>(
        'emits [Loading, Loaded] when preferences are fetched successfully',
        build: () {
          when(() => mockRepository.fetchPreferences())
              .thenAnswer((_) async => defaultPreferences);
          return PreferencesBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const LoadPreferencesRequested()),
        expect: () => [
          isA<PreferencesLoading>(),
          isA<PreferencesLoaded>()
              .having((s) => s.preferences, 'preferences', defaultPreferences),
        ],
      );

      blocTest<PreferencesBloc, PreferencesState>(
        'emits [Loading, Error] when fetch fails',
        build: () {
          when(() => mockRepository.fetchPreferences())
              .thenThrow(const FetchPreferencesException());
          return PreferencesBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const LoadPreferencesRequested()),
        expect: () => [
          isA<PreferencesLoading>(),
          isA<PreferencesError>()
              .having((s) => s.message, 'message', contains('Failed')),
        ],
      );
    });

    group('AddPrimaryKpiRequested', () {
      blocTest<PreferencesBloc, PreferencesState>(
        'adds KPI to preferences',
        build: () => PreferencesBloc(repository: mockRepository),
        seed: () => PreferencesLoaded(preferences: customPreferences),
        act: (bloc) =>
            bloc.add(const AddPrimaryKpiRequested(KpiType.renewalSuccessRate)),
        expect: () => [
          isA<PreferencesLoaded>()
              .having((s) => s.preferences.primaryKpis.length, 'kpi count', 3)
              .having((s) => s.hasUnsavedChanges, 'hasUnsavedChanges', true),
        ],
      );

      blocTest<PreferencesBloc, PreferencesState>(
        'does not add duplicate KPI',
        build: () => PreferencesBloc(repository: mockRepository),
        seed: () => PreferencesLoaded(preferences: customPreferences),
        act: (bloc) =>
            bloc.add(const AddPrimaryKpiRequested(KpiType.activeMrr)),
        expect: () => [
          isA<PreferencesLoaded>()
              .having((s) => s.preferences.primaryKpis.length, 'kpi count', 2),
        ],
      );

      blocTest<PreferencesBloc, PreferencesState>(
        'does not exceed max 4 KPIs',
        build: () => PreferencesBloc(repository: mockRepository),
        seed: () => PreferencesLoaded(preferences: defaultPreferences),
        act: (bloc) =>
            bloc.add(const AddPrimaryKpiRequested(KpiType.usageRevenue)),
        expect: () => [
          isA<PreferencesLoaded>()
              .having((s) => s.preferences.primaryKpis.length, 'kpi count', 4),
        ],
      );
    });

    group('RemovePrimaryKpiRequested', () {
      blocTest<PreferencesBloc, PreferencesState>(
        'removes KPI from preferences',
        build: () => PreferencesBloc(repository: mockRepository),
        seed: () => PreferencesLoaded(preferences: defaultPreferences),
        act: (bloc) =>
            bloc.add(const RemovePrimaryKpiRequested(KpiType.churned)),
        expect: () => [
          isA<PreferencesLoaded>()
              .having((s) => s.preferences.primaryKpis.length, 'kpi count', 3)
              .having((s) => s.hasUnsavedChanges, 'hasUnsavedChanges', true),
        ],
      );

      blocTest<PreferencesBloc, PreferencesState>(
        'does nothing when removing non-existent KPI',
        build: () => PreferencesBloc(repository: mockRepository),
        seed: () => PreferencesLoaded(preferences: customPreferences),
        act: (bloc) =>
            bloc.add(const RemovePrimaryKpiRequested(KpiType.churned)),
        expect: () => [
          isA<PreferencesLoaded>()
              .having((s) => s.preferences.primaryKpis.length, 'kpi count', 2),
        ],
      );
    });

    group('ReorderPrimaryKpiRequested', () {
      blocTest<PreferencesBloc, PreferencesState>(
        'reorders KPIs correctly',
        build: () => PreferencesBloc(repository: mockRepository),
        seed: () => PreferencesLoaded(preferences: defaultPreferences),
        act: (bloc) => bloc.add(const ReorderPrimaryKpiRequested(
          oldIndex: 0,
          newIndex: 2,
        )),
        expect: () => [
          isA<PreferencesLoaded>()
              .having((s) => s.preferences.primaryKpis[0], 'first kpi',
                  KpiType.activeMrr)
              .having((s) => s.preferences.primaryKpis[2], 'third kpi',
                  KpiType.renewalSuccessRate)
              .having((s) => s.hasUnsavedChanges, 'hasUnsavedChanges', true),
        ],
      );
    });

    group('ToggleSecondaryWidgetRequested', () {
      blocTest<PreferencesBloc, PreferencesState>(
        'disables enabled widget',
        build: () => PreferencesBloc(repository: mockRepository),
        seed: () => PreferencesLoaded(preferences: defaultPreferences),
        act: (bloc) => bloc.add(const ToggleSecondaryWidgetRequested(
            SecondaryWidget.revenueMixChart)),
        expect: () => [
          isA<PreferencesLoaded>()
              .having(
                  (s) => s.preferences.enabledSecondaryWidgets
                      .contains(SecondaryWidget.revenueMixChart),
                  'has revenue mix chart',
                  false)
              .having((s) => s.hasUnsavedChanges, 'hasUnsavedChanges', true),
        ],
      );

      blocTest<PreferencesBloc, PreferencesState>(
        'enables disabled widget',
        build: () => PreferencesBloc(repository: mockRepository),
        seed: () => PreferencesLoaded(preferences: customPreferences),
        act: (bloc) => bloc.add(const ToggleSecondaryWidgetRequested(
            SecondaryWidget.riskDistributionChart)),
        expect: () => [
          isA<PreferencesLoaded>()
              .having(
                  (s) => s.preferences.enabledSecondaryWidgets
                      .contains(SecondaryWidget.riskDistributionChart),
                  'has risk distribution chart',
                  true)
              .having((s) => s.hasUnsavedChanges, 'hasUnsavedChanges', true),
        ],
      );
    });

    group('SavePreferencesRequested', () {
      blocTest<PreferencesBloc, PreferencesState>(
        'emits [Loaded(saving), Saved, Loaded] when save succeeds',
        build: () {
          when(() => mockRepository.savePreferences(any()))
              .thenAnswer((_) async {});
          return PreferencesBloc(repository: mockRepository);
        },
        seed: () => PreferencesLoaded(
          preferences: customPreferences,
          hasUnsavedChanges: true,
        ),
        act: (bloc) => bloc.add(const SavePreferencesRequested()),
        expect: () => [
          isA<PreferencesLoaded>()
              .having((s) => s.isSaving, 'isSaving', true),
          isA<PreferencesSaved>(),
          isA<PreferencesLoaded>()
              .having((s) => s.hasUnsavedChanges, 'hasUnsavedChanges', false),
        ],
      );

      blocTest<PreferencesBloc, PreferencesState>(
        'emits [Loaded(saving), Error, Loaded] when save fails',
        build: () {
          when(() => mockRepository.savePreferences(any()))
              .thenThrow(const SavePreferencesException());
          return PreferencesBloc(repository: mockRepository);
        },
        seed: () => PreferencesLoaded(
          preferences: customPreferences,
          hasUnsavedChanges: true,
        ),
        act: (bloc) => bloc.add(const SavePreferencesRequested()),
        expect: () => [
          isA<PreferencesLoaded>()
              .having((s) => s.isSaving, 'isSaving', true),
          isA<PreferencesError>(),
          isA<PreferencesLoaded>()
              .having((s) => s.isSaving, 'isSaving', false),
        ],
      );
    });

    group('ResetPreferencesRequested', () {
      blocTest<PreferencesBloc, PreferencesState>(
        'resets to default preferences',
        build: () => PreferencesBloc(repository: mockRepository),
        seed: () => PreferencesLoaded(preferences: customPreferences),
        act: (bloc) => bloc.add(const ResetPreferencesRequested()),
        expect: () => [
          isA<PreferencesLoaded>()
              .having((s) => s.preferences.primaryKpis.length, 'kpi count', 4)
              .having((s) => s.hasUnsavedChanges, 'hasUnsavedChanges', true),
        ],
      );
    });
  });
}
