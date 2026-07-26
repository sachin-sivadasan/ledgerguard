import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../providers/analytics_provider.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_cohort_heatmap.dart';
import '../../widgets/lg_empty_state.dart';

class CohortTab extends StatelessWidget {
  const CohortTab({super.key});

  @override
  Widget build(BuildContext context) {
    final provider = context.watch<AnalyticsProvider>();
    final cohorts = provider.cohorts;

    if (cohorts.isEmpty) {
      return const LgEmptyState(
        icon: Icons.group_work,
        heading: 'Cohort data not yet available',
        description:
            'Cohort retention analysis requires at least two months of subscription data.',
      );
    }

    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          LgCard(
            title: 'Cohort Retention',
            child: CohortHeatmap(cohorts: cohorts),
          ),
        ],
      ),
    );
  }
}
