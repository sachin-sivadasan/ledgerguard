import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/services/fee_audit_service.dart';

void main() {
  group('FeeAuditReport.fromJson', () {
    test('parses verdict, tier signal, and rows', () {
      final r = FeeAuditReport.fromJson({
        'currency': 'USD',
        'configured_tier': 'SMALL_DEV_0',
        'configured_fee_pct': 0,
        'detected_fee_pct': 15,
        'tier_matches': false,
        'total_gross_cents': 1000000,
        'total_cut_cents': 150000,
        'effective_fee_pct': 15,
        'flagged_months': 1,
        'months_audited': 3,
        'savings_vs_default_cents': 50000,
        'months': [
          {'month': 'Jun', 'gross_cents': 100, 'shopify_cut_cents': 10, 'effective_fee_pct': 10, 'expected_cut_cents': 15, 'fee_variance_cents': -5, 'fee_guard_ok': false},
        ],
      });
      expect(r.detectedFeePct, 15);
      expect(r.tierMatches, isFalse);
      expect(r.flaggedMonths, 1);
      expect(r.allClear, isFalse);
      expect(r.savingsVsDefaultCents, 50000);
      expect(r.months.single.feeGuardOk, isFalse);
    });

    test('allClear true when no flagged months', () {
      final r = FeeAuditReport.fromJson({
        'months_audited': 4,
        'flagged_months': 0,
        'months': [],
      });
      expect(r.allClear, isTrue);
      expect(r.tierMatches, isTrue); // defaults safe
    });
  });
}
