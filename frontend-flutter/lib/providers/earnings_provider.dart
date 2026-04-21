import 'package:flutter/foundation.dart';
import '../mock_data/mock_earnings.dart';
import '../models/earning_model.dart';

class EarningsProvider extends ChangeNotifier {
  String? _selectedAppId;

  String? get selectedAppId => _selectedAppId;

  void setSelectedApp(String? appId) {
    _selectedAppId = appId;
    notifyListeners();
  }

  List<EarningPeriod> get periods =>
      List.of(mockEarningPeriods)..sort((a, b) => b.startDate.compareTo(a.startDate));

  String get totalEarned {
    final cents = mockEarningPeriods
        .where((p) => p.status == EarningStatus.paidOut)
        .fold<int>(0, (sum, p) => sum + p.netEarningsCents);
    return '\$${(cents / 100).toStringAsFixed(2)}';
  }

  String get pendingAmount {
    final cents = mockEarningPeriods
        .where((p) => p.status == EarningStatus.pending)
        .fold<int>(0, (sum, p) => sum + p.netEarningsCents);
    return '\$${(cents / 100).toStringAsFixed(2)}';
  }

  String get availableAmount {
    final cents = mockEarningPeriods
        .where((p) => p.status == EarningStatus.available)
        .fold<int>(0, (sum, p) => sum + p.netEarningsCents);
    return '\$${(cents / 100).toStringAsFixed(2)}';
  }

  FeeBreakdown calculateFees(int grossCents) {
    final tier = currentTier;
    final shopifyPct = tier.ratePct / 100;
    final processingPct = 0.029;
    final shopifyFee = (grossCents * shopifyPct).round();
    final processingFee = (grossCents * processingPct).round();
    return FeeBreakdown(
      grossCents: grossCents,
      shopifyFeePct: tier.ratePct,
      shopifyFeeCents: shopifyFee,
      processingFeePct: processingPct * 100,
      processingFeeCents: processingFee,
      netCents: grossCents - shopifyFee - processingFee,
    );
  }

  List<RevenueShareTier> get tiers => mockRevenueShareTiers;

  RevenueShareTier get currentTier =>
      mockRevenueShareTiers.firstWhere((t) => t.isCurrentTier);
}
