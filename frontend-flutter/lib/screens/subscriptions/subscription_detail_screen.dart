import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../models/payment_history_model.dart';
import '../../models/transaction_model.dart';
import '../../providers/subscription_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_badge.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_metric_card.dart';
import '../../widgets/lg_risk_badge.dart';
import '../../widgets/lg_status_badge.dart';

class SubscriptionDetailScreen extends StatelessWidget {
  final String subscriptionId;
  const SubscriptionDetailScreen({super.key, required this.subscriptionId});

  @override
  Widget build(BuildContext context) {
    final provider = context.watch<SubscriptionProvider>();
    final sub = provider.getById(subscriptionId);
    final dateFmt = DateFormat('MMM d, y');
    final theme = Theme.of(context);

    if (sub == null) {
      return Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 1200),
          child: const Padding(
            padding: EdgeInsets.all(LgSpacing.s600),
            child: Center(child: Text('Subscription not found')),
          ),
        ),
      );
    }

    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 1200),
        child: Padding(
          padding: const EdgeInsets.all(LgSpacing.s600),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Header
              Row(
                children: [
                  IconButton(
                    icon: const Icon(Icons.arrow_back),
                    onPressed: () => context.go('/subscriptions'),
                    tooltip: 'Back',
                  ),
                  const SizedBox(width: LgSpacing.s200),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        sub.shopDomain.replaceAll('.myshopify.com', ''),
                        style: theme.textTheme.headlineSmall,
                      ),
                      const SizedBox(height: LgSpacing.s100),
                      Text('Subscription detail',
                          style: theme.textTheme.bodySmall),
                    ],
                  ),
                ],
              ),
              const SizedBox(height: LgSpacing.s600),

              // Summary cards — responsive layout
              LayoutBuilder(
                builder: (context, constraints) {
                  final billingCard = LgCard(
                    title: 'Billing Info',
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        _InfoRow('Plan', sub.planName),
                        _InfoRow('Price', sub.priceFormatted),
                        _InfoRow('MRR', sub.mrrFormatted),
                        _InfoRow('Interval', sub.billingInterval.name),
                        _InfoRow('Created', dateFmt.format(sub.createdAt)),
                      ],
                    ),
                  );
                  final statusCard = LgCard(
                    title: 'Status',
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        _InfoRow('Status', '',
                            badge: LgStatusBadge(status: sub.status)),
                        _InfoRow('Risk', '',
                            badge: LgRiskBadge(riskState: sub.riskState)),
                        _InfoRow('Period End', dateFmt.format(sub.periodEnd)),
                        _InfoRow('Next Charge',
                            dateFmt.format(sub.expectedNextCharge)),
                      ],
                    ),
                  );
                  final storeCard = LgCard(
                    title: 'Store',
                    child: InkWell(
                      onTap: () => context.go('/stores'),
                      child: Row(
                        children: [
                          const Icon(Icons.store,
                              color: LgColors.textSecondary),
                          const SizedBox(width: LgSpacing.s200),
                          Expanded(
                            child: Text(
                              sub.shopDomain,
                              style: theme.textTheme.bodyLarge,
                              overflow: TextOverflow.ellipsis,
                            ),
                          ),
                          const Icon(Icons.chevron_right,
                              color: LgColors.textSecondary),
                        ],
                      ),
                    ),
                  );

                  if (constraints.maxWidth > 900) {
                    return IntrinsicHeight(
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          Expanded(child: billingCard),
                          const SizedBox(width: LgSpacing.s400),
                          Expanded(child: statusCard),
                          const SizedBox(width: LgSpacing.s400),
                          Expanded(child: storeCard),
                        ],
                      ),
                    );
                  }
                  return Column(
                    children: [
                      billingCard,
                      const SizedBox(height: LgSpacing.s300),
                      statusCard,
                      const SizedBox(height: LgSpacing.s300),
                      storeCard,
                    ],
                  );
                },
              ),
              const SizedBox(height: LgSpacing.s600),

              // Tabs: Payment History | Risk Timeline
              Expanded(
                child: DefaultTabController(
                  length: 2,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const TabBar(
                        isScrollable: true,
                        tabAlignment: TabAlignment.start,
                        padding: EdgeInsets.zero,
                        tabs: [
                          Tab(text: 'Payment History'),
                          Tab(text: 'Risk Timeline'),
                        ],
                      ),
                      const SizedBox(height: LgSpacing.s400),
                      Expanded(
                        child: TabBarView(
                          children: [
                            _PaymentHistoryTab(
                                subscriptionId: subscriptionId),
                            _RiskTimelineTab(
                                subscriptionId: subscriptionId),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

// ─── Payment History Tab ────────────────────────────────────────────

class _PaymentHistoryTab extends StatelessWidget {
  final String subscriptionId;
  const _PaymentHistoryTab({required this.subscriptionId});

  @override
  Widget build(BuildContext context) {
    final provider = context.read<SubscriptionProvider>();
    final history = provider.getPaymentHistory(subscriptionId);
    final dateFmt = DateFormat('MMM d, y');
    final theme = Theme.of(context);

    if (history.isEmpty) {
      return const Center(child: Text('No payment history'));
    }

    final totalGross = history
        .where((e) => e.grossAmountCents > 0)
        .fold<int>(0, (sum, e) => sum + e.grossAmountCents);
    final totalNet = history
        .where((e) => e.netAmountCents > 0)
        .fold<int>(0, (sum, e) => sum + e.netAmountCents);
    final paymentCount =
        history.where((e) => e.chargeType == ChargeType.recurring).length;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Metric cards
        Row(
          children: [
            LgMetricCard(
              label: 'Payments',
              value: '$paymentCount',
              icon: Icons.autorenew,
            ),
            const SizedBox(width: LgSpacing.s400),
            LgMetricCard(
              label: 'Total Gross',
              value: '\$${(totalGross / 100).toStringAsFixed(2)}',
              icon: Icons.payments,
            ),
            const SizedBox(width: LgSpacing.s400),
            LgMetricCard(
              label: 'Total Net',
              value: '\$${(totalNet / 100).toStringAsFixed(2)}',
              icon: Icons.account_balance,
            ),
          ],
        ),
        const SizedBox(height: LgSpacing.s400),

        // Scrollable list
        Expanded(
          child: ListView.builder(
            itemCount: history.length,
            itemBuilder: (context, index) {
              final entry = history[index];
              final isRefund = entry.chargeType == ChargeType.refund;
              return Padding(
                padding: const EdgeInsets.only(bottom: LgSpacing.s200),
                child: LgCard(
                  padding: const EdgeInsets.symmetric(
                      horizontal: LgSpacing.s400,
                      vertical: LgSpacing.s300),
                  child: Row(
                    children: [
                      Icon(Icons.circle,
                          size: 8,
                          color: isRefund
                              ? LgColors.critical
                              : LgColors.success),
                      const SizedBox(width: LgSpacing.s300),
                      SizedBox(
                        width: 120,
                        child: Text(
                            dateFmt.format(entry.transactionDate),
                            style: theme.textTheme.bodyMedium),
                      ),
                      SizedBox(
                        width: 100,
                        child: LgBadge(
                          label: entry.chargeTypeLabel,
                          tone: isRefund
                              ? BadgeTone.critical
                              : BadgeTone.defaultTone,
                        ),
                      ),
                      const Spacer(),
                      SizedBox(
                        width: 90,
                        child: Text(
                          entry.grossFormatted,
                          textAlign: TextAlign.right,
                          style: TextStyle(
                            fontWeight: FontWeight.w600,
                            color: isRefund
                                ? LgColors.critical
                                : LgColors.textPrimary,
                          ),
                        ),
                      ),
                      const SizedBox(width: LgSpacing.s300),
                      SizedBox(
                        width: 80,
                        child: Text(
                          entry.netFormatted,
                          textAlign: TextAlign.right,
                          style: TextStyle(
                              fontSize: 12,
                              color: LgColors.textSecondary),
                        ),
                      ),
                      const SizedBox(width: LgSpacing.s400),
                      SizedBox(
                        width: 80,
                        child: Align(
                          alignment: Alignment.centerRight,
                          child: LgBadge(
                            label: entry.earningsStatusLabel
                                .toUpperCase(),
                            tone: _earningsTone(entry.earningsStatus),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              );
            },
          ),
        ),
      ],
    );
  }

  BadgeTone _earningsTone(EarningsStatus status) => switch (status) {
        EarningsStatus.pending => BadgeTone.warning,
        EarningsStatus.available => BadgeTone.info,
        EarningsStatus.paidOut => BadgeTone.success,
      };
}

// ─── Risk Timeline Tab ──────────────────────────────────────────────

class _RiskTimelineTab extends StatelessWidget {
  final String subscriptionId;
  const _RiskTimelineTab({required this.subscriptionId});

  @override
  Widget build(BuildContext context) {
    final provider = context.read<SubscriptionProvider>();
    final timeline = provider.getRiskTimeline(subscriptionId);
    final dateFmt = DateFormat('MMM d, y – h:mm a');
    final theme = Theme.of(context);

    if (timeline.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.check_circle_outline,
                size: 48, color: LgColors.success),
            const SizedBox(height: LgSpacing.s300),
            Text('No risk state changes',
                style: theme.textTheme.titleSmall),
            const SizedBox(height: LgSpacing.s100),
            Text('This subscription has remained safe',
                style: TextStyle(color: LgColors.textSecondary)),
          ],
        ),
      );
    }

    // KPI summary
    final escalations = timeline.where((e) => e.isEscalation).length;
    final recoveries = timeline.where((e) => e.isRecovery).length;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Metric cards
        Row(
          children: [
            LgMetricCard(
              label: 'State Changes',
              value: '${timeline.length}',
              icon: Icons.timeline,
            ),
            const SizedBox(width: LgSpacing.s400),
            LgMetricCard(
              label: 'Escalations',
              value: '$escalations',
              icon: Icons.trending_up,
            ),
            const SizedBox(width: LgSpacing.s400),
            LgMetricCard(
              label: 'Recoveries',
              value: '$recoveries',
              icon: Icons.trending_down,
            ),
          ],
        ),
        const SizedBox(height: LgSpacing.s400),

        // Timeline list
        Expanded(
          child: ListView.builder(
            itemCount: timeline.length,
            itemBuilder: (context, index) {
              final entry = timeline[index];
              return Padding(
                padding: const EdgeInsets.only(bottom: LgSpacing.s200),
                child: LgCard(
                  padding: const EdgeInsets.symmetric(
                      horizontal: LgSpacing.s400,
                      vertical: LgSpacing.s300),
                  child: Row(
                    children: [
                      Icon(
                        entry.isEscalation
                            ? Icons.arrow_upward
                            : Icons.arrow_downward,
                        size: 16,
                        color: entry.isEscalation
                            ? LgColors.critical
                            : LgColors.success,
                      ),
                      const SizedBox(width: LgSpacing.s300),
                      // From → To badges
                      LgRiskBadge(riskState: entry.fromRiskState),
                      Padding(
                        padding: const EdgeInsets.symmetric(
                            horizontal: LgSpacing.s200),
                        child: Icon(Icons.arrow_forward,
                            size: 14, color: LgColors.textSecondary),
                      ),
                      LgRiskBadge(riskState: entry.toRiskState),
                      const SizedBox(width: LgSpacing.s400),
                      // Reason
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(entry.reason,
                                style: theme.textTheme.bodyMedium),
                            const SizedBox(height: 2),
                            Row(
                              children: [
                                LgBadge(
                                  label: entry.eventTypeLabel,
                                  tone: BadgeTone.defaultTone,
                                ),
                                const SizedBox(width: LgSpacing.s300),
                                Text(
                                  dateFmt.format(entry.occurredAt),
                                  style: TextStyle(
                                      fontSize: 11,
                                      color: LgColors.textSecondary),
                                ),
                              ],
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              );
            },
          ),
        ),
      ],
    );
  }
}

// ─── Shared Widgets ─────────────────────────────────────────────────

class _InfoRow extends StatelessWidget {
  final String label;
  final String value;
  final Widget? badge;
  const _InfoRow(this.label, this.value, {this.badge});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: LgSpacing.s200),
      child: Row(
        children: [
          SizedBox(
            width: 100,
            child: Text(label, style: Theme.of(context).textTheme.bodySmall),
          ),
          if (badge != null)
            badge!
          else
            Text(value, style: Theme.of(context).textTheme.bodyMedium),
        ],
      ),
    );
  }
}
