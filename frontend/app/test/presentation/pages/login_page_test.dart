import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/user_entity.dart';
import 'package:ledgerguard/presentation/blocs/auth/auth.dart';
import 'package:ledgerguard/presentation/pages/login_page.dart';

class MockAuthBloc extends Mock implements AuthBloc {}

class FakeAuthEvent extends Fake implements AuthEvent {}

class FakeAuthState extends Fake implements AuthState {}

void main() {
  late MockAuthBloc mockAuthBloc;

  setUpAll(() {
    registerFallbackValue(FakeAuthEvent());
    registerFallbackValue(FakeAuthState());
  });

  setUp(() {
    mockAuthBloc = MockAuthBloc();
  });

  Widget buildTestWidget() {
    return MaterialApp(
      home: BlocProvider<AuthBloc>.value(
        value: mockAuthBloc,
        child: const LoginPage(),
      ),
    );
  }

  group('LoginPage', () {
    testWidgets('renders email and password fields', (tester) async {
      when(() => mockAuthBloc.state).thenReturn(const AuthInitial());
      when(() => mockAuthBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget());

      expect(find.byType(TextFormField), findsNWidgets(2));
      expect(find.text('Email'), findsOneWidget);
      expect(find.text('Password'), findsOneWidget);
    });

    testWidgets('renders login button', (tester) async {
      when(() => mockAuthBloc.state).thenReturn(const AuthInitial());
      when(() => mockAuthBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget());

      expect(find.widgetWithText(ElevatedButton, 'Sign In'), findsOneWidget);
    });

    testWidgets('renders Google sign in button', (tester) async {
      when(() => mockAuthBloc.state).thenReturn(const AuthInitial());
      when(() => mockAuthBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget());

      expect(find.widgetWithText(OutlinedButton, 'Continue with Google'), findsOneWidget);
    });

    testWidgets('renders link to signup page', (tester) async {
      when(() => mockAuthBloc.state).thenReturn(const AuthInitial());
      when(() => mockAuthBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget());

      expect(find.text("Don't have an account?"), findsOneWidget);
      expect(find.text('Sign Up'), findsOneWidget);
    });

    testWidgets('shows loading indicator when AuthLoading', (tester) async {
      when(() => mockAuthBloc.state).thenReturn(const AuthLoading());
      when(() => mockAuthBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget());

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });

    testWidgets('shows error message when AuthError', (tester) async {
      when(() => mockAuthBloc.state).thenReturn(const AuthError('Invalid credentials'));
      when(() => mockAuthBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Invalid credentials'), findsOneWidget);
    });

    testWidgets('dispatches SignInWithEmailRequested on login button tap', (tester) async {
      when(() => mockAuthBloc.state).thenReturn(const AuthInitial());
      when(() => mockAuthBloc.stream).thenAnswer((_) => const Stream.empty());
      when(() => mockAuthBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget());

      await tester.enterText(find.byType(TextFormField).first, 'test@example.com');
      await tester.enterText(find.byType(TextFormField).last, 'password123');
      await tester.tap(find.widgetWithText(ElevatedButton, 'Sign In'));
      await tester.pump();

      verify(() => mockAuthBloc.add(
        const SignInWithEmailRequested(
          email: 'test@example.com',
          password: 'password123',
        ),
      )).called(1);
    });

    testWidgets('dispatches SignInWithGoogleRequested on Google button tap', (tester) async {
      when(() => mockAuthBloc.state).thenReturn(const AuthInitial());
      when(() => mockAuthBloc.stream).thenAnswer((_) => const Stream.empty());
      when(() => mockAuthBloc.add(any())).thenReturn(null);

      await tester.pumpWidget(buildTestWidget());

      await tester.tap(find.widgetWithText(OutlinedButton, 'Continue with Google'));
      await tester.pump();

      verify(() => mockAuthBloc.add(const SignInWithGoogleRequested())).called(1);
    });

    testWidgets('disables buttons when loading', (tester) async {
      when(() => mockAuthBloc.state).thenReturn(const AuthLoading());
      when(() => mockAuthBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget());

      // When loading, sign in button shows CircularProgressIndicator
      expect(find.byType(CircularProgressIndicator), findsOneWidget);

      // Find buttons by type (text may not be visible during loading)
      final elevatedButtons = tester.widgetList<ElevatedButton>(find.byType(ElevatedButton));
      final outlinedButtons = tester.widgetList<OutlinedButton>(find.byType(OutlinedButton));

      // All buttons should be disabled
      for (final button in elevatedButtons) {
        expect(button.onPressed, isNull);
      }
      for (final button in outlinedButtons) {
        expect(button.onPressed, isNull);
      }
    });
  });
}
