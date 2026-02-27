import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/daily_insight.dart';
import 'package:ledgerguard/domain/repositories/insight_repository.dart';
import 'package:ledgerguard/presentation/blocs/insight/insight.dart';

class MockInsightRepository extends Mock implements InsightRepository {}

void main() {
  late MockInsightRepository mockRepository;

  final testInsight = DailyInsight(
    summary:
        'Your renewal rate is trending up 5% this month. Focus on the 12 at-risk subscriptions worth \$4.2K MRR.',
    generatedAt: DateTime.now().subtract(const Duration(hours: 2)),
    keyPoints: [
      'Renewal rate up 5% month-over-month',
      '12 subscriptions need attention this week',
      'Usage revenue grew 8% from new customers',
    ],
  );

  final testInsightNoKeyPoints = DailyInsight(
    summary: 'Simple summary without key points.',
    generatedAt: DateTime.now(),
    keyPoints: const [],
  );

  setUp(() {
    mockRepository = MockInsightRepository();
  });

  group('InsightBloc', () {
    test('initial state is InsightInitial', () {
      final bloc = InsightBloc(repository: mockRepository);
      expect(bloc.state, const InsightInitial());
      bloc.close();
    });

    group('LoadInsightRequested', () {
      blocTest<InsightBloc, InsightState>(
        'emits [Loading, Loaded] when insight is fetched successfully',
        build: () {
          when(() => mockRepository.fetchDailyInsight())
              .thenAnswer((_) async => testInsight);
          return InsightBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const LoadInsightRequested()),
        expect: () => [
          isA<InsightLoading>(),
          isA<InsightLoaded>().having((s) => s.insight, 'insight', testInsight),
        ],
      );

      blocTest<InsightBloc, InsightState>(
        'emits [Loading, Empty] when insight is null',
        build: () {
          when(() => mockRepository.fetchDailyInsight())
              .thenAnswer((_) async => null);
          return InsightBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const LoadInsightRequested()),
        expect: () => [
          isA<InsightLoading>(),
          isA<InsightEmpty>(),
        ],
      );

      blocTest<InsightBloc, InsightState>(
        'emits [Loading, Error] when InsightException is thrown',
        build: () {
          when(() => mockRepository.fetchDailyInsight())
              .thenThrow(const InsightException('Network error'));
          return InsightBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const LoadInsightRequested()),
        expect: () => [
          isA<InsightLoading>(),
          isA<InsightError>()
              .having((s) => s.message, 'message', 'Network error'),
        ],
      );

      blocTest<InsightBloc, InsightState>(
        'emits [Loading, Error] when generic exception is thrown',
        build: () {
          when(() => mockRepository.fetchDailyInsight())
              .thenThrow(Exception('Unexpected error'));
          return InsightBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const LoadInsightRequested()),
        expect: () => [
          isA<InsightLoading>(),
          isA<InsightError>().having(
              (s) => s.message, 'message', contains('Failed to load insight')),
        ],
      );
    });

    group('RefreshInsightRequested', () {
      blocTest<InsightBloc, InsightState>(
        'emits [Loaded(refreshing), Loaded] when refresh succeeds',
        build: () {
          when(() => mockRepository.fetchDailyInsight())
              .thenAnswer((_) async => testInsight);
          return InsightBloc(repository: mockRepository);
        },
        seed: () => InsightLoaded(insight: testInsight),
        act: (bloc) => bloc.add(const RefreshInsightRequested()),
        expect: () => [
          isA<InsightLoaded>().having((s) => s.isRefreshing, 'isRefreshing', true),
          isA<InsightLoaded>()
              .having((s) => s.isRefreshing, 'isRefreshing', false),
        ],
      );

      blocTest<InsightBloc, InsightState>(
        'keeps current data when refresh fails with InsightException',
        build: () {
          when(() => mockRepository.fetchDailyInsight())
              .thenThrow(const InsightException('Refresh failed'));
          return InsightBloc(repository: mockRepository);
        },
        seed: () => InsightLoaded(insight: testInsight),
        act: (bloc) => bloc.add(const RefreshInsightRequested()),
        expect: () => [
          isA<InsightLoaded>().having((s) => s.isRefreshing, 'isRefreshing', true),
          isA<InsightLoaded>()
              .having((s) => s.isRefreshing, 'isRefreshing', false)
              .having((s) => s.insight, 'insight', testInsight),
        ],
      );

      blocTest<InsightBloc, InsightState>(
        'keeps current data when refresh fails with generic exception',
        build: () {
          when(() => mockRepository.fetchDailyInsight())
              .thenThrow(Exception('Network error'));
          return InsightBloc(repository: mockRepository);
        },
        seed: () => InsightLoaded(insight: testInsight),
        act: (bloc) => bloc.add(const RefreshInsightRequested()),
        expect: () => [
          isA<InsightLoaded>().having((s) => s.isRefreshing, 'isRefreshing', true),
          isA<InsightLoaded>()
              .having((s) => s.isRefreshing, 'isRefreshing', false),
        ],
      );

      blocTest<InsightBloc, InsightState>(
        'triggers load when not in loaded state',
        build: () {
          when(() => mockRepository.fetchDailyInsight())
              .thenAnswer((_) async => testInsight);
          return InsightBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const RefreshInsightRequested()),
        expect: () => [
          isA<InsightLoading>(),
          isA<InsightLoaded>(),
        ],
      );

      blocTest<InsightBloc, InsightState>(
        'emits [Loaded(refreshing), Empty] when refresh returns null',
        build: () {
          when(() => mockRepository.fetchDailyInsight())
              .thenAnswer((_) async => null);
          return InsightBloc(repository: mockRepository);
        },
        seed: () => InsightLoaded(insight: testInsight),
        act: (bloc) => bloc.add(const RefreshInsightRequested()),
        expect: () => [
          isA<InsightLoaded>().having((s) => s.isRefreshing, 'isRefreshing', true),
          isA<InsightEmpty>(),
        ],
      );
    });
  });

  group('DailyInsight', () {
    test('fromJson parses correctly', () {
      final json = {
        'summary': 'Test summary',
        'generated_at': '2024-01-15T10:30:00Z',
        'key_points': ['Point 1', 'Point 2'],
      };

      final insight = DailyInsight.fromJson(json);
      expect(insight.summary, 'Test summary');
      expect(insight.generatedAt, DateTime.parse('2024-01-15T10:30:00Z'));
      expect(insight.keyPoints, ['Point 1', 'Point 2']);
    });

    test('fromJson handles missing key_points', () {
      final json = {
        'summary': 'Test summary',
        'generated_at': '2024-01-15T10:30:00Z',
      };

      final insight = DailyInsight.fromJson(json);
      expect(insight.keyPoints, isEmpty);
    });

    test('fromJson handles missing summary', () {
      final json = {
        'generated_at': '2024-01-15T10:30:00Z',
      };

      final insight = DailyInsight.fromJson(json);
      expect(insight.summary, '');
    });

    test('hasKeyPoints returns correct value', () {
      expect(testInsight.hasKeyPoints, true);
      expect(testInsightNoKeyPoints.hasKeyPoints, false);
    });

    test('formattedGeneratedAt returns correct format for recent times', () {
      final recentInsight = DailyInsight(
        summary: 'Test',
        generatedAt: DateTime.now().subtract(const Duration(minutes: 30)),
      );
      expect(recentInsight.formattedGeneratedAt, '30 min ago');
    });

    test('formattedGeneratedAt returns correct format for hours ago', () {
      final hoursAgoInsight = DailyInsight(
        summary: 'Test',
        generatedAt: DateTime.now().subtract(const Duration(hours: 3)),
      );
      expect(hoursAgoInsight.formattedGeneratedAt, '3 hours ago');
    });
  });

  group('InsightException', () {
    test('NoAppSelectedInsightException has correct message', () {
      const exception = NoAppSelectedInsightException();
      expect(exception.message, contains('No app selected'));
      expect(exception.code, 'no-app-selected');
    });

    test('UnauthorizedInsightException has correct message', () {
      const exception = UnauthorizedInsightException();
      expect(exception.message, contains('authenticated'));
      expect(exception.code, 'unauthorized');
    });

    test('ProRequiredInsightException has correct message', () {
      const exception = ProRequiredInsightException();
      expect(exception.message, contains('PRO'));
      expect(exception.code, 'pro-required');
    });
  });
}
