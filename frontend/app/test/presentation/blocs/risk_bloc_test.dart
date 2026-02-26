import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/risk_summary.dart';
import 'package:ledgerguard/domain/repositories/risk_repository.dart';
import 'package:ledgerguard/presentation/blocs/risk/risk.dart';

class MockRiskRepository extends Mock implements RiskRepository {}

void main() {
  late MockRiskRepository mockRepository;

  const testSummary = RiskSummary(
    safeCount: 842,
    oneCycleMissedCount: 45,
    twoCyclesMissedCount: 18,
    churnedCount: 12,
    revenueAtRiskCents: 1850000,
  );

  const emptySummary = RiskSummary(
    safeCount: 0,
    oneCycleMissedCount: 0,
    twoCyclesMissedCount: 0,
    churnedCount: 0,
  );

  setUp(() {
    mockRepository = MockRiskRepository();
  });

  group('RiskBloc', () {
    test('initial state is RiskInitial', () {
      final bloc = RiskBloc(repository: mockRepository);
      expect(bloc.state, const RiskInitial());
      bloc.close();
    });

    group('LoadRiskSummaryRequested', () {
      blocTest<RiskBloc, RiskState>(
        'emits [Loading, Loaded] when summary is fetched successfully',
        build: () {
          when(() => mockRepository.fetchRiskSummary())
              .thenAnswer((_) async => testSummary);
          return RiskBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const LoadRiskSummaryRequested()),
        expect: () => [
          isA<RiskLoading>(),
          isA<RiskLoaded>().having((s) => s.summary, 'summary', testSummary),
        ],
      );

      blocTest<RiskBloc, RiskState>(
        'emits [Loading, Empty] when summary is null',
        build: () {
          when(() => mockRepository.fetchRiskSummary())
              .thenAnswer((_) async => null);
          return RiskBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const LoadRiskSummaryRequested()),
        expect: () => [
          isA<RiskLoading>(),
          isA<RiskEmpty>(),
        ],
      );

      blocTest<RiskBloc, RiskState>(
        'emits [Loading, Empty] when summary has no data',
        build: () {
          when(() => mockRepository.fetchRiskSummary())
              .thenAnswer((_) async => emptySummary);
          return RiskBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const LoadRiskSummaryRequested()),
        expect: () => [
          isA<RiskLoading>(),
          isA<RiskEmpty>(),
        ],
      );

      blocTest<RiskBloc, RiskState>(
        'emits [Loading, Error] when fetch fails',
        build: () {
          when(() => mockRepository.fetchRiskSummary())
              .thenThrow(const RiskException('Network error'));
          return RiskBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const LoadRiskSummaryRequested()),
        expect: () => [
          isA<RiskLoading>(),
          isA<RiskError>().having((s) => s.message, 'message', 'Network error'),
        ],
      );
    });

    group('RefreshRiskSummaryRequested', () {
      blocTest<RiskBloc, RiskState>(
        'emits [Loaded(refreshing), Loaded] when refresh succeeds',
        build: () {
          when(() => mockRepository.fetchRiskSummary())
              .thenAnswer((_) async => testSummary);
          return RiskBloc(repository: mockRepository);
        },
        seed: () => RiskLoaded(summary: testSummary),
        act: (bloc) => bloc.add(const RefreshRiskSummaryRequested()),
        expect: () => [
          isA<RiskLoaded>().having((s) => s.isRefreshing, 'isRefreshing', true),
          isA<RiskLoaded>()
              .having((s) => s.isRefreshing, 'isRefreshing', false),
        ],
      );

      blocTest<RiskBloc, RiskState>(
        'keeps current data when refresh fails',
        build: () {
          when(() => mockRepository.fetchRiskSummary())
              .thenThrow(const RiskException('Refresh failed'));
          return RiskBloc(repository: mockRepository);
        },
        seed: () => RiskLoaded(summary: testSummary),
        act: (bloc) => bloc.add(const RefreshRiskSummaryRequested()),
        expect: () => [
          isA<RiskLoaded>().having((s) => s.isRefreshing, 'isRefreshing', true),
          isA<RiskLoaded>()
              .having((s) => s.isRefreshing, 'isRefreshing', false)
              .having((s) => s.summary, 'summary', testSummary),
        ],
      );

      blocTest<RiskBloc, RiskState>(
        'triggers load when not in loaded state',
        build: () {
          when(() => mockRepository.fetchRiskSummary())
              .thenAnswer((_) async => testSummary);
          return RiskBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const RefreshRiskSummaryRequested()),
        expect: () => [
          isA<RiskLoading>(),
          isA<RiskLoaded>(),
        ],
      );
    });
  });

  group('RiskSummary', () {
    test('calculates totalSubscriptions correctly', () {
      expect(testSummary.totalSubscriptions, 917);
    });

    test('calculates percentages correctly', () {
      expect(testSummary.percentFor(RiskLevel.safe), closeTo(91.82, 0.01));
      expect(
          testSummary.percentFor(RiskLevel.oneCycleMissed), closeTo(4.91, 0.01));
      expect(testSummary.percentFor(RiskLevel.twoCyclesMissed),
          closeTo(1.96, 0.01));
      expect(testSummary.percentFor(RiskLevel.churned), closeTo(1.31, 0.01));
    });

    test('calculates atRiskCount correctly', () {
      expect(testSummary.atRiskCount, 63);
    });

    test('formats revenue at risk correctly', () {
      expect(testSummary.formattedRevenueAtRisk, '\$18.5K');
    });

    test('hasData returns true when totalSubscriptions > 0', () {
      expect(testSummary.hasData, true);
      expect(emptySummary.hasData, false);
    });
  });
}
