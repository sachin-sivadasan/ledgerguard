import 'package:flutter/material.dart';
import 'lg_badge.dart';

enum RiskState { safe, oneCycleMissed, twoCycleMissed, churned }

class LgRiskBadge extends StatelessWidget {
  final RiskState riskState;

  const LgRiskBadge({super.key, required this.riskState});

  @override
  Widget build(BuildContext context) {
    return LgBadge(label: _label, tone: _tone);
  }

  String get _label => switch (riskState) {
        RiskState.safe => 'Safe',
        RiskState.oneCycleMissed => '1 Cycle Missed',
        RiskState.twoCycleMissed => '2 Cycles Missed',
        RiskState.churned => 'Churned',
      };

  BadgeTone get _tone => switch (riskState) {
        RiskState.safe => BadgeTone.success,
        RiskState.oneCycleMissed => BadgeTone.warning,
        RiskState.twoCycleMissed => BadgeTone.attention,
        RiskState.churned => BadgeTone.critical,
      };
}
