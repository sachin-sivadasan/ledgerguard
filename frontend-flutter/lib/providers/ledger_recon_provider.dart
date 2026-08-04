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

  // Each month closes gross = net + revenue share + processing (~3%). Jun carries a
  // residual — a refund whose fee reversal never synced — to exercise the flagged path.
  ReconReport _mockReport() => const ReconReport(
        currency: 'USD',
        totalGrossCents: 20890218,
        totalNetCents: 17657676,
        totalRevenueShareCents: 2606271,
        totalProcessingCents: 626271,
        residualCents: 0,
        reconciled: false,
        monthsReconciled: 5,
        monthsFlagged: 1,
        monthsAudited: 6,
        months: [
          ReconMonth(month: 'Mar', grossCents: 4100000, netCents: 3362000, revenueShareCents: 615000, processingCents: 123000, accountedCents: 4100000, processingPct: 3.0, residualCents: 0, txCount: 210, reconciled: true),
          ReconMonth(month: 'Apr', grossCents: 4300000, netCents: 3526000, revenueShareCents: 645000, processingCents: 129000, accountedCents: 4300000, processingPct: 3.0, residualCents: 0, txCount: 221, reconciled: true),
          ReconMonth(month: 'May', grossCents: 3900000, netCents: 3198000, revenueShareCents: 585000, processingCents: 117000, accountedCents: 3900000, processingPct: 3.0, residualCents: 0, txCount: 198, reconciled: true),
          ReconMonth(month: 'Jun', grossCents: 3200000, netCents: 2624000, revenueShareCents: 480000, processingCents: 96000, accountedCents: 3200000, processingPct: 3.0, residualCents: 90000, txCount: 176, reconciled: false),
          ReconMonth(month: 'Jul', grossCents: 4680218, netCents: 3838779, revenueShareCents: 702033, processingCents: 139406, accountedCents: 4680218, processingPct: 2.98, residualCents: 0, txCount: 240, reconciled: true),
          ReconMonth(month: 'Aug', grossCents: 710000, netCents: 582300, revenueShareCents: 106500, processingCents: 21300, accountedCents: 710000, processingPct: 3.0, residualCents: 0, txCount: 33, reconciled: true),
        ],
      );
}
