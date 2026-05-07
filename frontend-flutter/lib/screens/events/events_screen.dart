import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../models/event_model.dart';
import '../../providers/apps_provider.dart';
import '../../providers/events_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_badge.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_metric_card.dart';
import '../../widgets/lg_metric_grid.dart';
import '../../widgets/lg_page.dart';

class EventsScreen extends StatefulWidget {
  const EventsScreen({super.key});

  @override
  State<EventsScreen> createState() => _EventsScreenState();
}

class _EventsScreenState extends State<EventsScreen> {
  bool _wasDemoMode = true;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _maybeLoadData());
  }

  void _maybeLoadData() {
    final apps = context.read<AppsProvider>();
    final provider = context.read<EventsProvider>();
    if (!apps.demoMode && apps.apps.isNotEmpty && !provider.isLoading) {
      provider.loadEvents(apps.apps.first.id);
    }
  }

  @override
  Widget build(BuildContext context) {
    final appsProvider = context.watch<AppsProvider>();
    final hasApps = appsProvider.apps.isNotEmpty;

    // One-shot: detect demo→live transition (initState won't re-fire in indexedStack)
    if (_wasDemoMode && !appsProvider.demoMode && hasApps) {
      _wasDemoMode = false;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) context.read<EventsProvider>().loadEvents(appsProvider.apps.first.id);
      });
    }
    if (appsProvider.demoMode) _wasDemoMode = true;

    if (!hasApps) {
      return LgPage(
        title: 'Events',
        child: LgEmptyState(
          icon: Icons.notifications_none,
          heading: 'No events yet',
          description: 'Connect your Shopify app to see events.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<EventsProvider>();
    final events = provider.events;
    final dateFmt = DateFormat('MMM d, y – h:mm a');
    final theme = Theme.of(context);

    return LgPage(
      title: 'Events',
      subtitle: '${events.length} events',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Time range toggle
          SegmentedButton<TimeRange>(
            segments: const [
              ButtonSegment(value: TimeRange.today, label: Text('Today')),
              ButtonSegment(value: TimeRange.thisWeek, label: Text('This Week')),
              ButtonSegment(value: TimeRange.thisMonth, label: Text('This Month')),
            ],
            selected: {provider.timeRange},
            onSelectionChanged: (s) => provider.setTimeRange(s.first),
          ),
          const SizedBox(height: LgSpacing.s400),

          // KPI cards
          LgMetricGrid(
            children: [
              LgMetricCard(label: 'Installs ${_rangeLabel(provider.timeRange)}', value: '${provider.installs}', icon: Icons.download),
              LgMetricCard(label: 'Uninstalls ${_rangeLabel(provider.timeRange)}', value: '${provider.uninstalls}', icon: Icons.delete_outline),
              LgMetricCard(label: 'Churns ${_rangeLabel(provider.timeRange)}', value: '${provider.churns}', icon: Icons.cancel_outlined),
              LgMetricCard(label: 'Billing Failures ${_rangeLabel(provider.timeRange)}', value: '${provider.billingFailures}', icon: Icons.error_outline),
            ],
          ),
          const SizedBox(height: LgSpacing.s600),

          // Filters
          Wrap(
            spacing: LgSpacing.s300,
            runSpacing: LgSpacing.s200,
            children: [
              _TypeFilter(
                value: provider.typeFilter,
                onChanged: provider.setTypeFilter,
              ),
              _AppFilter(
                value: provider.appFilter,
                onChanged: provider.setAppFilter,
              ),
              _StoreFilter(
                value: provider.storeFilter,
                onChanged: provider.setStoreFilter,
              ),
              if (provider.typeFilter != null ||
                  provider.appFilter != null ||
                  provider.storeFilter != null)
                TextButton(
                    onPressed: provider.clearFilters,
                    child: const Text('Clear')),
            ],
          ),
          const SizedBox(height: LgSpacing.s300),

          // Pagination label
          if (events.length > 50)
            Padding(
              padding: const EdgeInsets.only(bottom: LgSpacing.s200),
              child: Text('Showing 50 of ${events.length} events',
                  style: TextStyle(fontSize: 12, color: LgColors.textSecondary)),
            ),

          // Empty filter state
          if (events.isEmpty)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: LgSpacing.s800),
              child: Center(
                child: Column(
                  children: [
                    Icon(Icons.filter_list_off, size: 40, color: LgColors.textDisabled),
                    const SizedBox(height: LgSpacing.s300),
                    Text('No events match your filters',
                        style: theme.textTheme.bodyMedium?.copyWith(color: LgColors.textSecondary)),
                  ],
                ),
              ),
            ),

          // Event list
          ...events.take(50).map((event) {
            return Padding(
              padding: const EdgeInsets.only(bottom: LgSpacing.s300),
              child: LgCard(
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Icon(
                      _eventIcon(event.type),
                      size: 20,
                      color: _eventColor(event.type),
                    ),
                    const SizedBox(width: LgSpacing.s300),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              Expanded(
                                child: Text(event.title,
                                    style: theme.textTheme.titleSmall),
                              ),
                              LgBadge(
                                label: _eventBadgeLabel(event.type),
                                tone: _eventBadgeTone(event.type),
                              ),
                            ],
                          ),
                          const SizedBox(height: LgSpacing.s100),
                          Text(event.description,
                              style: theme.textTheme.bodyMedium),
                          const SizedBox(height: LgSpacing.s100),
                          Row(
                            children: [
                              Text(
                                event.storeDomain
                                    .replaceAll('.myshopify.com', ''),
                                style: TextStyle(
                                    fontSize: 12, color: LgColors.textSecondary),
                              ),
                              const SizedBox(width: LgSpacing.s300),
                              Text(
                                dateFmt.format(event.date),
                                style: TextStyle(
                                    fontSize: 12, color: LgColors.textSecondary),
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
          }),
        ],
      ),
    );
  }
}

String _rangeLabel(TimeRange range) => switch (range) {
      TimeRange.today => 'Today',
      TimeRange.thisWeek => 'This Week',
      TimeRange.thisMonth => 'This Month',
    };

IconData _eventIcon(EventType type) => switch (type) {
      EventType.appInstall => Icons.download,
      EventType.appUninstall => Icons.delete_outline,
      EventType.subscriptionActivated => Icons.check_circle_outline,
      EventType.subscriptionCancelled => Icons.cancel_outlined,
      EventType.planUpgrade => Icons.arrow_upward,
      EventType.planDowngrade => Icons.arrow_downward,
      EventType.billingFailure => Icons.error_outline,
      EventType.billingSuccess => Icons.paid,
      EventType.riskStateChange => Icons.warning_amber,
      EventType.reviewSubmitted => Icons.star_outline,
      EventType.usageCharge => Icons.data_usage,
    };

Color _eventColor(EventType type) => switch (type) {
      EventType.appInstall => LgColors.success,
      EventType.appUninstall => LgColors.critical,
      EventType.subscriptionActivated => LgColors.success,
      EventType.subscriptionCancelled => LgColors.critical,
      EventType.planUpgrade => LgColors.info,
      EventType.planDowngrade => LgColors.warning,
      EventType.billingFailure => LgColors.critical,
      EventType.billingSuccess => LgColors.success,
      EventType.riskStateChange => LgColors.warning,
      EventType.reviewSubmitted => LgColors.info,
      EventType.usageCharge => LgColors.info,
    };

String _eventBadgeLabel(EventType type) => switch (type) {
      EventType.appInstall => 'INSTALL',
      EventType.appUninstall => 'UNINSTALL',
      EventType.subscriptionActivated => 'ACTIVATED',
      EventType.subscriptionCancelled => 'CANCELLED',
      EventType.planUpgrade => 'UPGRADE',
      EventType.planDowngrade => 'DOWNGRADE',
      EventType.billingFailure => 'BILLING FAIL',
      EventType.billingSuccess => 'BILLING OK',
      EventType.riskStateChange => 'RISK CHANGE',
      EventType.reviewSubmitted => 'REVIEW',
      EventType.usageCharge => 'USAGE',
    };

BadgeTone _eventBadgeTone(EventType type) => switch (type) {
      EventType.appInstall => BadgeTone.success,
      EventType.appUninstall => BadgeTone.critical,
      EventType.subscriptionActivated => BadgeTone.success,
      EventType.subscriptionCancelled => BadgeTone.critical,
      EventType.planUpgrade => BadgeTone.info,
      EventType.planDowngrade => BadgeTone.warning,
      EventType.billingFailure => BadgeTone.critical,
      EventType.billingSuccess => BadgeTone.success,
      EventType.riskStateChange => BadgeTone.warning,
      EventType.reviewSubmitted => BadgeTone.info,
      EventType.usageCharge => BadgeTone.info,
    };

class _TypeFilter extends StatelessWidget {
  final EventType? value;
  final ValueChanged<EventType?> onChanged;
  const _TypeFilter({this.value, required this.onChanged});

  @override
  Widget build(BuildContext context) {
    return PopupMenuButton<EventType?>(
      onSelected: onChanged,
      itemBuilder: (ctx) => [
        const PopupMenuItem(value: null, child: Text('All Types')),
        ...EventType.values.map((t) => PopupMenuItem(
              value: t,
              child: Text(_eventBadgeLabel(t)),
            )),
      ],
      child: Chip(
        label: Text(value != null ? _eventBadgeLabel(value!) : 'Type'),
        deleteIcon: value != null ? const Icon(Icons.close, size: 14) : null,
        onDeleted: value != null ? () => onChanged(null) : null,
      ),
    );
  }
}

class _AppFilter extends StatelessWidget {
  final String? value;
  final ValueChanged<String?> onChanged;
  const _AppFilter({this.value, required this.onChanged});

  @override
  Widget build(BuildContext context) {
    final apps = context.watch<AppsProvider>().apps;
    return PopupMenuButton<String?>(
      onSelected: onChanged,
      itemBuilder: (ctx) => [
        const PopupMenuItem(value: null, child: Text('All Apps')),
        ...apps.map((app) => PopupMenuItem(value: app.id, child: Text(app.name))),
      ],
      child: Chip(
        label: Text(value != null
            ? apps.firstWhere((a) => a.id == value, orElse: () => apps.first).name
            : 'App'),
        deleteIcon: value != null ? const Icon(Icons.close, size: 14) : null,
        onDeleted: value != null ? () => onChanged(null) : null,
      ),
    );
  }
}

class _StoreFilter extends StatelessWidget {
  final String? value;
  final ValueChanged<String?> onChanged;
  const _StoreFilter({this.value, required this.onChanged});

  @override
  Widget build(BuildContext context) {
    return PopupMenuButton<String?>(
      onSelected: onChanged,
      itemBuilder: (ctx) => [
        const PopupMenuItem(value: null, child: Text('All Stores')),
        const PopupMenuItem(value: 'acme-store', child: Text('acme-store')),
        const PopupMenuItem(value: 'bright-gadgets', child: Text('bright-gadgets')),
        const PopupMenuItem(value: 'eco-shop', child: Text('eco-shop')),
        const PopupMenuItem(value: 'daily-deals', child: Text('daily-deals')),
        const PopupMenuItem(value: 'fresh-foods', child: Text('fresh-foods')),
        const PopupMenuItem(value: 'glow-beauty', child: Text('glow-beauty')),
        const PopupMenuItem(value: 'alpha-outlet', child: Text('alpha-outlet')),
        const PopupMenuItem(value: 'beta-mart', child: Text('beta-mart')),
      ],
      child: Chip(
        label: Text(value ?? 'Store'),
        deleteIcon: value != null ? const Icon(Icons.close, size: 14) : null,
        onDeleted: value != null ? () => onChanged(null) : null,
      ),
    );
  }
}
