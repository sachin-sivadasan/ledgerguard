import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:ledgerguard/domain/entities/subscription.dart';
import 'package:ledgerguard/presentation/widgets/risk_badge.dart';

void main() {
  group('RiskBadge', () {
    Widget buildTestWidget(RiskState riskState, {bool isCompact = false}) {
      return MaterialApp(
        home: Scaffold(
          body: RiskBadge(
            riskState: riskState,
            isCompact: isCompact,
          ),
        ),
      );
    }

    group('displays correct text', () {
      testWidgets('shows "Safe" for safe state', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.safe));
        expect(find.text('Safe'), findsOneWidget);
      });

      testWidgets('shows "At Risk" for oneCycleMissed state', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.oneCycleMissed));
        expect(find.text('At Risk'), findsOneWidget);
      });

      testWidgets('shows "High Risk" for twoCyclesMissed state', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.twoCyclesMissed));
        expect(find.text('High Risk'), findsOneWidget);
      });

      testWidgets('shows "Churned" for churned state', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.churned));
        expect(find.text('Churned'), findsOneWidget);
      });
    });

    group('displays correct colors', () {
      testWidgets('safe state has green color', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.safe));

        final textWidget = tester.widget<Text>(find.text('Safe'));
        final textStyle = textWidget.style!;
        expect(textStyle.color, const Color(0xFF22C55E));
      });

      testWidgets('oneCycleMissed state has yellow color', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.oneCycleMissed));

        final textWidget = tester.widget<Text>(find.text('At Risk'));
        final textStyle = textWidget.style!;
        expect(textStyle.color, const Color(0xFFEAB308));
      });

      testWidgets('twoCyclesMissed state has orange color', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.twoCyclesMissed));

        final textWidget = tester.widget<Text>(find.text('High Risk'));
        final textStyle = textWidget.style!;
        expect(textStyle.color, const Color(0xFFF97316));
      });

      testWidgets('churned state has red color', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.churned));

        final textWidget = tester.widget<Text>(find.text('Churned'));
        final textStyle = textWidget.style!;
        expect(textStyle.color, const Color(0xFFEF4444));
      });
    });

    group('displays correct icons', () {
      testWidgets('safe state shows check circle icon', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.safe));
        expect(find.byIcon(Icons.check_circle_outline), findsOneWidget);
      });

      testWidgets('oneCycleMissed state shows warning icon', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.oneCycleMissed));
        expect(find.byIcon(Icons.warning_amber_outlined), findsOneWidget);
      });

      testWidgets('twoCyclesMissed state shows error icon', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.twoCyclesMissed));
        expect(find.byIcon(Icons.error_outline), findsOneWidget);
      });

      testWidgets('churned state shows cancel icon', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.churned));
        expect(find.byIcon(Icons.cancel_outlined), findsOneWidget);
      });
    });

    group('compact mode', () {
      testWidgets('compact mode does not show icon', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.safe, isCompact: true));
        expect(find.byIcon(Icons.check_circle_outline), findsNothing);
        expect(find.text('Safe'), findsOneWidget);
      });

      testWidgets('compact mode has smaller font size', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.safe, isCompact: true));

        final textWidget = tester.widget<Text>(find.text('Safe'));
        final textStyle = textWidget.style!;
        expect(textStyle.fontSize, 11);
      });

      testWidgets('full mode has larger font size', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.safe, isCompact: false));

        final textWidget = tester.widget<Text>(find.text('Safe'));
        final textStyle = textWidget.style!;
        expect(textStyle.fontSize, 13);
      });
    });
  });

  group('RiskStateIndicator', () {
    Widget buildTestWidget(RiskState riskState) {
      return MaterialApp(
        home: Scaffold(
          body: SizedBox(
            width: 400,
            child: RiskStateIndicator(riskState: riskState),
          ),
        ),
      );
    }

    group('displays correct description', () {
      testWidgets('safe state shows healthy description', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.safe));
        expect(find.text('Subscription is healthy with recent payments'), findsOneWidget);
      });

      testWidgets('oneCycleMissed state shows monitor description', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.oneCycleMissed));
        expect(find.text('One billing cycle missed - monitor closely'), findsOneWidget);
      });

      testWidgets('twoCyclesMissed state shows action needed description', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.twoCyclesMissed));
        expect(find.text('Two billing cycles missed - action needed'), findsOneWidget);
      });

      testWidgets('churned state shows churned description', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.churned));
        expect(find.text('Subscription has churned'), findsOneWidget);
      });
    });

    group('displays correct icons', () {
      testWidgets('safe state shows verified icon', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.safe));
        expect(find.byIcon(Icons.verified), findsOneWidget);
      });

      testWidgets('oneCycleMissed state shows warning icon', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.oneCycleMissed));
        expect(find.byIcon(Icons.warning), findsOneWidget);
      });

      testWidgets('twoCyclesMissed state shows error icon', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.twoCyclesMissed));
        expect(find.byIcon(Icons.error), findsOneWidget);
      });

      testWidgets('churned state shows cancel icon', (tester) async {
        await tester.pumpWidget(buildTestWidget(RiskState.churned));
        expect(find.byIcon(Icons.cancel), findsOneWidget);
      });
    });
  });
}
