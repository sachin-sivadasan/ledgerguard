import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/models/analytics_model.dart';

void main() {
  group('ExpenseBreakdown.fromJson — Fee Guard (RPT-FEES-1)', () {
    test('parses real shopify cut + guard fields', () {
      final e = ExpenseBreakdown.fromJson({
        'month': 'Mar',
        'gross_cents': 100000,
        'shopify_cut_cents': 15000,
        'expected_cut_cents': 15000,
        'fee_variance_cents': 0,
        'fee_guard_ok': true,
        'effective_fee_pct': 15.0,
      });
      expect(e.shopifyCutCents, 15000);
      expect(e.expectedCutCents, 15000);
      expect(e.feeGuardOk, isTrue);
      expect(e.effectiveFeePct, 15.0);
    });

    test('flags an overcharge variance', () {
      final e = ExpenseBreakdown.fromJson({
        'month': 'Apr',
        'gross_cents': 100000,
        'shopify_cut_cents': 20000,
        'expected_cut_cents': 15000,
        'fee_variance_cents': 5000,
        'fee_guard_ok': false,
      });
      expect(e.feeVarianceCents, 5000);
      expect(e.feeGuardOk, isFalse);
    });

    // Pre-Guard/empty payloads must not raise a false alarm.
    test('defaults feeGuardOk to true when absent', () {
      final e = ExpenseBreakdown.fromJson({'month': 'x', 'gross_cents': 0});
      expect(e.feeGuardOk, isTrue);
      expect(e.expectedCutCents, 0);
    });
  });
}
