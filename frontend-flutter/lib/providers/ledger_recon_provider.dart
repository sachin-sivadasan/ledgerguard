import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../services/ledger_recon_service.dart';

class LedgerReconProvider extends ChangeNotifier {
  final LedgerReconService _service;

  bool _isLoading = false;
  bool _demoMode = false;
  String? _error;
  bool _isServiceUnavailable = false;
  String? _selectedAppId;
  CancelToken? _cancelToken;
  ReconReport? _report;

  LedgerReconProvider(this._service);

  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isServiceUnavailable => _isServiceUnavailable;
  String? get selectedAppId => _selectedAppId;
  ReconReport? get report => _report;

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

  ReconReport _mockReport() => const ReconReport(
        currency: 'USD',
        totalGrossCents: 20890218,
        totalFeeCents: 2606271,
        totalNetCents: 18283947,
        residualCents: 0,
        reconciled: false,
        monthsReconciled: 5,
        monthsFlagged: 1,
        monthsAudited: 6,
        months: [
          ReconMonth(month: 'Mar', grossCents: 4100000, feeCents: 615000, netCents: 3485000, expectedNetCents: 3485000, residualCents: 0, txCount: 210, reconciled: true),
          ReconMonth(month: 'Apr', grossCents: 4300000, feeCents: 645000, netCents: 3655000, expectedNetCents: 3655000, residualCents: 0, txCount: 221, reconciled: true),
          ReconMonth(month: 'May', grossCents: 3900000, feeCents: 585000, netCents: 3315000, expectedNetCents: 3315000, residualCents: 0, txCount: 198, reconciled: true),
          ReconMonth(month: 'Jun', grossCents: 3200000, feeCents: 0, netCents: 2720000, expectedNetCents: 3200000, residualCents: -480000, txCount: 176, reconciled: false),
          ReconMonth(month: 'Jul', grossCents: 4680218, feeCents: 702033, netCents: 3978185, expectedNetCents: 3978185, residualCents: 0, txCount: 240, reconciled: true),
          ReconMonth(month: 'Aug', grossCents: 710000, feeCents: 106500, netCents: 603500, expectedNetCents: 603500, residualCents: 0, txCount: 33, reconciled: true),
        ],
      );
}
