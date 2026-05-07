import '../widgets/lg_risk_badge.dart';
import 'timeline_event.dart';

class Store {
  final String id;
  final String shopDomain;
  final List<String> installedAppIds;
  final int healthScore;
  final int lifetimeValueCents;
  final DateTime firstInstallDate;
  final DateTime lastInteraction;
  final RiskState riskState;
  final List<TimelineEvent> timeline;

  const Store({
    required this.id,
    required this.shopDomain,
    required this.installedAppIds,
    required this.healthScore,
    required this.lifetimeValueCents,
    required this.firstInstallDate,
    required this.lastInteraction,
    required this.riskState,
    this.timeline = const [],
  });

  String get ltvFormatted =>
      '\$${(lifetimeValueCents / 100).toStringAsFixed(2)}';

  factory Store.fromJson(Map<String, dynamic> json) {
    return Store(
      id: json['id'].toString(),
      shopDomain: json['shop_domain'] as String? ?? '',
      installedAppIds: (json['installed_app_ids'] as List<dynamic>?)
              ?.map((e) => e.toString())
              .toList() ??
          [],
      healthScore: json['health_score'] as int? ?? 0,
      lifetimeValueCents: json['lifetime_value_cents'] as int? ?? 0,
      firstInstallDate: DateTime.parse(json['first_install_date'] as String? ??
          DateTime.now().toIso8601String()),
      lastInteraction: DateTime.parse(json['last_interaction'] as String? ??
          DateTime.now().toIso8601String()),
      riskState: _parseRiskState(json['risk_state'] as String? ?? 'SAFE'),
    );
  }

  static RiskState _parseRiskState(String s) {
    switch (s.toUpperCase()) {
      case 'SAFE':
        return RiskState.safe;
      case 'ONE_CYCLE_MISSED':
        return RiskState.oneCycleMissed;
      case 'TWO_CYCLES_MISSED':
        return RiskState.twoCycleMissed;
      case 'CHURNED':
        return RiskState.churned;
      default:
        return RiskState.safe;
    }
  }
}
