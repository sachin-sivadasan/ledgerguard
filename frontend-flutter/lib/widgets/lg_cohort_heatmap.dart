import 'package:flutter/material.dart';

import '../models/analytics_model.dart';
import '../theme/app_colors.dart';

// The heatmap gradient endpoints: a light indigo tint (low retention) to solid primary
// (high). Matches the wireframe's purple 6-step scale.
const Color _heatLow = Color(0xFFE1E4F8);
const Color _heatHigh = LgColors.primary;

/// Reusable cohort retention heatmap. Renders a horizontally-scrollable
/// [Table] where each row is a signup-month cohort and each month column
/// (M0 through M of the longest-lived cohort) is the percentage of that cohort
/// still retained N months later; M0 is the 100% signup baseline. Cell fill is a
/// purple intensity scale (light → solid indigo) with white text on the darker cells;
/// ragged rows (shorter than the widest cohort) render a grey "no data" cell. A
/// Low→High legend renders below the grid.
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

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        SingleChildScrollView(
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
                        return const _NoDataCell();
                      }),
                    ],
                  )),
            ],
          ),
        ),
        const SizedBox(height: 12),
        const _HeatmapLegend(),
      ],
    );
  }
}

/// Low→High gradient legend + a "no data" swatch, matching the wireframe.
class _HeatmapLegend extends StatelessWidget {
  const _HeatmapLegend();
  @override
  Widget build(BuildContext context) {
    const labelStyle = TextStyle(fontSize: 10, color: LgColors.textSecondary);
    return Row(
      children: [
        const Text('Low', style: labelStyle),
        const SizedBox(width: 6),
        Container(
          width: 96,
          height: 10,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(2),
            gradient: const LinearGradient(colors: [_heatLow, _heatHigh]),
          ),
        ),
        const SizedBox(width: 6),
        const Text('High', style: labelStyle),
        const SizedBox(width: 20),
        Container(width: 12, height: 12, color: LgColors.surfaceSecondary),
        const SizedBox(width: 4),
        const Text('No data', style: labelStyle),
      ],
    );
  }
}

/// A grey cell for a month a cohort hasn't reached yet (ragged rows).
class _NoDataCell extends StatelessWidget {
  const _NoDataCell();
  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.all(8),
        color: LgColors.surfaceSecondary,
        child: const Text('—',
            style: TextStyle(fontSize: 12, color: LgColors.textDisabled)),
      );
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
    // Purple intensity scale: light indigo (low) → solid primary (high). Blending toward
    // the tint keeps even low values readably purple rather than washed-out white.
    final bg = Color.lerp(_heatLow, _heatHigh, intensity)!;
    // White text on the darker (high-retention) cells for contrast; dark on the light ones.
    final fg = intensity > 0.55 ? Colors.white : LgColors.textPrimary;
    return Container(
      padding: const EdgeInsets.all(8),
      color: bg,
      child: Text(
        '${pct.toStringAsFixed(0)}%',
        style: TextStyle(
          fontSize: 12,
          color: fg,
          fontWeight: intensity > 0.7 ? FontWeight.w600 : FontWeight.w400,
        ),
      ),
    );
  }
}
