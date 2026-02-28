import 'package:bloc_test/bloc_test.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/dashboard_preferences.dart';
import 'package:ledgerguard/presentation/blocs/preferences/preferences.dart';
import 'package:ledgerguard/presentation/widgets/dashboard_config_dialog.dart';

class MockPreferencesBloc extends MockBloc<PreferencesEvent, PreferencesState>
    implements PreferencesBloc {}

void main() {
  late MockPreferencesBloc mockBloc;

  final defaultPreferences = DashboardPreferences.defaults();
  final customPreferences = DashboardPreferences(
    primaryKpis: [KpiType.activeMrr, KpiType.totalRevenue],
    enabledSecondaryWidgets: {SecondaryWidget.revenueMixChart},
  );

  setUpAll(() {
    registerFallbackValue(const LoadPreferencesRequested());
    registerFallbackValue(const AddPrimaryKpiRequested(KpiType.activeMrr));
    registerFallbackValue(const RemovePrimaryKpiRequested(KpiType.activeMrr));
    registerFallbackValue(const ToggleSecondaryWidgetRequested(
        SecondaryWidget.revenueMixChart));
    registerFallbackValue(const SavePreferencesRequested());
    registerFallbackValue(const ResetPreferencesRequested());
    registerFallbackValue(
        const ReorderPrimaryKpiRequested(oldIndex: 0, newIndex: 1));
  });

  setUp(() {
    mockBloc = MockPreferencesBloc();
  });

  Widget buildTestWidget() {
    return MaterialApp(
      home: BlocProvider<PreferencesBloc>.value(
        value: mockBloc,
        child: const Scaffold(
          body: DashboardConfigDialog(),
        ),
      ),
    );
  }

  group('DashboardConfigDialog', () {
    testWidgets('shows loading state', (tester) async {
      when(() => mockBloc.state).thenReturn(const PreferencesLoading());

      await tester.pumpWidget(buildTestWidget());

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });

    testWidgets('shows error state with retry button', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(const PreferencesError('Network error'));

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Error: Network error'), findsOneWidget);
      expect(find.text('Retry'), findsOneWidget);

      await tester.tap(find.text('Retry'));
      await tester.pump();

      verify(() => mockBloc.add(const LoadPreferencesRequested())).called(1);
    });

    testWidgets('shows preferences when loaded', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(PreferencesLoaded(preferences: defaultPreferences));

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Dashboard Configuration'), findsOneWidget);
      expect(find.text('Primary KPIs'), findsOneWidget);
      expect(find.text('Secondary Widgets'), findsOneWidget);
    });

    testWidgets('displays all primary KPIs', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(PreferencesLoaded(preferences: defaultPreferences));

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Renewal Success Rate'), findsOneWidget);
      expect(find.text('Active MRR'), findsOneWidget);
      expect(find.text('Revenue at Risk'), findsOneWidget);
      expect(find.text('Churned'), findsOneWidget);
    });

    testWidgets('displays all secondary widgets', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(PreferencesLoaded(preferences: defaultPreferences));

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Usage Revenue'), findsOneWidget);
      expect(find.text('Total Revenue'), findsOneWidget);
      expect(find.text('Revenue Mix'), findsOneWidget);
      expect(find.text('Risk Distribution'), findsOneWidget);
      expect(find.text('Earnings Timeline'), findsOneWidget);
    });

    testWidgets('shows unsaved changes indicator', (tester) async {
      when(() => mockBloc.state).thenReturn(PreferencesLoaded(
        preferences: defaultPreferences,
        hasUnsavedChanges: true,
      ));

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Unsaved changes'), findsOneWidget);
    });

    testWidgets('removes KPI when remove button tapped', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(PreferencesLoaded(preferences: defaultPreferences));

      await tester.pumpWidget(buildTestWidget());

      // Find and tap the first remove button
      final removeButtons = find.byIcon(Icons.remove_circle_outline);
      expect(removeButtons, findsNWidgets(4));

      await tester.tap(removeButtons.first);
      await tester.pump();

      verify(() => mockBloc
          .add(const RemovePrimaryKpiRequested(KpiType.renewalSuccessRate))).called(1);
    });

    testWidgets('toggles secondary widget when switch tapped', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(PreferencesLoaded(preferences: defaultPreferences));

      await tester.pumpWidget(buildTestWidget());

      // Scroll down to make the switches visible
      await tester.drag(find.byType(SingleChildScrollView), const Offset(0, -300));
      await tester.pumpAndSettle();

      // Find the Usage Revenue switch (first switch) and toggle it
      final switches = find.byType(Switch);
      expect(switches, findsNWidgets(5));

      await tester.tap(switches.first);
      await tester.pump();

      verify(() => mockBloc.add(
          const ToggleSecondaryWidgetRequested(SecondaryWidget.usageRevenue))).called(1);
    });

    testWidgets('save button is disabled when no unsaved changes',
        (tester) async {
      when(() => mockBloc.state).thenReturn(PreferencesLoaded(
        preferences: defaultPreferences,
        hasUnsavedChanges: false,
      ));

      await tester.pumpWidget(buildTestWidget());

      final saveButton =
          tester.widget<ElevatedButton>(find.widgetWithText(ElevatedButton, 'Save'));
      expect(saveButton.onPressed, isNull);
    });

    testWidgets('save button is enabled when there are unsaved changes',
        (tester) async {
      when(() => mockBloc.state).thenReturn(PreferencesLoaded(
        preferences: defaultPreferences,
        hasUnsavedChanges: true,
      ));

      await tester.pumpWidget(buildTestWidget());

      final saveButton =
          tester.widget<ElevatedButton>(find.widgetWithText(ElevatedButton, 'Save'));
      expect(saveButton.onPressed, isNotNull);
    });

    testWidgets('reset button resets preferences to default', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(PreferencesLoaded(preferences: customPreferences));

      await tester.pumpWidget(buildTestWidget());

      await tester.tap(find.text('Reset to Default'));
      await tester.pump();

      verify(() => mockBloc.add(const ResetPreferencesRequested())).called(1);
    });

    testWidgets('shows KPI count', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(PreferencesLoaded(preferences: customPreferences));

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('(2/4)'), findsOneWidget);
    });

    testWidgets('shows Add KPI button when under max', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(PreferencesLoaded(preferences: customPreferences));

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Add KPI'), findsOneWidget);
    });

    testWidgets('hides Add KPI button when at max', (tester) async {
      when(() => mockBloc.state)
          .thenReturn(PreferencesLoaded(preferences: defaultPreferences));

      await tester.pumpWidget(buildTestWidget());

      // At 4 KPIs, the Add KPI button should not be visible
      expect(find.text('Add KPI'), findsNothing);
    });

    testWidgets('shows saving indicator when saving', (tester) async {
      when(() => mockBloc.state).thenReturn(PreferencesLoaded(
        preferences: defaultPreferences,
        isSaving: true,
      ));

      await tester.pumpWidget(buildTestWidget());

      // There should be a circular progress indicator with key 'save-progress'
      expect(find.byKey(const Key('save-progress')), findsOneWidget);
    });
  });
}
