import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import '../../mock_data/mock_apps.dart';
import '../../providers/apps_provider.dart';
import '../../providers/risk_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_risk_badge.dart';

class RiskScreen extends StatelessWidget {
  const RiskScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final hasApps = context.watch<AppsProvider>().apps.isNotEmpty;
    if (!hasApps) {
      return LgPage(
        title: 'Risk Breakdown',
        child: LgEmptyState(
          icon: Icons.shield_outlined,
          heading: 'No risk data yet',
          description:
              'Connect your Shopify app to see risk analysis.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<RiskProvider>();
    final dist = provider.distribution;
    final theme = Theme.of(context);
    final showAppFilter = mockApps.length > 1;

    return LgPage(
      title: 'Risk Breakdown',
      subtitle: 'Subscription risk funnel and recovery playbooks',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (showAppFilter) ...[
            PopupMenuButton<String?>(
              onSelected: provider.setSelectedApp,
              itemBuilder: (_) => [
                const PopupMenuItem(value: null, child: Text('All Apps')),
                ...mockApps.map((app) => PopupMenuItem(
                      value: app.id,
                      child: Text(app.name),
                    )),
              ],
              child: Chip(
                label: Text(provider.selectedAppId != null
                    ? mockApps.firstWhere((a) => a.id == provider.selectedAppId).name
                    : 'All Apps'),
                deleteIcon: provider.selectedAppId != null
                    ? const Icon(Icons.close, size: 14)
                    : null,
                onDeleted: provider.selectedAppId != null
                    ? () => provider.setSelectedApp(null)
                    : null,
              ),
            ),
            const SizedBox(height: LgSpacing.s300),
          ],

          // Risk funnel
          LgCard(
            title: 'Risk Funnel',
            child: Row(
              children: [
                _FunnelSegment('Safe', dist.safe, LgColors.riskSafe, flex: dist.safe),
                _FunnelSegment('1 Cycle', dist.oneCycle, LgColors.riskOneCycle, flex: dist.oneCycle),
                _FunnelSegment('2 Cycles', dist.twoCycle, LgColors.riskTwoCycle, flex: dist.twoCycle),
                _FunnelSegment('Churned', dist.churned, LgColors.riskChurned, flex: dist.churned),
              ],
            ),
          ),
          const SizedBox(height: LgSpacing.s600),

          // At-risk stores
          Text('At-Risk Stores (${provider.atRiskStores.length})', style: theme.textTheme.titleMedium),
          const SizedBox(height: LgSpacing.s300),
          if (provider.atRiskStores.isEmpty)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: LgSpacing.s600),
              child: Center(
                child: Column(
                  children: [
                    Icon(Icons.check_circle_outline, size: 40, color: LgColors.success),
                    const SizedBox(height: LgSpacing.s300),
                    Text('No stores at risk',
                        style: theme.textTheme.bodyMedium?.copyWith(color: LgColors.textSecondary)),
                  ],
                ),
              ),
            ),
          ...provider.atRiskStores.map((store) {
            return Padding(
              padding: const EdgeInsets.only(bottom: LgSpacing.s200),
              child: MouseRegion(
                cursor: SystemMouseCursors.click,
                child: LgCard(
                  child: InkWell(
                    onTap: () => context.go('/stores/${store.id}'),
                    child: Row(
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(store.shopDomain.replaceAll('.myshopify.com', ''),
                                style: theme.textTheme.titleSmall),
                            const SizedBox(height: LgSpacing.s100),
                            Text('Health: ${store.healthScore}% | LTV: ${store.ltvFormatted}',
                                style: theme.textTheme.bodySmall),
                          ],
                        ),
                      ),
                      LgRiskBadge(riskState: store.riskState),
                      const SizedBox(width: LgSpacing.s200),
                      const Icon(Icons.chevron_right, color: LgColors.textSecondary),
                    ],
                  ),
                ),
              ),
              ),
            );
          }),
          const SizedBox(height: LgSpacing.s600),

          // Recovery playbooks
          Text('Recovery Playbooks', style: theme.textTheme.titleMedium),
          const SizedBox(height: LgSpacing.s300),
          ...provider.playbooks.map((pb) {
            return Padding(
              padding: const EdgeInsets.only(bottom: LgSpacing.s300),
              child: LgCard(
                title: pb.name,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(pb.description, style: theme.textTheme.bodySmall),
                    const SizedBox(height: LgSpacing.s300),
                    Row(
                      children: [
                        Icon(Icons.check_circle, size: 16, color: LgColors.success),
                        const SizedBox(width: LgSpacing.s100),
                        Text('${(pb.successRate * 100).toStringAsFixed(0)}% success rate',
                            style: theme.textTheme.bodyMedium),
                      ],
                    ),
                    const SizedBox(height: LgSpacing.s300),
                    ...pb.steps.asMap().entries.map((entry) {
                      return Padding(
                        padding: const EdgeInsets.only(bottom: LgSpacing.s200),
                        child: Row(
                          children: [
                            CircleAvatar(
                              radius: 10,
                              backgroundColor: LgColors.surfaceSecondary,
                              child: Text('${entry.key + 1}',
                                  style: const TextStyle(fontSize: 10, color: LgColors.textSecondary)),
                            ),
                            const SizedBox(width: LgSpacing.s200),
                            Expanded(child: Text(entry.value.label, style: theme.textTheme.bodyMedium)),
                          ],
                        ),
                      );
                    }),
                  ],
                ),
              ),
            );
          }),
        ],
      ),
    );
  }
}

class _FunnelSegment extends StatelessWidget {
  final String label;
  final int count;
  final Color color;
  final int flex;
  const _FunnelSegment(this.label, this.count, this.color, {required this.flex});

  @override
  Widget build(BuildContext context) {
    return Expanded(
      flex: flex.clamp(1, 100),
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: LgSpacing.s400),
        color: color.withValues(alpha: 0.15),
        child: Column(
          children: [
            Text('$count', style: TextStyle(fontSize: 20, fontWeight: FontWeight.w700, color: color)),
            const SizedBox(height: 4),
            Text(label, style: TextStyle(fontSize: 11, color: color, fontWeight: FontWeight.w500)),
          ],
        ),
      ),
    );
  }
}
