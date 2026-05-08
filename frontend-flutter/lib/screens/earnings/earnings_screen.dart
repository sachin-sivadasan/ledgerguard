import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../models/earning_model.dart';
import '../../providers/apps_provider.dart';
import '../../providers/earnings_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_badge.dart';
import '../../theme/app_breakpoints.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_metric_card.dart';
import '../../widgets/lg_metric_grid.dart';
import '../../widgets/lg_page.dart';

class EarningsScreen extends StatefulWidget {
  const EarningsScreen({super.key});

  @override
  State<EarningsScreen> createState() => _EarningsScreenState();
}

class _EarningsScreenState extends State<EarningsScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<EarningsProvider>().loadEarnings(appId);
  }

  @override
  Widget build(BuildContext context) {
    final appsProvider = context.watch<AppsProvider>();
    final hasApps = appsProvider.apps.isNotEmpty;

    if (!hasApps) {
      return LgPage(
        title: 'Earnings',
        child: LgEmptyState(
          icon: Icons.account_balance_wallet_outlined,
          heading: 'No earnings yet',
          description:
              'Connect your Shopify app to track earnings.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<EarningsProvider>();

    if (provider.error != null) {
      return LgPage(
        title: 'Earnings',
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    final appsList = context.watch<AppsProvider>().apps;
    final showAppFilter = appsList.length > 1;

    return LgPage(
      title: 'Earnings',
      subtitle: 'Revenue and fee breakdown',
      scrollable: false,
      child: DefaultTabController(
        length: 2,
        child: Column(
          children: [
            if (showAppFilter) ...[
              Align(
                alignment: Alignment.centerLeft,
                child: PopupMenuButton<String?>(
                  onSelected: provider.setSelectedApp,
                  itemBuilder: (_) => [
                    const PopupMenuItem(value: null, child: Text('All Apps')),
                    ...appsList.map((app) => PopupMenuItem(
                          value: app.id,
                          child: Text(app.name),
                        )),
                  ],
                  child: Chip(
                    label: Text(provider.selectedAppId != null
                        ? appsList.firstWhere((a) => a.id == provider.selectedAppId, orElse: () => appsList.first).name
                        : 'All Apps'),
                    deleteIcon: provider.selectedAppId != null
                        ? const Icon(Icons.close, size: 14)
                        : null,
                    onDeleted: provider.selectedAppId != null
                        ? () => provider.setSelectedApp(null)
                        : null,
                  ),
                ),
              ),
              const SizedBox(height: LgSpacing.s300),
            ],
            const TabBar(
              isScrollable: true,
              tabAlignment: TabAlignment.start,
              padding: EdgeInsets.zero,
              tabs: [
                Tab(text: 'Earnings'),
                Tab(text: 'Fees & Tiers'),
              ],
            ),
            const SizedBox(height: LgSpacing.s400),
            Expanded(
              child: TabBarView(
                children: [
                  _EarningsTab(),
                  _FeesAndTiersTab(),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _EarningsTab extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    final provider = context.watch<EarningsProvider>();
    final periods = provider.periods;
    final dateFmt = DateFormat('MMM d, y');
    final theme = Theme.of(context);

    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          LgMetricGrid(
            children: [
              LgMetricCard(label: 'Total Earned', value: provider.totalEarned, icon: Icons.account_balance_wallet),
              LgMetricCard(label: 'Pending', value: provider.pendingAmount, icon: Icons.hourglass_empty),
              LgMetricCard(label: 'Available', value: provider.availableAmount, icon: Icons.check_circle_outline),
            ],
          ),
          const SizedBox(height: LgSpacing.s600),
          ...periods.map((period) {
            return Padding(
              padding: const EdgeInsets.only(bottom: LgSpacing.s300),
              child: LgCard(
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              Expanded(
                                child: Text(period.month,
                                    style: theme.textTheme.titleSmall),
                              ),
                              LgBadge(
                                label: period.statusLabel.toUpperCase(),
                                tone: _statusTone(period.status),
                              ),
                            ],
                          ),
                          const SizedBox(height: LgSpacing.s200),
                          LgBreakpoints.isMobile(context)
                              ? Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text('Gross: ${period.grossFormatted}', style: theme.textTheme.bodyMedium),
                                    const SizedBox(height: LgSpacing.s100),
                                    Text('Shopify: ${period.shopifyCutFormatted}', style: TextStyle(fontSize: 13, color: LgColors.textSecondary)),
                                    const SizedBox(height: LgSpacing.s100),
                                    Text('Net: ${period.netFormatted}', style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: LgColors.success)),
                                  ],
                                )
                              : Row(
                                  children: [
                                    Text('Gross: ${period.grossFormatted}', style: theme.textTheme.bodyMedium),
                                    const SizedBox(width: LgSpacing.s400),
                                    Text('Shopify: ${period.shopifyCutFormatted}', style: TextStyle(fontSize: 13, color: LgColors.textSecondary)),
                                    const SizedBox(width: LgSpacing.s400),
                                    Text('Net: ${period.netFormatted}', style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: LgColors.success)),
                                  ],
                                ),
                          if (period.paidOutDate != null) ...[
                            const SizedBox(height: LgSpacing.s100),
                            Text(
                              'Paid out ${dateFmt.format(period.paidOutDate!)}',
                              style: TextStyle(
                                  fontSize: 12,
                                  color: LgColors.textSecondary),
                            ),
                          ],
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            );
          }),
        ],
      ),
    );
  }

  BadgeTone _statusTone(EarningStatus status) => switch (status) {
        EarningStatus.pending => BadgeTone.warning,
        EarningStatus.available => BadgeTone.info,
        EarningStatus.paidOut => BadgeTone.success,
      };
}

class _FeesAndTiersTab extends StatefulWidget {
  @override
  State<_FeesAndTiersTab> createState() => _FeesAndTiersTabState();
}

class _FeesAndTiersTabState extends State<_FeesAndTiersTab> {
  final _grossController = TextEditingController(text: '1000');
  FeeBreakdown? _breakdown;

  @override
  void initState() {
    super.initState();
    _calculate();
  }

  void _calculate() {
    final gross = double.tryParse(_grossController.text);
    if (gross != null && gross > 0) {
      final provider = context.read<EarningsProvider>();
      setState(() {
        _breakdown = provider.calculateFees((gross * 100).round());
      });
    }
  }

  @override
  void dispose() {
    _grossController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final provider = context.watch<EarningsProvider>();
    final currentTier = provider.currentTier;
    final theme = Theme.of(context);

    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Current tier
          LgCard(
            title: 'Current Tier',
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text(currentTier.name,
                        style: theme.textTheme.titleMedium),
                    const SizedBox(width: LgSpacing.s300),
                    LgBadge(
                        label: currentTier.rateLabel,
                        tone: BadgeTone.success),
                  ],
                ),
                const SizedBox(height: LgSpacing.s200),
                Text(currentTier.description,
                    style: theme.textTheme.bodyMedium),
              ],
            ),
          ),
          const SizedBox(height: LgSpacing.s600),

          // Fee calculator
          LgCard(
            title: 'Fee Calculator',
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    SizedBox(
                      width: 200,
                      child: TextField(
                        controller: _grossController,
                        decoration: const InputDecoration(
                          labelText: 'Gross Amount (\$)',
                          prefixText: '\$ ',
                        ),
                        keyboardType: TextInputType.number,
                        onChanged: (_) => _calculate(),
                      ),
                    ),
                  ],
                ),
                if (_breakdown != null) ...[
                  const SizedBox(height: LgSpacing.s400),
                  _feeRow('Gross Revenue', _breakdown!.grossFormatted,
                      theme),
                  _feeRow(
                      'Shopify Fee (${_breakdown!.shopifyFeePct.toStringAsFixed(0)}%)',
                      '- ${_breakdown!.shopifyFeeFormatted}',
                      theme),
                  _feeRow(
                      'Processing Fee (${_breakdown!.processingFeePct.toStringAsFixed(1)}%)',
                      '- ${_breakdown!.processingFeeFormatted}',
                      theme),
                  const Divider(),
                  _feeRow('Net Earnings', _breakdown!.netFormatted, theme,
                      bold: true),
                ],
              ],
            ),
          ),
          const SizedBox(height: LgSpacing.s600),

          // All tiers
          LgCard(
            title: 'Revenue Share Tiers',
            child: Column(
              children: provider.tiers.map((tier) {
                return Padding(
                  padding:
                      const EdgeInsets.only(bottom: LgSpacing.s300),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      SizedBox(
                        width: 140,
                        child: Text(tier.name,
                            style: theme.textTheme.titleSmall),
                      ),
                      const SizedBox(width: LgSpacing.s300),
                      LgBadge(
                          label: tier.rateLabel,
                          tone: tier.isCurrentTier
                              ? BadgeTone.success
                              : BadgeTone.defaultTone),
                      const SizedBox(width: LgSpacing.s300),
                      Expanded(
                        child: Text(tier.description,
                            style: TextStyle(
                                fontSize: 13,
                                color: LgColors.textSecondary)),
                      ),
                      if (tier.isCurrentTier)
                        const LgBadge(
                            label: 'CURRENT', tone: BadgeTone.newTone),
                    ],
                  ),
                );
              }).toList(),
            ),
          ),
        ],
      ),
    );
  }

  Widget _feeRow(String label, String value, ThemeData theme,
      {bool bold = false}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: LgSpacing.s100),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label,
              style: bold
                  ? theme.textTheme.titleSmall
                  : theme.textTheme.bodyMedium),
          Text(value,
              style: bold
                  ? TextStyle(
                      fontWeight: FontWeight.w600, color: LgColors.success)
                  : theme.textTheme.bodyMedium),
        ],
      ),
    );
  }
}
