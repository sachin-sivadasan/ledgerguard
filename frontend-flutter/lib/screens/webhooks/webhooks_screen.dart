import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../models/webhook_model.dart';
import '../../providers/apps_provider.dart';
import '../../providers/webhook_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_badge.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_metric_card.dart';
import '../../widgets/lg_metric_grid.dart';
import '../../widgets/lg_page.dart';

class WebhooksScreen extends StatelessWidget {
  const WebhooksScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final hasApps = context.watch<AppsProvider>().apps.isNotEmpty;
    if (!hasApps) {
      return LgPage(
        title: 'Webhooks',
        child: LgEmptyState(
          icon: Icons.webhook,
          heading: 'No webhooks yet',
          description:
              'Connect your Shopify app to see webhook events.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<WebhookProvider>();
    final webhooks = provider.webhooks;
    final dateFmt = DateFormat('MMM d, y – h:mm a');
    final theme = Theme.of(context);

    return LgPage(
      title: 'Webhooks',
      subtitle: '${webhooks.length} webhook events',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // KPI cards
          LgMetricGrid(
            children: [
              LgMetricCard(label: 'Webhooks Today', value: '${provider.totalToday}', icon: Icons.webhook),
              LgMetricCard(label: 'Failed Today', value: '${provider.failedToday}', icon: Icons.error_outline),
              LgMetricCard(label: '7d Success Rate', value: provider.successRate7d, icon: Icons.check_circle_outline),
            ],
          ),
          const SizedBox(height: LgSpacing.s600),

          // Filters
          Wrap(
            spacing: LgSpacing.s300,
            runSpacing: LgSpacing.s200,
            children: [
              _SourceFilter(
                value: provider.sourceFilter,
                onChanged: provider.setSourceFilter,
              ),
              _StatusFilter(
                value: provider.statusFilter,
                onChanged: provider.setStatusFilter,
              ),
              if (provider.sourceFilter != null ||
                  provider.statusFilter != null)
                TextButton(
                    onPressed: provider.clearFilters,
                    child: const Text('Clear')),
            ],
          ),
          const SizedBox(height: LgSpacing.s300),

          // Pagination label
          if (webhooks.length > 50)
            Padding(
              padding: const EdgeInsets.only(bottom: LgSpacing.s200),
              child: Text('Showing 50 of ${webhooks.length} webhook events',
                  style: TextStyle(fontSize: 12, color: LgColors.textSecondary)),
            ),

          // Empty filter state
          if (webhooks.isEmpty)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: LgSpacing.s800),
              child: Center(
                child: Column(
                  children: [
                    Icon(Icons.filter_list_off, size: 40, color: LgColors.textDisabled),
                    const SizedBox(height: LgSpacing.s300),
                    Text('No webhooks match your filters',
                        style: theme.textTheme.bodyMedium?.copyWith(color: LgColors.textSecondary)),
                  ],
                ),
              ),
            ),

          // Webhook list
          ...webhooks.take(50).map((wh) {
            final isFailed = wh.status == WebhookStatus.failed;
            return Padding(
              padding: const EdgeInsets.only(bottom: LgSpacing.s300),
              child: LgCard(
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Icon(
                      wh.source == WebhookSource.shopify
                          ? Icons.shopping_cart
                          : Icons.payment,
                      size: 20,
                      color: isFailed ? LgColors.critical : LgColors.info,
                    ),
                    const SizedBox(width: LgSpacing.s300),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              Expanded(
                                child: Text(wh.topic,
                                    style: theme.textTheme.titleSmall),
                              ),
                              LgBadge(
                                label: wh.status == WebhookStatus.success
                                    ? 'SUCCESS'
                                    : 'FAILED',
                                tone: wh.status == WebhookStatus.success
                                    ? BadgeTone.success
                                    : BadgeTone.critical,
                              ),
                            ],
                          ),
                          const SizedBox(height: LgSpacing.s100),
                          Text(wh.payloadSummary,
                              style: theme.textTheme.bodyMedium),
                          const SizedBox(height: LgSpacing.s100),
                          Row(
                            children: [
                              LgBadge(
                                  label: wh.sourceLabel,
                                  tone: BadgeTone.defaultTone),
                              const SizedBox(width: LgSpacing.s300),
                              if (wh.storeDomain != null) ...[
                                Text(
                                  wh.storeDomain!
                                      .replaceAll('.myshopify.com', ''),
                                  style: TextStyle(
                                      fontSize: 12,
                                      color: LgColors.textSecondary),
                                ),
                                const SizedBox(width: LgSpacing.s300),
                              ],
                              Text(
                                dateFmt.format(wh.receivedAt),
                                style: TextStyle(
                                    fontSize: 12,
                                    color: LgColors.textSecondary),
                              ),
                              if (wh.httpStatus != 200) ...[
                                const SizedBox(width: LgSpacing.s300),
                                LgBadge(
                                    label: 'HTTP ${wh.httpStatus}',
                                    tone: BadgeTone.critical),
                              ],
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

class _SourceFilter extends StatelessWidget {
  final WebhookSource? value;
  final ValueChanged<WebhookSource?> onChanged;
  const _SourceFilter({this.value, required this.onChanged});

  @override
  Widget build(BuildContext context) {
    return PopupMenuButton<WebhookSource?>(
      onSelected: onChanged,
      itemBuilder: (ctx) => [
        const PopupMenuItem(value: null, child: Text('All Sources')),
        const PopupMenuItem(
            value: WebhookSource.shopify, child: Text('Shopify')),
        const PopupMenuItem(
            value: WebhookSource.razorpay, child: Text('Razorpay')),
      ],
      child: Chip(
        label: Text(value != null
            ? (value == WebhookSource.shopify ? 'Shopify' : 'Razorpay')
            : 'Source'),
        deleteIcon: value != null ? const Icon(Icons.close, size: 14) : null,
        onDeleted: value != null ? () => onChanged(null) : null,
      ),
    );
  }
}

class _StatusFilter extends StatelessWidget {
  final WebhookStatus? value;
  final ValueChanged<WebhookStatus?> onChanged;
  const _StatusFilter({this.value, required this.onChanged});

  @override
  Widget build(BuildContext context) {
    return PopupMenuButton<WebhookStatus?>(
      onSelected: onChanged,
      itemBuilder: (ctx) => [
        const PopupMenuItem(value: null, child: Text('All Statuses')),
        const PopupMenuItem(
            value: WebhookStatus.success, child: Text('Success')),
        const PopupMenuItem(
            value: WebhookStatus.failed, child: Text('Failed')),
      ],
      child: Chip(
        label: Text(value != null
            ? (value == WebhookStatus.success ? 'Success' : 'Failed')
            : 'Status'),
        deleteIcon: value != null ? const Icon(Icons.close, size: 14) : null,
        onDeleted: value != null ? () => onChanged(null) : null,
      ),
    );
  }
}
