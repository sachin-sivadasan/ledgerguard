import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/presentation/blocs/auth/auth.dart';
import 'package:ledgerguard/presentation/pages/signup_page.dart';

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
        child: const SignupPage(),
      ),
    );
  }

  group('SignupPage', () {
    testWidgets('renders email and password fields', (tester) async {
      when(() => mockAuthBloc.state).thenReturn(const AuthInitial());
      when(() => mockAuthBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget());

      expect(find.byType(TextFormField), findsNWidgets(2));
      expect(find.text('Email'), findsOneWidget);
      expect(find.text('Password'), findsOneWidget);
    });

    testWidgets('renders create account button', (tester) async {
      when(() => mockAuthBloc.state).thenReturn(const AuthInitial());
      when(() => mockAuthBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget());

      expect(find.widgetWithText(ElevatedButton, 'Create Account'), findsOneWidget);
    });

    testWidgets('renders Google sign up button', (tester) async {
      when(() => mockAuthBloc.state).thenReturn(const AuthInitial());
      when(() => mockAuthBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget());

      expect(find.widgetWithText(OutlinedButton, 'Continue with Google'), findsOneWidget);
    });

    testWidgets('renders link to login page', (tester) async {
      when(() => mockAuthBloc.state).thenReturn(const AuthInitial());
      when(() => mockAuthBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Already have an account?'), findsOneWidget);
      expect(find.text('Sign In'), findsOneWidget);
    });

    testWidgets('shows loading indicator when AuthLoading', (tester) async {
      when(() => mockAuthBloc.state).thenReturn(const AuthLoading());
      when(() => mockAuthBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget());

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });

    testWidgets('shows error message when AuthError', (tester) async {
      when(() => mockAuthBloc.state).thenReturn(const AuthError('Email already in use'));
      when(() => mockAuthBloc.stream).thenAnswer((_) => const Stream.empty());

      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Email already in use'), findsOneWidget);
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

      // When loading, create account button shows CircularProgressIndicator
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
