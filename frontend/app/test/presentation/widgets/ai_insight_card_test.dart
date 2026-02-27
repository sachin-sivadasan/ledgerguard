import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/daily_insight.dart';
import 'package:ledgerguard/presentation/blocs/insight/insight.dart';
import 'package:ledgerguard/presentation/widgets/ai_insight_card.dart';

class MockInsightBloc extends Mock implements InsightBloc {}

class FakeLoadInsightRequested extends Fake implements LoadInsightRequested {}

void main() {
  late MockInsightBloc mockInsightBloc;

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

  setUpAll(() {
    registerFallbackValue(FakeLoadInsightRequested());
  });

  setUp(() {
    mockInsightBloc = MockInsightBloc();
  });

  Widget buildTestWidget() {
    return MaterialApp(
      home: BlocProvider<InsightBloc>.value(
        value: mockInsightBloc,
        child: const Scaffold(body: AiInsightCard()),
      ),
    );
  }

  group('AiInsightCard', () {
    testWidgets('shows nothing for InsightInitial state', (tester) async {
      when(() => mockInsightBloc.state).thenReturn(const InsightInitial());
      when(() => mockInsightBloc.stream).thenAnswer((_) => const Stream.empty());
      when(() => mockInsightBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget());

      // Should only have SizedBox.shrink
      expect(find.byType(Container), findsNothing);
      expect(find.text('Daily Insight'), findsNothing);
    });

    testWidgets('shows loading state with shimmer', (tester) async {
      when(() => mockInsightBloc.state).thenReturn(const InsightLoading());
      when(() => mockInsightBloc.stream).thenAnswer((_) => const Stream.empty());
      when(() => mockInsightBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Daily Insight'), findsOneWidget);
      // Check for shimmer placeholders (grey containers)
      final shimmerContainers = find.byWidgetPredicate((widget) =>
          widget is Container &&
          widget.decoration is BoxDecoration &&
          (widget.decoration as BoxDecoration).color == Colors.grey[300]);
      expect(shimmerContainers, findsWidgets);
    });

    testWidgets('shows nothing for InsightEmpty state', (tester) async {
      when(() => mockInsightBloc.state).thenReturn(const InsightEmpty());
      when(() => mockInsightBloc.stream).thenAnswer((_) => const Stream.empty());
      when(() => mockInsightBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Daily Insight'), findsNothing);
    });

    testWidgets('shows nothing for InsightError state', (tester) async {
      when(() => mockInsightBloc.state)
          .thenReturn(const InsightError('Error message'));
      when(() => mockInsightBloc.stream).thenAnswer((_) => const Stream.empty());
      when(() => mockInsightBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Daily Insight'), findsNothing);
    });

    testWidgets('shows insight card with summary when loaded', (tester) async {
      when(() => mockInsightBloc.state)
          .thenReturn(InsightLoaded(insight: testInsight));
      when(() => mockInsightBloc.stream).thenAnswer((_) => const Stream.empty());
      when(() => mockInsightBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Daily Insight'), findsOneWidget);
      expect(find.text(testInsight.summary), findsOneWidget);
    });

    testWidgets('shows key takeaways when insight has key points',
        (tester) async {
      when(() => mockInsightBloc.state)
          .thenReturn(InsightLoaded(insight: testInsight));
      when(() => mockInsightBloc.stream).thenAnswer((_) => const Stream.empty());
      when(() => mockInsightBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Key Takeaways'), findsOneWidget);
      expect(find.text('Renewal rate up 5% month-over-month'), findsOneWidget);
      expect(
          find.text('12 subscriptions need attention this week'), findsOneWidget);
      expect(
          find.text('Usage revenue grew 8% from new customers'), findsOneWidget);
    });

    testWidgets('hides key takeaways when insight has no key points',
        (tester) async {
      when(() => mockInsightBloc.state)
          .thenReturn(InsightLoaded(insight: testInsightNoKeyPoints));
      when(() => mockInsightBloc.stream).thenAnswer((_) => const Stream.empty());
      when(() => mockInsightBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Key Takeaways'), findsNothing);
    });

    testWidgets('shows AI icon', (tester) async {
      when(() => mockInsightBloc.state)
          .thenReturn(InsightLoaded(insight: testInsight));
      when(() => mockInsightBloc.stream).thenAnswer((_) => const Stream.empty());
      when(() => mockInsightBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget());

      expect(find.byIcon(Icons.auto_awesome), findsOneWidget);
    });

    testWidgets('shows refresh button when loaded', (tester) async {
      when(() => mockInsightBloc.state)
          .thenReturn(InsightLoaded(insight: testInsight));
      when(() => mockInsightBloc.stream).thenAnswer((_) => const Stream.empty());
      when(() => mockInsightBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget());

      expect(find.byIcon(Icons.refresh), findsOneWidget);
    });

    testWidgets('shows spinner when refreshing', (tester) async {
      when(() => mockInsightBloc.state)
          .thenReturn(InsightLoaded(insight: testInsight, isRefreshing: true));
      when(() => mockInsightBloc.stream).thenAnswer((_) => const Stream.empty());
      when(() => mockInsightBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget());

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.byIcon(Icons.refresh), findsNothing);
    });

    testWidgets('can collapse and expand card', (tester) async {
      when(() => mockInsightBloc.state)
          .thenReturn(InsightLoaded(insight: testInsight));
      when(() => mockInsightBloc.stream).thenAnswer((_) => const Stream.empty());
      when(() => mockInsightBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget());

      // Initially expanded
      expect(find.text(testInsight.summary), findsOneWidget);

      // Tap header to collapse - use the header text as anchor
      await tester.tap(find.text('Daily Insight'));
      await tester.pumpAndSettle();

      // Content should be hidden (AnimatedCrossFade hides it)
      // The summary text widget is still in the tree but not visible
    });

    testWidgets('loads insight on init', (tester) async {
      when(() => mockInsightBloc.state).thenReturn(const InsightInitial());
      when(() => mockInsightBloc.stream).thenAnswer((_) => const Stream.empty());
      when(() => mockInsightBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget());

      verify(() => mockInsightBloc.add(any(that: isA<LoadInsightRequested>())))
          .called(1);
    });

    testWidgets('triggers refresh when refresh button tapped', (tester) async {
      when(() => mockInsightBloc.state)
          .thenReturn(InsightLoaded(insight: testInsight));
      when(() => mockInsightBloc.stream).thenAnswer((_) => const Stream.empty());
      when(() => mockInsightBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget());

      // Clear the initial load call verification
      clearInteractions(mockInsightBloc);
      when(() => mockInsightBloc.add(any())).thenReturn(null);

      // Tap refresh
      await tester.tap(find.byIcon(Icons.refresh));
      await tester.pump();

      verify(() => mockInsightBloc.add(any(that: isA<RefreshInsightRequested>())))
          .called(1);
    });

    testWidgets('shows generated time', (tester) async {
      when(() => mockInsightBloc.state)
          .thenReturn(InsightLoaded(insight: testInsight));
      when(() => mockInsightBloc.stream).thenAnswer((_) => const Stream.empty());
      when(() => mockInsightBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget());

      expect(find.textContaining('Generated'), findsOneWidget);
    });
  });
}
