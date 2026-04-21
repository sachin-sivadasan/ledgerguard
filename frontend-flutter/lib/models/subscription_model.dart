import '../widgets/lg_risk_badge.dart';
import '../widgets/lg_status_badge.dart';

enum BillingInterval { monthly, annual }

class Subscription {
  final String id;
  final String shopDomain;
  final String appId;
  final String planName;
  final int priceCents;
  final SubscriptionStatus status;
  final RiskState riskState;
  final BillingInterval billingInterval;
  final DateTime periodEnd;
  final DateTime expectedNextCharge;
  final DateTime createdAt;

  const Subscription({
    required this.id,
    required this.shopDomain,
    required this.appId,
    required this.planName,
    required this.priceCents,
    required this.status,
    required this.riskState,
    required this.billingInterval,
    required this.periodEnd,
    required this.expectedNextCharge,
    required this.createdAt,
  });

  String get priceFormatted =>
      '\$${(priceCents / 100).toStringAsFixed(2)}';

  int get mrrCents => billingInterval == BillingInterval.annual
      ? (priceCents / 12).round()
      : priceCents;

  String get mrrFormatted =>
      '\$${(mrrCents / 100).toStringAsFixed(2)}/mo';
}
