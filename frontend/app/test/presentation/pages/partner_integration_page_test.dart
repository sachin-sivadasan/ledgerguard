import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/partner_integration.dart';
import 'package:ledgerguard/domain/entities/user_profile.dart';
import 'package:ledgerguard/presentation/blocs/partner_integration/partner_integration.dart';
import 'package:ledgerguard/presentation/blocs/role/role.dart';
import 'package:ledgerguard/presentation/pages/partner_integration_page.dart';

class MockPartnerIntegrationBloc extends Mock
    implements PartnerIntegrationBloc {}

class MockRoleBloc extends Mock implements RoleBloc {}

class FakePartnerIntegrationEvent extends Fake
    implements PartnerIntegrationEvent {}

void main() {
  late MockPartnerIntegrationBloc mockIntegrationBloc;
  late MockRoleBloc mockRoleBloc;

  const ownerProfile = UserProfile(
    id: 'user-1',
    email: 'owner@example.com',
    role: UserRole.owner,
    planTier: PlanTier.pro,
  );

  const adminProfile = UserProfile(
    id: 'user-2',
    email: 'admin@example.com',
    role: UserRole.admin,
    planTier: PlanTier.starter,
  );

  final connectedIntegration = PartnerIntegration(
    partnerId: 'partner-123',
    status: IntegrationStatus.connected,
    connectedAt: DateTime(2024, 1, 1),
  );

  setUpAll(() {
    registerFallbackValue(FakePartnerIntegrationEvent());
  });

  setUp(() {
    mockIntegrationBloc = MockPartnerIntegrationBloc();
    mockRoleBloc = MockRoleBloc();
  });

  Widget buildTestWidget({
    PartnerIntegrationState? integrationState,
    RoleState? roleState,
  }) {
    when(() => mockIntegrationBloc.state)
        .thenReturn(integrationState ?? const PartnerIntegrationInitial());
    when(() => mockIntegrationBloc.stream)
        .thenAnswer((_) => const Stream.empty());

    when(() => mockRoleBloc.state)
        .thenReturn(roleState ?? const RoleLoaded(ownerProfile));
    when(() => mockRoleBloc.stream).thenAnswer((_) => const Stream.empty());

    return MaterialApp(
      home: MultiBlocProvider(
        providers: [
          BlocProvider<PartnerIntegrationBloc>.value(
            value: mockIntegrationBloc,
          ),
          BlocProvider<RoleBloc>.value(value: mockRoleBloc),
        ],
        child: const PartnerIntegrationPage(),
      ),
    );
  }

  group('PartnerIntegrationPage', () {
    testWidgets('renders page title', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Partner Integration'), findsOneWidget);
      expect(find.text('Connect Shopify Partner Account'), findsOneWidget);
    });

    testWidgets('renders Connect Shopify Partner button', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        integrationState: const PartnerIntegrationNotConnected(),
      ));

      expect(find.text('Connect Shopify Partner'), findsOneWidget);
      expect(find.text('Connect with OAuth'), findsOneWidget);
    });

    testWidgets('dispatches ConnectWithOAuthRequested on OAuth button tap',
        (tester) async {
      when(() => mockIntegrationBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget(
        integrationState: const PartnerIntegrationNotConnected(),
      ));

      await tester.tap(find.text('Connect Shopify Partner'));
      await tester.pump();

      verify(() => mockIntegrationBloc.add(const ConnectWithOAuthRequested()))
          .called(1);
    });

    testWidgets('shows Manual Token section for admin users', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        integrationState: const PartnerIntegrationNotConnected(),
        roleState: const RoleLoaded(adminProfile),
      ));

      expect(find.text('Manual Token Entry'), findsOneWidget);
      expect(find.text('Partner ID'), findsOneWidget);
      expect(find.text('API Token'), findsOneWidget);
      expect(find.text('Save Token'), findsOneWidget);
    });

    testWidgets('shows Manual Token section for owner users', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        integrationState: const PartnerIntegrationNotConnected(),
        roleState: const RoleLoaded(ownerProfile),
      ));

      expect(find.text('Manual Token Entry'), findsOneWidget);
    });

    testWidgets('hides Manual Token section when role not loaded',
        (tester) async {
      await tester.pumpWidget(buildTestWidget(
        integrationState: const PartnerIntegrationNotConnected(),
        roleState: const RoleInitial(),
      ));

      expect(find.text('Manual Token Entry'), findsNothing);
    });

    testWidgets('dispatches SaveManualTokenRequested on Save Token tap',
        (tester) async {
      when(() => mockIntegrationBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget(
        integrationState: const PartnerIntegrationNotConnected(),
        roleState: const RoleLoaded(ownerProfile),
      ));

      // Scroll to make Save Token visible
      await tester.scrollUntilVisible(
        find.text('Save Token'),
        200,
        scrollable: find.byType(Scrollable).first,
      );

      // Enter values
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Partner ID'),
        'my-partner-id',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'API Token'),
        'my-api-token',
      );

      // Tap save button
      await tester.tap(find.text('Save Token'));
      await tester.pump();

      verify(() => mockIntegrationBloc.add(
            const SaveManualTokenRequested(
              partnerId: 'my-partner-id',
              apiToken: 'my-api-token',
            ),
          )).called(1);
    });

    testWidgets('shows validation error when Partner ID is empty',
        (tester) async {
      await tester.pumpWidget(buildTestWidget(
        integrationState: const PartnerIntegrationNotConnected(),
        roleState: const RoleLoaded(ownerProfile),
      ));

      // Scroll to make Save Token visible
      await tester.scrollUntilVisible(
        find.text('Save Token'),
        200,
        scrollable: find.byType(Scrollable).first,
      );

      // Tap save without entering values
      await tester.tap(find.text('Save Token'));
      await tester.pump();

      expect(find.text('Partner ID is required'), findsOneWidget);
    });

    testWidgets('shows validation error when API Token is empty',
        (tester) async {
      await tester.pumpWidget(buildTestWidget(
        integrationState: const PartnerIntegrationNotConnected(),
        roleState: const RoleLoaded(ownerProfile),
      ));

      // Scroll to make Save Token visible
      await tester.scrollUntilVisible(
        find.text('Save Token'),
        200,
        scrollable: find.byType(Scrollable).first,
      );

      // Enter only Partner ID
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Partner ID'),
        'my-partner-id',
      );

      await tester.tap(find.text('Save Token'));
      await tester.pump();

      expect(find.text('API Token is required'), findsOneWidget);
    });

    testWidgets('shows loading indicator when loading', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        integrationState:
            const PartnerIntegrationLoading(message: 'Connecting...'),
      ));

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.text('Connecting...'), findsOneWidget);
    });

    testWidgets('shows connected state with partner ID', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        integrationState: PartnerIntegrationConnected(connectedIntegration),
      ));

      expect(find.text('Connected'), findsOneWidget);
      expect(find.text('Partner ID: partner-123'), findsOneWidget);
      expect(find.text('Disconnect'), findsOneWidget);
    });

    testWidgets('dispatches DisconnectRequested on Disconnect tap',
        (tester) async {
      when(() => mockIntegrationBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget(
        integrationState: PartnerIntegrationConnected(connectedIntegration),
      ));

      await tester.tap(find.text('Disconnect'));
      await tester.pump();

      verify(() => mockIntegrationBloc.add(const DisconnectRequested()))
          .called(1);
    });

    testWidgets('shows error message when error state', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        integrationState:
            const PartnerIntegrationError('Connection failed'),
      ));

      expect(find.text('Connection failed'), findsOneWidget);
    });

    testWidgets('shows success state with connected card', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        integrationState: PartnerIntegrationSuccess(
          integration: connectedIntegration,
          message: 'Successfully connected!',
        ),
      ));

      expect(find.text('Connected'), findsOneWidget);
      expect(find.text('Partner ID: partner-123'), findsOneWidget);
    });

    testWidgets('shows only loading indicator when loading', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        integrationState: const PartnerIntegrationLoading(),
        roleState: const RoleLoaded(ownerProfile),
      ));

      // When loading, the OAuth and Manual sections are not shown
      // Only the loading indicator is shown
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.text('Connect Shopify Partner'), findsNothing);
      expect(find.text('Save Token'), findsNothing);
    });

    testWidgets('checks integration status on init', (tester) async {
      when(() => mockIntegrationBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget());

      verify(() => mockIntegrationBloc
          .add(const CheckIntegrationStatusRequested())).called(1);
    });
  });
}
