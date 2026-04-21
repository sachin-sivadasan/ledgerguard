import 'transaction_model.dart';

enum EarningsStatus { pending, available, paidOut }

class PaymentHistoryEntry {
  final String id;
  final DateTime transactionDate;
  final ChargeType chargeType;
  final int grossAmountCents;
  final int netAmountCents;
  final EarningsStatus earningsStatus;

  const PaymentHistoryEntry({
    required this.id,
    required this.transactionDate,
    required this.chargeType,
    required this.grossAmountCents,
    required this.netAmountCents,
    required this.earningsStatus,
  });

  String get grossFormatted =>
      '\$${(grossAmountCents / 100).toStringAsFixed(2)}';
  String get netFormatted =>
      '\$${(netAmountCents / 100).toStringAsFixed(2)}';

  String get chargeTypeLabel => switch (chargeType) {
        ChargeType.recurring => 'RECURRING',
        ChargeType.usage => 'USAGE',
        ChargeType.oneTime => 'ONE_TIME',
        ChargeType.refund => 'REFUND',
      };

  String get earningsStatusLabel => switch (earningsStatus) {
        EarningsStatus.pending => 'Pending',
        EarningsStatus.available => 'Available',
        EarningsStatus.paidOut => 'Paid Out',
      };
}
