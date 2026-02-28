import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/user_entity.dart';
import 'package:ledgerguard/domain/entities/user_profile.dart';
import 'package:ledgerguard/presentation/blocs/auth/auth.dart';
import 'package:ledgerguard/presentation/blocs/role/role.dart';
import 'package:ledgerguard/presentation/pages/profile_page.dart';

class MockAuthBloc extends Mock implements AuthBloc {}

class MockRoleBloc extends Mock implements RoleBloc {}

class MockGoRouter extends Mock implements GoRouter {}

class FakeAuthEvent extends Fake implements AuthEvent {}

void main() {
  late MockAuthBloc mockAuthBloc;
  late MockRoleBloc mockRoleBloc;

  const testUser = UserEntity(
    id: 'user-123',
    email: 'test@example.com',
  );

  const ownerProfile = UserProfile(
    id: 'user-123',
    email: 'test@example.com',
    role: UserRole.owner,
    planTier: PlanTier.pro,
  );

  const adminProfile = UserProfile(
    id: 'user-123',
    email: 'test@example.com',
    role: UserRole.admin,
    planTier: PlanTier.starter,
  );

  setUpAll(() {
    registerFallbackValue(FakeAuthEvent());
  });

  setUp(() {
    mockAuthBloc = MockAuthBloc();
    mockRoleBloc = MockRoleBloc();
  });

  Widget buildTestWidget({
    AuthState? authState,
    RoleState? roleState,
  }) {
    when(() => mockAuthBloc.state)
        .thenReturn(authState ?? const Authenticated(testUser));
    when(() => mockAuthBloc.stream).thenAnswer((_) => const Stream.empty());
    when(() => mockAuthBloc.add(any())).thenReturn(null);

    when(() => mockRoleBloc.state)
        .thenReturn(roleState ?? const RoleLoaded(ownerProfile));
    when(() => mockRoleBloc.stream).thenAnswer((_) => const Stream.empty());

    return MaterialApp(
      home: MultiBlocProvider(
        providers: [
          BlocProvider<AuthBloc>.value(value: mockAuthBloc),
          BlocProvider<RoleBloc>.value(value: mockRoleBloc),
        ],
        child: const ProfilePage(),
      ),
    );
  }

  group('ProfilePage', () {
    testWidgets('shows app bar title', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Profile'), findsOneWidget);
    });

    testWidgets('shows loading when auth not authenticated', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        authState: const Unauthenticated(),
      ));

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });

    testWidgets('shows loading when role is loading', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        roleState: const RoleLoading(),
      ));

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });

    testWidgets('shows error state when role fails', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        roleState: const RoleError('Network error'),
      ));

      expect(find.text('Failed to load profile'), findsOneWidget);
      expect(find.text('Network error'), findsOneWidget);
    });

    testWidgets('shows user email', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      expect(find.text('test@example.com'), findsWidgets);
    });

    testWidgets('shows user initials in avatar', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      // Initials are derived from email: test@example.com -> 'T'
      expect(find.text('T'), findsOneWidget);
    });

    testWidgets('shows role badge for owner', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        roleState: const RoleLoaded(ownerProfile),
      ));

      expect(find.text('Owner'), findsWidgets);
    });

    testWidgets('shows role badge for admin', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        roleState: const RoleLoaded(adminProfile),
      ));

      expect(find.text('Admin'), findsWidgets);
    });

    testWidgets('shows plan badge for pro', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        roleState: const RoleLoaded(ownerProfile),
      ));

      expect(find.text('Pro'), findsWidgets);
    });

    testWidgets('shows plan badge for free', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        roleState: const RoleLoaded(adminProfile),
      ));

      expect(find.text('Free'), findsWidgets);
    });

    testWidgets('shows upgrade card for free tier', (tester) async {
      // Use a larger screen to ensure content is visible
      tester.view.physicalSize = const Size(800, 1200);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(() {
        tester.view.resetPhysicalSize();
        tester.view.resetDevicePixelRatio();
      });

      await tester.pumpWidget(buildTestWidget(
        roleState: const RoleLoaded(adminProfile),
      ));
      await tester.pumpAndSettle();

      // Upgrade card should be visible for free tier - shows "Upgrade to Pro" button
      // and "Free Plan" text
      expect(find.text('Free Plan'), findsOneWidget);
      expect(find.text('Upgrade to Pro'), findsOneWidget);
    });

    testWidgets('hides upgrade card for pro tier', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        roleState: const RoleLoaded(ownerProfile),
      ));
      await tester.pumpAndSettle();

      expect(find.text('Upgrade to Pro'), findsNothing);
    });

    testWidgets('shows upgrade coming soon snackbar on tap', (tester) async {
      // Use a larger screen to ensure content is visible
      tester.view.physicalSize = const Size(800, 1200);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(() {
        tester.view.resetPhysicalSize();
        tester.view.resetDevicePixelRatio();
      });

      await tester.pumpWidget(buildTestWidget(
        roleState: const RoleLoaded(adminProfile),
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Upgrade to Pro'));
      await tester.pumpAndSettle();

      expect(find.text('Upgrade coming soon!'), findsOneWidget);
    });

    testWidgets('shows notification settings link', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      // Scroll to find Notifications in Settings section
      await tester.pumpAndSettle();
      final notificationsFinder = find.text('Notifications');
      if (notificationsFinder.evaluate().isEmpty) {
        await tester.dragUntilVisible(
          notificationsFinder,
          find.byType(Scrollable),
          const Offset(0, -100),
        );
      }
      await tester.pumpAndSettle();

      expect(notificationsFinder, findsOneWidget);
    });

    testWidgets('shows logout button', (tester) async {
      // Use a larger screen to ensure all content is visible
      tester.view.physicalSize = const Size(800, 2000);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(() {
        tester.view.resetPhysicalSize();
        tester.view.resetDevicePixelRatio();
      });

      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      expect(find.text('Log Out'), findsOneWidget);
    });

    testWidgets('shows logout confirmation dialog on tap', (tester) async {
      // Use a larger screen to ensure all content is visible
      tester.view.physicalSize = const Size(800, 2000);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(() {
        tester.view.resetPhysicalSize();
        tester.view.resetDevicePixelRatio();
      });

      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      await tester.tap(find.text('Log Out').first);
      await tester.pumpAndSettle();

      expect(find.text('Are you sure you want to log out?'), findsOneWidget);
      expect(find.text('Cancel'), findsOneWidget);
    });

    testWidgets('dismisses dialog on cancel', (tester) async {
      // Use a larger screen to ensure all content is visible
      tester.view.physicalSize = const Size(800, 2000);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(() {
        tester.view.resetPhysicalSize();
        tester.view.resetDevicePixelRatio();
      });

      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      await tester.tap(find.text('Log Out').first);
      await tester.pumpAndSettle();

      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();

      expect(find.text('Are you sure you want to log out?'), findsNothing);
    });

    testWidgets('dispatches SignOutRequested on confirm logout', (tester) async {
      // Use a larger screen to ensure all content is visible
      tester.view.physicalSize = const Size(800, 2000);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(() {
        tester.view.resetPhysicalSize();
        tester.view.resetDevicePixelRatio();
      });

      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      await tester.tap(find.text('Log Out').first);
      await tester.pumpAndSettle();

      // Tap the logout button in the dialog (ElevatedButton, not the page button)
      await tester.tap(find.widgetWithText(ElevatedButton, 'Log Out'));
      await tester.pumpAndSettle();

      verify(() => mockAuthBloc.add(any(that: isA<SignOutRequested>()))).called(1);
    });

    testWidgets('shows account section', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      // Account section header
      expect(find.text('Account'), findsOneWidget);
      // Email Address tile within account section
      expect(find.text('Email Address'), findsOneWidget);
      // Account Security tile
      expect(find.text('Account Security'), findsOneWidget);
    });

    testWidgets('shows settings section', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      // Scroll down to find Settings section
      final settingsFinder = find.text('Settings');
      await tester.dragUntilVisible(
        settingsFinder,
        find.byType(Scrollable),
        const Offset(0, -200),
      );
      await tester.pumpAndSettle();

      expect(settingsFinder, findsOneWidget);
    });

    testWidgets('shows integrations section', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      // Scroll to find Integrations section
      final integrationsFinder = find.text('Integrations');
      await tester.dragUntilVisible(
        integrationsFinder,
        find.byType(Scrollable),
        const Offset(0, -100),
      );
      await tester.pumpAndSettle();

      expect(integrationsFinder, findsOneWidget);
      expect(find.text('Shopify Partner'), findsOneWidget);
    });
  });
}
