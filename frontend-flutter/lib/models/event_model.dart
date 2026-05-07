enum EventType {
  appInstall,
  appUninstall,
  subscriptionActivated,
  subscriptionCancelled,
  planUpgrade,
  planDowngrade,
  billingFailure,
  billingSuccess,
  riskStateChange,
  reviewSubmitted,
  usageCharge,
}

class AppEvent {
  final String id;
  final DateTime date;
  final EventType type;
  final String appId;
  final String storeDomain;
  final String title;
  final String description;
  final Map<String, String>? metadata;

  const AppEvent({
    required this.id,
    required this.date,
    required this.type,
    required this.appId,
    required this.storeDomain,
    required this.title,
    required this.description,
    this.metadata,
  });

  factory AppEvent.fromJson(Map<String, dynamic> json) {
    return AppEvent(
      id: json['id'].toString(),
      date: DateTime.parse(
          json['date'] as String? ?? DateTime.now().toIso8601String()),
      type: _parseEventType(json['type'] as String? ?? ''),
      appId: json['app_id'].toString(),
      storeDomain: json['store_domain'] as String? ?? '',
      title: json['title'] as String? ?? '',
      description: json['description'] as String? ?? '',
      metadata: json['metadata'] != null
          ? Map<String, String>.from(json['metadata'] as Map)
          : null,
    );
  }

  static EventType _parseEventType(String s) {
    switch (s.toUpperCase()) {
      case 'APP_INSTALL':
        return EventType.appInstall;
      case 'APP_UNINSTALL':
        return EventType.appUninstall;
      case 'SUBSCRIPTION_ACTIVATED':
        return EventType.subscriptionActivated;
      case 'SUBSCRIPTION_CANCELLED':
        return EventType.subscriptionCancelled;
      case 'PLAN_UPGRADE':
        return EventType.planUpgrade;
      case 'PLAN_DOWNGRADE':
        return EventType.planDowngrade;
      case 'BILLING_FAILURE':
        return EventType.billingFailure;
      case 'BILLING_SUCCESS':
        return EventType.billingSuccess;
      case 'RISK_STATE_CHANGE':
        return EventType.riskStateChange;
      case 'REVIEW_SUBMITTED':
        return EventType.reviewSubmitted;
      case 'USAGE_CHARGE':
        return EventType.usageCharge;
      default:
        return EventType.appInstall;
    }
  }
}
