import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../services/net_new_subs_service.dart';

class NetNewSubsProvider extends ChangeNotifier {
  final NetNewSubsService _service;

  bool _isLoading = false;
  bool _demoMode = false;
  String? _error;
  bool _isServiceUnavailable = false;
  String? _selectedAppId;
  CancelToken? _cancelToken;

  NetNewSubsReport? _report;

  NetNewSubsProvider(this._service);

  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isServiceUnavailable => _isServiceUnavailable;
  String? get selectedAppId => _selectedAppId;
  NetNewSubsReport? get report => _report;

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

  NetNewSubsReport _mockReport() {
    final base = DateTime(2026, 7, 1);
    return NetNewSubsReport(
      currency: 'USD',
      newSubs: 38,
      churned: 9,
      net: 29,
      trend: List.generate(8, (i) {
        final n = 4 + (i % 3) + (i ~/ 2);
        final c = 1 + (i % 2);
        return NetNewTrendPoint(
          date: base.add(Duration(days: i * 3)),
          newSubs: n,
          churned: c,
          net: n - c,
        );
      }),
      newStores: [
        NewSubRow(
            domain: 'acme-widgets.myshopify.com',
            shopName: 'Acme Widgets',
            planName: 'Pro',
            mrrCents: 4900,
            started: '2026-07-24'),
        NewSubRow(
            domain: 'blue-ocean-store.myshopify.com',
            shopName: 'Blue Ocean',
            planName: 'Growth',
            mrrCents: 3900,
            started: '2026-07-23'),
        NewSubRow(
            domain: 'quick-commerce.myshopify.com',
            shopName: 'Quick Commerce',
            planName: 'Starter',
            mrrCents: 1900,
            started: '2026-07-22'),
      ],
    );
  }
}
