import 'package:equatable/equatable.dart';

import 'subscription.dart';
import 'transaction.dart';

/// Represents store-specific earnings summary
class StoreEarnings extends Equatable {
  final int pendingCents;
  final int availableCents;
  final int paidOutCents;
  final int totalCents;

  const StoreEarnings({
    required this.pendingCents,
    required this.availableCents,
    required this.paidOutCents,
    required this.totalCents,
  });

  /// Format pending as currency
  String get formattedPending => _formatCurrency(pendingCents);

  /// Format available as currency
  String get formattedAvailable => _formatCurrency(availableCents);

  /// Format paid out as currency
  String get formattedPaidOut => _formatCurrency(paidOutCents);

  /// Format total as currency
  String get formattedTotal => _formatCurrency(totalCents);

  String _formatCurrency(int cents) {
    final dollars = cents / 100;
    if (dollars >= 1000) {
      return '\$${(dollars / 1000).toStringAsFixed(1)}K';
    }
    return '\$${dollars.toStringAsFixed(2)}';
  }

  factory StoreEarnings.fromJson(Map<String, dynamic> json) {
    return StoreEarnings(
      pendingCents: json['pending_cents'] as int,
      availableCents: json['available_cents'] as int,
      paidOutCents: json['paid_out_cents'] as int,
      totalCents: json['total_cents'] as int,
    );
  }

  static const empty = StoreEarnings(
    pendingCents: 0,
    availableCents: 0,
    paidOutCents: 0,
    totalCents: 0,
  );

  @override
  List<Object?> get props => [
        pendingCents,
        availableCents,
        paidOutCents,
        totalCents,
      ];
}

/// Represents complete health data for a store
class StoreHealth extends Equatable {
  final Subscription subscription;
  final List<Transaction> transactions;
  final StoreEarnings earnings;

  const StoreHealth({
    required this.subscription,
    required this.transactions,
    required this.earnings,
  });

  /// Get recurring transactions only
  List<Transaction> get recurringTransactions => transactions
      .where((tx) => tx.chargeType == ChargeType.recurring)
      .toList();

  /// Get usage transactions only
  List<Transaction> get usageTransactions =>
      transactions.where((tx) => tx.chargeType == ChargeType.usage).toList();

  /// Total revenue from all transactions (net)
  int get totalRevenueCents =>
      transactions.fold(0, (sum, tx) => sum + tx.netAmountCents);

  /// Format total revenue
  String get formattedTotalRevenue {
    final dollars = totalRevenueCents / 100;
    if (dollars >= 1000) {
      return '\$${(dollars / 1000).toStringAsFixed(1)}K';
    }
    return '\$${dollars.toStringAsFixed(2)}';
  }

  /// Average transaction value
  int get avgTransactionCents {
    if (transactions.isEmpty) return 0;
    return totalRevenueCents ~/ transactions.length;
  }

  factory StoreHealth.fromJson(Map<String, dynamic> json) {
    return StoreHealth(
      subscription: Subscription.fromJson(
        json['subscription'] as Map<String, dynamic>,
      ),
      transactions: (json['transactions'] as List<dynamic>)
          .map((e) => Transaction.fromJson(e as Map<String, dynamic>))
          .toList(),
      earnings: StoreEarnings.fromJson(
        json['earnings'] as Map<String, dynamic>,
      ),
    );
  }

  @override
  List<Object?> get props => [subscription, transactions, earnings];
}
