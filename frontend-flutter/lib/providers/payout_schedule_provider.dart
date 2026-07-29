import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../core/date_range.dart';
import '../services/payout_schedule_service.dart';

class PayoutScheduleProvider extends ChangeNotifier {
  final PayoutScheduleService _service;

  bool _isLoading = false;
  bool _demoMode = false;
  String? _error;
  bool _isServiceUnavailable = false;
  String? _selectedAppId;
  DateRangePreset _dateRange = DateRangePreset.defaultPreset;
  CancelToken? _cancelToken;

  PayoutScheduleReport? _report;

  PayoutScheduleProvider(this._service);

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

  PayoutScheduleReport? get report => _report;

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
          limit: kPayoutSchedulePreview,
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

  /// Fetches a single server-paged window of the payouts table for the dedicated
  /// detail screen, using the currently selected app + date range.
  Future<PayoutScheduleReport> fetchPayoutsPage({
    required int limit,
    required int offset,
  }) {
    final appId = _selectedAppId;
    if (appId == null) return Future.value(PayoutScheduleReport.empty());
    final range = resolveDateRange(_dateRange, DateTime.now());
    return _service.fetchReport(appId,
        from: range.from, to: range.to, limit: limit, offset: offset);
  }

  PayoutScheduleReport _mockReport() {
    return PayoutScheduleReport(
      currency: 'USD',
      // Reconciles with the rows below (the report's invariant): available row
      // (298000) + pending rows (124000 + 98000 + 61000 = 283000).
      upcomingPayoutCents: 298000,
      pendingCents: 283000,
      nextPayoutDate: '2026-07-30',
      rowsTotal: 4,
      rows: [
        PayoutScheduleRow(
          availableDate: '2026-07-30',
          amountCents: 298000,
          chargeCount: 42,
          status: 'Available',
        ),
        PayoutScheduleRow(
          availableDate: '2026-08-06',
          amountCents: 124000,
          chargeCount: 28,
          status: 'Pending',
        ),
        PayoutScheduleRow(
          availableDate: '2026-08-13',
          amountCents: 98000,
          chargeCount: 19,
          status: 'Pending',
        ),
        PayoutScheduleRow(
          availableDate: '2026-08-20',
          amountCents: 61000,
          chargeCount: 11,
          status: 'Pending',
        ),
      ],
    );
  }
}
