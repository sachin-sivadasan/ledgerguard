import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/user_profile.dart';
import 'package:ledgerguard/presentation/blocs/role/role.dart';
import 'package:ledgerguard/presentation/widgets/role_guard.dart';

class MockRoleBloc extends Mock implements RoleBloc {}

void main() {
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

  setUp(() {
    mockRoleBloc = MockRoleBloc();
  });

  Widget buildTestWidget({
    required Widget child,
  }) {
    return MaterialApp(
      home: BlocProvider<RoleBloc>.value(
        value: mockRoleBloc,
        child: Scaffold(body: child),
      ),
    );
  }

  group('RoleGuard', () {
    testWidgets('shows child for owner when owner required', (tester) async {
      when(() => mockRoleBloc.state).thenReturn(const RoleLoaded(ownerProfile));
      when(() => mockRoleBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget(
        child: const RoleGuard.ownerOnly(
          child: Text('Owner Content'),
        ),
      ));

      expect(find.text('Owner Content'), findsOneWidget);
    });

    testWidgets('hides child for admin when owner required', (tester) async {
      when(() => mockRoleBloc.state).thenReturn(const RoleLoaded(adminProfile));
      when(() => mockRoleBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget(
        child: const RoleGuard.ownerOnly(
          child: Text('Owner Content'),
        ),
      ));

      expect(find.text('Owner Content'), findsNothing);
    });

    testWidgets('shows child for owner when admin required', (tester) async {
      when(() => mockRoleBloc.state).thenReturn(const RoleLoaded(ownerProfile));
      when(() => mockRoleBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget(
        child: const RoleGuard.adminOnly(
          child: Text('Admin Content'),
        ),
      ));

      expect(find.text('Admin Content'), findsOneWidget);
    });

    testWidgets('shows child for admin when admin required', (tester) async {
      when(() => mockRoleBloc.state).thenReturn(const RoleLoaded(adminProfile));
      when(() => mockRoleBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget(
        child: const RoleGuard.adminOnly(
          child: Text('Admin Content'),
        ),
      ));

      expect(find.text('Admin Content'), findsOneWidget);
    });

    testWidgets('shows fallback when role not met', (tester) async {
      when(() => mockRoleBloc.state).thenReturn(const RoleLoaded(adminProfile));
      when(() => mockRoleBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget(
        child: const RoleGuard.ownerOnly(
          child: Text('Owner Content'),
          fallback: Text('Access Denied'),
        ),
      ));

      expect(find.text('Owner Content'), findsNothing);
      expect(find.text('Access Denied'), findsOneWidget);
    });

    testWidgets('shows nothing when role not loaded and no fallback', (tester) async {
      when(() => mockRoleBloc.state).thenReturn(const RoleInitial());
      when(() => mockRoleBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget(
        child: const RoleGuard.ownerOnly(
          child: Text('Owner Content'),
        ),
      ));

      expect(find.text('Owner Content'), findsNothing);
      expect(find.byType(SizedBox), findsWidgets);
    });

    testWidgets('shows loading indicator when loading and showLoading true', (tester) async {
      when(() => mockRoleBloc.state).thenReturn(const RoleLoading());
      when(() => mockRoleBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget(
        child: const RoleGuard.ownerOnly(
          child: Text('Owner Content'),
          showLoading: true,
        ),
      ));

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });
  });

  group('ProGuard', () {
    testWidgets('shows child for Pro tier user', (tester) async {
      when(() => mockRoleBloc.state).thenReturn(const RoleLoaded(ownerProfile));
      when(() => mockRoleBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget(
        child: const ProGuard(
          child: Text('Pro Content'),
        ),
      ));

      expect(find.text('Pro Content'), findsOneWidget);
    });

    testWidgets('hides child for Starter tier user', (tester) async {
      when(() => mockRoleBloc.state).thenReturn(const RoleLoaded(adminProfile));
      when(() => mockRoleBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget(
        child: const ProGuard(
          child: Text('Pro Content'),
        ),
      ));

      expect(find.text('Pro Content'), findsNothing);
    });

    testWidgets('shows fallback for Starter tier user', (tester) async {
      when(() => mockRoleBloc.state).thenReturn(const RoleLoaded(adminProfile));
      when(() => mockRoleBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget(
        child: const ProGuard(
          child: Text('Pro Content'),
          fallback: Text('Upgrade to Pro'),
        ),
      ));

      expect(find.text('Pro Content'), findsNothing);
      expect(find.text('Upgrade to Pro'), findsOneWidget);
    });
  });
}
