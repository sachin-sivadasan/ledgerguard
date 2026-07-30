import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/models/earning_model.dart';

// Regression guard for the "Earnings page shows all $0.00" bug: the models must
// decode the LIVE API field names, not the legacy/demo short names.
void main() {
  group('EarningsStatus.fromJson', () {
    test('maps live total_* keys and upcoming_availability', () {
      final s = EarningsStatus.fromJson({
        'total_pending_cents': 1615352,
        'total_available_cents': 116257226,
        'total_paid_out_cents': 0,
        'upcoming_availability': [
          {'date': '2026-08-01', 'amount_cents': 4104},
        ],
      });
      expect(s.pendingCents, 1615352);
      expect(s.availableCents, 116257226);
      expect(s.paidOutCents, 0);
      expect(s.upcoming.single.amountCents, 4104);
    });

    test('falls back to legacy short keys (demo/back-compat)', () {
      final s = EarningsStatus.fromJson({
        'pending_cents': 100,
        'available_cents': 200,
        'paid_out_cents': 300,
        'upcoming': [
          {'date': '2026-08-01', 'amount_cents': 50},
        ],
      });
      expect(s.pendingCents, 100);
      expect(s.availableCents, 200);
      expect(s.paidOutCents, 300);
      expect(s.upcoming.single.amountCents, 50);
    });
  });

  group('EarningPeriod.fromJson', () {
    test('live per-date row maps total_amount_cents to net, no breakdown', () {
      final p = EarningPeriod.fromJson({
        'date': '2026-07-30',
        'total_amount_cents': 138422,
      });
      expect(p.netEarningsCents, 138422);
      expect(p.grossCents, 0);
      expect(p.shopifyCutCents, 0);
      expect(p.hasFeeBreakdown, isFalse);
      expect(p.startDate, DateTime.parse('2026-07-30'));
    });

    test('rich row keeps gross/shopify and reports a breakdown', () {
      final p = EarningPeriod.fromJson({
        'id': 'jul',
        'month': 'July 2026',
        'start_date': '2026-07-01',
        'end_date': '2026-07-31',
        'gross_cents': 10000,
        'shopify_cut_cents': 2000,
        'net_earnings_cents': 8000,
        'status': 'AVAILABLE',
      });
      expect(p.grossCents, 10000);
      expect(p.shopifyCutCents, 2000);
      expect(p.netEarningsCents, 8000);
      expect(p.hasFeeBreakdown, isTrue);
      expect(p.status, EarningStatus.available);
    });
  });
}
