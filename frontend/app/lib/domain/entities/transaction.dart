import 'package:equatable/equatable.dart';

/// Represents a charge type for transactions
enum ChargeType {
  recurring,
  usage,
  oneTime,
  refund;

  String get displayName {
    switch (this) {
      case ChargeType.recurring:
        return 'Subscription';
      case ChargeType.usage:
        return 'Usage';
      case ChargeType.oneTime:
        return 'One-Time';
      case ChargeType.refund:
        return 'Refund';
    }
  }

  static ChargeType fromString(String value) {
    switch (value.toUpperCase()) {
      case 'RECURRING':
        return ChargeType.recurring;
      case 'USAGE':
        return ChargeType.usage;
      case 'ONE_TIME':
        return ChargeType.oneTime;
      case 'REFUND':
        return ChargeType.refund;
      default:
        return ChargeType.recurring;
    }
  }
}

/// Represents an earnings status for transactions
enum EarningsStatus {
  pending,
  available,
  paidOut;

  String get displayName {
    switch (this) {
      case EarningsStatus.pending:
        return 'Pending';
      case EarningsStatus.available:
        return 'Available';
      case EarningsStatus.paidOut:
        return 'Paid Out';
    }
  }

  static EarningsStatus fromString(String value) {
    switch (value.toUpperCase()) {
      case 'PENDING':
        return EarningsStatus.pending;
      case 'AVAILABLE':
        return EarningsStatus.available;
      case 'PAID_OUT':
        return EarningsStatus.paidOut;
      default:
        return EarningsStatus.pending;
    }
  }
}

/// Represents a transaction from the Shopify Partner API
class Transaction extends Equatable {
  final String id;
  final String shopifyGid;
  final ChargeType chargeType;
  final int grossAmountCents;
  final int netAmountCents;
  final String currency;
  final DateTime transactionDate;
  final EarningsStatus earningsStatus;
  final DateTime? availableDate;

  const Transaction({
    required this.id,
    required this.shopifyGid,
    required this.chargeType,
    required this.grossAmountCents,
    required this.netAmountCents,
    required this.currency,
    required this.transactionDate,
    required this.earningsStatus,
    this.availableDate,
  });

  /// Format gross amount as currency string
  String get formattedGrossAmount => _formatCurrency(grossAmountCents);

  /// Format net amount as currency string
  String get formattedNetAmount => _formatCurrency(netAmountCents);

  /// Calculate fees (gross - net)
  int get feesCents => grossAmountCents - netAmountCents;

  /// Format fees as currency string
  String get formattedFees => _formatCurrency(feesCents);

  /// Days until available (null if already available or paid out)
  int? get daysUntilAvailable {
    if (earningsStatus != EarningsStatus.pending || availableDate == null) {
      return null;
    }
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final availDate = DateTime(
      availableDate!.year,
      availableDate!.month,
      availableDate!.day,
    );
    final diff = availDate.difference(today).inDays;
    return diff >= 0 ? diff : 0;
  }

  String _formatCurrency(int cents) {
    final dollars = cents / 100;
    if (dollars >= 1000) {
      return '\$${(dollars / 1000).toStringAsFixed(1)}K';
    }
    return '\$${dollars.toStringAsFixed(2)}';
  }

  factory Transaction.fromJson(Map<String, dynamic> json) {
    return Transaction(
      id: json['id'] as String,
      shopifyGid: json['shopify_gid'] as String,
      chargeType: ChargeType.fromString(json['charge_type'] as String),
      grossAmountCents: json['gross_amount_cents'] as int,
      netAmountCents: json['net_amount_cents'] as int,
      currency: json['currency'] as String,
      transactionDate: DateTime.parse(json['transaction_date'] as String),
      earningsStatus:
          EarningsStatus.fromString(json['earnings_status'] as String),
      availableDate: json['available_date'] != null
          ? DateTime.parse(json['available_date'] as String)
          : null,
    );
  }

  @override
  List<Object?> get props => [
        id,
        shopifyGid,
        chargeType,
        grossAmountCents,
        netAmountCents,
        currency,
        transactionDate,
        earningsStatus,
        availableDate,
      ];
}
