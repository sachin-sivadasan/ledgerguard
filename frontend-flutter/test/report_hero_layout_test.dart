import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/widgets/lg_table.dart';

// Regression guard: a hero Row with crossAxisAlignment.stretch + Expanded cards,
// followed by more content, inside an unbounded-height scroll view MUST NOT throw
// or blank the siblings below it. stretch needs a bounded cross-axis, so the hero
// must be wrapped in IntrinsicHeight (see _HeroRow in the report screens).
//
// This mirrors LgPage's layout chain (Center > ConstrainedBox > Padding > Column >
// Expanded > SingleChildScrollView > Column) by hand — it does NOT pump LgPage/the
// real screens, so it guards the *pattern*, not any specific screen. Removing
// IntrinsicHeight from a real _HeroRow would still regress prod without failing here.
void main() {
  Widget heroRow() => IntrinsicHeight(
        child: Row(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          for (var i = 0; i < 4; i++) ...[
            if (i > 0) const SizedBox(width: 12),
            Expanded(
              child: Container(
                decoration: BoxDecoration(border: Border.all(color: Colors.grey)),
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [Text('Label $i'), Text('\$99,999.99'), Text('footnote')],
                ),
              ),
            ),
          ],
        ],
      ),
      );

  Widget chargesTable() {
    final rows = List.generate(60, (i) => <Widget>[
          Text('Jul ${i % 28 + 1}'),
          Text('Store $i'),
          Text('\$49.99'),
          Text('\$41.04'),
          const Text('Pending'),
          Text('Aug 4'),
        ]);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('Charges'),
        LgTable(
          columns: const [
            LgTableColumn('DATE', flex: 2),
            LgTableColumn('STORE', flex: 4),
            LgTableColumn('GROSS', flex: 2, numeric: true),
            LgTableColumn('NET', flex: 2, numeric: true),
            LgTableColumn('STATUS', flex: 2),
            LgTableColumn('AVAILABLE DATE', flex: 3, numeric: true),
          ],
          rows: rows,
        ),
      ],
    );
  }

  testWidgets('hero Row(stretch) + charges table inside LgPage layout', (tester) async {
    await tester.pumpWidget(MaterialApp(
      home: Scaffold(
        body: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 1200),
            child: Padding(
              padding: const EdgeInsets.all(24),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('Header'),
                  const SizedBox(height: 24),
                  Expanded(
                    child: SingleChildScrollView(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          heroRow(),
                          const SizedBox(height: 24),
                          chargesTable(),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    ));

    expect(tester.takeException(), isNull, reason: 'hero+table must lay out with no exception');
    expect(find.text('Charges'), findsOneWidget, reason: 'charges heading missing');
    expect(find.text('DATE'), findsOneWidget);
    expect(find.text('Store 0'), findsOneWidget);
  });
}
