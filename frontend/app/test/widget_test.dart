import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/presentation/blocs/auth/auth.dart';
import 'package:ledgerguard/presentation/pages/login_page.dart';

class MockAuthBloc extends Mock implements AuthBloc {}

void main() {
  testWidgets('LoginPage renders without error', (WidgetTester tester) async {
    final mockAuthBloc = MockAuthBloc();
    when(() => mockAuthBloc.state).thenReturn(const AuthInitial());
    when(() => mockAuthBloc.stream).thenAnswer((_) => const Stream.empty());

    await tester.pumpWidget(
      MaterialApp(
        home: BlocProvider<AuthBloc>.value(
          value: mockAuthBloc,
          child: const LoginPage(),
        ),
      ),
    );

    expect(find.byType(MaterialApp), findsOneWidget);
    expect(find.text('LedgerGuard'), findsOneWidget);
  });
}
