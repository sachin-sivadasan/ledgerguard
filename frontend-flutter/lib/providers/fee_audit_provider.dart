import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../services/fee_audit_service.dart';

class FeeAuditProvider extends ChangeNotifier {
  final FeeAuditService _service;

  bool _isLoading = false;
  bool _demoMode = false;
  String? _error;
  bool _isServiceUnavailable = false;
  String? _selectedAppId;
  CancelToken? _cancelToken;
  FeeAuditReport? _report;

  FeeAuditProvider(this._service);

  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isServiceUnavailable => _isServiceUnavailable;
  String? get selectedAppId => _selectedAppId;
  FeeAuditReport? get report => _report;

  /// Wired via DemoModeCoordinator — shows a mock dataset, no API call.
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
    if (token == _cancelToken) {
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<Uint8List?> fetchCsvBytes() {
    final appId = _selectedAppId;
    if (appId == null) return Future.value(null);
    return _service.fetchCsvBytes(appId);
  }

  FeeAuditReport _mockReport() => FeeAuditReport(
        currency: 'USD',
        configuredTier: 'SMALL_DEV_0',
        configuredFeePct: 0,
        detectedFeePct: 15,
        tierMatches: false,
        totalGrossCents: 20890218,
        totalCutCents: 2606271,
        effectiveFeePct: 12.5,
        flaggedMonths: 1,
        monthsAudited: 6,
        savingsVsDefaultCents: 1571773,
        months: const [
          FeeAuditRow(month: 'Mar', grossCents: 4100000, shopifyCutCents: 615000, effectiveFeePct: 15, expectedCutCents: 615000, feeVarianceCents: 0, feeGuardOk: true),
          FeeAuditRow(month: 'Apr', grossCents: 4300000, shopifyCutCents: 645000, effectiveFeePct: 15, expectedCutCents: 645000, feeVarianceCents: 0, feeGuardOk: true),
          FeeAuditRow(month: 'May', grossCents: 3900000, shopifyCutCents: 585000, effectiveFeePct: 15, expectedCutCents: 585000, feeVarianceCents: 0, feeGuardOk: true),
          FeeAuditRow(month: 'Jun', grossCents: 3200000, shopifyCutCents: 310000, effectiveFeePct: 9.7, expectedCutCents: 480000, feeVarianceCents: -170000, feeGuardOk: false),
          FeeAuditRow(month: 'Jul', grossCents: 720000, shopifyCutCents: 108000, effectiveFeePct: 15, expectedCutCents: 108000, feeVarianceCents: 0, feeGuardOk: true),
          FeeAuditRow(month: 'Aug', grossCents: 470218, shopifyCutCents: 70271, effectiveFeePct: 15, expectedCutCents: 70532, feeVarianceCents: -261, feeGuardOk: true),
        ],
      );
}
