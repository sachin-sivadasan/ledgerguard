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

  factory Subscription.fromJson(Map<String, dynamic> json) {
    return Subscription(
      id: json['id'].toString(),
      shopDomain: json['myshopify_domain'] as String? ??
          json['shop_domain'] as String? ??
          '',
      appId: (json['app_id'] ?? '').toString(),
      planName: json['plan_name'] as String? ?? '',
      priceCents: json['base_price_cents'] as int? ??
          json['price_cents'] as int? ??
          0,
      status: _parseStatus(json['status'] as String? ?? 'ACTIVE'),
      riskState: parseRiskState(json['risk_state'] as String? ?? 'SAFE'),
      billingInterval: _parseBillingInterval(
          json['billing_interval'] as String? ?? 'EVERY_30_DAYS'),
      periodEnd: _parseDate(json['period_end'] ?? json['expected_next_charge']),
      expectedNextCharge: _parseDate(json['expected_next_charge']),
      createdAt: _parseDate(json['created_at']),
    );
  }

  static BillingInterval _parseBillingInterval(String s) {
    final upper = s.toUpperCase();
    if (upper.contains('ANNUAL') || upper.contains('365')) {
      return BillingInterval.annual;
    }
    return BillingInterval.monthly;
  }

  static DateTime _parseDate(dynamic value) {
    if (value == null) return DateTime.now();
    if (value is String) return DateTime.tryParse(value) ?? DateTime.now();
    return DateTime.now();
  }

  static SubscriptionStatus _parseStatus(String s) {
    switch (s.toUpperCase()) {
      case 'ACTIVE':
        return SubscriptionStatus.active;
      case 'FROZEN':
        return SubscriptionStatus.frozen;
      case 'CANCELLED':
        return SubscriptionStatus.cancelled;
      case 'PENDING':
        return SubscriptionStatus.pending;
      case 'UNINSTALLED':
        return SubscriptionStatus.uninstalled;
      default:
        return SubscriptionStatus.active;
    }
  }

  static RiskState parseRiskState(String s) {
    switch (s.toUpperCase()) {
      case 'SAFE':
        return RiskState.safe;
      case 'ONE_CYCLE_MISSED':
        return RiskState.oneCycleMissed;
      case 'TWO_CYCLES_MISSED':
        return RiskState.twoCycleMissed;
      case 'CHURNED':
        return RiskState.churned;
      default:
        return RiskState.safe;
    }
  }

  /// Server API value for a risk state (matches backend parseRiskState).
  static String riskStateToApi(RiskState r) => switch (r) {
        RiskState.safe => 'SAFE',
        RiskState.oneCycleMissed => 'ONE_CYCLE_MISSED',
        RiskState.twoCycleMissed => 'TWO_CYCLES_MISSED',
        RiskState.churned => 'CHURNED',
      };

  /// Server API value for a subscription status.
  static String statusToApi(SubscriptionStatus s) => switch (s) {
        SubscriptionStatus.active => 'ACTIVE',
        SubscriptionStatus.frozen => 'FROZEN',
        SubscriptionStatus.cancelled => 'CANCELLED',
        SubscriptionStatus.pending => 'PENDING',
        SubscriptionStatus.uninstalled => 'UNINSTALLED',
      };

  String get priceFormatted =>
      '\$${(priceCents / 100).toStringAsFixed(2)}';

  int get mrrCents => billingInterval == BillingInterval.annual
      ? (priceCents / 12).round()
      : priceCents;

  String get mrrFormatted =>
      '\$${(mrrCents / 100).toStringAsFixed(2)}/mo';
}
