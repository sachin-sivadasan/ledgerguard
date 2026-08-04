import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/services/customer_insights_service.dart';

void main() {
  group('CustomerInsights.fromJson', () {
    test('parses KPIs and all four segment lists', () {
      final r = CustomerInsights.fromJson({
        'currency': 'USD',
        'totalCustomers': 4,
        'activeMrrCents': 16000,
        'atRiskCustomers': 1,
        'atRiskMrrCents': 5000,
        'revenueBands': [
          {'label': '< \$25', 'customers': 1, 'mrrCents': 1000, 'pctOfCustomers': 0.25},
          {'label': '\$50–\$100', 'customers': 3, 'mrrCents': 15000, 'pctOfCustomers': 0.75},
        ],
        'riskSegments': [
          {'riskState': 'SAFE', 'customers': 3, 'mrrCents': 11000},
          {'riskState': 'AT_RISK', 'customers': 1, 'mrrCents': 5000},
          {'riskState': 'CHURNED', 'customers': 1, 'mrrCents': 1000},
        ],
        'planRisk': [
          {'planName': 'Pro', 'customers': 3, 'safeCount': 2, 'atRiskCount': 1, 'mrrCents': 15000, 'atRiskMrrCents': 5000},
        ],
        'topCustomers': [
          {'shopName': 'a', 'planName': 'Pro', 'mrrCents': 5000, 'riskState': 'SAFE'},
        ],
      });

      expect(r.totalCustomers, 4);
      expect(r.activeMrrCents, 16000);
      expect(r.atRiskMrrCents, 5000);
      expect(r.revenueBands.length, 2);
      expect(r.revenueBands[1].pctOfCustomers, 0.75);
      expect(r.riskSegments.firstWhere((s) => s.riskState == 'CHURNED').customers, 1);
      final pro = r.planRisk.single;
      expect(pro.atRiskCount, 1);
      expect(pro.atRiskMrrCents, 5000);
      expect(r.topCustomers.single.shopName, 'a');
    });

    test('empty payload yields safe defaults', () {
      final r = CustomerInsights.fromJson({});
      expect(r.totalCustomers, 0);
      expect(r.revenueBands, isEmpty);
      expect(r.planRisk, isEmpty);
      expect(r.currency, 'USD');
    });
  });
}
