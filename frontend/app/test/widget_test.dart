import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:ledgerguard/app.dart';

void main() {
  testWidgets('App renders without error', (WidgetTester tester) async {
    await tester.pumpWidget(const LedgerGuardApp());

    // App should render with placeholder page
    expect(find.byType(MaterialApp), findsOneWidget);
  });
}
