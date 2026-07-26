import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../services/churn_service.dart';

class ChurnProvider extends ChangeNotifier {
  final ChurnService _service;

  bool _isLoading = false;
  bool _demoMode = false;
  String? _error;
  bool _isServiceUnavailable = false;
  String? _selectedAppId;
  CancelToken? _cancelToken;

  ChurnReport? _report;

  ChurnProvider(this._service);

  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isServiceUnavailable => _isServiceUnavailable;
  String? get selectedAppId => _selectedAppId;
  ChurnReport? get report => _report;

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

  ChurnReport _mockReport() {
    final now = DateTime.now();
    return ChurnReport(
      currency: 'USD',
      churnRate: 0.042,
      churnedMrrLostCents: 112000,
      churnedCount: 9,
      trend: List.generate(
        8,
        (i) => ChurnTrendPoint(
          date: now.subtract(Duration(days: (7 - i) * 4)),
          // Gentle downward-ish drift with a little noise.
          churnRate: 0.058 - i * 0.002 + (i.isEven ? 0.001 : -0.001),
        ),
      ),
      stores: [
        ChurnStore(
          domain: 'acme-widgets.myshopify.com',
          shopName: 'Acme Widgets',
          mrrLostCents: 4900,
          churnedDate: now.subtract(const Duration(days: 24)),
          tenureDays: 252,
          planName: 'Pro',
        ),
        ChurnStore(
          domain: 'blue-ocean-store.myshopify.com',
          shopName: 'Blue Ocean',
          mrrLostCents: 3900,
          churnedDate: now.subtract(const Duration(days: 28)),
          tenureDays: 153,
          planName: 'Pro',
        ),
        ChurnStore(
          domain: 'quick-commerce.myshopify.com',
          shopName: 'Quick Commerce',
          mrrLostCents: 2900,
          churnedDate: now.subtract(const Duration(days: 35)),
          tenureDays: 87,
          planName: 'Starter',
        ),
        ChurnStore(
          domain: 'retro-goods.myshopify.com',
          shopName: 'Retro Goods',
          mrrLostCents: 1900,
          churnedDate: now.subtract(const Duration(days: 42)),
          tenureDays: 48,
          planName: 'Starter',
        ),
      ],
    );
  }
}
