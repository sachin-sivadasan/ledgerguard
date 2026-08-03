import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/services/activation_service.dart';

void main() {
  group('ActivationReport.fromJson — 2-stage funnel (RPT-ACTIVATION-1)', () {
    test('parses installs/paid/overallPct and 2 stages', () {
      final r = ActivationReport.fromJson({
        'installs': 10469,
        'paid': 2931,
        'overallPct': 0.28,
        'stages': [
          {'key': 'installs', 'label': 'Installs', 'count': 10469, 'pctOfPrior': 1.0},
          {'key': 'paid', 'label': 'Paid / Recurring', 'count': 2931, 'pctOfPrior': 0.28},
        ],
      });
      expect(r.installs, 10469);
      expect(r.paid, 2931);
      expect(r.overallPct, closeTo(0.28, 1e-9));
      expect(r.stages.length, 2);
      expect(r.stages[1].key, 'paid');
    });

    test('overallPct clamps to [0,1]', () {
      final r = ActivationReport.fromJson({'installs': 1, 'paid': 1, 'overallPct': 1.5, 'stages': []});
      expect(r.overallPct, 1.0);
    });

    // Missing fields degrade to zeroed defaults, not throw.
    test('defaults when fields missing', () {
      final r = ActivationReport.fromJson({});
      expect(r.installs, 0);
      expect(r.paid, 0);
      expect(r.overallPct, 0);
      expect(r.stages, isEmpty);
    });
  });
}
