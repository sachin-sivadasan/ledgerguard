import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/services/ledger_recon_service.dart';

void main() {
  group('ReconReport.fromJson', () {
    test('parses totals, verdict and rows', () {
      final r = ReconReport.fromJson({
        'currency': 'USD',
        'total_gross_cents': 1000,
        'total_fee_cents': 150,
        'total_net_cents': 850,
        'residual_cents': 0,
        'reconciled': false,
        'months_reconciled': 5,
        'months_flagged': 1,
        'months_audited': 6,
        'months': [
          {'month': 'Jun', 'gross_cents': 100, 'fee_cents': 0, 'net_cents': 85, 'expected_net_cents': 100, 'residual_cents': -15, 'tx_count': 3, 'reconciled': false},
        ],
      });
      expect(r.reconciled, isFalse);
      expect(r.monthsFlagged, 1);
      expect(r.totalNetCents, 850);
      expect(r.months.single.residualCents, -15);
      expect(r.months.single.reconciled, isFalse);
    });

    test('defaults reconciled true when absent', () {
      final r = ReconReport.fromJson({'months': []});
      expect(r.reconciled, isTrue);
      expect(r.monthsAudited, 0);
    });
  });
}
