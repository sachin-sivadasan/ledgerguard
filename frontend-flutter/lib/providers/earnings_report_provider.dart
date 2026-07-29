import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../core/date_range.dart';
import '../services/earnings_report_service.dart';

class EarningsReportProvider extends ChangeNotifier {
  final EarningsReportService _service;

  bool _isLoading = false;
  bool _demoMode = false;
  String? _error;
  bool _isServiceUnavailable = false;
  String? _selectedAppId;
  DateRangePreset _dateRange = DateRangePreset.defaultPreset;
  CancelToken? _cancelToken;

  EarningsReport? _report;

  EarningsReportProvider(this._service);

  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isServiceUnavailable => _isServiceUnavailable;
  String? get selectedAppId => _selectedAppId;
  DateRangePreset get dateRange => _dateRange;
  EarningsReport? get report => _report;

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
      // Report page shows a preview; the full table lives on the dedicated
      // charges detail screen (server-paged). KPIs stay full regardless of limit.
      _report = await _service.fetchReport(appId,
          from: range.from,
          to: range.to,
          limit: kEarningsChargesPreview,
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

  /// Fetches a single server-paged window of the charges table for the dedicated
  /// detail screen, using the currently selected app + date range. KPIs come back
  /// too (full-set) but the detail screen only uses the paged rows + chargesTotal.
  Future<EarningsReport> fetchChargesPage({
    required int limit,
    required int offset,
  }) {
    final appId = _selectedAppId;
    if (appId == null) return Future.value(EarningsReport.empty());
    final range = resolveDateRange(_dateRange, DateTime.now());
    return _service.fetchReport(appId,
        from: range.from, to: range.to, limit: limit, offset: offset);
  }

  EarningsReport _mockReport() {
    return EarningsReport(
      currency: 'USD',
      netEarningsCents: 842000,
      pendingCents: 124000,
      availableCents: 298000,
      paidOutCents: 420000,
      chargesTotal: 4,
      charges: [
        EarningsCharge(
          date: DateTime(2026, 7, 22),
          domain: 'acme-widgets.myshopify.com',
          shopName: 'Acme Widgets',
          grossCents: 4900,
          netCents: 3920,
          status: 'Pending',
          availableDate: DateTime(2026, 8, 6),
        ),
        EarningsCharge(
          date: DateTime(2026, 7, 18),
          domain: 'blue-ocean.myshopify.com',
          shopName: 'Blue Ocean',
          grossCents: 3900,
          netCents: 3120,
          status: 'Available',
          availableDate: DateTime(2026, 7, 25),
        ),
        EarningsCharge(
          date: DateTime(2026, 7, 12),
          domain: 'quick-commerce.myshopify.com',
          shopName: 'Quick Commerce',
          grossCents: 2900,
          netCents: 2320,
          status: 'Available',
          availableDate: DateTime(2026, 7, 19),
        ),
        EarningsCharge(
          date: DateTime(2026, 7, 5),
          domain: 'north-star.myshopify.com',
          shopName: 'North Star',
          grossCents: 9900,
          netCents: 7920,
          status: 'Paid',
          availableDate: DateTime(2026, 7, 12),
        ),
      ],
    );
  }
}
