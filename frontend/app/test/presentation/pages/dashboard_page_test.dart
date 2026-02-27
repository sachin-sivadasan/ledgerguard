import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:get_it/get_it.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/dashboard_metrics.dart';
import 'package:ledgerguard/domain/entities/time_range.dart';
import 'package:ledgerguard/domain/entities/user_profile.dart';
import 'package:ledgerguard/presentation/blocs/dashboard/dashboard.dart';
import 'package:ledgerguard/presentation/blocs/earnings/earnings.dart';
import 'package:ledgerguard/presentation/blocs/insight/insight.dart';
import 'package:ledgerguard/presentation/blocs/preferences/preferences.dart';
import 'package:ledgerguard/presentation/blocs/role/role.dart';
import 'package:ledgerguard/presentation/pages/dashboard_page.dart';

class MockDashboardBloc extends Mock implements DashboardBloc {}

class MockRoleBloc extends Mock implements RoleBloc {}

class MockInsightBloc extends Mock implements InsightBloc {}

class MockPreferencesBloc extends Mock implements PreferencesBloc {}

class MockEarningsBloc extends Mock implements EarningsBloc {}

class FakeDashboardEvent extends Fake implements DashboardEvent {}

class FakeInsightEvent extends Fake implements InsightEvent {}

class FakePreferencesEvent extends Fake implements PreferencesEvent {}

class FakeEarningsEvent extends Fake implements EarningsEvent {}

void main() {
  late MockDashboardBloc mockBloc;
  late MockRoleBloc mockRoleBloc;
  late MockInsightBloc mockInsightBloc;
  late MockPreferencesBloc mockPreferencesBloc;
  late MockEarningsBloc mockEarningsBloc;

  const proUserProfile = UserProfile(
    id: 'user-1',
    email: 'pro@example.com',
    role: UserRole.owner,
    planTier: PlanTier.pro,
  );

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

  setUpAll(() {
    registerFallbackValue(FakeDashboardEvent());
    registerFallbackValue(FakeInsightEvent());
    registerFallbackValue(FakePreferencesEvent());
    registerFallbackValue(FakeEarningsEvent());
  });

  setUp(() {
    mockBloc = MockDashboardBloc();
    mockRoleBloc = MockRoleBloc();
    mockInsightBloc = MockInsightBloc();
    mockPreferencesBloc = MockPreferencesBloc();
    mockEarningsBloc = MockEarningsBloc();

    // Setup RoleBloc defaults
    when(() => mockRoleBloc.state).thenReturn(const RoleLoaded(proUserProfile));
    when(() => mockRoleBloc.stream).thenAnswer((_) => const Stream.empty());

    // Setup InsightBloc defaults - empty state so card is hidden
    when(() => mockInsightBloc.state).thenReturn(const InsightEmpty());
    when(() => mockInsightBloc.stream).thenAnswer((_) => const Stream.empty());
    when(() => mockInsightBloc.add(any())).thenReturn(null);

    // Setup PreferencesBloc defaults
    when(() => mockPreferencesBloc.state).thenReturn(const PreferencesInitial());
    when(() => mockPreferencesBloc.stream).thenAnswer((_) => const Stream.empty());
    when(() => mockPreferencesBloc.add(any())).thenReturn(null);

    // Setup EarningsBloc defaults
    when(() => mockEarningsBloc.state).thenReturn(const EarningsInitial());
    when(() => mockEarningsBloc.stream).thenAnswer((_) => const Stream.empty());
    when(() => mockEarningsBloc.add(any())).thenReturn(null);
    when(() => mockEarningsBloc.close()).thenAnswer((_) async {});

    // Register EarningsBloc in GetIt for the dashboard page
    GetIt.instance.registerFactory<EarningsBloc>(() => mockEarningsBloc);
  });

  tearDown(() {
    // Clean up GetIt
    if (GetIt.instance.isRegistered<EarningsBloc>()) {
      GetIt.instance.unregister<EarningsBloc>();
    }
  });

  Widget buildTestWidget({DashboardState? state}) {
    when(() => mockBloc.state).thenReturn(state ?? const DashboardInitial());
    when(() => mockBloc.stream).thenAnswer((_) => const Stream.empty());
    when(() => mockBloc.add(any())).thenReturn(null);

    return MaterialApp(
      home: MultiBlocProvider(
        providers: [
          BlocProvider<DashboardBloc>.value(value: mockBloc),
          BlocProvider<RoleBloc>.value(value: mockRoleBloc),
          BlocProvider<InsightBloc>.value(value: mockInsightBloc),
          BlocProvider<PreferencesBloc>.value(value: mockPreferencesBloc),
        ],
        child: const DashboardPage(),
      ),
    );
  }

  Future<void> setLargeScreen(WidgetTester tester) async {
    tester.view.physicalSize = const Size(1200, 900);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
  }

  group('DashboardPage', () {
    testWidgets('renders page title', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: DashboardLoaded(metrics: testMetrics, timeRange: TimeRange.thisMonth()),
      ));

      expect(find.text('Dashboard'), findsOneWidget);
    });

    testWidgets('fetches metrics on init when in initial state', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      verify(() => mockBloc.add(const LoadDashboardRequested())).called(1);
    });

    testWidgets('shows loading indicator when loading', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: const DashboardLoading(),
      ));

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });

    testWidgets('shows error message when error state', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: const DashboardError('Network error'),
      ));

      expect(find.text('Failed to load dashboard'), findsOneWidget);
      expect(find.text('Network error'), findsOneWidget);
      expect(find.text('Retry'), findsOneWidget);
    });

    testWidgets('dispatches LoadDashboardRequested on retry tap', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        state: const DashboardError('Network error'),
      ));

      await tester.tap(find.text('Retry'));
      await tester.pump();

      verify(() => mockBloc.add(const LoadDashboardRequested())).called(1);
    });

    group('Empty state', () {
      testWidgets('shows empty state message', (tester) async {
        await tester.pumpWidget(buildTestWidget(
          state: const DashboardEmpty(),
        ));

        expect(find.text('No Metrics Yet'), findsOneWidget);
        expect(find.text('No metrics available. Sync your app data to see metrics.'), findsOneWidget);
      });

      testWidgets('shows sync button in empty state', (tester) async {
        await tester.pumpWidget(buildTestWidget(
          state: const DashboardEmpty(),
        ));

        expect(find.text('Sync Data'), findsOneWidget);
      });

      testWidgets('dispatches RefreshDashboardRequested on sync tap', (tester) async {
        await tester.pumpWidget(buildTestWidget(
          state: const DashboardEmpty(),
        ));

        await tester.tap(find.text('Sync Data'));
        await tester.pump();

        verify(() => mockBloc.add(const RefreshDashboardRequested())).called(1);
      });

      testWidgets('shows custom empty message', (tester) async {
        await tester.pumpWidget(buildTestWidget(
          state: const DashboardEmpty(message: 'Custom empty message'),
        ));

        expect(find.text('Custom empty message'), findsOneWidget);
      });
    });

    group('Primary KPIs', () {
      testWidgets('displays Renewal Success Rate', (tester) async {
        await tester.pumpWidget(buildTestWidget(
          state: DashboardLoaded(metrics: testMetrics, timeRange: TimeRange.thisMonth()),
        ));

        expect(find.text('Renewal Success Rate'), findsOneWidget);
        expect(find.text('94.2%'), findsOneWidget);
      });

      testWidgets('displays Active MRR', (tester) async {
        await tester.pumpWidget(buildTestWidget(
          state: DashboardLoaded(metrics: testMetrics, timeRange: TimeRange.thisMonth()),
        ));

        expect(find.text('Active MRR'), findsOneWidget);
        expect(find.text('\$124.5K'), findsOneWidget);
      });

      testWidgets('displays Revenue at Risk', (tester) async {
        await tester.pumpWidget(buildTestWidget(
          state: DashboardLoaded(metrics: testMetrics, timeRange: TimeRange.thisMonth()),
        ));

        expect(find.text('Revenue at Risk'), findsOneWidget);
        expect(find.text('\$18.5K'), findsOneWidget);
      });

      testWidgets('displays Churned metrics', (tester) async {
        await setLargeScreen(tester);
        await tester.pumpWidget(buildTestWidget(
          state: DashboardLoaded(metrics: testMetrics, timeRange: TimeRange.thisMonth()),
        ));

        // 'Churned' appears in both primary KPI and risk distribution
        expect(find.text('Churned'), findsAtLeastNWidgets(1));
        expect(find.text('\$3.2K'), findsOneWidget);
        expect(find.text('12 subscriptions'), findsOneWidget);
      });
    });

    group('Secondary Section', () {
      testWidgets('displays Usage Revenue', (tester) async {
        await tester.pumpWidget(buildTestWidget(
          state: DashboardLoaded(metrics: testMetrics, timeRange: TimeRange.thisMonth()),
        ));

        expect(find.text('Usage Revenue'), findsOneWidget);
        expect(find.text('\$23.4K'), findsOneWidget);
      });

      testWidgets('displays Total Revenue', (tester) async {
        await tester.pumpWidget(buildTestWidget(
          state: DashboardLoaded(metrics: testMetrics, timeRange: TimeRange.thisMonth()),
        ));

        expect(find.text('Total Revenue'), findsOneWidget);
        expect(find.text('\$152.4K'), findsOneWidget);
      });

      testWidgets('displays Revenue Mix chart', (tester) async {
        await tester.pumpWidget(buildTestWidget(
          state: DashboardLoaded(metrics: testMetrics, timeRange: TimeRange.thisMonth()),
        ));

        expect(find.text('Revenue Mix'), findsOneWidget);
        expect(find.text('Recurring'), findsOneWidget);
        expect(find.text('Usage'), findsOneWidget);
        expect(find.text('One-time'), findsOneWidget);
      });

      testWidgets('displays Risk Distribution chart', (tester) async {
        await tester.pumpWidget(buildTestWidget(
          state: DashboardLoaded(metrics: testMetrics, timeRange: TimeRange.thisMonth()),
        ));

        expect(find.text('Risk Distribution'), findsOneWidget);
        expect(find.text('Safe'), findsOneWidget);
        expect(find.text('At Risk'), findsOneWidget);
        expect(find.text('Critical'), findsOneWidget);
        // Note: 'Churned' appears in both primary KPI and risk distribution
      });

      testWidgets('displays risk counts', (tester) async {
        await tester.pumpWidget(buildTestWidget(
          state: DashboardLoaded(metrics: testMetrics, timeRange: TimeRange.thisMonth()),
        ));

        expect(find.text('842'), findsOneWidget); // Safe
        expect(find.text('45'), findsOneWidget); // At Risk
        expect(find.text('18'), findsOneWidget); // Critical
      });
    });

    group('Refresh functionality', () {
      testWidgets('shows refresh button in app bar', (tester) async {
        await tester.pumpWidget(buildTestWidget(
          state: DashboardLoaded(metrics: testMetrics, timeRange: TimeRange.thisMonth()),
        ));

        expect(find.byIcon(Icons.refresh), findsOneWidget);
      });

      testWidgets('dispatches RefreshDashboardRequested on refresh tap',
          (tester) async {
        await tester.pumpWidget(buildTestWidget(
          state: DashboardLoaded(metrics: testMetrics, timeRange: TimeRange.thisMonth()),
        ));

        await tester.tap(find.byIcon(Icons.refresh));
        await tester.pump();

        verify(() => mockBloc.add(const RefreshDashboardRequested())).called(1);
      });

      testWidgets('shows loading indicator when refreshing', (tester) async {
        await tester.pumpWidget(buildTestWidget(
          state: DashboardLoaded(metrics: testMetrics, timeRange: TimeRange.thisMonth(), isRefreshing: true),
        ));

        // Should show progress indicators (one in app bar, possibly others in child widgets)
        expect(find.byType(CircularProgressIndicator), findsAtLeastNWidgets(1));
      });

      testWidgets('disables refresh button when refreshing', (tester) async {
        await tester.pumpWidget(buildTestWidget(
          state: DashboardLoaded(metrics: testMetrics, timeRange: TimeRange.thisMonth(), isRefreshing: true),
        ));

        // Find the refresh icon button - it should be disabled
        final refreshButton = find.byIcon(Icons.refresh);
        expect(refreshButton, findsNothing); // Icon is replaced by progress
      });
    });

    group('Section headers', () {
      testWidgets('displays Primary KPIs section header', (tester) async {
        await tester.pumpWidget(buildTestWidget(
          state: DashboardLoaded(metrics: testMetrics, timeRange: TimeRange.thisMonth()),
        ));

        expect(find.text('Primary KPIs'), findsOneWidget);
      });

      testWidgets('displays Revenue & Risk section header', (tester) async {
        await tester.pumpWidget(buildTestWidget(
          state: DashboardLoaded(metrics: testMetrics, timeRange: TimeRange.thisMonth()),
        ));

        expect(find.text('Revenue & Risk'), findsOneWidget);
      });
    });
  });
}
