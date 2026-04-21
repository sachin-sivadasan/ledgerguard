import '../widgets/lg_risk_badge.dart';

class RiskTimelineEntry {
  final String id;
  final RiskState fromRiskState;
  final RiskState toRiskState;
  final String fromStatus;
  final String toStatus;
  final String eventType;
  final String reason;
  final DateTime occurredAt;

  const RiskTimelineEntry({
    required this.id,
    required this.fromRiskState,
    required this.toRiskState,
    required this.fromStatus,
    required this.toStatus,
    required this.eventType,
    required this.reason,
    required this.occurredAt,
  });

  String get eventTypeLabel => switch (eventType) {
        'sync' => 'Sync',
        'webhook' => 'Webhook',
        'billing_failure' => 'Billing Failure',
        'app_uninstalled' => 'App Uninstalled',
        'manual' => 'Manual',
        _ => eventType,
      };

  bool get isEscalation =>
      toRiskState.index > fromRiskState.index;

  bool get isRecovery =>
      toRiskState.index < fromRiskState.index;
}
