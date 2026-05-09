import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../models/analytics_model.dart';
import '../../providers/analytics_provider.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';

class MultiAppTab extends StatelessWidget {
  const MultiAppTab({super.key});

  @override
  Widget build(BuildContext context) {
    final provider = context.watch<AnalyticsProvider>();
    final apps = provider.apps;
    final theme = Theme.of(context);

    if (apps.isEmpty) {
      return const LgEmptyState(
        icon: Icons.apps,
        heading: 'Multi-app comparison not yet available',
        description:
            'Connect multiple Shopify apps to compare performance across your portfolio.',
      );
    }

    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          LgCard(
            title: 'App Comparison',
            child: LgBreakpoints.isMobile(context)
                ? SingleChildScrollView(
                    scrollDirection: Axis.horizontal,
                    child: ConstrainedBox(
                      constraints: const BoxConstraints(minWidth: 500),
                      child: _buildComparisonTable(apps, theme),
                    ),
                  )
                : _buildComparisonTable(apps, theme),
          ),
          const SizedBox(height: LgSpacing.s600),

          // Per-app MRR cards
          LayoutBuilder(
            builder: (context, constraints) {
              final cols = LgBreakpoints.isMobile(context) ? 1 : apps.length.clamp(1, 4);
              final spacing = LgSpacing.s400;
              final totalSpacing = spacing * (cols - 1);
              final cardWidth = cols == 1
                  ? constraints.maxWidth
                  : (constraints.maxWidth - totalSpacing) / cols;

              return Wrap(
                spacing: spacing,
                runSpacing: spacing,
                children: apps.map((app) {
                  return SizedBox(
                    width: cardWidth,
                    child: LgCard(
                      title: app.name,
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('\$${app.mrrDollars.toStringAsFixed(0)}/mo', style: theme.textTheme.headlineSmall),
                          const SizedBox(height: LgSpacing.s200),
                          Row(
                            children: [
                              const Icon(Icons.subscriptions, size: 14, color: LgColors.textSecondary),
                              const SizedBox(width: 4),
                              Text('${app.subscriptionCount} subscriptions', style: theme.textTheme.bodySmall),
                            ],
                          ),
                          if (app.renewalRate > 0) ...[
                            const SizedBox(height: LgSpacing.s100),
                            Row(
                              children: [
                                const Icon(Icons.autorenew, size: 14, color: LgColors.success),
                                const SizedBox(width: 4),
                                Text('${(app.renewalRate * 100).toStringAsFixed(0)}% renewal', style: theme.textTheme.bodySmall),
                              ],
                            ),
                          ],
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

  Widget _buildComparisonTable(List<AppComparison> apps, ThemeData theme) {
    return Table(
      columnWidths: const {
        0: FlexColumnWidth(2),
        1: FlexColumnWidth(1),
        2: FlexColumnWidth(1),
        3: FlexColumnWidth(1),
      },
      children: [
        TableRow(
          decoration: BoxDecoration(color: LgColors.surfaceSecondary),
          children: ['App', 'Subs', 'MRR', 'At Risk']
              .map((h) => Padding(
                    padding: const EdgeInsets.all(8),
                    child: Text(h, style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: LgColors.textSecondary)),
                  ))
              .toList(),
        ),
        ...apps.map((app) {
          return TableRow(
            children: [
              Padding(padding: const EdgeInsets.all(8), child: Text(app.name, style: theme.textTheme.bodyMedium)),
              _Cell('${app.subscriptionCount}'),
              _Cell('\$${app.mrrDollars.toStringAsFixed(0)}'),
              _Cell('\$${(app.atRiskCents / 100).toStringAsFixed(0)}'),
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
