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

      expect(find.text('TE'), findsOneWidget); // First 2 chars of 'test'
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
      await tester.pumpWidget(buildTestWidget(
        roleState: const RoleLoaded(adminProfile),
      ));

      expect(find.text('Upgrade to Pro'), findsOneWidget);
      expect(find.text('Upgrade Now'), findsOneWidget);
    });

    testWidgets('hides upgrade card for pro tier', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        roleState: const RoleLoaded(ownerProfile),
      ));

      expect(find.text('Upgrade to Pro'), findsNothing);
    });

    testWidgets('shows upgrade coming soon snackbar on tap', (tester) async {
      await tester.pumpWidget(buildTestWidget(
        roleState: const RoleLoaded(adminProfile),
      ));

      // Scroll to make the upgrade button visible
      await tester.ensureVisible(find.text('Upgrade Now'));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Upgrade Now'));
      await tester.pumpAndSettle();

      expect(find.text('Upgrade functionality coming soon!'), findsOneWidget);
    });

    testWidgets('shows notification settings link', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Notification Settings'), findsOneWidget);
    });

    testWidgets('shows logout button', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Log Out'), findsOneWidget);
    });

    testWidgets('shows logout confirmation dialog on tap', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      // Scroll to make the logout button visible
      await tester.ensureVisible(find.widgetWithText(OutlinedButton, 'Log Out'));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(OutlinedButton, 'Log Out'));
      await tester.pumpAndSettle();

      expect(find.text('Are you sure you want to log out?'), findsOneWidget);
      expect(find.text('Cancel'), findsOneWidget);
    });

    testWidgets('dismisses dialog on cancel', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      // Scroll to make the logout button visible
      await tester.ensureVisible(find.widgetWithText(OutlinedButton, 'Log Out'));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(OutlinedButton, 'Log Out'));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();

      expect(find.text('Are you sure you want to log out?'), findsNothing);
    });

    testWidgets('dispatches SignOutRequested on confirm logout', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      // Scroll to make the logout button visible
      await tester.ensureVisible(find.widgetWithText(OutlinedButton, 'Log Out'));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(OutlinedButton, 'Log Out'));
      await tester.pumpAndSettle();

      // Tap the logout button in the dialog (not the page button)
      await tester.tap(find.widgetWithText(TextButton, 'Log Out'));
      await tester.pumpAndSettle();

      verify(() => mockAuthBloc.add(any(that: isA<SignOutRequested>()))).called(1);
    });

    testWidgets('shows account section', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Account'), findsOneWidget);
      expect(find.text('Email'), findsOneWidget);
      expect(find.text('Role'), findsOneWidget);
      expect(find.text('Plan'), findsOneWidget);
    });

    testWidgets('shows settings section', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Settings'), findsOneWidget);
    });
  });
}
