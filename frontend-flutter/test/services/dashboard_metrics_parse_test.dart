import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/services/metrics_service.dart';

void main() {
  group('DashboardMetrics.fromJson', () {
    test('parses current values and period-over-period deltas', () {
      final m = DashboardMetrics.fromJson({
        'current': {
          'active_mrr_cents': 5847989,
          'revenue_at_risk_cents': 97018,
          'usage_revenue_cents': 721406,
          'renewal_success_rate': 0.3675,
          'safe_count': 1079,
          'churned_count': 1833,
        },
        'delta': {
          'active_mrr_percent': -7.38,
          'renewal_success_rate_percent': -8.51,
          'usage_revenue_percent': 196.87,
          'revenue_at_risk_percent': -84.52,
        },
      });
      expect(m.mrrCents, 5847989);
      expect(m.renewalRate, closeTo(36.75, 0.01)); // 0.3675 → 36.75
      expect(m.riskDistribution.churned, 1833);
      expect(m.mrrDeltaPct, -7.38);
      expect(m.renewalDeltaPct, -8.51);
      expect(m.usageDeltaPct, 196.87);
      expect(m.riskDeltaPct, -84.52);
    });

    test('deltas are null when the backend omits the delta block', () {
      final m = DashboardMetrics.fromJson({
        'current': {'active_mrr_cents': 1000},
      });
      expect(m.mrrDeltaPct, isNull);
      expect(m.riskDeltaPct, isNull);
    });
  });
}
