import 'package:flutter/material.dart';

import '../models/analytics_model.dart';
import '../theme/app_colors.dart';

/// Reusable cohort retention heatmap. Renders a horizontally-scrollable
/// [Table] where each row is a signup-month cohort and each month column
/// (M0 through M of the longest-lived cohort) is the percentage of that cohort
/// still retained N months later; M0 is the 100% signup baseline. Green intensity
/// scales with the percentage; ragged rows (shorter than the widest cohort)
/// render '—' for the missing months.
///
/// Extracted from analytics/cohort_tab.dart so the analytics tab and the
/// Retention Cohorts report share a single rendering implementation.
class CohortHeatmap extends StatelessWidget {
  final List<CohortData> cohorts;
  const CohortHeatmap({super.key, required this.cohorts});

  @override
  Widget build(BuildContext context) {
    final maxMonths = cohorts.fold<int>(
        0, (m, c) => c.retentionPcts.length > m ? c.retentionPcts.length : m);

    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: Table(
        defaultColumnWidth: const FixedColumnWidth(80),
        columnWidths: const {
          0: FixedColumnWidth(100),
          1: FixedColumnWidth(60),
        },
        children: [
          // Header
          TableRow(
            decoration: BoxDecoration(color: LgColors.surfaceSecondary),
            children: [
              const _HeaderCell('Cohort'),
              const _HeaderCell('N'),
              ...List.generate(maxMonths, (i) => _HeaderCell('M$i')),
            ],
          ),
          // Data
          ...cohorts.map((c) => TableRow(
                children: [
                  _DataCell(c.cohortMonth),
                  _DataCell('${c.initialStores}'),
                  ...List.generate(maxMonths, (i) {
                    if (i < c.retentionPcts.length) {
                      return _HeatCell(c.retentionPcts[i]);
                    }
                    return const _DataCell('—');
                  }),
                ],
              )),
        ],
      ),
    );
  }
}

class _HeaderCell extends StatelessWidget {
  final String text;
  const _HeaderCell(this.text);
  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.all(8),
        child: Text(text,
            style: const TextStyle(
                fontSize: 11,
                fontWeight: FontWeight.w600,
                color: LgColors.textSecondary)),
      );
}

class _DataCell extends StatelessWidget {
  final String text;
  const _DataCell(this.text);
  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.all(8),
        child: Text(text, style: const TextStyle(fontSize: 12)),
      );
}

class _HeatCell extends StatelessWidget {
  final double pct;
  const _HeatCell(this.pct);

  @override
  Widget build(BuildContext context) {
    final intensity = (pct / 100).clamp(0.0, 1.0);
    return Container(
      padding: const EdgeInsets.all(8),
      color: LgColors.success.withValues(alpha: intensity * 0.3),
      child: Text(
        '${pct.toStringAsFixed(0)}%',
        style: TextStyle(
          fontSize: 12,
          fontWeight: intensity > 0.7 ? FontWeight.w600 : FontWeight.w400,
        ),
      ),
    );
  }
}
