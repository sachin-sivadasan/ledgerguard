import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../services/subscriptions_service.dart';

class SubscriptionsProvider extends ChangeNotifier {
  final SubscriptionsService _service;

  bool _isLoading = false;
  bool _demoMode = false;
  String? _error;
  bool _isServiceUnavailable = false;
  String? _selectedAppId;
  CancelToken? _cancelToken;

  SubscriptionsReport? _report;

  SubscriptionsProvider(this._service);

  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isServiceUnavailable => _isServiceUnavailable;
  String? get selectedAppId => _selectedAppId;
  SubscriptionsReport? get report => _report;

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

  SubscriptionsReport _mockReport() {
    return SubscriptionsReport(
      currency: 'USD',
      activeSubs: 148,
      activeMrrCents: 1243200, // ARPU ≈ $84
      arpuCents: 8400,
      ltvCents: 234000, // ARPU ÷ ~3.6% churn
      churnRate: 0.036,
      plans: [
        SubscriptionsPlan(
          planName: 'Starter',
          activeSubs: 92,
          mrrCents: 266800,
          arpuCents: 2900,
          ltvCents: 81000,
          pctOfSubs: 92 / 148,
        ),
        SubscriptionsPlan(
          planName: 'Pro',
          activeSubs: 44,
          mrrCents: 435600,
          arpuCents: 9900,
          ltvCents: 276000,
          pctOfSubs: 44 / 148,
        ),
        SubscriptionsPlan(
          planName: 'Enterprise',
          activeSubs: 12,
          mrrCents: 358800,
          arpuCents: 29900,
          ltvCents: 834000,
          pctOfSubs: 12 / 148,
        ),
      ],
    );
  }
}
