import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../providers/analytics_provider.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';

class MultiAppTab extends StatelessWidget {
  const MultiAppTab({super.key});

  @override
  Widget build(BuildContext context) {
    final provider = context.watch<AnalyticsProvider>();
    final apps = provider.apps;
    final mrrMap = provider.appMrrCents();
    final subMap = provider.appSubCount();
    final theme = Theme.of(context);

    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          LgCard(
            title: 'App Comparison',
            child: LgBreakpoints.isMobile(context)
                ? SingleChildScrollView(
                    scrollDirection: Axis.horizontal,
                    child: _buildComparisonTable(apps, mrrMap, subMap, theme),
                  )
                : _buildComparisonTable(apps, mrrMap, subMap, theme),
          ),
          const SizedBox(height: LgSpacing.s600),

          // Per-app MRR cards
          LayoutBuilder(
            builder: (context, constraints) {
              final cols = LgBreakpoints.isMobile(context) ? 1 : apps.length;
              final spacing = LgSpacing.s400;
              final totalSpacing = spacing * (cols - 1);
              final cardWidth = cols == 1
                  ? constraints.maxWidth
                  : (constraints.maxWidth - totalSpacing) / cols;

              return Wrap(
                spacing: spacing,
                runSpacing: spacing,
                children: apps.map((app) {
                  final mrr = (mrrMap[app.id] ?? 0) / 100;
                  return SizedBox(
                    width: cardWidth,
                    child: LgCard(
                      title: app.name,
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('\$${mrr.toStringAsFixed(0)}/mo', style: theme.textTheme.headlineSmall),
                          const SizedBox(height: LgSpacing.s200),
                          Row(
                            children: [
                              const Icon(Icons.people, size: 14, color: LgColors.textSecondary),
                              const SizedBox(width: 4),
                              Text('${app.installCount} installs', style: theme.textTheme.bodySmall),
                            ],
                          ),
                          const SizedBox(height: LgSpacing.s100),
                          Row(
                            children: [
                              const Icon(Icons.star, size: 14, color: LgColors.warning),
                              const SizedBox(width: 4),
                              Text('${app.avgRating} rating', style: theme.textTheme.bodySmall),
                            ],
                          ),
                        ],
                      ),
                    ),
                  );
                }).toList(),
              );
            },
          ),
        ],
      ),
    );
  }

  Widget _buildComparisonTable(List apps, Map mrrMap, Map subMap, ThemeData theme) {
    return Table(
      columnWidths: const {
        0: FlexColumnWidth(2),
        1: FlexColumnWidth(1),
        2: FlexColumnWidth(1),
        3: FlexColumnWidth(1),
        4: FlexColumnWidth(1),
      },
      children: [
        TableRow(
          decoration: BoxDecoration(color: LgColors.surfaceSecondary),
          children: ['App', 'Installs', 'Subs', 'MRR', 'Avg Rating']
              .map((h) => Padding(
                    padding: const EdgeInsets.all(8),
                    child: Text(h, style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: LgColors.textSecondary)),
                  ))
              .toList(),
        ),
        ...apps.map((app) {
          final mrr = ((mrrMap[app.id] ?? 0) as int) / 100;
          final subs = subMap[app.id] ?? 0;
          return TableRow(
            children: [
              Padding(padding: const EdgeInsets.all(8), child: Text(app.name, style: theme.textTheme.bodyMedium)),
              _Cell('${app.installCount}'),
              _Cell('$subs'),
              _Cell('\$${mrr.toStringAsFixed(0)}'),
              Padding(
                padding: const EdgeInsets.all(8),
                child: Row(
                  children: [
                    const Icon(Icons.star, size: 14, color: LgColors.warning),
                    const SizedBox(width: 2),
                    Text(app.avgRating.toStringAsFixed(1), style: const TextStyle(fontSize: 13)),
                  ],
                ),
              ),
            ],
          );
        }),
      ],
    );
  }
}

class _Cell extends StatelessWidget {
  final String text;
  const _Cell(this.text);
  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.all(8),
        child: Text(text, style: const TextStyle(fontSize: 13)),
      );
}
