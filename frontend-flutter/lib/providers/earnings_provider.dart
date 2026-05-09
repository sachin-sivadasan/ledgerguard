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
  List<RevenueShareTier>? _liveTiers;
  FeeBreakdownResponse? _liveFeeBreakdown;
  FeeSummary? _liveFeeSummary;
  EarningsStatus? _liveEarningsStatus;

  EarningsProvider(this._earningsService);

  bool get demoMode => _demoMode;
  bool get isLoading => _isLoading;
  String? get error => _error;
  String? get selectedAppId => _selectedAppId;
  FeeSummary? get feeSummary => _demoMode ? null : _liveFeeSummary;
  EarningsStatus? get earningsStatus =>
      _demoMode ? null : _liveEarningsStatus;
  FeeBreakdownResponse? get feeBreakdownResponse =>
      _demoMode ? null : _liveFeeBreakdown;

  void setDemoMode(bool value) {
    _demoMode = value;
    notifyListeners();
  }

  void setSelectedApp(String? appId) {
    _selectedAppId = appId;
    notifyListeners();
    if (!_demoMode && appId != null) {
      loadEarnings(appId);
      loadTiers();
      loadFeeSummary(appId);
      loadEarningsStatus(appId);
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

  Future<void> loadTiers() async {
    if (_demoMode) return;
    _liveTiers = await _earningsService.fetchTiers();
    notifyListeners();
  }

  Future<void> loadFeeSummary(String appId) async {
    if (_demoMode) return;
    _liveFeeSummary = await _earningsService.fetchFeeSummary(appId);
    notifyListeners();
  }

  Future<void> loadEarningsStatus(String appId) async {
    if (_demoMode) return;
    _liveEarningsStatus =
        await _earningsService.fetchEarningsStatus(appId);
    notifyListeners();
  }

  Future<void> loadFeeBreakdown(String appId, int amountCents) async {
    if (_demoMode) return;
    _liveFeeBreakdown =
        await _earningsService.fetchFeeBreakdown(appId, amountCents);
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
    if (!_demoMode && _liveEarningsStatus != null) {
      return _liveEarningsStatus!.pendingFormatted;
    }
    final cents = _allPeriods
        .where((p) => p.status == EarningStatus.pending)
        .fold<int>(0, (sum, p) => sum + p.netEarningsCents);
    return '\$${(cents / 100).toStringAsFixed(2)}';
  }

  String get availableAmount {
    if (!_demoMode && _liveEarningsStatus != null) {
      return _liveEarningsStatus!.availableFormatted;
    }
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

  List<RevenueShareTier> get tiers {
    if (!_demoMode && _liveTiers != null && _liveTiers!.isNotEmpty) {
      return _liveTiers!;
    }
    return mockRevenueShareTiers;
  }

  RevenueShareTier get currentTier =>
      tiers.firstWhere((t) => t.isCurrentTier,
          orElse: () => tiers.first);
}
