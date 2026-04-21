import '../models/store_model.dart';
import '../models/timeline_event.dart';
import '../widgets/lg_risk_badge.dart';

final _now = DateTime.now();

final mockStores = <Store>[
  Store(
    id: 'store-1', shopDomain: 'acme-store.myshopify.com',
    installedAppIds: ['app-1', 'app-2'], healthScore: 95,
    lifetimeValueCents: 125900, firstInstallDate: _now.subtract(const Duration(days: 365)),
    lastInteraction: _now.subtract(const Duration(days: 2)), riskState: RiskState.safe,
    timeline: [
      TimelineEvent(date: _now.subtract(const Duration(days: 365)), type: TimelineEventType.install, description: 'Installed InventorySync Pro'),
      TimelineEvent(date: _now.subtract(const Duration(days: 300)), type: TimelineEventType.install, description: 'Installed ReviewBoost'),
      TimelineEvent(date: _now.subtract(const Duration(days: 30)), type: TimelineEventType.transaction, description: 'Recurring charge \$49.99'),
      TimelineEvent(date: _now.subtract(const Duration(days: 2)), type: TimelineEventType.transaction, description: 'Recurring charge \$19.99'),
    ],
  ),
  Store(
    id: 'store-2', shopDomain: 'bright-gadgets.myshopify.com',
    installedAppIds: ['app-1', 'app-3'], healthScore: 88,
    lifetimeValueCents: 98700, firstInstallDate: _now.subtract(const Duration(days: 280)),
    lastInteraction: _now.subtract(const Duration(days: 5)), riskState: RiskState.safe,
    timeline: [
      TimelineEvent(date: _now.subtract(const Duration(days: 280)), type: TimelineEventType.install, description: 'Installed InventorySync Pro'),
      TimelineEvent(date: _now.subtract(const Duration(days: 200)), type: TimelineEventType.install, description: 'Installed ShipTracker'),
      TimelineEvent(date: _now.subtract(const Duration(days: 5)), type: TimelineEventType.transaction, description: 'Recurring charge \$49.99'),
    ],
  ),
  Store(
    id: 'store-3', shopDomain: 'cool-threads.myshopify.com',
    installedAppIds: ['app-2'], healthScore: 92,
    lifetimeValueCents: 47940, firstInstallDate: _now.subtract(const Duration(days: 200)),
    lastInteraction: _now.subtract(const Duration(days: 1)), riskState: RiskState.safe,
    timeline: [
      TimelineEvent(date: _now.subtract(const Duration(days: 200)), type: TimelineEventType.install, description: 'Installed ReviewBoost'),
      TimelineEvent(date: _now.subtract(const Duration(days: 1)), type: TimelineEventType.transaction, description: 'Usage charge \$12.50'),
    ],
  ),
  Store(
    id: 'store-4', shopDomain: 'daily-deals.myshopify.com',
    installedAppIds: ['app-1'], healthScore: 78,
    lifetimeValueCents: 35960, firstInstallDate: _now.subtract(const Duration(days: 180)),
    lastInteraction: _now.subtract(const Duration(days: 10)), riskState: RiskState.safe,
    timeline: [
      TimelineEvent(date: _now.subtract(const Duration(days: 180)), type: TimelineEventType.install, description: 'Installed InventorySync Pro'),
      TimelineEvent(date: _now.subtract(const Duration(days: 10)), type: TimelineEventType.transaction, description: 'Recurring charge \$19.99'),
    ],
  ),
  Store(
    id: 'store-5', shopDomain: 'eco-shop.myshopify.com',
    installedAppIds: ['app-1', 'app-2', 'app-3'], healthScore: 97,
    lifetimeValueCents: 189500, firstInstallDate: _now.subtract(const Duration(days: 400)),
    lastInteraction: _now.subtract(const Duration(days: 1)), riskState: RiskState.safe,
    timeline: [
      TimelineEvent(date: _now.subtract(const Duration(days: 400)), type: TimelineEventType.install, description: 'Installed InventorySync Pro'),
      TimelineEvent(date: _now.subtract(const Duration(days: 350)), type: TimelineEventType.install, description: 'Installed ReviewBoost'),
      TimelineEvent(date: _now.subtract(const Duration(days: 250)), type: TimelineEventType.install, description: 'Installed ShipTracker'),
      TimelineEvent(date: _now.subtract(const Duration(days: 1)), type: TimelineEventType.transaction, description: 'Recurring charge \$99.99'),
    ],
  ),
  Store(
    id: 'store-6', shopDomain: 'alpha-outlet.myshopify.com',
    installedAppIds: ['app-1'], healthScore: 52,
    lifetimeValueCents: 23940, firstInstallDate: _now.subtract(const Duration(days: 250)),
    lastInteraction: _now.subtract(const Duration(days: 38)), riskState: RiskState.oneCycleMissed,
    timeline: [
      TimelineEvent(date: _now.subtract(const Duration(days: 250)), type: TimelineEventType.install, description: 'Installed InventorySync Pro'),
      TimelineEvent(date: _now.subtract(const Duration(days: 38)), type: TimelineEventType.riskChange, description: 'Risk changed: Safe → 1 Cycle Missed'),
    ],
  ),
  Store(
    id: 'store-7', shopDomain: 'beta-mart.myshopify.com',
    installedAppIds: ['app-1', 'app-2'], healthScore: 45,
    lifetimeValueCents: 41900, firstInstallDate: _now.subtract(const Duration(days: 300)),
    lastInteraction: _now.subtract(const Duration(days: 42)), riskState: RiskState.oneCycleMissed,
    timeline: [
      TimelineEvent(date: _now.subtract(const Duration(days: 300)), type: TimelineEventType.install, description: 'Installed InventorySync Pro'),
      TimelineEvent(date: _now.subtract(const Duration(days: 42)), type: TimelineEventType.riskChange, description: 'Risk changed: Safe → 1 Cycle Missed'),
      TimelineEvent(date: _now.subtract(const Duration(days: 40)), type: TimelineEventType.note, description: 'Sent recovery email'),
    ],
  ),
  Store(
    id: 'store-8', shopDomain: 'craft-corner.myshopify.com',
    installedAppIds: ['app-2'], healthScore: 48,
    lifetimeValueCents: 15960, firstInstallDate: _now.subtract(const Duration(days: 200)),
    lastInteraction: _now.subtract(const Duration(days: 45)), riskState: RiskState.oneCycleMissed,
    timeline: [
      TimelineEvent(date: _now.subtract(const Duration(days: 200)), type: TimelineEventType.install, description: 'Installed ReviewBoost'),
      TimelineEvent(date: _now.subtract(const Duration(days: 45)), type: TimelineEventType.riskChange, description: 'Risk changed: Safe → 1 Cycle Missed'),
    ],
  ),
  Store(
    id: 'store-9', shopDomain: 'dusk-decor.myshopify.com',
    installedAppIds: ['app-2'], healthScore: 40,
    lifetimeValueCents: 11970, firstInstallDate: _now.subtract(const Duration(days: 150)),
    lastInteraction: _now.subtract(const Duration(days: 50)), riskState: RiskState.oneCycleMissed,
    timeline: [
      TimelineEvent(date: _now.subtract(const Duration(days: 150)), type: TimelineEventType.install, description: 'Installed ReviewBoost'),
      TimelineEvent(date: _now.subtract(const Duration(days: 50)), type: TimelineEventType.riskChange, description: 'Risk changed: Safe → 1 Cycle Missed'),
    ],
  ),
  Store(
    id: 'store-10', shopDomain: 'ink-press.myshopify.com',
    installedAppIds: ['app-1'], healthScore: 22,
    lifetimeValueCents: 29940, firstInstallDate: _now.subtract(const Duration(days: 350)),
    lastInteraction: _now.subtract(const Duration(days: 70)), riskState: RiskState.twoCycleMissed,
    timeline: [
      TimelineEvent(date: _now.subtract(const Duration(days: 350)), type: TimelineEventType.install, description: 'Installed InventorySync Pro'),
      TimelineEvent(date: _now.subtract(const Duration(days: 70)), type: TimelineEventType.riskChange, description: 'Risk changed: 1 Cycle Missed → 2 Cycles Missed'),
    ],
  ),
  Store(
    id: 'store-11', shopDomain: 'jade-spa.myshopify.com',
    installedAppIds: ['app-1'], healthScore: 18,
    lifetimeValueCents: 24950, firstInstallDate: _now.subtract(const Duration(days: 320)),
    lastInteraction: _now.subtract(const Duration(days: 75)), riskState: RiskState.twoCycleMissed,
    timeline: [
      TimelineEvent(date: _now.subtract(const Duration(days: 320)), type: TimelineEventType.install, description: 'Installed InventorySync Pro'),
      TimelineEvent(date: _now.subtract(const Duration(days: 75)), type: TimelineEventType.riskChange, description: 'Risk changed: 1 Cycle Missed → 2 Cycles Missed'),
    ],
  ),
  Store(
    id: 'store-12', shopDomain: 'maple-candles.myshopify.com',
    installedAppIds: ['app-1'], healthScore: 5,
    lifetimeValueCents: 11970, firstInstallDate: _now.subtract(const Duration(days: 400)),
    lastInteraction: _now.subtract(const Duration(days: 110)), riskState: RiskState.churned,
    timeline: [
      TimelineEvent(date: _now.subtract(const Duration(days: 400)), type: TimelineEventType.install, description: 'Installed InventorySync Pro'),
      TimelineEvent(date: _now.subtract(const Duration(days: 110)), type: TimelineEventType.riskChange, description: 'Risk changed: 2 Cycles Missed → Churned'),
    ],
  ),
  Store(
    id: 'store-13', shopDomain: 'neon-arcade.myshopify.com',
    installedAppIds: ['app-2'], healthScore: 3,
    lifetimeValueCents: 7980, firstInstallDate: _now.subtract(const Duration(days: 350)),
    lastInteraction: _now.subtract(const Duration(days: 120)), riskState: RiskState.churned,
    timeline: [
      TimelineEvent(date: _now.subtract(const Duration(days: 350)), type: TimelineEventType.install, description: 'Installed ReviewBoost'),
      TimelineEvent(date: _now.subtract(const Duration(days: 120)), type: TimelineEventType.riskChange, description: 'Risk changed: 2 Cycles Missed → Churned'),
    ],
  ),
  Store(
    id: 'store-14', shopDomain: 'fresh-foods.myshopify.com',
    installedAppIds: ['app-3'], healthScore: 85,
    lifetimeValueCents: 59940, firstInstallDate: _now.subtract(const Duration(days: 240)),
    lastInteraction: _now.subtract(const Duration(days: 3)), riskState: RiskState.safe,
    timeline: [
      TimelineEvent(date: _now.subtract(const Duration(days: 240)), type: TimelineEventType.install, description: 'Installed ShipTracker'),
      TimelineEvent(date: _now.subtract(const Duration(days: 3)), type: TimelineEventType.transaction, description: 'Recurring charge \$49.99'),
    ],
  ),
  Store(
    id: 'store-15', shopDomain: 'glow-beauty.myshopify.com',
    installedAppIds: ['app-2', 'app-3'], healthScore: 90,
    lifetimeValueCents: 71920, firstInstallDate: _now.subtract(const Duration(days: 260)),
    lastInteraction: _now.subtract(const Duration(days: 4)), riskState: RiskState.safe,
    timeline: [
      TimelineEvent(date: _now.subtract(const Duration(days: 260)), type: TimelineEventType.install, description: 'Installed ReviewBoost'),
      TimelineEvent(date: _now.subtract(const Duration(days: 180)), type: TimelineEventType.install, description: 'Installed ShipTracker'),
      TimelineEvent(date: _now.subtract(const Duration(days: 4)), type: TimelineEventType.transaction, description: 'Recurring charge \$49.99'),
    ],
  ),
];
