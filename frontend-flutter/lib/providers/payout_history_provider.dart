import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../core/date_range.dart';
import '../services/payout_history_service.dart';

class PayoutHistoryProvider extends ChangeNotifier {
  final PayoutHistoryService _service;

  bool _isLoading = false;
  bool _demoMode = false;
  String? _error;
  bool _isServiceUnavailable = false;
  String? _selectedAppId;
  DateRangePreset _dateRange = DateRangePreset.defaultPreset;
  CancelToken? _cancelToken;

  PayoutHistoryReport? _report;

  PayoutHistoryProvider(this._service);

  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isServiceUnavailable => _isServiceUnavailable;
  String? get selectedAppId => _selectedAppId;
  DateRangePreset get dateRange => _dateRange;

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

  PayoutHistoryReport? get report => _report;

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
      // Report page shows a preview; the full table lives on the dedicated
      // payouts detail screen (server-paged). KPIs stay full regardless of limit.
      _report = await _service.fetchReport(appId,
          from: range.from,
          to: range.to,
          limit: kPayoutHistoryPreview,
          cancelToken: token);
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

  /// Fetches a single server-paged window of the payout-periods table for the
  /// dedicated detail screen, using the currently selected app + date range. KPIs
  /// come back too (full-set) but the detail screen only uses the paged rows +
  /// rowsTotal.
  Future<PayoutHistoryReport> fetchPayoutsPage({
    required int limit,
    required int offset,
  }) {
    final appId = _selectedAppId;
    if (appId == null) return Future.value(PayoutHistoryReport.empty());
    final range = resolveDateRange(_dateRange, DateTime.now());
    return _service.fetchReport(appId,
        from: range.from, to: range.to, limit: limit, offset: offset);
  }

  PayoutHistoryReport _mockReport() {
    // Rows sum to totalPaid; avg = total ÷ payoutCount (the report's invariant).
    return PayoutHistoryReport(
      currency: 'USD',
      totalPaidCents: 1050000, // 312000 + 278000 + 241000 + 219000
      payoutCount: 4,
      avgPayoutCents: 262500,
      rowsTotal: 4,
      rows: [
        PayoutHistoryRow(
          period: '2026-06',
          amountCents: 312000,
          chargeCount: 58,
          availableDate: '2026-07-05',
        ),
        PayoutHistoryRow(
          period: '2026-05',
          amountCents: 278000,
          chargeCount: 51,
          availableDate: '2026-06-05',
        ),
        PayoutHistoryRow(
          period: '2026-04',
          amountCents: 241000,
          chargeCount: 47,
          availableDate: '2026-05-05',
        ),
        PayoutHistoryRow(
          period: '2026-03',
          amountCents: 219000,
          chargeCount: 44,
          availableDate: '2026-04-05',
        ),
      ],
    );
  }
}
