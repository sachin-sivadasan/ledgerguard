import 'package:flutter/material.dart';
import '../theme/app_breakpoints.dart';
import '../theme/app_spacing.dart';

/// Arranges [LgMetricCard] widgets in a responsive grid.
///
/// Mobile: 2 columns, Tablet: 3 columns, Desktop: 4 columns.
class LgMetricGrid extends StatelessWidget {
  final List<Widget> children;

  const LgMetricGrid({super.key, required this.children});

  @override
  Widget build(BuildContext context) {
    final columns = LgBreakpoints.metricColumns(context);
    return LayoutBuilder(
      builder: (context, constraints) {
        final spacing = LgSpacing.s400;
        final totalSpacing = spacing * (columns - 1);
        final cardWidth = (constraints.maxWidth - totalSpacing) / columns;

        return Wrap(
          spacing: spacing,
          runSpacing: spacing,
          children: children.map((child) {
            return SizedBox(width: cardWidth, child: child);
          }).toList(),
        );
      },
    );
  }
}
