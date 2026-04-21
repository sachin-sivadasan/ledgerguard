import 'dart:math';
import '../models/transaction_model.dart';

final _now = DateTime.now();
final _rng = Random(42);

final mockTransactions = _generateTransactions();

List<Transaction> _generateTransactions() {
  final txns = <Transaction>[];
  final stores = [
    'acme-store', 'bright-gadgets', 'cool-threads', 'daily-deals',
    'eco-shop', 'fresh-foods', 'glow-beauty', 'home-essentials',
    'iconic-wear', 'just-pets', 'alpha-outlet', 'beta-mart',
    'craft-corner', 'ink-press', 'maple-candles',
  ];
  final apps = ['app-1', 'app-2', 'app-3'];
  int id = 1;

  // Generate ~120 transactions over 3 months
  for (int dayOffset = 90; dayOffset >= 0; dayOffset -= 1) {
    // ~1.3 txns per day
    if (_rng.nextDouble() > 0.7) continue;

    final date = _now.subtract(Duration(days: dayOffset));
    final numTxns = 1 + _rng.nextInt(3);

    for (int t = 0; t < numTxns && txns.length < 120; t++) {
      final store = stores[_rng.nextInt(stores.length)];
      final app = apps[_rng.nextInt(apps.length)];

      // 80% recurring, 15% usage, 5% one-time (with occasional refund)
      final roll = _rng.nextDouble();
      ChargeType type;
      int grossCents;

      if (roll < 0.80) {
        type = ChargeType.recurring;
        grossCents = [1999, 4999, 9999][_rng.nextInt(3)];
      } else if (roll < 0.95) {
        type = ChargeType.usage;
        grossCents = 200 + _rng.nextInt(1800); // $2-$20
      } else if (roll < 0.98) {
        type = ChargeType.oneTime;
        grossCents = 2999 + _rng.nextInt(7000); // $30-$100
      } else {
        type = ChargeType.refund;
        grossCents = -(1999 + _rng.nextInt(3000));
      }

      final netCents = (grossCents * 0.80).round(); // After Shopify 20% cut

      txns.add(Transaction(
        id: 'txn-$id',
        date: date,
        shopDomain: '$store.myshopify.com',
        chargeType: type,
        appId: app,
        grossAmountCents: grossCents,
        netAmountCents: netCents,
      ));
      id++;
    }
  }

  return txns;
}
