import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../services/usage_trends_service.dart';

class UsageTrendsProvider extends ChangeNotifier {
  final UsageTrendsService _service;

  bool _isLoading = false;
  bool _demoMode = false;
  String? _error;
  bool _isServiceUnavailable = false;
  String? _selectedAppId;
  CancelToken? _cancelToken;

  UsageTrendsReport? _report;

  UsageTrendsProvider(this._service);

  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isServiceUnavailable => _isServiceUnavailable;
  String? get selectedAppId => _selectedAppId;
  UsageTrendsReport? get report => _report;

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

  UsageTrendsReport _mockReport() {
    final now = DateTime.now();
    return UsageTrendsReport(
      currency: 'USD',
      usageMrrEquivCents: 104000,
      wowChangePct: 0.084,
      activeStores: 37,
      weeklyTrend: List.generate(
        8,
        (i) => UsageWeekPoint(
          // One point per week, oldest first.
          weekStart: now.subtract(Duration(days: (7 - i) * 7)),
          // Gentle upward drift around 24000 with a little noise.
          usageCents: 24000 + i * 400 + (i.isEven ? 600 : -600),
        ),
      ),
      stores: [
        UsageTrendStore(
          domain: 'acme-widgets.myshopify.com',
          shopName: 'Acme Widgets',
          usageCents: 98000,
          wowPct: 0.14,
        ),
        UsageTrendStore(
          domain: 'blue-ocean.myshopify.com',
          shopName: 'Blue Ocean',
          usageCents: 74000,
          wowPct: 0.06,
        ),
        UsageTrendStore(
          domain: 'quick-commerce.myshopify.com',
          shopName: 'Quick Commerce',
          usageCents: 52000,
          wowPct: -0.03,
        ),
      ],
    );
  }
}
