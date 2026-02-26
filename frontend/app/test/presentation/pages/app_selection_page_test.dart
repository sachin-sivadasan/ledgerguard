import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/shopify_app.dart';
import 'package:ledgerguard/presentation/blocs/app_selection/app_selection.dart';
import 'package:ledgerguard/presentation/pages/app_selection_page.dart';

class MockAppSelectionBloc extends Mock implements AppSelectionBloc {}

class FakeAppSelectionEvent extends Fake implements AppSelectionEvent {}

void main() {
  late MockAppSelectionBloc mockBloc;

  const testApps = [
    ShopifyApp(
      id: 'app-1',
      name: 'App One',
      description: 'First app description',
      installCount: 100,
    ),
    ShopifyApp(
      id: 'app-2',
      name: 'App Two',
      description: 'Second app description',
      installCount: 200,
    ),
    ShopifyApp(
      id: 'app-3',
      name: 'App Three',
      description: 'Third app description',
    ),
  ];

  setUpAll(() {
    registerFallbackValue(FakeAppSelectionEvent());
  });

  setUp(() {
    mockBloc = MockAppSelectionBloc();
  });

  Widget buildTestWidget({AppSelectionState? state}) {
    when(() => mockBloc.state).thenReturn(state ?? const AppSelectionInitial());
    when(() => mockBloc.stream).thenAnswer((_) => const Stream.empty());

    return MaterialApp(
      home: BlocProvider<AppSelectionBloc>.value(
        value: mockBloc,
        child: const AppSelectionPage(),
      ),
    );
  }

  group('AppSelectionPage', () {
    testWidgets('renders page title', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Select App'), findsOneWidget);
      expect(find.text('Choose Your App'), findsOneWidget);
    });

    testWidgets('fetches apps on init', (tester) async {
      when(() => mockBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget());

      verify(() => mockBloc.add(const FetchAppsRequested())).called(1);
    });

    testWidgets('shows loading indicator when loading', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: const AppSelectionLoading(),
      ));

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.text('Loading apps...'), findsOneWidget);
    });

    testWidgets('shows error message when error state', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: const AppSelectionError('Failed to load apps'),
      ));

      expect(find.text('Failed to load apps'), findsOneWidget);
      expect(find.text('Retry'), findsOneWidget);
    });

    testWidgets('dispatches FetchAppsRequested on retry tap', (tester) async {
      when(() => mockBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget(
        state: const AppSelectionError('Failed to load apps'),
      ));

      await tester.tap(find.text('Retry'));
      await tester.pump();

      // Called once on init and once on retry
      verify(() => mockBloc.add(const FetchAppsRequested())).called(2);
    });

    testWidgets('shows list of apps when loaded', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: const AppSelectionLoaded(apps: testApps),
      ));

      expect(find.text('App One'), findsOneWidget);
      expect(find.text('App Two'), findsOneWidget);
      expect(find.text('App Three'), findsOneWidget);
    });

    testWidgets('shows app descriptions', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: const AppSelectionLoaded(apps: testApps),
      ));

      expect(find.text('First app description'), findsOneWidget);
      expect(find.text('Second app description'), findsOneWidget);
    });

    testWidgets('shows install counts', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: const AppSelectionLoaded(apps: testApps),
      ));

      expect(find.text('100 installs'), findsOneWidget);
      expect(find.text('200 installs'), findsOneWidget);
    });

    testWidgets('dispatches AppSelected on app tap', (tester) async {
      when(() => mockBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget(
        state: const AppSelectionLoaded(apps: testApps),
      ));

      await tester.tap(find.text('App Two'));
      await tester.pump();

      verify(() => mockBloc.add(AppSelected(testApps[1]))).called(1);
    });

    testWidgets('shows Confirm Selection button when app selected',
        (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: AppSelectionLoaded(apps: testApps, selectedApp: testApps[0]),
      ));

      expect(find.text('Confirm Selection'), findsOneWidget);
    });

    testWidgets('hides Confirm Selection button when no app selected',
        (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: const AppSelectionLoaded(apps: testApps),
      ));

      expect(find.text('Confirm Selection'), findsNothing);
    });

    testWidgets('dispatches ConfirmSelectionRequested on confirm tap',
        (tester) async {
      when(() => mockBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget(
        state: AppSelectionLoaded(apps: testApps, selectedApp: testApps[0]),
      ));

      await tester.tap(find.text('Confirm Selection'));
      await tester.pump();

      verify(() => mockBloc.add(const ConfirmSelectionRequested())).called(1);
    });

    testWidgets('shows saving indicator when saving', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: AppSelectionSaving(apps: testApps, selectedApp: testApps[0]),
      ));

      expect(find.text('Saving...'), findsOneWidget);
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });

    testWidgets('disables app selection when saving', (tester) async {
      when(() => mockBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget(
        state: AppSelectionSaving(apps: testApps, selectedApp: testApps[0]),
      ));

      // Tap on an app - should not dispatch event
      await tester.tap(find.text('App Two'));
      await tester.pump();

      // Only FetchAppsRequested from init should be called
      verifyNever(() => mockBloc.add(AppSelected(testApps[1])));
    });

    testWidgets('shows selected app with visual indicator', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: AppSelectionLoaded(apps: testApps, selectedApp: testApps[1]),
      ));

      // Selected app should have check icon
      expect(find.byIcon(Icons.check_circle), findsOneWidget);
    });
  });
}
