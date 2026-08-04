import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../services/customer_insights_service.dart';

class CustomerInsightsProvider extends ChangeNotifier {
  final CustomerInsightsService _service;

  bool _isLoading = false;
  bool _demoMode = false;
  String? _error;
  bool _isServiceUnavailable = false;
  String? _selectedAppId;
  CancelToken? _cancelToken;
  CustomerInsights? _report;

  CustomerInsightsProvider(this._service);

  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isServiceUnavailable => _isServiceUnavailable;
  String? get selectedAppId => _selectedAppId;
  CustomerInsights? get report => _report;

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

  CustomerInsights _mockReport() => const CustomerInsights(
        currency: 'USD',
        totalCustomers: 2765,
        activeMrrCents: 8940000,
        atRiskCustomers: 214,
        atRiskMrrCents: 612000,
        revenueBands: [
          RevenueBand(label: '< \$25', customers: 1180, mrrCents: 1770000, pctOfCustomers: 0.4268),
          RevenueBand(label: '\$25–\$50', customers: 820, mrrCents: 2870000, pctOfCustomers: 0.2966),
          RevenueBand(label: '\$50–\$100', customers: 540, mrrCents: 3240000, pctOfCustomers: 0.1953),
          RevenueBand(label: '\$100–\$250', customers: 190, mrrCents: 2850000, pctOfCustomers: 0.0687),
          RevenueBand(label: '\$250+', customers: 35, mrrCents: 1210000, pctOfCustomers: 0.0127),
        ],
        riskSegments: [
          RiskSegment(riskState: 'SAFE', customers: 2551, mrrCents: 8328000),
          RiskSegment(riskState: 'AT_RISK', customers: 214, mrrCents: 612000),
          RiskSegment(riskState: 'CHURNED', customers: 389, mrrCents: 0),
        ],
        planRisk: [
          PlanRiskRow(planName: 'Growth', customers: 640, safeCount: 585, atRiskCount: 55, mrrCents: 4480000, atRiskMrrCents: 385000),
          PlanRiskRow(planName: 'Starter', customers: 1520, safeCount: 1418, atRiskCount: 102, mrrCents: 2660000, atRiskMrrCents: 143000),
          PlanRiskRow(planName: 'Pro', customers: 605, safeCount: 548, atRiskCount: 57, mrrCents: 1800000, atRiskMrrCents: 84000),
        ],
        topCustomers: [
          TopCustomer(shopName: 'northwind-supply', planName: 'Pro', mrrCents: 49900, riskState: 'SAFE'),
          TopCustomer(shopName: 'acme-outfitters', planName: 'Pro', mrrCents: 49900, riskState: 'ONE_CYCLE_MISSED'),
          TopCustomer(shopName: 'globex-store', planName: 'Growth', mrrCents: 29900, riskState: 'SAFE'),
          TopCustomer(shopName: 'umbrella-goods', planName: 'Growth', mrrCents: 29900, riskState: 'SAFE'),
          TopCustomer(shopName: 'initech-shop', planName: 'Growth', mrrCents: 29900, riskState: 'TWO_CYCLES_MISSED'),
        ],
      );
}
