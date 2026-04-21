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
}
