import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard/core/services/snackbar_service.dart';

void main() {
  group('SnackbarService', () {
    late SnackbarService service;
    late GlobalKey<ScaffoldMessengerState> messengerKey;

    setUp(() {
      service = SnackbarService();
      messengerKey = GlobalKey<ScaffoldMessengerState>();
    });

    Widget buildTestWidget({required Widget child}) {
      return MaterialApp(
        scaffoldMessengerKey: messengerKey,
        home: Scaffold(body: child),
      );
    }

    testWidgets('showError displays error snackbar', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        child: Builder(
          builder: (context) {
            return ElevatedButton(
              onPressed: () {
                service.init(messengerKey);
                service.showError('Error message');
              },
              child: const Text('Show Error'),
            );
          },
        ),
      ));

      await tester.tap(find.text('Show Error'));
      await tester.pump();

      expect(find.text('Error message'), findsOneWidget);
      expect(find.byIcon(Icons.error_outline), findsOneWidget);
    });

    testWidgets('showSuccess displays success snackbar', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        child: Builder(
          builder: (context) {
            return ElevatedButton(
              onPressed: () {
                service.init(messengerKey);
                service.showSuccess('Success message');
              },
              child: const Text('Show Success'),
            );
          },
        ),
      ));

      await tester.tap(find.text('Show Success'));
      await tester.pump();

      expect(find.text('Success message'), findsOneWidget);
      expect(find.byIcon(Icons.check_circle_outline), findsOneWidget);
    });

    testWidgets('showInfo displays info snackbar', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        child: Builder(
          builder: (context) {
            return ElevatedButton(
              onPressed: () {
                service.init(messengerKey);
                service.showInfo('Info message');
              },
              child: const Text('Show Info'),
            );
          },
        ),
      ));

      await tester.tap(find.text('Show Info'));
      await tester.pump();

      expect(find.text('Info message'), findsOneWidget);
      expect(find.byIcon(Icons.info_outline), findsOneWidget);
    });

    testWidgets('showWarning displays warning snackbar', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        child: Builder(
          builder: (context) {
            return ElevatedButton(
              onPressed: () {
                service.init(messengerKey);
                service.showWarning('Warning message');
              },
              child: const Text('Show Warning'),
            );
          },
        ),
      ));

      await tester.tap(find.text('Show Warning'));
      await tester.pump();

      expect(find.text('Warning message'), findsOneWidget);
      expect(find.byIcon(Icons.warning_amber_outlined), findsOneWidget);
    });

    testWidgets('hide removes current snackbar', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        child: Builder(
          builder: (context) {
            return Column(
              children: [
                ElevatedButton(
                  onPressed: () {
                    service.init(messengerKey);
                    service.showInfo('Info message');
                  },
                  child: const Text('Show'),
                ),
                ElevatedButton(
                  onPressed: () => service.hide(),
                  child: const Text('Hide'),
                ),
              ],
            );
          },
        ),
      ));

      await tester.tap(find.text('Show'));
      await tester.pump();
      expect(find.text('Info message'), findsOneWidget);

      await tester.tap(find.text('Hide'));
      await tester.pumpAndSettle();
      expect(find.text('Info message'), findsNothing);
    });

    testWidgets('does nothing when not initialized', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        child: Builder(
          builder: (context) {
            return ElevatedButton(
              onPressed: () {
                // Don't initialize, just call showError
                service.showError('Error message');
              },
              child: const Text('Show Error'),
            );
          },
        ),
      ));

      await tester.tap(find.text('Show Error'));
      await tester.pump();

      // Should not crash, just do nothing
      expect(find.text('Error message'), findsNothing);
    });
  });
}
