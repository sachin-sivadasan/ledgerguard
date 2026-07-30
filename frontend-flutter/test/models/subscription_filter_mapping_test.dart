import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/models/subscription_model.dart';
import 'package:ledgerguard_flutter/widgets/lg_risk_badge.dart';
import 'package:ledgerguard_flutter/widgets/lg_status_badge.dart';

// Guards the enum→API string contract used to send server-side filter params. These MUST
// match what the backend accepts (parseRiskState / subscription_status filter), else the
// filter is silently dropped and the user sees an unfiltered list.
void main() {
  group('riskStateToApi', () {
    test('maps every RiskState to the backend value', () {
      expect(Subscription.riskStateToApi(RiskState.safe), 'SAFE');
      expect(Subscription.riskStateToApi(RiskState.oneCycleMissed),
          'ONE_CYCLE_MISSED');
      // The plural "CYCLES" is the easy-to-break one — backend expects TWO_CYCLES_MISSED.
      expect(Subscription.riskStateToApi(RiskState.twoCycleMissed),
          'TWO_CYCLES_MISSED');
      expect(Subscription.riskStateToApi(RiskState.churned), 'CHURNED');
    });

    test('round-trips through parseRiskState', () {
      for (final r in RiskState.values) {
        expect(Subscription.parseRiskState(Subscription.riskStateToApi(r)), r);
      }
    });
  });

  group('statusToApi', () {
    test('maps every SubscriptionStatus to the backend value', () {
      expect(Subscription.statusToApi(SubscriptionStatus.active), 'ACTIVE');
      expect(Subscription.statusToApi(SubscriptionStatus.frozen), 'FROZEN');
      expect(Subscription.statusToApi(SubscriptionStatus.cancelled), 'CANCELLED');
      expect(Subscription.statusToApi(SubscriptionStatus.pending), 'PENDING');
      expect(Subscription.statusToApi(SubscriptionStatus.uninstalled),
          'UNINSTALLED');
    });
  });
}
