import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/models/store_model.dart';
import 'package:ledgerguard_flutter/providers/store_provider.dart';
import 'package:ledgerguard_flutter/widgets/lg_risk_badge.dart';

// SD-1: store detail resolves a store by domain from a server `search` result set
// (a substring match), preferring the exact-domain hit.
void main() {
  Store store(String domain) => Store(
        id: domain,
        shopDomain: domain,
        installedAppIds: const [],
        healthScore: 90,
        lifetimeValueCents: 0,
        firstInstallDate: DateTime(2026, 1, 1),
        lastInteraction: DateTime(2026, 1, 1),
        riskState: RiskState.safe,
        timeline: const [],
      );

  test('prefers the exact-domain match over other substring hits', () {
    final results = [store('shoply.myshopify.com'), store('shop.myshopify.com')];
    final m = StoreProvider.matchStoreByDomain(results, 'shop.myshopify.com');
    expect(m?.shopDomain, 'shop.myshopify.com');
  });

  test('falls back to the first hit when no exact match', () {
    final results = [store('a-shop.myshopify.com'), store('b-shop.myshopify.com')];
    final m = StoreProvider.matchStoreByDomain(results, 'shop');
    expect(m?.shopDomain, 'a-shop.myshopify.com');
  });

  test('empty results → null (store not found)', () {
    expect(StoreProvider.matchStoreByDomain(const [], 'x.myshopify.com'), isNull);
  });
}
