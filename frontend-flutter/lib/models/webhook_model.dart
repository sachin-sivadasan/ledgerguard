enum WebhookSource { shopify, razorpay }

enum WebhookStatus { success, failed }

class WebhookEvent {
  final String id;
  final DateTime receivedAt;
  final WebhookSource source;
  final String topic;
  final String? storeDomain;
  final String? appId;
  final WebhookStatus status;
  final String payloadSummary;
  final int httpStatus;

  const WebhookEvent({
    required this.id,
    required this.receivedAt,
    required this.source,
    required this.topic,
    this.storeDomain,
    this.appId,
    required this.status,
    required this.payloadSummary,
    required this.httpStatus,
  });

  String get sourceLabel => switch (source) {
        WebhookSource.shopify => 'Shopify',
        WebhookSource.razorpay => 'Razorpay',
      };
}
