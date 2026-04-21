import '../models/subscription_model.dart';
import '../widgets/lg_risk_badge.dart';
import '../widgets/lg_status_badge.dart';

final _now = DateTime.now();

final mockSubscriptions = <Subscription>[
  // === SAFE subscriptions (25) ===
  ..._safeSubscriptions(),
  // === ONE_CYCLE_MISSED (8) ===
  ..._oneCycleSubscriptions(),
  // === TWO_CYCLE_MISSED (4) ===
  ..._twoCycleSubscriptions(),
  // === CHURNED (3) ===
  ..._churnedSubscriptions(),
];

List<Subscription> _safeSubscriptions() {
  final stores = [
    'acme-store', 'bright-gadgets', 'cool-threads', 'daily-deals',
    'eco-shop', 'fresh-foods', 'glow-beauty', 'home-essentials',
    'iconic-wear', 'just-pets', 'keen-outdoors', 'luxe-living',
    'metro-style', 'nova-tech', 'olive-garden-supply',
    'prime-picks', 'quick-fix', 'retro-vibes', 'stellar-goods',
    'trendy-tots', 'urban-craft', 'vista-home', 'wave-sports',
    'xpress-gifts', 'zen-wellness',
  ];
  final plans = ['Basic', 'Pro', 'Enterprise'];
  final prices = [1999, 4999, 9999];
  final apps = ['app-1', 'app-2', 'app-3'];

  return List.generate(25, (i) {
    final planIdx = i % 3;
    return Subscription(
      id: 'sub-safe-${i + 1}',
      shopDomain: '${stores[i]}.myshopify.com',
      appId: apps[i % 3],
      planName: plans[planIdx],
      priceCents: prices[planIdx],
      status: SubscriptionStatus.active,
      riskState: RiskState.safe,
      billingInterval: i % 5 == 0 ? BillingInterval.annual : BillingInterval.monthly,
      periodEnd: _now.add(Duration(days: 10 + i)),
      expectedNextCharge: _now.add(Duration(days: 10 + i)),
      createdAt: _now.subtract(Duration(days: 90 + i * 10)),
    );
  });
}

List<Subscription> _oneCycleSubscriptions() {
  final stores = [
    'alpha-outlet', 'beta-mart', 'craft-corner', 'dusk-decor',
    'ever-bloom', 'flair-fashion', 'gem-jewelry', 'hive-honey',
  ];
  return List.generate(8, (i) {
    return Subscription(
      id: 'sub-one-${i + 1}',
      shopDomain: '${stores[i]}.myshopify.com',
      appId: i < 3 ? 'app-1' : (i < 6 ? 'app-2' : 'app-3'),
      planName: i % 2 == 0 ? 'Pro' : 'Basic',
      priceCents: i % 2 == 0 ? 4999 : 1999,
      status: SubscriptionStatus.active,
      riskState: RiskState.oneCycleMissed,
      billingInterval: BillingInterval.monthly,
      periodEnd: _now.subtract(Duration(days: 35 + i * 3)),
      expectedNextCharge: _now.subtract(Duration(days: 35 + i * 3)),
      createdAt: _now.subtract(Duration(days: 200 + i * 15)),
    );
  });
}

List<Subscription> _twoCycleSubscriptions() {
  final stores = ['ink-press', 'jade-spa', 'kite-kids', 'leaf-organic'];
  return List.generate(4, (i) {
    return Subscription(
      id: 'sub-two-${i + 1}',
      shopDomain: '${stores[i]}.myshopify.com',
      appId: i < 2 ? 'app-1' : 'app-3',
      planName: 'Pro',
      priceCents: 4999,
      status: SubscriptionStatus.frozen,
      riskState: RiskState.twoCycleMissed,
      billingInterval: BillingInterval.monthly,
      periodEnd: _now.subtract(Duration(days: 65 + i * 5)),
      expectedNextCharge: _now.subtract(Duration(days: 65 + i * 5)),
      createdAt: _now.subtract(Duration(days: 300 + i * 20)),
    );
  });
}

List<Subscription> _churnedSubscriptions() {
  final stores = ['maple-candles', 'neon-arcade', 'opal-rings'];
  return List.generate(3, (i) {
    return Subscription(
      id: 'sub-churn-${i + 1}',
      shopDomain: '${stores[i]}.myshopify.com',
      appId: ['app-1', 'app-2', 'app-3'][i],
      planName: 'Basic',
      priceCents: 1999,
      status: SubscriptionStatus.cancelled,
      riskState: RiskState.churned,
      billingInterval: BillingInterval.monthly,
      periodEnd: _now.subtract(Duration(days: 100 + i * 10)),
      expectedNextCharge: _now.subtract(Duration(days: 100 + i * 10)),
      createdAt: _now.subtract(Duration(days: 400 + i * 30)),
    );
  });
}
