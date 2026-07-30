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
    context.read<EarningsProvider>().setSelectedApp(appId);
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

    if (provider.isLoading && provider.periods.isEmpty) {
      return LgPage(
        title: 'Earnings',
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final appsList = context.watch<AppsProvider>().apps;
    final showAppFilter = appsList.length > 1;

    return LgPage(
      title: 'Earnings',
      subtitle: 'Revenue and fee breakdown',
      onRefresh: refreshData,
      scrollable: false,
      child: DefaultTabController(
        length: 2,
        child: Column(
          children: [
            if (showAppFilter) ...[
              Align(
                alignment: Alignment.centerLeft,
                child: PopupMenuButton<String>(
                  onSelected: provider.setSelectedApp,
                  itemBuilder: (_) => [
                    ...appsList.map((app) => PopupMenuItem(
                          value: app.id,
                          child: Text(app.name),
                        )),
                  ],
                  child: Chip(
                    label: Text(appsList.firstWhere((a) => a.id == provider.selectedAppId, orElse: () => appsList.first).name),
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

          // Upcoming 30 days (from API)
          if (provider.earningsStatus != null &&
              provider.earningsStatus!.upcoming.isNotEmpty) ...[
            const SizedBox(height: LgSpacing.s600),
            LgCard(
              title: 'Upcoming 30 Days',
              child: Column(
                children: provider.earningsStatus!.upcoming.map((entry) {
                  return Padding(
                    padding: const EdgeInsets.symmetric(
                        vertical: LgSpacing.s100),
                    child: Row(
                      mainAxisAlignment:
                          MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          DateFormat('MMM d').format(entry.date),
                          style: theme.textTheme.bodyMedium,
                        ),
                        Text(
                          entry.amountFormatted,
                          style: const TextStyle(
                            fontWeight: FontWeight.w600,
                            color: LgColors.success,
                          ),
                        ),
                      ],
                    ),
                  );
                }).toList(),
              ),
            ),
          ],

          // Fee savings card
          if (provider.feeSummary != null &&
              provider.feeSummary!.savingsCents > 0) ...[
            const SizedBox(height: LgSpacing.s600),
            LgCard(
              title: 'Fee Savings',
              child: Row(
                children: [
                  Icon(Icons.savings, color: LgColors.success, size: 32),
                  const SizedBox(width: LgSpacing.s300),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'You saved ${provider.feeSummary!.savingsFormatted}',
                          style: theme.textTheme.titleSmall,
                        ),
                        const SizedBox(height: LgSpacing.s100),
                        Text(
                          'compared to the default 20% tier over ${provider.feeSummary!.transactionCount} transactions',
                          style: TextStyle(
                              fontSize: 13,
                              color: LgColors.textSecondary),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ],

          const SizedBox(height: LgSpacing.s600),
          ...periods.map((period) {
            // Live per-date rows have no month label; fall back to the date.
            final title = period.month.isNotEmpty
                ? period.month
                : dateFmt.format(period.startDate);
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
                                child: Text(title,
                                    style: theme.textTheme.titleSmall),
                              ),
                              // Per-date status isn't provided for live rows; only
                              // badge the rich (demo/breakdown) source to avoid
                              // labelling every historical row "PENDING".
                              if (period.hasFeeBreakdown)
                                LgBadge(
                                  label: period.statusLabel.toUpperCase(),
                                  tone: _statusTone(period.status),
                                )
                              else
                                Text('Net: ${period.netFormatted}',
                                    style: const TextStyle(
                                        fontSize: 14,
                                        fontWeight: FontWeight.w600,
                                        color: LgColors.success)),
                            ],
                          ),
                          // Full gross/fee breakdown only when the source has it
                          // (monthly period cards). Matches wireframe 14-earnings:
                          // stacked label/value columns — Gross, Shopify Fee (rate%)
                          // in red as a negative, Net in green.
                          if (period.hasFeeBreakdown) ...[
                            const SizedBox(height: LgSpacing.s300),
                            _PeriodBreakdown(period: period),
                          ],
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

/// The Gross / Shopify Fee (rate%) / Net breakdown for one earnings period card,
/// per wireframe 14-earnings. Row on desktop, stacked on mobile.
class _PeriodBreakdown extends StatelessWidget {
  final EarningPeriod period;
  const _PeriodBreakdown({required this.period});

  @override
  Widget build(BuildContext context) {
    final cut = period.shopifyCutCents;
    // Guard the derived rate: clamp to [0,100] so a legacy/refund month with an
    // odd gross/net ratio can't render e.g. "-12%" or "140%".
    final rate = period.grossCents > 0
        ? ((cut * 100 / period.grossCents).round()).clamp(0, 100)
        : 0;
    // Only show the fee as a negative when there actually is a cut; never emit a
    // double-minus for a (shouldn't-happen) net > gross month.
    final feeValue =
        cut > 0 ? '-${period.shopifyCutFormatted}' : period.shopifyCutFormatted;
    final cells = <Widget>[
      _EarningMetric(label: 'Gross', value: period.grossFormatted),
      _EarningMetric(
          label: 'Shopify Fee ($rate%)',
          value: feeValue,
          color: LgColors.critical),
      _EarningMetric(
          label: 'Net', value: period.netFormatted, color: LgColors.success),
    ];
    if (LgBreakpoints.isMobile(context)) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          for (final c in cells)
            Padding(
                padding: const EdgeInsets.only(bottom: LgSpacing.s200),
                child: c),
        ],
      );
    }
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [for (final c in cells) Expanded(child: c)],
    );
  }
}

/// A stacked label-over-value cell for the earnings period card (Gross / Shopify
/// Fee / Net), per wireframe 14-earnings.
class _EarningMetric extends StatelessWidget {
  final String label;
  final String value;
  final Color? color;
  const _EarningMetric({required this.label, required this.value, this.color});

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label,
            style: const TextStyle(fontSize: 12, color: LgColors.textSecondary)),
        const SizedBox(height: LgSpacing.s100),
        Text(value,
            style: TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w600,
                color: color ?? LgColors.textPrimary)),
      ],
    );
  }
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
      final amountCents = (gross * 100).round();
      if (!provider.demoMode && provider.selectedAppId != null) {
        provider.loadFeeBreakdown(provider.selectedAppId!, amountCents);
      }
      setState(() {
        _breakdown = provider.calculateFees(amountCents);
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
    final feeSummary = provider.feeSummary;
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

          // Fee savings (from API)
          if (feeSummary != null && feeSummary.savingsCents > 0) ...[
            LgCard(
              title: 'Fee Savings',
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                      'You saved ${feeSummary.savingsFormatted} compared to the default 20% tier',
                      style: theme.textTheme.bodyMedium),
                  const SizedBox(height: LgSpacing.s200),
                  Text(
                      '${feeSummary.transactionCount} transactions processed',
                      style: TextStyle(
                          fontSize: 13, color: LgColors.textSecondary)),
                ],
              ),
            ),
            const SizedBox(height: LgSpacing.s600),
          ],

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
                // Show API breakdown comparison if available
                if (provider.feeBreakdownResponse != null) ...[
                  const SizedBox(height: LgSpacing.s400),
                  Text('Breakdown by tier:',
                      style: theme.textTheme.titleSmall),
                  const SizedBox(height: LgSpacing.s200),
                  ...provider.feeBreakdownResponse!.tiers.map((tb) =>
                      _feeRow(
                          '${tb.tierName} (${tb.ratePct.toStringAsFixed(0)}%)',
                          '\$${(tb.netCents / 100).toStringAsFixed(2)} net',
                          theme)),
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
