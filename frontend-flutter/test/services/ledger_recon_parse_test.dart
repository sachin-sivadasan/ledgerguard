import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/services/ledger_recon_service.dart';

void main() {
  group('ReconReport.fromJson', () {
    test('parses totals, decomposition and rows', () {
      final r = ReconReport.fromJson({
        'currency': 'USD',
        'total_gross_cents': 1000,
        'total_net_cents': 820,
        'total_revenue_share_cents': 150,
        'total_processing_cents': 30,
        'residual_cents': 0,
        'reconciled': false,
        'months_reconciled': 5,
        'months_flagged': 1,
        'months_audited': 6,
        'months': [
          {'month': 'Jun', 'gross_cents': 100, 'net_cents': 82, 'revenue_share_cents': 15, 'processing_cents': 0, 'accounted_cents': 97, 'processing_pct': 0.0, 'processing_suspect': false, 'residual_cents': 3, 'tx_count': 3, 'reconciled': false},
        ],
      });
      expect(r.reconciled, isFalse);
      expect(r.monthsFlagged, 1);
      expect(r.totalNetCents, 820);
      expect(r.totalRevenueShareCents, 150);
      expect(r.totalProcessingCents, 30);
      final m = r.months.single;
      expect(m.processingCents, 0);
      expect(m.accountedCents, 97);
      expect(m.residualCents, 3);
      expect(m.reconciled, isFalse);
    });

    test('parses processing_suspect anomaly flag', () {
      final r = ReconReport.fromJson({
        'reconciled': false,
        'months': [
          {'month': 'Jul', 'gross_cents': 10000, 'net_cents': 8200, 'revenue_share_cents': 0, 'processing_cents': 1800, 'accounted_cents': 10000, 'processing_pct': 18.0, 'processing_suspect': true, 'residual_cents': 0, 'tx_count': 5, 'reconciled': false},
        ],
      });
      final m = r.months.single;
      expect(m.processingSuspect, isTrue);
      expect(m.processingPct, 18.0);
      expect(m.residualCents, 0); // buckets close, yet flagged
      expect(m.reconciled, isFalse);
    });

    test('defaults reconciled true when absent', () {
      final r = ReconReport.fromJson({'months': []});
      expect(r.reconciled, isTrue);
      expect(r.monthsAudited, 0);
    });
  });
}
