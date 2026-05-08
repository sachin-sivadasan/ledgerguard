import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../mock_data/mock_earnings.dart';
import '../models/earning_model.dart';
import '../services/earnings_service.dart';

class EarningsProvider extends ChangeNotifier {
  final EarningsService _earningsService;

  bool _demoMode = false;
  bool _isLoading = false;
  String? _error;
  String? _selectedAppId;
  CancelToken? _cancelToken;

  List<EarningPeriod> _liveEarnings = [];

  EarningsProvider(this._earningsService);

  bool get demoMode => _demoMode;
  bool get isLoading => _isLoading;
  String? get error => _error;
  String? get selectedAppId => _selectedAppId;

  void setDemoMode(bool value) {
    _demoMode = value;
    notifyListeners();
  }

  void setSelectedApp(String? appId) {
    _selectedAppId = appId;
    notifyListeners();
    if (!_demoMode && appId != null) {
      loadEarnings(appId);
    }
  }

  Future<void> loadEarnings(String appId) async {
    if (_demoMode) return;
    _cancelToken?.cancel('Superseded');
    _cancelToken = CancelToken();
    _isLoading = true;
    _error = null;
    notifyListeners();
    try {
      _liveEarnings = await _earningsService.fetchEarnings(appId,
          cancelToken: _cancelToken);
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) return;
      _error = e.message;
    } catch (e) {
      _error = e.toString();
    }
    _isLoading = false;
    notifyListeners();
  }

  List<EarningPeriod> get _allPeriods =>
      _demoMode ? mockEarningPeriods : _liveEarnings;

  List<EarningPeriod> get periods =>
      List.of(_allPeriods)
        ..sort((a, b) => b.startDate.compareTo(a.startDate));

  String get totalEarned {
    final cents = _allPeriods
        .where((p) => p.status == EarningStatus.paidOut)
        .fold<int>(0, (sum, p) => sum + p.netEarningsCents);
    return '\$${(cents / 100).toStringAsFixed(2)}';
  }

  String get pendingAmount {
    final cents = _allPeriods
        .where((p) => p.status == EarningStatus.pending)
        .fold<int>(0, (sum, p) => sum + p.netEarningsCents);
    return '\$${(cents / 100).toStringAsFixed(2)}';
  }

  String get availableAmount {
    final cents = _allPeriods
        .where((p) => p.status == EarningStatus.available)
        .fold<int>(0, (sum, p) => sum + p.netEarningsCents);
    return '\$${(cents / 100).toStringAsFixed(2)}';
  }

  FeeBreakdown calculateFees(int grossCents) {
    final tier = currentTier;
    final shopifyPct = tier.ratePct / 100;
    const processingPct = 0.029;
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
