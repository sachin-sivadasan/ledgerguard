import 'package:flutter/foundation.dart';
import '../mock_data/mock_playbooks.dart';
import '../mock_data/mock_stores.dart';
import '../models/analytics_model.dart';
import '../models/playbook_model.dart';
import '../models/store_model.dart';
import '../services/risk_service.dart';
import '../widgets/lg_risk_badge.dart';

class RiskProvider extends ChangeNotifier {
  final RiskService _riskService;

  bool _demoMode = false;
  bool _isLoading = false;
  String? _error;
  String? _selectedAppId;

  RiskSummary? _liveSummary;

  RiskProvider(this._riskService);

  bool get demoMode => _demoMode;
  bool get isLoading => _isLoading;
  String? get error => _error;
  String? get selectedAppId => _selectedAppId;

  void setDemoMode(bool value) {
    _demoMode = value;
    notifyListeners();
  }

  void setSelectedApp(String? appId) {
    _selectedAppId = appId;
    notifyListeners();
    if (!_demoMode && appId != null) {
      loadRiskSummary(appId);
    }
  }

  Future<void> loadRiskSummary(String appId) async {
    if (_demoMode || _isLoading) return;
    _isLoading = true;
    _error = null;
    notifyListeners();
    try {
      _liveSummary = await _riskService.fetchRiskSummary(appId);
    } catch (e) {
      _error = e.toString();
    }
    _isLoading = false;
    notifyListeners();
  }

  List<Store> get _filteredStores {
    if (_selectedAppId == null) return mockStores;
    return mockStores
        .where((s) => s.installedAppIds.contains(_selectedAppId))
        .toList();
  }

  RiskDistribution get distribution {
    if (!_demoMode) {
      return _liveSummary?.distribution ??
          const RiskDistribution(
              safe: 0, oneCycle: 0, twoCycle: 0, churned: 0);
    }
    final stores = _filteredStores;
    return RiskDistribution(
      safe: stores.where((s) => s.riskState == RiskState.safe).length,
      oneCycle:
          stores.where((s) => s.riskState == RiskState.oneCycleMissed).length,
      twoCycle:
          stores.where((s) => s.riskState == RiskState.twoCycleMissed).length,
      churned: stores.where((s) => s.riskState == RiskState.churned).length,
    );
  }

  List<Store> get atRiskStores {
    if (!_demoMode) return _liveSummary?.atRiskStores ?? [];
    return _filteredStores
        .where((s) => s.riskState != RiskState.safe)
        .toList()
      ..sort((a, b) => a.healthScore.compareTo(b.healthScore));
  }

  List<RecoveryPlaybook> get playbooks => mockPlaybooks;
}
