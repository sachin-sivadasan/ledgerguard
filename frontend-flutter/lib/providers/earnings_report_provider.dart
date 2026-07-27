import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../services/earnings_report_service.dart';

class EarningsReportProvider extends ChangeNotifier {
  final EarningsReportService _service;

  bool _isLoading = false;
  bool _demoMode = false;
  String? _error;
  bool _isServiceUnavailable = false;
  String? _selectedAppId;
  CancelToken? _cancelToken;

  EarningsReport? _report;

  EarningsReportProvider(this._service);

  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isServiceUnavailable => _isServiceUnavailable;
  String? get selectedAppId => _selectedAppId;
  EarningsReport? get report => _report;

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
    try {
      _report = await _service.fetchReport(appId, cancelToken: token);
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
    return _service.fetchCsvBytes(appId);
  }

  EarningsReport _mockReport() {
    return EarningsReport(
      currency: 'USD',
      netEarningsCents: 842000,
      pendingCents: 124000,
      availableCents: 298000,
      paidOutCents: 420000,
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
