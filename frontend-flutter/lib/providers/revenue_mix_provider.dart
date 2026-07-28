import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../core/date_range.dart';
import '../services/revenue_mix_service.dart';

class RevenueMixProvider extends ChangeNotifier {
  final RevenueMixService _service;

  bool _isLoading = false;
  bool _demoMode = false;
  String? _error;
  bool _isServiceUnavailable = false;
  String? _selectedAppId;
  DateRangePreset _dateRange = DateRangePreset.defaultPreset;
  CancelToken? _cancelToken;

  RevenueMixReport? _report;

  RevenueMixProvider(this._service);

  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isServiceUnavailable => _isServiceUnavailable;
  String? get selectedAppId => _selectedAppId;
  DateRangePreset get dateRange => _dateRange;
  RevenueMixReport? get report => _report;

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

  RevenueMixReport _mockReport() {
    return RevenueMixReport(
      currency: 'USD',
      recurringCents: 1248000,
      usageCents: 440500,
      oneTimeCents: 146800,
      refundCents: 0,
      grossCents: 1835300,
      netCents: 1835300,
      segments: [
        RevenueMixSegment(type: 'Recurring', amountCents: 1248000, pct: 0.68),
        RevenueMixSegment(type: 'Usage', amountCents: 440500, pct: 0.24),
        RevenueMixSegment(type: 'One-time', amountCents: 146800, pct: 0.08),
      ],
    );
  }
}
