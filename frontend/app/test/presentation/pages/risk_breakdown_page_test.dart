import 'package:bloc_test/bloc_test.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/risk_summary.dart';
import 'package:ledgerguard/presentation/blocs/risk/risk.dart';
import 'package:ledgerguard/presentation/pages/risk_breakdown_page.dart';

class MockRiskBloc extends MockBloc<RiskEvent, RiskState> implements RiskBloc {}

void main() {
  late MockRiskBloc mockBloc;

  const testSummary = RiskSummary(
    safeCount: 842,
    oneCycleMissedCount: 45,
    twoCyclesMissedCount: 18,
    churnedCount: 12,
    revenueAtRiskCents: 1850000,
  );

  setUpAll(() {
    registerFallbackValue(const LoadRiskSummaryRequested());
  });

  setUp(() {
    mockBloc = MockRiskBloc();
  });

  Widget buildTestWidget() {
    return MaterialApp(
      home: BlocProvider<RiskBloc>.value(
        value: mockBloc,
        child: const RiskBreakdownPage(),
      ),
    );
  }

  group('RiskBreakdownPage', () {
    testWidgets('shows loading state', (tester) async {
      when(() => mockBloc.state).thenReturn(const RiskLoading());

      await tester.pumpWidget(buildTestWidget());

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });

    testWidgets('shows empty state', (tester) async {
      when(() => mockBloc.state).thenReturn(const RiskEmpty());

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('No Risk Data'), findsOneWidget);
      expect(find.byIcon(Icons.pie_chart_outline), findsOneWidget);
    });

    testWidgets('shows error state with retry button', (tester) async {
      when(() => mockBloc.state).thenReturn(const RiskError('Network error'));

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Failed to load risk data'), findsOneWidget);
      expect(find.text('Network error'), findsOneWidget);
      expect(find.text('Retry'), findsOneWidget);

      await tester.tap(find.text('Retry'));
      await tester.pump();

      verify(() => mockBloc.add(const LoadRiskSummaryRequested())).called(1);
    });

    testWidgets('shows app bar title', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(RiskLoaded(summary: testSummary));

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Risk Breakdown'), findsOneWidget);
    });

    testWidgets('shows total subscriptions', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(RiskLoaded(summary: testSummary));

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Total Subscriptions'), findsOneWidget);
      expect(find.text('917'), findsOneWidget);
    });

    testWidgets('shows revenue at risk', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(RiskLoaded(summary: testSummary));

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Revenue at Risk'), findsOneWidget);
      expect(find.text('\$18.5K'), findsOneWidget);
    });

    testWidgets('shows distribution section', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(RiskLoaded(summary: testSummary));

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Distribution'), findsOneWidget);
    });

    testWidgets('shows breakdown by state section', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(RiskLoaded(summary: testSummary));

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Breakdown by State'), findsOneWidget);
    });

    testWidgets('shows all risk states with counts', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(RiskLoaded(summary: testSummary));

      await tester.pumpWidget(buildTestWidget());

      // Scroll to see all content
      await tester.drag(
          find.byType(SingleChildScrollView), const Offset(0, -300));
      await tester.pumpAndSettle();

      expect(find.text('SAFE'), findsOneWidget);
      expect(find.text('842'), findsOneWidget);
      expect(find.text('ONE_CYCLE_MISSED'), findsOneWidget);
      expect(find.text('45'), findsOneWidget);
      expect(find.text('TWO_CYCLES_MISSED'), findsOneWidget);
      expect(find.text('18'), findsOneWidget);
      expect(find.text('CHURNED'), findsOneWidget);
      expect(find.text('12'), findsOneWidget);
    });

    testWidgets('shows legend items', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(RiskLoaded(summary: testSummary));

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Safe'), findsWidgets);
      expect(find.text('One Cycle Missed'), findsOneWidget);
      expect(find.text('Two Cycles Missed'), findsOneWidget);
      expect(find.text('Churned'), findsWidgets);
    });

    testWidgets('shows refresh button', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(RiskLoaded(summary: testSummary));

      await tester.pumpWidget(buildTestWidget());

      expect(find.byIcon(Icons.refresh), findsOneWidget);
    });

    testWidgets('dispatches refresh on refresh button tap', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(RiskLoaded(summary: testSummary));

      await tester.pumpWidget(buildTestWidget());

      await tester.tap(find.byIcon(Icons.refresh));
      await tester.pump();

      verify(() => mockBloc.add(const RefreshRiskSummaryRequested())).called(1);
    });

    testWidgets('shows loading indicator when refreshing', (tester) async {
      when(() => mockBloc.state).thenReturn(
          RiskLoaded(summary: testSummary, isRefreshing: true));

      await tester.pumpWidget(buildTestWidget());

      // Should show progress indicator in app bar
      expect(
        find.descendant(
          of: find.byType(AppBar),
          matching: find.byType(CircularProgressIndicator),
        ),
        findsOneWidget,
      );
    });

    testWidgets('shows pie chart', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(RiskLoaded(summary: testSummary));

      await tester.pumpWidget(buildTestWidget());

      expect(find.byType(CustomPaint), findsWidgets);
    });

    testWidgets('shows risk state descriptions', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(RiskLoaded(summary: testSummary));

      await tester.pumpWidget(buildTestWidget());

      // Scroll to see descriptions
      await tester.drag(
          find.byType(SingleChildScrollView), const Offset(0, -300));
      await tester.pumpAndSettle();

      expect(find.text('Active and healthy subscriptions'), findsOneWidget);
      expect(find.text('Missed 1 billing cycle (31-60 days)'), findsOneWidget);
      expect(find.text('Missed 2 billing cycles (61-90 days)'), findsOneWidget);
      expect(find.text('Inactive for 90+ days'), findsOneWidget);
    });
  });
}
