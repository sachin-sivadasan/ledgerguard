import '../models/payment_history_model.dart';
import '../models/transaction_model.dart';

final _now = DateTime.now();

/// Generates mock payment history for any subscription.
/// Uses subscription ID as seed for deterministic results.
List<PaymentHistoryEntry> generatePaymentHistory(String subscriptionId) {
  // Derive a seed from the subscription ID for consistent data
  final seed = subscriptionId.hashCode.abs();
  final isChurned = subscriptionId.contains('churn');
  final isTwoCycle = subscriptionId.contains('two');
  final isOneCycle = subscriptionId.contains('one');

  // Determine price from subscription ID pattern
  final priceCents = subscriptionId.contains('safe')
      ? [1999, 4999, 9999][seed % 3]
      : subscriptionId.contains('one')
          ? (seed % 2 == 0 ? 4999 : 1999)
          : subscriptionId.contains('two')
              ? 4999
              : 1999;

  final entries = <PaymentHistoryEntry>[];
  int id = 1;

  // Generate 12 months of payment history
  for (int monthsAgo = 11; monthsAgo >= 0; monthsAgo--) {
    final txDate = DateTime(_now.year, _now.month - monthsAgo, 1 + (seed % 15));

    // Skip recent months for at-risk/churned subscriptions
    if (isChurned && monthsAgo < 4) continue;
    if (isTwoCycle && monthsAgo < 2) continue;
    if (isOneCycle && monthsAgo < 1) continue;

    // Recurring charge
    final earningsStatus = monthsAgo > 2
        ? EarningsStatus.paidOut
        : monthsAgo > 1
            ? EarningsStatus.available
            : EarningsStatus.pending;

    entries.add(PaymentHistoryEntry(
      id: '$subscriptionId-pay-${id++}',
      transactionDate: txDate,
      chargeType: ChargeType.recurring,
      grossAmountCents: priceCents,
      netAmountCents: (priceCents * 0.8).round(),
      earningsStatus: earningsStatus,
    ));

    // Add occasional usage charges for safe subscriptions
    if (!isChurned && !isTwoCycle && monthsAgo % 3 == 0) {
      final usageAmount = 500 + (seed + monthsAgo) % 2000;
      entries.add(PaymentHistoryEntry(
        id: '$subscriptionId-pay-${id++}',
        transactionDate: txDate.add(const Duration(days: 10)),
        chargeType: ChargeType.usage,
        grossAmountCents: usageAmount,
        netAmountCents: (usageAmount * 0.8).round(),
        earningsStatus: earningsStatus,
      ));
    }

    // Add a refund for churned subscriptions in their last active month
    if (isChurned && monthsAgo == 4) {
      entries.add(PaymentHistoryEntry(
        id: '$subscriptionId-pay-${id++}',
        transactionDate: txDate.add(const Duration(days: 20)),
        chargeType: ChargeType.refund,
        grossAmountCents: -priceCents,
        netAmountCents: -(priceCents * 0.8).round(),
        earningsStatus: EarningsStatus.paidOut,
      ));
    }
  }

  // Sort newest first
  entries.sort((a, b) => b.transactionDate.compareTo(a.transactionDate));
  return entries;
}
