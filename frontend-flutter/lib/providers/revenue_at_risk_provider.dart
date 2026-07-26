import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../services/revenue_at_risk_service.dart';

class RevenueAtRiskProvider extends ChangeNotifier {
  final RevenueAtRiskService _service;

  bool _isLoading = false;
  String? _error;
  bool _isServiceUnavailable = false;
  String? _selectedAppId;
  CancelToken? _cancelToken;

  RevenueAtRiskReport? _report;

  RevenueAtRiskProvider(this._service);

  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isServiceUnavailable => _isServiceUnavailable;
  String? get selectedAppId => _selectedAppId;
  RevenueAtRiskReport? get report => _report;

  void setSelectedApp(String? appId) {
    _selectedAppId = appId;
    notifyListeners();
    if (appId != null) {
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
    try {
      _report = await _service.fetchReport(appId, cancelToken: _cancelToken);
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) return;
      if (e.response?.statusCode == 503) {
        _isServiceUnavailable = true;
        _error = 'Service temporarily unavailable.';
      } else {
        _error = e.message ?? e.toString();
      }
    } catch (e) {
      _error = e.toString();
    }
    _isLoading = false;
    notifyListeners();
  }

  /// Absolute CSV export URL for the currently selected app.
  String? csvExportUrl() {
    final appId = _selectedAppId;
    if (appId == null) return null;
    return _service.csvExportUrl(appId);
  }
}
