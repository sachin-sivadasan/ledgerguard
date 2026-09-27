import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/services/push_deep_link.dart';

void main() {
  test('risk_alert with subscription_id -> subscription detail route', () {
    expect(routeForData({'type': 'risk_alert', 'subscription_id': 'sub-9'}),
        '/subscriptions/sub-9');
  });

  test('daily_summary -> dashboard', () {
    expect(routeForData({'type': 'daily_summary', 'app_id': 'a1'}), '/');
  });

  test('missing/unknown/malformed data -> dashboard fallback (Review Focus #2/#3)', () {
    expect(routeForData({}), '/');
    expect(routeForData({'type': 'nonsense'}), '/');
    expect(routeForData({'type': 'risk_alert'}), '/'); // no subscription_id
    expect(routeForData({'type': 'risk_alert', 'subscription_id': ''}), '/'); // empty id
    expect(routeForData({'type': 'risk_alert', 'subscription_id': 42}), '/'); // non-string id
  });
}
