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

  factory Transaction.fromJson(Map<String, dynamic> json) {
    return Transaction(
      id: json['id'].toString(),
      date: DateTime.parse(
          json['date'] as String? ?? DateTime.now().toIso8601String()),
      shopDomain: json['shop_domain'] as String? ?? '',
      chargeType: _parseChargeType(json['charge_type'] as String? ?? ''),
      appId: json['app_id'].toString(),
      grossAmountCents: json['gross_amount_cents'] as int? ?? 0,
      netAmountCents: json['net_amount_cents'] as int? ?? 0,
    );
  }

  static ChargeType _parseChargeType(String s) {
    switch (s.toUpperCase()) {
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
