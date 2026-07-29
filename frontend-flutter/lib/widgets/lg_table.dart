import 'package:flutter/material.dart';
import '../theme/app_colors.dart';
import '../theme/app_spacing.dart';

/// A column definition for [LgTable].
class LgTableColumn {
  final String label;
  final int flex;

  /// Right-aligns the header + cells (for money / counts).
  final bool numeric;
  const LgTableColumn(this.label, {this.flex = 2, this.numeric = false});
}

/// A data table matching the wireframes: a shaded header row (rx=8) followed by flat
/// data rows separated by thin dividers — replacing the earlier card-per-row stacks.
/// Each row is a list of cell widgets, one per column (same order + count as [columns]).
class LgTable extends StatelessWidget {
  final List<LgTableColumn> columns;
  final List<List<Widget>> rows;
  const LgTable({super.key, required this.columns, required this.rows});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Header row.
        Container(
          padding: const EdgeInsets.symmetric(
              horizontal: LgSpacing.s400, vertical: LgSpacing.s300),
          decoration: BoxDecoration(
            color: LgColors.surfaceSecondary,
            borderRadius: BorderRadius.circular(8),
          ),
          child: Row(
            children: [
              for (var i = 0; i < columns.length; i++) ...[
                if (i > 0) const SizedBox(width: LgSpacing.s400),
                Expanded(
                  flex: columns[i].flex,
                  child: Text(
                    columns[i].label,
                    textAlign: columns[i].numeric ? TextAlign.end : TextAlign.start,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: LgColors.textSecondary,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
              ],
            ],
          ),
        ),
        // Data rows.
        for (final row in rows)
          Container(
            padding: const EdgeInsets.symmetric(
                horizontal: LgSpacing.s400, vertical: LgSpacing.s300),
            decoration: const BoxDecoration(
              border: Border(bottom: BorderSide(color: LgColors.border)),
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                for (var i = 0; i < columns.length; i++) ...[
                  if (i > 0) const SizedBox(width: LgSpacing.s400),
                  Expanded(
                    flex: columns[i].flex,
                    child: Align(
                      alignment: columns[i].numeric
                          ? Alignment.centerRight
                          : Alignment.centerLeft,
                      child: row[i],
                    ),
                  ),
                ],
              ],
            ),
          ),
      ],
    );
  }
}
