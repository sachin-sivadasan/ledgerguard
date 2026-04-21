enum ChargeType { recurring, usage, oneTime, refund }

class Transaction {
  final String id;
  final DateTime date;
  final String shopDomain;
  final ChargeType chargeType;
  final String appId;
  final int grossAmountCents;
  final int netAmountCents;

  const Transaction({
    required this.id,
    required this.date,
    required this.shopDomain,
    required this.chargeType,
    required this.appId,
    required this.grossAmountCents,
    required this.netAmountCents,
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
}
