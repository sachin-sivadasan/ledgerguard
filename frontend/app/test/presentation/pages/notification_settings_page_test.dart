import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/notification_preferences.dart';
import 'package:ledgerguard/presentation/blocs/notification_preferences/notification_preferences.dart';
import 'package:ledgerguard/presentation/pages/notification_settings_page.dart';

class MockNotificationPreferencesBloc extends Mock
    implements NotificationPreferencesBloc {}

class FakeNotificationPreferencesEvent extends Fake
    implements NotificationPreferencesEvent {}

void main() {
  late MockNotificationPreferencesBloc mockBloc;

  const testPreferences = NotificationPreferences(
    criticalAlertsEnabled: true,
    dailySummaryEnabled: true,
    dailySummaryTime: TimeOfDay(hour: 9, minute: 0),
  );

  setUpAll(() {
    registerFallbackValue(FakeNotificationPreferencesEvent());
  });

  setUp(() {
    mockBloc = MockNotificationPreferencesBloc();
  });

  Widget buildTestWidget({NotificationPreferencesState? state}) {
    when(() => mockBloc.state).thenReturn(
        state ?? const NotificationPreferencesInitial());
    when(() => mockBloc.stream).thenAnswer((_) => const Stream.empty());
    when(() => mockBloc.add(any())).thenReturn(null);

    return MaterialApp(
      home: BlocProvider<NotificationPreferencesBloc>.value(
        value: mockBloc,
        child: const NotificationSettingsPage(),
      ),
    );
  }

  group('NotificationSettingsPage', () {
    testWidgets('shows app bar title', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: NotificationPreferencesLoaded(preferences: testPreferences),
      ));

      expect(find.text('Notification Settings'), findsOneWidget);
    });

    testWidgets('loads preferences on init', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      verify(() => mockBloc.add(any(that: isA<LoadNotificationPreferencesRequested>())))
          .called(1);
    });

    testWidgets('shows loading indicator when loading', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: const NotificationPreferencesLoading(),
      ));

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });

    testWidgets('shows error state with retry button', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: const NotificationPreferencesError('Network error'),
      ));

      expect(find.text('Failed to load settings'), findsOneWidget);
      expect(find.text('Network error'), findsOneWidget);
      expect(find.text('Retry'), findsOneWidget);
    });

    testWidgets('dispatches load on retry tap', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: const NotificationPreferencesError('Network error'),
      ));

      // Clear initial load call
      clearInteractions(mockBloc);
      when(() => mockBloc.add(any())).thenReturn(null);

      await tester.tap(find.text('Retry'));
      await tester.pump();

      verify(() => mockBloc.add(any(that: isA<LoadNotificationPreferencesRequested>())))
          .called(1);
    });

    testWidgets('shows critical alerts section', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: NotificationPreferencesLoaded(preferences: testPreferences),
      ));

      expect(find.text('Critical Alerts'), findsOneWidget);
      expect(find.text('Enable Critical Alerts'), findsOneWidget);
    });

    testWidgets('shows daily summary section', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: NotificationPreferencesLoaded(preferences: testPreferences),
      ));

      expect(find.text('Daily Summary'), findsOneWidget);
      expect(find.text('Enable Daily Summary'), findsOneWidget);
      expect(find.text('Summary Time'), findsOneWidget);
    });

    testWidgets('shows save button', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: NotificationPreferencesLoaded(preferences: testPreferences),
      ));

      expect(find.text('Save Changes'), findsOneWidget);
    });

    testWidgets('save button is disabled when no unsaved changes', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: NotificationPreferencesLoaded(
          preferences: testPreferences,
          hasUnsavedChanges: false,
        ),
      ));

      final button = tester.widget<ElevatedButton>(
        find.widgetWithText(ElevatedButton, 'Save Changes'),
      );
      expect(button.onPressed, isNull);
    });

    testWidgets('save button is enabled when has unsaved changes', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: NotificationPreferencesLoaded(
          preferences: testPreferences,
          hasUnsavedChanges: true,
        ),
      ));

      final button = tester.widget<ElevatedButton>(
        find.widgetWithText(ElevatedButton, 'Save Changes'),
      );
      expect(button.onPressed, isNotNull);
    });

    testWidgets('shows unsaved changes indicator', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: NotificationPreferencesLoaded(
          preferences: testPreferences,
          hasUnsavedChanges: true,
        ),
      ));

      expect(find.text('You have unsaved changes'), findsOneWidget);
    });

    testWidgets('shows saving indicator when saving', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: NotificationPreferencesLoaded(
          preferences: testPreferences,
          isSaving: true,
        ),
      ));

      // Should show a small CircularProgressIndicator in the button
      expect(
        find.descendant(
          of: find.byType(ElevatedButton),
          matching: find.byType(CircularProgressIndicator),
        ),
        findsOneWidget,
      );
    });

    testWidgets('dispatches toggle critical alerts on switch tap', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: NotificationPreferencesLoaded(preferences: testPreferences),
      ));

      clearInteractions(mockBloc);
      when(() => mockBloc.add(any())).thenReturn(null);

      // Find the switch for critical alerts
      final switches = find.byType(Switch);
      await tester.tap(switches.first);
      await tester.pump();

      verify(() => mockBloc.add(any(that: isA<ToggleCriticalAlertsRequested>())))
          .called(1);
    });

    testWidgets('dispatches toggle daily summary on switch tap', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: NotificationPreferencesLoaded(preferences: testPreferences),
      ));

      clearInteractions(mockBloc);
      when(() => mockBloc.add(any())).thenReturn(null);

      // Find the switches - second one is daily summary
      final switches = find.byType(Switch);
      await tester.tap(switches.at(1));
      await tester.pump();

      verify(() => mockBloc.add(any(that: isA<ToggleDailySummaryRequested>())))
          .called(1);
    });

    testWidgets('dispatches save on save button tap', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: NotificationPreferencesLoaded(
          preferences: testPreferences,
          hasUnsavedChanges: true,
        ),
      ));

      clearInteractions(mockBloc);
      when(() => mockBloc.add(any())).thenReturn(null);

      await tester.tap(find.text('Save Changes'));
      await tester.pump();

      verify(() => mockBloc.add(any(that: isA<SaveNotificationPreferencesRequested>())))
          .called(1);
    });

    testWidgets('time picker button shows formatted time', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: NotificationPreferencesLoaded(preferences: testPreferences),
      ));

      expect(find.text('9:00 AM'), findsOneWidget);
    });

    testWidgets('time picker button is disabled when daily summary is off',
        (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: NotificationPreferencesLoaded(
          preferences: testPreferences.copyWith(dailySummaryEnabled: false),
        ),
      ));

      final textButton = tester.widget<TextButton>(
        find.widgetWithText(TextButton, '9:00 AM'),
      );
      expect(textButton.onPressed, isNull);
    });

    testWidgets('time picker button is enabled when daily summary is on',
        (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: NotificationPreferencesLoaded(preferences: testPreferences),
      ));

      final textButton = tester.widget<TextButton>(
        find.widgetWithText(TextButton, '9:00 AM'),
      );
      expect(textButton.onPressed, isNotNull);
    });

    testWidgets('shows switches with correct initial values', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: NotificationPreferencesLoaded(preferences: testPreferences),
      ));

      final switches = tester.widgetList<Switch>(find.byType(Switch)).toList();
      expect(switches[0].value, true); // Critical alerts
      expect(switches[1].value, true); // Daily summary
    });

    testWidgets('shows switches with disabled values', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: NotificationPreferencesLoaded(
          preferences: testPreferences.copyWith(
            criticalAlertsEnabled: false,
            dailySummaryEnabled: false,
          ),
        ),
      ));

      final switches = tester.widgetList<Switch>(find.byType(Switch)).toList();
      expect(switches[0].value, false); // Critical alerts
      expect(switches[1].value, false); // Daily summary
    });
  });
}
