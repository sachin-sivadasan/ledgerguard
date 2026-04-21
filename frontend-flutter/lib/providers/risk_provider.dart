import 'package:flutter/foundation.dart';
import '../mock_data/mock_playbooks.dart';
import '../mock_data/mock_stores.dart';
import '../models/analytics_model.dart';
import '../models/playbook_model.dart';
import '../models/store_model.dart';
import '../widgets/lg_risk_badge.dart';

class RiskProvider extends ChangeNotifier {
  String? _selectedAppId;

  String? get selectedAppId => _selectedAppId;

  void setSelectedApp(String? appId) {
    _selectedAppId = appId;
    notifyListeners();
  }

  List<Store> get _filteredStores {
    if (_selectedAppId == null) return mockStores;
    return mockStores
        .where((s) => s.installedAppIds.contains(_selectedAppId))
        .toList();
  }

  RiskDistribution get distribution {
    final stores = _filteredStores;
    return RiskDistribution(
      safe: stores.where((s) => s.riskState == RiskState.safe).length,
      oneCycle: stores.where((s) => s.riskState == RiskState.oneCycleMissed).length,
      twoCycle: stores.where((s) => s.riskState == RiskState.twoCycleMissed).length,
      churned: stores.where((s) => s.riskState == RiskState.churned).length,
    );
  }

  List<Store> get atRiskStores => _filteredStores
      .where((s) => s.riskState != RiskState.safe)
      .toList()
    ..sort((a, b) => a.healthScore.compareTo(b.healthScore));

  List<RecoveryPlaybook> get playbooks => mockPlaybooks;
}
