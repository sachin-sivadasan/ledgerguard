import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/services/plan_label_service.dart';

void main() {
  group('PlanTier.fromJson', () {
    test('parses a tier with a saved label', () {
      final t = PlanTier.fromJson({
        'billingInterval': 'MONTHLY',
        'priceCents': 2900,
        'key': 'MONTHLY:2900',
        'pseudoLabel': '\$29.00/mo',
        'label': 'Starter',
        'customers': 42,
      });
      expect(t.priceCents, 2900);
      expect(t.key, 'MONTHLY:2900');
      expect(t.pseudoLabel, '\$29.00/mo');
      expect(t.label, 'Starter');
      expect(t.customers, 42);
    });

    test('defaults an unlabeled tier to an empty label', () {
      final t = PlanTier.fromJson({
        'billingInterval': 'ANNUAL',
        'priceCents': 140000,
        'key': 'ANNUAL:140000',
        'pseudoLabel': '\$1400.00/yr',
        'customers': 3,
      });
      expect(t.label, '');
      expect(t.customers, 3);
    });
  });
}
