import 'package:equatable/equatable.dart';

/// Risk state for a subscription
enum RiskState {
  safe,
  oneCycleMissed,
  twoCyclesMissed,
  churned;

  String get displayName {
    switch (this) {
      case RiskState.safe:
        return 'Safe';
      case RiskState.oneCycleMissed:
        return 'At Risk';
      case RiskState.twoCyclesMissed:
        return 'High Risk';
      case RiskState.churned:
        return 'Churned';
    }
  }

  String get apiValue {
    switch (this) {
      case RiskState.safe:
        return 'SAFE';
      case RiskState.oneCycleMissed:
        return 'ONE_CYCLE_MISSED';
      case RiskState.twoCyclesMissed:
        return 'TWO_CYCLES_MISSED';
      case RiskState.churned:
        return 'CHURNED';
    }
  }

  static RiskState fromString(String value) {
    switch (value) {
      case 'SAFE':
        return RiskState.safe;
      case 'ONE_CYCLE_MISSED':
        return RiskState.oneCycleMissed;
      case 'TWO_CYCLES_MISSED':
        return RiskState.twoCyclesMissed;
      case 'CHURNED':
        return RiskState.churned;
      default:
        return RiskState.safe;
    }
  }
}

/// Billing interval for subscriptions
enum BillingInterval {
  monthly,
  annual;

  String get displayName {
    switch (this) {
      case BillingInterval.monthly:
        return 'Monthly';
      case BillingInterval.annual:
        return 'Annual';
    }
  }

  String get apiValue {
    switch (this) {
      case BillingInterval.monthly:
        return 'MONTHLY';
      case BillingInterval.annual:
        return 'ANNUAL';
    }
  }

  static BillingInterval fromString(String value) {
    switch (value.toUpperCase()) {
      case 'ANNUAL':
        return BillingInterval.annual;
      case 'MONTHLY':
      default:
        return BillingInterval.monthly;
    }
  }
}

/// Represents a subscription in the system
class Subscription extends Equatable {
  final String id;
  final String shopifyGid;
  final String myshopifyDomain;
  final String? shopName;
  final String planName;
  final int basePriceCents;
  final BillingInterval billingInterval;
  final RiskState riskState;
  final String status;
  final DateTime createdAt;
  final DateTime? expectedNextCharge;
  final DateTime? lastChargeDate;

  const Subscription({
    required this.id,
    required this.shopifyGid,
    required this.myshopifyDomain,
    this.shopName,
    required this.planName,
    required this.basePriceCents,
    required this.billingInterval,
    required this.riskState,
    required this.status,
    required this.createdAt,
    this.expectedNextCharge,
    this.lastChargeDate,
  });

  /// Display name for the subscription (shop name or domain)
  String get displayName => shopName?.isNotEmpty == true ? shopName! : myshopifyDomain;

  /// Monthly recurring revenue in cents
  int get mrrCents {
    if (billingInterval == BillingInterval.annual) {
      return basePriceCents ~/ 12;
    }
    return basePriceCents;
  }

  /// Formatted price string
  String get formattedPrice {
    final dollars = basePriceCents / 100;
    return '\$${dollars.toStringAsFixed(2)}/${billingInterval == BillingInterval.monthly ? 'mo' : 'yr'}';
  }

  /// Is this subscription active?
  bool get isActive => status == 'ACTIVE';

  /// Factory constructor from JSON
  factory Subscription.fromJson(Map<String, dynamic> json) {
    return Subscription(
      id: json['id'] as String,
      shopifyGid: json['shopify_gid'] as String? ?? '',
      myshopifyDomain: json['myshopify_domain'] as String? ?? '',
      shopName: json['shop_name'] as String?,
      planName: json['plan_name'] as String? ?? 'Unknown Plan',
      basePriceCents: json['base_price_cents'] as int? ?? 0,
      billingInterval: BillingInterval.fromString(
        json['billing_interval'] as String? ?? 'MONTHLY',
      ),
      riskState: RiskState.fromString(
        json['risk_state'] as String? ?? 'SAFE',
      ),
      status: json['status'] as String? ?? 'ACTIVE',
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : DateTime.now(),
      expectedNextCharge: json['expected_next_charge'] != null
          ? DateTime.parse(json['expected_next_charge'] as String)
          : null,
      lastChargeDate: json['last_charge_date'] != null
          ? DateTime.parse(json['last_charge_date'] as String)
          : null,
    );
  }

  @override
  List<Object?> get props => [
        id,
        shopifyGid,
        myshopifyDomain,
        shopName,
        planName,
        basePriceCents,
        billingInterval,
        riskState,
        status,
        createdAt,
        expectedNextCharge,
        lastChargeDate,
      ];
}

/// Response wrapper for paginated subscription list
class SubscriptionListResponse extends Equatable {
  final List<Subscription> subscriptions;
  final int total;
  final int limit;
  final int offset;

  const SubscriptionListResponse({
    required this.subscriptions,
    required this.total,
    required this.limit,
    required this.offset,
  });

  bool get hasMore => offset + subscriptions.length < total;

  factory SubscriptionListResponse.fromJson(Map<String, dynamic> json) {
    final subscriptionsList = (json['subscriptions'] as List<dynamic>? ?? [])
        .map((s) => Subscription.fromJson(s as Map<String, dynamic>))
        .toList();

    return SubscriptionListResponse(
      subscriptions: subscriptionsList,
      total: json['total'] as int? ?? subscriptionsList.length,
      limit: json['limit'] as int? ?? 50,
      offset: json['offset'] as int? ?? 0,
    );
  }

  @override
  List<Object?> get props => [subscriptions, total, limit, offset];
}
