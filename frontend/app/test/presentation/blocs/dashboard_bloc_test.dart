import 'package:bloc_test/bloc_test.dart';
import 'package:fake_async/fake_async.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/dashboard_metrics.dart';
import 'package:ledgerguard/domain/entities/time_range.dart';
import 'package:ledgerguard/domain/repositories/dashboard_repository.dart';
import 'package:ledgerguard/presentation/blocs/dashboard/dashboard.dart';

class MockDashboardRepository extends Mock implements DashboardRepository {}

void main() {
  late MockDashboardRepository mockRepository;

  const testMetrics = DashboardMetrics(
    renewalSuccessRate: 94.2,
    activeMrr: 12450000,
    revenueAtRisk: 1850000,
    churnedRevenue: 320000,
    churnedCount: 12,
    usageRevenue: 2340000,
    totalRevenue: 15240000,
    revenueMix: RevenueMix(
      recurring: 12450000,
      usage: 2340000,
      oneTime: 450000,
    ),
    riskDistribution: RiskDistribution(
      safe: 842,
      atRisk: 45,
      critical: 18,
      churned: 12,
    ),
  );

  setUp(() {
    mockRepository = MockDashboardRepository();
  });

  group('DashboardBloc', () {
    test('initial state is DashboardInitial', () {
      final bloc = DashboardBloc(repository: mockRepository);
      expect(bloc.state, const DashboardInitial());
      bloc.close();
    });

    group('LoadDashboardRequested', () {
      blocTest<DashboardBloc, DashboardState>(
        'emits [Loading, Loaded] when metrics are fetched successfully',
        build: () {
          when(() => mockRepository.fetchMetrics(timeRange: any(named: 'timeRange')))
              .thenAnswer((_) async => testMetrics);
          return DashboardBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const LoadDashboardRequested()),
        expect: () => [
          isA<DashboardLoading>(),
          isA<DashboardLoaded>()
              .having((s) => s.metrics, 'metrics', testMetrics),
        ],
      );

      blocTest<DashboardBloc, DashboardState>(
        'emits [Loading, Error] when fetch fails',
        build: () {
          when(() => mockRepository.fetchMetrics(timeRange: any(named: 'timeRange')))
              .thenThrow(const DashboardException('Network error'));
          return DashboardBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const LoadDashboardRequested()),
        expect: () => [
          isA<DashboardLoading>(),
          isA<DashboardError>()
              .having((s) => s.message, 'message', 'Network error'),
        ],
      );

      blocTest<DashboardBloc, DashboardState>(
        'emits [Loading, Empty] when no metrics available',
        build: () {
          when(() => mockRepository.fetchMetrics(timeRange: any(named: 'timeRange')))
              .thenAnswer((_) async => null);
          return DashboardBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const LoadDashboardRequested()),
        expect: () => [
          isA<DashboardLoading>(),
          isA<DashboardEmpty>(),
        ],
      );
    });

    group('RefreshDashboardRequested', () {
      blocTest<DashboardBloc, DashboardState>(
        'emits [Loaded(refreshing), Loaded] when refresh succeeds',
        build: () {
          when(() => mockRepository.refreshMetrics(timeRange: any(named: 'timeRange')))
              .thenAnswer((_) async => testMetrics);
          return DashboardBloc(repository: mockRepository);
        },
        seed: () => DashboardLoaded(metrics: testMetrics, timeRange: TimeRange.thisMonth()),
        act: (bloc) => bloc.add(const RefreshDashboardRequested()),
        expect: () => [
          isA<DashboardLoaded>()
              .having((s) => s.isRefreshing, 'isRefreshing', true),
          isA<DashboardLoaded>()
              .having((s) => s.isRefreshing, 'isRefreshing', false),
        ],
      );

      blocTest<DashboardBloc, DashboardState>(
        'keeps current data when refresh fails',
        build: () {
          when(() => mockRepository.refreshMetrics(timeRange: any(named: 'timeRange')))
              .thenThrow(const DashboardException('Refresh failed'));
          return DashboardBloc(repository: mockRepository);
        },
        seed: () => DashboardLoaded(metrics: testMetrics, timeRange: TimeRange.thisMonth()),
        act: (bloc) => bloc.add(const RefreshDashboardRequested()),
        expect: () => [
          isA<DashboardLoaded>()
              .having((s) => s.isRefreshing, 'isRefreshing', true),
          isA<DashboardLoaded>()
              .having((s) => s.isRefreshing, 'isRefreshing', false)
              .having((s) => s.metrics, 'metrics', testMetrics),
        ],
      );

      blocTest<DashboardBloc, DashboardState>(
        'triggers load when not in loaded state',
        build: () {
          when(() => mockRepository.fetchMetrics(timeRange: any(named: 'timeRange')))
              .thenAnswer((_) async => testMetrics);
          return DashboardBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const RefreshDashboardRequested()),
        expect: () => [
          isA<DashboardLoading>(),
          isA<DashboardLoaded>(),
        ],
      );

      blocTest<DashboardBloc, DashboardState>(
        'emits [Loaded(refreshing), Empty] when refresh returns null',
        build: () {
          when(() => mockRepository.refreshMetrics(timeRange: any(named: 'timeRange')))
              .thenAnswer((_) async => null);
          return DashboardBloc(repository: mockRepository);
        },
        seed: () => DashboardLoaded(metrics: testMetrics, timeRange: TimeRange.thisMonth()),
        act: (bloc) => bloc.add(const RefreshDashboardRequested()),
        expect: () => [
          isA<DashboardLoaded>()
              .having((s) => s.isRefreshing, 'isRefreshing', true),
          isA<DashboardEmpty>(),
        ],
      );
    });

    group('ServiceUnavailable', () {
      blocTest<DashboardBloc, DashboardState>(
        'emits [Loading, ServiceUnavailable] when 503 received',
        build: () {
          when(() => mockRepository.fetchMetrics(timeRange: any(named: 'timeRange')))
              .thenThrow(const ServiceUnavailableException());
          return DashboardBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const LoadDashboardRequested()),
        expect: () => [
          isA<DashboardLoading>(),
          isA<DashboardServiceUnavailable>(),
        ],
      );

      test('auto-retry fires after interval and emits Loaded on success', () {
        fakeAsync((async) {
          int callCount = 0;
          final repo = MockDashboardRepository();
          when(() => repo.fetchMetrics(timeRange: any(named: 'timeRange')))
              .thenAnswer((_) async {
            callCount++;
            if (callCount == 1) {
              throw const ServiceUnavailableException();
            }
            return testMetrics;
          });

          final bloc = DashboardBloc(repository: repo);
          bloc.add(const LoadDashboardRequested());

          // Process first load (will fail with 503)
          async.elapse(Duration.zero);
          expect(bloc.state, isA<DashboardServiceUnavailable>());

          // Advance past retry interval
          async.elapse(const Duration(seconds: 16));
          expect(bloc.state, isA<DashboardLoaded>());

          bloc.close();
        });
      });

      test('stops retrying after max retries', () {
        fakeAsync((async) {
          final repo = MockDashboardRepository();
          when(() => repo.fetchMetrics(timeRange: any(named: 'timeRange')))
              .thenThrow(const ServiceUnavailableException());

          final bloc = DashboardBloc(repository: repo);
          bloc.add(const LoadDashboardRequested());

          // Process initial load
          async.elapse(Duration.zero);

          // Advance past 3 retry intervals (each 15s)
          async.elapse(const Duration(seconds: 16)); // retry 1
          async.elapse(const Duration(seconds: 16)); // retry 2
          async.elapse(const Duration(seconds: 16)); // retry 3

          // After 3 retries, no more retries scheduled
          // Total calls: 1 initial + 3 retries = 4
          verify(() => repo.fetchMetrics(timeRange: any(named: 'timeRange')))
              .called(4);

          // Advance more — should NOT retry again
          async.elapse(const Duration(seconds: 30));
          verifyNever(() => repo.fetchMetrics(timeRange: any(named: 'timeRange')));

          bloc.close();
        });
      });
    });
  });
}
