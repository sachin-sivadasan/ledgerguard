import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../core/date_range.dart';
import '../services/revenue_at_risk_service.dart';
import '../widgets/lg_risk_badge.dart';

class RevenueAtRiskProvider extends ChangeNotifier {
  final RevenueAtRiskService _service;

  bool _isLoading = false;
  bool _demoMode = false;
  String? _error;
  bool _isServiceUnavailable = false;
  String? _selectedAppId;
  DateRangePreset _dateRange = DateRangePreset.defaultPreset;
  CancelToken? _cancelToken;

  RevenueAtRiskReport? _report;

  RevenueAtRiskProvider(this._service);

  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isServiceUnavailable => _isServiceUnavailable;
  String? get selectedAppId => _selectedAppId;
  DateRangePreset get dateRange => _dateRange;
  RevenueAtRiskReport? get report => _report;

  /// Wired from the LgPage date-range selector. Re-loads the report for the new window.
  void setDateRange(DateRangePreset preset) {
    if (preset == _dateRange) return;
    _dateRange = preset;
    notifyListeners();
    final appId = _selectedAppId;
    if (!_demoMode && appId != null) {
      loadReport(appId);
    }
  }

  /// Wired via DemoModeCoordinator. In demo mode the report shows a mock
  /// dataset (consistent with every other screen) and no API call is made.
  void setDemoMode(bool value) {
    _demoMode = value;
    _cancelToken?.cancel('demo mode change');
    _isLoading = false;
    _error = null;
    _isServiceUnavailable = false;
    _report = value ? _mockReport() : null;
    notifyListeners();
  }

  void setSelectedApp(String? appId) {
    _selectedAppId = appId;
    notifyListeners();
    if (!_demoMode && appId != null) {
      loadReport(appId);
    }
  }

  Future<void> loadReport(String appId) async {
    _cancelToken?.cancel('Superseded');
    _cancelToken = CancelToken();
    _isLoading = true;
    _error = null;
    _isServiceUnavailable = false;
    notifyListeners();
    final token = _cancelToken;
    final range = resolveDateRange(_dateRange, DateTime.now());
    try {
      _report = await _service.fetchReport(appId,
          from: range.from, to: range.to, cancelToken: token);
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) {
        // A newer load superseded this one and will manage loading state.
        // But if this token is still the active one (a lone cancel with no
        // successor), fall through so we clear the spinner instead of leaving
        // it stuck forever.
        if (token != _cancelToken) return;
      } else if (e.response?.statusCode == 503) {
        _isServiceUnavailable = true;
        _error = 'Service temporarily unavailable.';
      } else {
        _error = e.message ?? e.toString();
      }
    } catch (e) {
      _error = e.toString();
    }
    // Only the most recent request is allowed to settle loading state, so a
    // superseded request can't stomp a newer in-flight load.
    if (token == _cancelToken) {
      _isLoading = false;
      notifyListeners();
    }
  }

  /// Fetches the CSV export bytes for the currently selected app through the
  /// authenticated ApiClient. Returns null if no app is selected.
  Future<Uint8List?> fetchCsvBytes() {
    final appId = _selectedAppId;
    if (appId == null) return Future.value(null);
    final range = resolveDateRange(_dateRange, DateTime.now());
    return _service.fetchCsvBytes(appId, from: range.from, to: range.to);
  }

  RevenueAtRiskReport _mockReport() {
    final now = DateTime.now();
    return RevenueAtRiskReport(
      currency: 'USD',
      totalAtRiskCents: 18420,
      recoverableCents: 9840,
      oneCycleCents: 12000,
      twoCycleCents: 6420,
      oneCycleCount: 8,
      twoCycleCount: 4,
      trend: List.generate(
        8,
        (i) => RevenueAtRiskTrendPoint(
          date: now.subtract(Duration(days: (7 - i) * 4)),
          atRiskCents: 15000 + i * 480,
        ),
      ),
      stores: [
        RevenueAtRiskStore(
          domain: 'acme-widgets.myshopify.com',
          shopName: 'Acme Widgets',
          mrrCents: 4900,
          riskState: RiskState.twoCycleMissed,
          daysLate: 47,
          expectedChargeDate: now.subtract(const Duration(days: 47)),
          planName: 'Pro',
          recoverableCents: 1225,
        ),
        RevenueAtRiskStore(
          domain: 'blue-ocean-store.myshopify.com',
          shopName: 'Blue Ocean',
          mrrCents: 3900,
          riskState: RiskState.oneCycleMissed,
          daysLate: 22,
          expectedChargeDate: now.subtract(const Duration(days: 22)),
          planName: 'Pro',
          recoverableCents: 2340,
        ),
        RevenueAtRiskStore(
          domain: 'quick-commerce.myshopify.com',
          shopName: 'Quick Commerce',
          mrrCents: 2900,
          riskState: RiskState.oneCycleMissed,
          daysLate: 18,
          expectedChargeDate: now.subtract(const Duration(days: 18)),
          planName: 'Starter',
          recoverableCents: 1740,
        ),
      ],
    );
  }
}
