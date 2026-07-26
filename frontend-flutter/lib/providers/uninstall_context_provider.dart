import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../services/uninstall_context_service.dart';

class UninstallContextProvider extends ChangeNotifier {
  final UninstallContextService _service;

  bool _isLoading = false;
  bool _demoMode = false;
  String? _error;
  bool _isServiceUnavailable = false;
  String? _selectedAppId;
  CancelToken? _cancelToken;

  UninstallContextReport? _report;

  UninstallContextProvider(this._service);

  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isServiceUnavailable => _isServiceUnavailable;
  String? get selectedAppId => _selectedAppId;
  UninstallContextReport? get report => _report;

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

  UninstallContextReport _mockReport() {
    final now = DateTime.now();
    return UninstallContextReport(
      uninstalls: 14,
      wereAtRiskPct: 0.71,
      medianTenureMonths: 3.2,
      stores: [
        UninstallStore(
          domain: 'acme-widgets.myshopify.com',
          stateBeforeUninstall: 'At-Risk',
          planName: 'Pro',
          tenureMonths: 8.4,
          uninstalledDate: now.subtract(const Duration(days: 3)),
        ),
        UninstallStore(
          domain: 'blue-ocean.myshopify.com',
          stateBeforeUninstall: 'Frozen',
          planName: 'Growth',
          tenureMonths: 5.1,
          uninstalledDate: now.subtract(const Duration(days: 7)),
        ),
        UninstallStore(
          domain: 'quick-commerce.myshopify.com',
          stateBeforeUninstall: 'At-Risk',
          planName: 'Starter',
          tenureMonths: 2.9,
          uninstalledDate: now.subtract(const Duration(days: 14)),
        ),
        UninstallStore(
          domain: 'retro-goods.myshopify.com',
          stateBeforeUninstall: 'Healthy',
          planName: 'Pro',
          tenureMonths: 1.6,
          uninstalledDate: now.subtract(const Duration(days: 21)),
        ),
      ],
    );
  }
}
