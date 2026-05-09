import '../models/webhook_model.dart';

List<WebhookEvent> get mockWebhooks {
  final now = DateTime.now();
  return <WebhookEvent>[
  // Day 0 – today
  WebhookEvent(
    id: 'wh-1', receivedAt: now.subtract(const Duration(hours: 1)),
    source: WebhookSource.shopify, topic: 'subscription/activate',
    storeDomain: 'bright-gadgets.myshopify.com', appId: 'app-1',
    status: WebhookStatus.success, payloadSummary: 'Subscription #12345 activated', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-2', receivedAt: now.subtract(const Duration(hours: 2)),
    source: WebhookSource.razorpay, topic: 'subscription.charged',
    storeDomain: null, appId: 'app-1',
    status: WebhookStatus.success, payloadSummary: 'sub_HxR4Kq charged INR 4999', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-3', receivedAt: now.subtract(const Duration(hours: 3)),
    source: WebhookSource.shopify, topic: 'billing_attempt/success',
    storeDomain: 'eco-shop.myshopify.com', appId: 'app-2',
    status: WebhookStatus.success, payloadSummary: 'Charge \$19.99 collected', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-4', receivedAt: now.subtract(const Duration(hours: 5)),
    source: WebhookSource.shopify, topic: 'billing_attempt/failure',
    storeDomain: 'alpha-outlet.myshopify.com', appId: 'app-1',
    status: WebhookStatus.failed, payloadSummary: 'Card declined for \$49.99', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-5', receivedAt: now.subtract(const Duration(hours: 7)),
    source: WebhookSource.razorpay, topic: 'payment.failed',
    storeDomain: null, appId: 'app-1',
    status: WebhookStatus.failed, payloadSummary: 'Payment pay_LmN failed: insufficient_funds', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-6', receivedAt: now.subtract(const Duration(hours: 9)),
    source: WebhookSource.shopify, topic: 'subscription/update',
    storeDomain: 'daily-deals.myshopify.com', appId: 'app-3',
    status: WebhookStatus.success, payloadSummary: 'Plan changed to Pro', httpStatus: 200,
  ),

  // Day 1
  WebhookEvent(
    id: 'wh-7', receivedAt: now.subtract(const Duration(days: 1, hours: 1)),
    source: WebhookSource.shopify, topic: 'app/uninstalled',
    storeDomain: 'maple-candles.myshopify.com', appId: 'app-1',
    status: WebhookStatus.success, payloadSummary: 'InventorySync Pro uninstalled', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-8', receivedAt: now.subtract(const Duration(days: 1, hours: 4)),
    source: WebhookSource.razorpay, topic: 'subscription.activated',
    storeDomain: null, appId: 'app-2',
    status: WebhookStatus.success, payloadSummary: 'sub_JkQ activated', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-9', receivedAt: now.subtract(const Duration(days: 1, hours: 8)),
    source: WebhookSource.shopify, topic: 'billing_attempt/success',
    storeDomain: 'acme-store.myshopify.com', appId: 'app-1',
    status: WebhookStatus.success, payloadSummary: 'Charge \$49.99 collected', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-10', receivedAt: now.subtract(const Duration(days: 1, hours: 10)),
    source: WebhookSource.shopify, topic: 'subscription/activate',
    storeDomain: 'fresh-foods.myshopify.com', appId: 'app-2',
    status: WebhookStatus.failed, payloadSummary: 'Webhook handler timeout', httpStatus: 504,
  ),

  // Day 2
  WebhookEvent(
    id: 'wh-11', receivedAt: now.subtract(const Duration(days: 2, hours: 2)),
    source: WebhookSource.shopify, topic: 'billing_attempt/success',
    storeDomain: 'bright-gadgets.myshopify.com', appId: 'app-1',
    status: WebhookStatus.success, payloadSummary: 'Charge \$49.99 collected', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-12', receivedAt: now.subtract(const Duration(days: 2, hours: 6)),
    source: WebhookSource.razorpay, topic: 'subscription.charged',
    storeDomain: null, appId: 'app-3',
    status: WebhookStatus.success, payloadSummary: 'sub_AbC charged INR 2999', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-13', receivedAt: now.subtract(const Duration(days: 2, hours: 9)),
    source: WebhookSource.shopify, topic: 'subscription/update',
    storeDomain: 'glow-beauty.myshopify.com', appId: 'app-3',
    status: WebhookStatus.success, payloadSummary: 'Plan upgraded to Enterprise', httpStatus: 200,
  ),

  // Day 3
  WebhookEvent(
    id: 'wh-14', receivedAt: now.subtract(const Duration(days: 3, hours: 1)),
    source: WebhookSource.shopify, topic: 'billing_attempt/failure',
    storeDomain: 'beta-mart.myshopify.com', appId: 'app-2',
    status: WebhookStatus.failed, payloadSummary: 'Insufficient funds for \$19.99', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-15', receivedAt: now.subtract(const Duration(days: 3, hours: 5)),
    source: WebhookSource.razorpay, topic: 'subscription.cancelled',
    storeDomain: null, appId: 'app-1',
    status: WebhookStatus.success, payloadSummary: 'sub_DeF cancelled by merchant', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-16', receivedAt: now.subtract(const Duration(days: 3, hours: 8)),
    source: WebhookSource.shopify, topic: 'subscription/activate',
    storeDomain: 'cool-threads.myshopify.com', appId: 'app-1',
    status: WebhookStatus.success, payloadSummary: 'Subscription #23456 activated', httpStatus: 200,
  ),

  // Day 4
  WebhookEvent(
    id: 'wh-17', receivedAt: now.subtract(const Duration(days: 4, hours: 3)),
    source: WebhookSource.shopify, topic: 'billing_attempt/success',
    storeDomain: 'fresh-foods.myshopify.com', appId: 'app-3',
    status: WebhookStatus.success, payloadSummary: 'Charge \$49.99 collected', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-18', receivedAt: now.subtract(const Duration(days: 4, hours: 7)),
    source: WebhookSource.razorpay, topic: 'payment.failed',
    storeDomain: null, appId: 'app-2',
    status: WebhookStatus.failed, payloadSummary: 'Payment pay_XyZ failed: card_expired', httpStatus: 200,
  ),

  // Day 5
  WebhookEvent(
    id: 'wh-19', receivedAt: now.subtract(const Duration(days: 5, hours: 2)),
    source: WebhookSource.shopify, topic: 'subscription/update',
    storeDomain: 'eco-shop.myshopify.com', appId: 'app-1',
    status: WebhookStatus.success, payloadSummary: 'Plan changed to Enterprise', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-20', receivedAt: now.subtract(const Duration(days: 5, hours: 6)),
    source: WebhookSource.shopify, topic: 'billing_attempt/success',
    storeDomain: 'cool-threads.myshopify.com', appId: 'app-2',
    status: WebhookStatus.success, payloadSummary: 'Charge \$19.99 collected', httpStatus: 200,
  ),

  // Day 6
  WebhookEvent(
    id: 'wh-21', receivedAt: now.subtract(const Duration(days: 6, hours: 1)),
    source: WebhookSource.shopify, topic: 'app/uninstalled',
    storeDomain: 'craft-corner.myshopify.com', appId: 'app-3',
    status: WebhookStatus.failed, payloadSummary: 'Handler returned 500', httpStatus: 500,
  ),
  WebhookEvent(
    id: 'wh-22', receivedAt: now.subtract(const Duration(days: 6, hours: 5)),
    source: WebhookSource.razorpay, topic: 'subscription.charged',
    storeDomain: null, appId: 'app-1',
    status: WebhookStatus.success, payloadSummary: 'sub_GhI charged INR 4999', httpStatus: 200,
  ),

  // Day 7
  WebhookEvent(
    id: 'wh-23', receivedAt: now.subtract(const Duration(days: 7, hours: 3)),
    source: WebhookSource.shopify, topic: 'subscription/activate',
    storeDomain: 'acme-store.myshopify.com', appId: 'app-3',
    status: WebhookStatus.success, payloadSummary: 'Subscription #34567 activated', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-24', receivedAt: now.subtract(const Duration(days: 7, hours: 8)),
    source: WebhookSource.shopify, topic: 'billing_attempt/failure',
    storeDomain: 'ink-press.myshopify.com', appId: 'app-2',
    status: WebhookStatus.failed, payloadSummary: 'Expired card for \$19.99', httpStatus: 200,
  ),

  // Day 8
  WebhookEvent(
    id: 'wh-25', receivedAt: now.subtract(const Duration(days: 8, hours: 2)),
    source: WebhookSource.razorpay, topic: 'subscription.activated',
    storeDomain: null, appId: 'app-3',
    status: WebhookStatus.success, payloadSummary: 'sub_JkL activated', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-26', receivedAt: now.subtract(const Duration(days: 8, hours: 6)),
    source: WebhookSource.shopify, topic: 'billing_attempt/success',
    storeDomain: 'daily-deals.myshopify.com', appId: 'app-1',
    status: WebhookStatus.success, payloadSummary: 'Charge \$19.99 collected', httpStatus: 200,
  ),

  // Day 9
  WebhookEvent(
    id: 'wh-27', receivedAt: now.subtract(const Duration(days: 9, hours: 4)),
    source: WebhookSource.shopify, topic: 'subscription/update',
    storeDomain: 'glow-beauty.myshopify.com', appId: 'app-2',
    status: WebhookStatus.success, payloadSummary: 'Plan downgraded to Basic', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-28', receivedAt: now.subtract(const Duration(days: 9, hours: 9)),
    source: WebhookSource.razorpay, topic: 'payment.failed',
    storeDomain: null, appId: 'app-1',
    status: WebhookStatus.failed, payloadSummary: 'Payment pay_MnO failed: bank_refused', httpStatus: 200,
  ),

  // Day 10
  WebhookEvent(
    id: 'wh-29', receivedAt: now.subtract(const Duration(days: 10, hours: 2)),
    source: WebhookSource.shopify, topic: 'billing_attempt/success',
    storeDomain: 'acme-store.myshopify.com', appId: 'app-1',
    status: WebhookStatus.success, payloadSummary: 'Charge \$49.99 collected', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-30', receivedAt: now.subtract(const Duration(days: 10, hours: 7)),
    source: WebhookSource.shopify, topic: 'subscription/activate',
    storeDomain: 'dusk-decor.myshopify.com', appId: 'app-1',
    status: WebhookStatus.success, payloadSummary: 'Subscription #45678 activated', httpStatus: 200,
  ),

  // Day 11
  WebhookEvent(
    id: 'wh-31', receivedAt: now.subtract(const Duration(days: 11, hours: 3)),
    source: WebhookSource.razorpay, topic: 'subscription.charged',
    storeDomain: null, appId: 'app-2',
    status: WebhookStatus.success, payloadSummary: 'sub_PqR charged INR 1999', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-32', receivedAt: now.subtract(const Duration(days: 11, hours: 8)),
    source: WebhookSource.shopify, topic: 'billing_attempt/failure',
    storeDomain: 'neon-arcade.myshopify.com', appId: 'app-2',
    status: WebhookStatus.failed, payloadSummary: 'Card declined for \$29.99', httpStatus: 200,
  ),

  // Day 12
  WebhookEvent(
    id: 'wh-33', receivedAt: now.subtract(const Duration(days: 12, hours: 1)),
    source: WebhookSource.shopify, topic: 'subscription/update',
    storeDomain: 'bright-gadgets.myshopify.com', appId: 'app-1',
    status: WebhookStatus.success, payloadSummary: 'Plan upgraded to Pro', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-34', receivedAt: now.subtract(const Duration(days: 12, hours: 5)),
    source: WebhookSource.shopify, topic: 'billing_attempt/success',
    storeDomain: 'glow-beauty.myshopify.com', appId: 'app-3',
    status: WebhookStatus.success, payloadSummary: 'Charge \$49.99 collected', httpStatus: 200,
  ),

  // Day 13
  WebhookEvent(
    id: 'wh-35', receivedAt: now.subtract(const Duration(days: 13, hours: 2)),
    source: WebhookSource.razorpay, topic: 'subscription.cancelled',
    storeDomain: null, appId: 'app-2',
    status: WebhookStatus.success, payloadSummary: 'sub_StU cancelled by admin', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-36', receivedAt: now.subtract(const Duration(days: 13, hours: 6)),
    source: WebhookSource.shopify, topic: 'app/uninstalled',
    storeDomain: 'neon-arcade.myshopify.com', appId: 'app-2',
    status: WebhookStatus.success, payloadSummary: 'ReviewBoost uninstalled', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-37', receivedAt: now.subtract(const Duration(days: 13, hours: 10)),
    source: WebhookSource.shopify, topic: 'billing_attempt/success',
    storeDomain: 'eco-shop.myshopify.com', appId: 'app-1',
    status: WebhookStatus.failed, payloadSummary: 'Handler returned 502', httpStatus: 502,
  ),

  // Day 14
  WebhookEvent(
    id: 'wh-38', receivedAt: now.subtract(const Duration(days: 14, hours: 1)),
    source: WebhookSource.shopify, topic: 'subscription/activate',
    storeDomain: 'ink-press.myshopify.com', appId: 'app-2',
    status: WebhookStatus.success, payloadSummary: 'Subscription #56789 activated', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-39', receivedAt: now.subtract(const Duration(days: 14, hours: 5)),
    source: WebhookSource.razorpay, topic: 'subscription.charged',
    storeDomain: null, appId: 'app-1',
    status: WebhookStatus.success, payloadSummary: 'sub_VwX charged INR 4999', httpStatus: 200,
  ),
  WebhookEvent(
    id: 'wh-40', receivedAt: now.subtract(const Duration(days: 14, hours: 9)),
    source: WebhookSource.shopify, topic: 'billing_attempt/success',
    storeDomain: 'beta-mart.myshopify.com', appId: 'app-1',
    status: WebhookStatus.success, payloadSummary: 'Charge \$19.99 collected', httpStatus: 200,
  ),
]..sort((a, b) => b.receivedAt.compareTo(a.receivedAt));
}
