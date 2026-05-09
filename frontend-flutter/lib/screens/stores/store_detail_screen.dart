import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../models/event_model.dart';
import '../../models/subscription_model.dart';
import '../../models/timeline_event.dart';
import '../../providers/events_provider.dart';
import '../../providers/store_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../theme/app_breakpoints.dart';
import '../../widgets/lg_badge.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_risk_badge.dart';
import '../../widgets/lg_status_badge.dart';

class StoreDetailScreen extends StatelessWidget {
  final String storeId;
  const StoreDetailScreen({super.key, required this.storeId});

  @override
  Widget build(BuildContext context) {
    final provider = context.watch<StoreProvider>();
    final store = provider.getById(storeId);
    final theme = Theme.of(context);
    final dateFmt = DateFormat('MMM d, y');

    if (store == null) {
      return LgPage(
        title: 'Store',
        backAction: () => context.go('/stores'),
        child: const Center(child: Text('Store not found')),
      );
    }

    final subscriptions = provider.getSubscriptionsForStore(store.shopDomain);

    return LgPage(
      title: store.shopDomain.replaceAll('.myshopify.com', ''),
      subtitle: 'Store CRM',
      backAction: () => context.go('/stores'),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Overview cards — responsive layout
          LayoutBuilder(
            builder: (context, constraints) {
              final overviewCard = LgCard(
                title: 'Overview',
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    _Info('Health Score', '${store.healthScore}%'),
                    _Info('Lifetime Value', store.ltvFormatted),
                    _Info('First Install', dateFmt.format(store.firstInstallDate)),
                    _Info('Last Interaction', dateFmt.format(store.lastInteraction)),
                    Row(
                      children: [
                        Text('Risk State: ', style: theme.textTheme.bodySmall),
                        LgRiskBadge(riskState: store.riskState),
                      ],
                    ),
                  ],
                ),
              );
              final appsCard = LgCard(
                title: 'Installed Apps',
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: store.installedAppIds.map((appId) {
                    final name = switch (appId) {
                      'app-1' => 'InventorySync Pro',
                      'app-2' => 'ReviewBoost',
                      'app-3' => 'ShipTracker',
                      _ => appId,
                    };
                    return Padding(
                      padding: const EdgeInsets.only(bottom: LgSpacing.s200),
                      child: Row(
                        children: [
                          const Icon(Icons.apps, size: 16, color: LgColors.primary),
                          const SizedBox(width: LgSpacing.s200),
                          Text(name, style: theme.textTheme.bodyMedium),
                        ],
                      ),
                    );
                  }).toList(),
                ),
              );
              final subsCard = LgCard(
                title: 'Subscriptions (${subscriptions.length})',
                child: subscriptions.isEmpty
                    ? Text('No subscriptions', style: theme.textTheme.bodySmall)
                    : Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: subscriptions.map((sub) {
                          return _SubscriptionRow(subscription: sub);
                        }).toList(),
                      ),
              );

              if (constraints.maxWidth > 900) {
                return IntrinsicHeight(
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      Expanded(child: overviewCard),
                      const SizedBox(width: LgSpacing.s400),
                      Expanded(child: appsCard),
                      const SizedBox(width: LgSpacing.s400),
                      Expanded(child: subsCard),
                    ],
                  ),
                );
              }
              return Column(
                children: [
                  overviewCard,
                  const SizedBox(height: LgSpacing.s300),
                  appsCard,
                  const SizedBox(height: LgSpacing.s300),
                  subsCard,
                ],
              );
            },
          ),
          const SizedBox(height: LgSpacing.s600),

          // Timeline (prefer events system, fall back to legacy timeline)
          _StoreTimeline(storeDomain: store.shopDomain, legacyTimeline: store.timeline),
        ],
      ),
    );
  }

}

class _SubscriptionRow extends StatelessWidget {
  final Subscription subscription;
  const _SubscriptionRow({required this.subscription});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.only(bottom: LgSpacing.s300),
      child: InkWell(
        borderRadius: BorderRadius.circular(6),
        onTap: () => context.go('/subscriptions/${subscription.id}'),
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: LgSpacing.s100),
          child: Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      '${subscription.planName} · ${subscription.priceFormatted}/mo',
                      style: theme.textTheme.bodyMedium,
                    ),
                    const SizedBox(height: LgSpacing.s100),
                    Row(
                      children: [
                        LgStatusBadge(status: subscription.status),
                        const SizedBox(width: LgSpacing.s200),
                        LgRiskBadge(riskState: subscription.riskState),
                      ],
                    ),
                  ],
                ),
              ),
              const Icon(Icons.chevron_right, size: 18, color: LgColors.textSecondary),
            ],
          ),
        ),
      ),
    );
  }
}

class _StoreTimeline extends StatelessWidget {
  final String storeDomain;
  final List<TimelineEvent> legacyTimeline;
  const _StoreTimeline({required this.storeDomain, required this.legacyTimeline});

  @override
  Widget build(BuildContext context) {
    final ep = context.watch<EventsProvider>();
    final storeEvents = ep.eventsForStore(storeDomain);
    final theme = Theme.of(context);
    final dateFmt = DateFormat('MMM d, y');

    if (storeEvents.isNotEmpty) {
      return LgCard(
        title: 'Timeline',
        child: Column(
          children: storeEvents.map((event) {
            return Padding(
              padding: const EdgeInsets.only(bottom: LgSpacing.s300),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Icon(_appEventIcon(event.type), size: 18, color: _appEventColor(event.type)),
                  const SizedBox(width: LgSpacing.s300),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Expanded(child: Text(event.title, style: theme.textTheme.bodyMedium)),
                            LgBadge(label: _appEventBadge(event.type), tone: _appEventTone(event.type)),
                          ],
                        ),
                        Text(dateFmt.format(event.date), style: theme.textTheme.bodySmall),
                      ],
                    ),
                  ),
                ],
              ),
            );
          }).toList(),
        ),
      );
    }

    // Fallback to legacy timeline
    return LgCard(
      title: 'Timeline',
      child: Column(
        children: legacyTimeline.reversed.map((event) {
          return Padding(
            padding: const EdgeInsets.only(bottom: LgSpacing.s300),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Icon(_legacyIcon(event.type), size: 18, color: _legacyColor(event.type)),
                const SizedBox(width: LgSpacing.s300),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(event.description, style: theme.textTheme.bodyMedium),
                      Text(dateFmt.format(event.date), style: theme.textTheme.bodySmall),
                    ],
                  ),
                ),
              ],
            ),
          );
        }).toList(),
      ),
    );
  }

  IconData _legacyIcon(TimelineEventType type) => switch (type) {
        TimelineEventType.install => Icons.download,
        TimelineEventType.transaction => Icons.receipt_long,
        TimelineEventType.riskChange => Icons.warning_amber,
        TimelineEventType.note => Icons.note,
      };

  Color _legacyColor(TimelineEventType type) => switch (type) {
        TimelineEventType.install => LgColors.success,
        TimelineEventType.transaction => LgColors.info,
        TimelineEventType.riskChange => LgColors.warning,
        TimelineEventType.note => LgColors.textSecondary,
      };

  IconData _appEventIcon(EventType type) => switch (type) {
        EventType.appInstall => Icons.download,
        EventType.appUninstall => Icons.delete_outline,
        EventType.appReactivated => Icons.refresh,
        EventType.appDeactivated => Icons.pause_circle_outline,
        EventType.subscriptionActivated => Icons.check_circle_outline,
        EventType.subscriptionCancelled => Icons.cancel_outlined,
        EventType.subscriptionFrozen => Icons.ac_unit,
        EventType.subscriptionUnfrozen => Icons.whatshot,
        EventType.planUpgrade => Icons.arrow_upward,
        EventType.planDowngrade => Icons.arrow_downward,
        EventType.billingFailure => Icons.error_outline,
        EventType.billingSuccess => Icons.paid,
        EventType.riskStateChange => Icons.warning_amber,
        EventType.reviewSubmitted => Icons.star_outline,
        EventType.usageCharge => Icons.data_usage,
      };

  Color _appEventColor(EventType type) => switch (type) {
        EventType.appInstall => LgColors.success,
        EventType.appUninstall => LgColors.critical,
        EventType.appReactivated => LgColors.success,
        EventType.appDeactivated => LgColors.warning,
        EventType.subscriptionActivated => LgColors.success,
        EventType.subscriptionCancelled => LgColors.critical,
        EventType.subscriptionFrozen => LgColors.warning,
        EventType.subscriptionUnfrozen => LgColors.success,
        EventType.planUpgrade => LgColors.info,
        EventType.planDowngrade => LgColors.warning,
        EventType.billingFailure => LgColors.critical,
        EventType.billingSuccess => LgColors.success,
        EventType.riskStateChange => LgColors.warning,
        EventType.reviewSubmitted => LgColors.info,
        EventType.usageCharge => LgColors.info,
      };

  String _appEventBadge(EventType type) => switch (type) {
        EventType.appInstall => 'INSTALL',
        EventType.appUninstall => 'UNINSTALL',
        EventType.appReactivated => 'REACTIVATED',
        EventType.appDeactivated => 'DEACTIVATED',
        EventType.subscriptionActivated => 'ACTIVATED',
        EventType.subscriptionCancelled => 'CANCELLED',
        EventType.subscriptionFrozen => 'FROZEN',
        EventType.subscriptionUnfrozen => 'UNFROZEN',
        EventType.planUpgrade => 'UPGRADE',
        EventType.planDowngrade => 'DOWNGRADE',
        EventType.billingFailure => 'BILLING FAIL',
        EventType.billingSuccess => 'BILLING OK',
        EventType.riskStateChange => 'RISK CHANGE',
        EventType.reviewSubmitted => 'REVIEW',
        EventType.usageCharge => 'USAGE',
      };

  BadgeTone _appEventTone(EventType type) => switch (type) {
        EventType.appInstall => BadgeTone.success,
        EventType.appUninstall => BadgeTone.critical,
        EventType.appReactivated => BadgeTone.success,
        EventType.appDeactivated => BadgeTone.warning,
        EventType.subscriptionActivated => BadgeTone.success,
        EventType.subscriptionCancelled => BadgeTone.critical,
        EventType.subscriptionFrozen => BadgeTone.warning,
        EventType.subscriptionUnfrozen => BadgeTone.success,
        EventType.planUpgrade => BadgeTone.info,
        EventType.planDowngrade => BadgeTone.warning,
        EventType.billingFailure => BadgeTone.critical,
        EventType.billingSuccess => BadgeTone.success,
        EventType.riskStateChange => BadgeTone.warning,
        EventType.reviewSubmitted => BadgeTone.info,
        EventType.usageCharge => BadgeTone.info,
      };
}

class _Info extends StatelessWidget {
  final String label;
  final String value;
  const _Info(this.label, this.value);

  @override
  Widget build(BuildContext context) {
    final labelWidth = LgBreakpoints.isMobile(context) ? 90.0 : 120.0;
    return Padding(
      padding: const EdgeInsets.only(bottom: LgSpacing.s200),
      child: Row(
        children: [
          SizedBox(width: labelWidth, child: Text(label, style: Theme.of(context).textTheme.bodySmall)),
          Expanded(child: Text(value, style: Theme.of(context).textTheme.bodyMedium)),
        ],
      ),
    );
  }
}
