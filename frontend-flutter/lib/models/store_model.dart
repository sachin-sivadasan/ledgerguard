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
}
