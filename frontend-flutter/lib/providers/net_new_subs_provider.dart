import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../core/date_range.dart';
import '../services/net_new_subs_service.dart';

class NetNewSubsProvider extends ChangeNotifier {
  final NetNewSubsService _service;

  bool _isLoading = false;
  bool _demoMode = false;
  String? _error;
  bool _isServiceUnavailable = false;
  String? _selectedAppId;
  DateRangePreset _dateRange = DateRangePreset.defaultPreset;
  CancelToken? _cancelToken;

  NetNewSubsReport? _report;

  NetNewSubsProvider(this._service);

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
    final range = resolveDateRange(_dateRange, DateTime.now());
    try {
      // Report page shows a preview; the full new-stores table lives on the
      // dedicated subscriptions detail screen (server-paged). KPIs/trend stay
      // full regardless of limit.
      _report = await _service.fetchReport(appId,
          from: range.from,
          to: range.to,
          limit: kNetNewSubsPreview,
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

  /// Fetches a single server-paged window of the new-stores table for the
  /// dedicated detail screen, using the currently selected app + date range.
  /// KPIs/trend come back too (full-set) but the detail screen only uses the
  /// paged rows + newStoresTotal.
  Future<NetNewSubsReport> fetchSubscriptionsPage({
    required int limit,
    required int offset,
  }) {
    final appId = _selectedAppId;
    if (appId == null) return Future.value(NetNewSubsReport.empty());
    final range = resolveDateRange(_dateRange, DateTime.now());
    return _service.fetchReport(appId,
        from: range.from, to: range.to, limit: limit, offset: offset);
  }

  NetNewSubsReport _mockReport() {
    final base = DateTime(2026, 7, 1);
    return NetNewSubsReport(
      currency: 'USD',
      newSubs: 38,
      churned: 9,
      net: 29,
      newStoresTotal: 3,
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
