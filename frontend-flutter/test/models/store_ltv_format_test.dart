import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/models/store_model.dart';
import 'package:ledgerguard_flutter/widgets/lg_risk_badge.dart';

Store _store(int ltvCents) => Store(
      id: 's',
      shopDomain: 'x.myshopify.com',
      installedAppIds: const [],
      healthScore: 50,
      lifetimeValueCents: ltvCents,
      firstInstallDate: DateTime(2025, 1, 1),
      lastInteraction: DateTime(2025, 1, 1),
      riskState: RiskState.safe,
    );

void main() {
  group('Store.ltvFormatted', () {
    test('formats positive value', () {
      expect(_store(125900).ltvFormatted, r'$1259.00');
    });

    test('formats zero', () {
      expect(_store(0).ltvFormatted, r'$0.00');
    });

    // RISK-3: net LTV can be negative (refunds > revenue); the sign must precede
    // the currency symbol, not sit between it and the digits.
    test('formats negative value as -\$X, not \$-X', () {
      expect(_store(-2800).ltvFormatted, r'-$28.00');
    });
  });
}
