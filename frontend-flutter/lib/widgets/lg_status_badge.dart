import 'package:flutter/material.dart';
import 'lg_badge.dart';

enum SubscriptionStatus { active, frozen, cancelled, pending }

class LgStatusBadge extends StatelessWidget {
  final SubscriptionStatus status;

  const LgStatusBadge({super.key, required this.status});

  @override
  Widget build(BuildContext context) {
    return LgBadge(label: _label, tone: _tone);
  }

  String get _label => switch (status) {
        SubscriptionStatus.active => 'Active',
        SubscriptionStatus.frozen => 'Frozen',
        SubscriptionStatus.cancelled => 'Cancelled',
        SubscriptionStatus.pending => 'Pending',
      };

  BadgeTone get _tone => switch (status) {
        SubscriptionStatus.active => BadgeTone.success,
        SubscriptionStatus.frozen => BadgeTone.warning,
        SubscriptionStatus.cancelled => BadgeTone.critical,
        SubscriptionStatus.pending => BadgeTone.info,
      };
}
