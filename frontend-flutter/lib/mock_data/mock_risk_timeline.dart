import '../models/risk_timeline_model.dart';
import '../widgets/lg_risk_badge.dart';

final _now = DateTime.now();

/// Generates mock risk timeline for a subscription.
/// Safe subscriptions get minimal history, at-risk/churned get richer timelines.
List<RiskTimelineEntry> generateRiskTimeline(String subscriptionId) {
  final entries = <RiskTimelineEntry>[];
  int id = 1;

  if (subscriptionId.contains('churn')) {
    // Churned: full escalation path
    entries.addAll([
      RiskTimelineEntry(
        id: '$subscriptionId-rt-${id++}',
        fromRiskState: RiskState.twoCycleMissed,
        toRiskState: RiskState.churned,
        fromStatus: 'FROZEN',
        toStatus: 'CANCELLED',
        eventType: 'sync',
        reason: 'No payment received for 90+ days',
        occurredAt: _now.subtract(const Duration(days: 5)),
      ),
      RiskTimelineEntry(
        id: '$subscriptionId-rt-${id++}',
        fromRiskState: RiskState.oneCycleMissed,
        toRiskState: RiskState.twoCycleMissed,
        fromStatus: 'ACTIVE',
        toStatus: 'FROZEN',
        eventType: 'sync',
        reason: 'No payment received for 60+ days',
        occurredAt: _now.subtract(const Duration(days: 35)),
      ),
      RiskTimelineEntry(
        id: '$subscriptionId-rt-${id++}',
        fromRiskState: RiskState.safe,
        toRiskState: RiskState.oneCycleMissed,
        fromStatus: 'ACTIVE',
        toStatus: 'ACTIVE',
        eventType: 'billing_failure',
        reason: 'Recurring charge failed — card declined',
        occurredAt: _now.subtract(const Duration(days: 65)),
      ),
    ]);
  } else if (subscriptionId.contains('two')) {
    // Two cycles missed: escalation in progress
    entries.addAll([
      RiskTimelineEntry(
        id: '$subscriptionId-rt-${id++}',
        fromRiskState: RiskState.oneCycleMissed,
        toRiskState: RiskState.twoCycleMissed,
        fromStatus: 'ACTIVE',
        toStatus: 'FROZEN',
        eventType: 'sync',
        reason: 'No payment received for 60+ days',
        occurredAt: _now.subtract(const Duration(days: 10)),
      ),
      RiskTimelineEntry(
        id: '$subscriptionId-rt-${id++}',
        fromRiskState: RiskState.safe,
        toRiskState: RiskState.oneCycleMissed,
        fromStatus: 'ACTIVE',
        toStatus: 'ACTIVE',
        eventType: 'billing_failure',
        reason: 'Recurring charge failed — insufficient funds',
        occurredAt: _now.subtract(const Duration(days: 40)),
      ),
    ]);
  } else if (subscriptionId.contains('one')) {
    // One cycle missed
    final seed = subscriptionId.hashCode.abs();
    entries.add(RiskTimelineEntry(
      id: '$subscriptionId-rt-${id++}',
      fromRiskState: RiskState.safe,
      toRiskState: RiskState.oneCycleMissed,
      fromStatus: 'ACTIVE',
      toStatus: 'ACTIVE',
      eventType: seed % 2 == 0 ? 'billing_failure' : 'sync',
      reason: seed % 2 == 0
          ? 'Recurring charge failed — expired card'
          : 'No payment received for 30+ days',
      occurredAt: _now.subtract(Duration(days: 15 + seed % 10)),
    ));
  }
  // Safe subscriptions: empty timeline (no risk changes)

  entries.sort((a, b) => b.occurredAt.compareTo(a.occurredAt));
  return entries;
}
