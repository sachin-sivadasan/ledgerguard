import 'package:equatable/equatable.dart';

/// Represents the current billing status for the authenticated user.
class BillingStatus extends Equatable {
  final String plan;
  final String status;
  final int amountCents;
  final String currency;
  final DateTime? currentPeriodStart;
  final DateTime? currentPeriodEnd;
  final String? shortUrl;

  const BillingStatus({
    required this.plan,
    required this.status,
    this.amountCents = 0,
    this.currency = 'USD',
    this.currentPeriodStart,
    this.currentPeriodEnd,
    this.shortUrl,
  });

  factory BillingStatus.fromJson(Map<String, dynamic> json) {
    return BillingStatus(
      plan: json['plan'] as String? ?? 'FREE',
      status: json['status'] as String? ?? 'NONE',
      amountCents: (json['amount_cents'] as num?)?.toInt() ?? 0,
      currency: json['currency'] as String? ?? 'USD',
      currentPeriodStart: json['current_period_start'] != null
          ? DateTime.parse(json['current_period_start'] as String)
          : null,
      currentPeriodEnd: json['current_period_end'] != null
          ? DateTime.parse(json['current_period_end'] as String)
          : null,
      shortUrl: json['short_url'] as String?,
    );
  }

  bool get isActive => status == 'ACTIVE';
  bool get isFree => plan == 'FREE' && status == 'NONE';
  bool get isPaid => plan == 'STARTER' || plan == 'PRO';

  String get formattedPrice {
    if (amountCents == 0) return 'Free';
    final dollars = amountCents / 100;
    return '\$${dollars.toStringAsFixed(0)}/mo';
  }

  String get planDisplayName {
    switch (plan) {
      case 'STARTER':
        return 'Starter';
      case 'PRO':
        return 'Pro';
      default:
        return 'Free';
    }
  }

  @override
  List<Object?> get props => [
        plan,
        status,
        amountCents,
        currency,
        currentPeriodStart,
        currentPeriodEnd,
        shortUrl,
      ];
}

/// Represents the result of creating a checkout session.
class CheckoutResult extends Equatable {
  final String subscriptionId;
  final String shortUrl;

  const CheckoutResult({
    required this.subscriptionId,
    required this.shortUrl,
  });

  factory CheckoutResult.fromJson(Map<String, dynamic> json) {
    return CheckoutResult(
      subscriptionId: json['subscription_id'] as String,
      shortUrl: json['short_url'] as String,
    );
  }

  @override
  List<Object?> get props => [subscriptionId, shortUrl];
}
