import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../mock_data/mock_apps.dart';
import '../../models/transaction_model.dart';
import '../../providers/apps_provider.dart';
import '../../providers/transaction_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../theme/app_breakpoints.dart';
import '../../widgets/lg_badge.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_data_table.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_page.dart';

class TransactionsScreen extends StatelessWidget {
  const TransactionsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final hasApps = context.watch<AppsProvider>().apps.isNotEmpty;
    if (!hasApps) {
      return LgPage(
        title: 'Transactions',
        child: LgEmptyState(
          icon: Icons.swap_horiz,
          heading: 'No transactions yet',
          description:
              'Connect your Shopify app to see transaction history.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<TransactionProvider>();
    final txns = provider.transactions;
    final dateFmt = DateFormat('MMM d, y');
    final theme = Theme.of(context);

    return LgPage(
      title: 'Transactions',
      subtitle: '${txns.length} transactions',
      secondaryActions: [
        LgPageAction(label: 'Export CSV', onPressed: () {}),
      ],
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
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
              if (provider.typeFilter != null || provider.appFilter != null)
                TextButton(onPressed: provider.clearFilters, child: const Text('Clear')),
            ],
          ),
          const SizedBox(height: LgSpacing.s300),

          // Summary
          LgCard(
            child: LgBreakpoints.isMobile(context)
                ? Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _SummaryRow('Gross Total', '\$${(provider.totalGrossCents / 100).toStringAsFixed(2)}', theme),
                      const SizedBox(height: LgSpacing.s200),
                      _SummaryRow('Net Total', '\$${(provider.totalNetCents / 100).toStringAsFixed(2)}', theme),
                      const SizedBox(height: LgSpacing.s200),
                      _SummaryRow('Shopify Cut', '\$${((provider.totalGrossCents - provider.totalNetCents) / 100).toStringAsFixed(2)}', theme, valueColor: LgColors.warning),
                    ],
                  )
                : Row(
                    children: [
                      Text('Gross Total: ', style: theme.textTheme.bodySmall),
                      Text('\$${(provider.totalGrossCents / 100).toStringAsFixed(2)}', style: theme.textTheme.titleSmall),
                      const SizedBox(width: LgSpacing.s600),
                      Text('Net Total: ', style: theme.textTheme.bodySmall),
                      Text('\$${(provider.totalNetCents / 100).toStringAsFixed(2)}', style: theme.textTheme.titleSmall),
                      const SizedBox(width: LgSpacing.s600),
                      Text('Shopify Cut: ', style: theme.textTheme.bodySmall),
                      Text(
                        '\$${((provider.totalGrossCents - provider.totalNetCents) / 100).toStringAsFixed(2)}',
                        style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: LgColors.warning),
                      ),
                    ],
                  ),
          ),
          const SizedBox(height: LgSpacing.s300),

          // Pagination label
          if (txns.length > 50)
            Padding(
              padding: const EdgeInsets.only(bottom: LgSpacing.s200),
              child: Text('Showing 50 of ${txns.length} transactions',
                  style: TextStyle(fontSize: 12, color: LgColors.textSecondary)),
            ),

          // Empty filter state
          if (txns.isEmpty)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: LgSpacing.s800),
              child: Center(
                child: Column(
                  children: [
                    Icon(Icons.filter_list_off, size: 40, color: LgColors.textDisabled),
                    const SizedBox(height: LgSpacing.s300),
                    Text('No transactions match your filters',
                        style: theme.textTheme.bodyMedium?.copyWith(color: LgColors.textSecondary)),
                  ],
                ),
              ),
            ),

          // Table
          if (txns.isNotEmpty)
          Card(
            child: LgDataTable(
              columns: const [
                LgColumn(title: 'Date'),
                LgColumn(title: 'Store'),
                LgColumn(title: 'Type'),
                LgColumn(title: 'App'),
                LgColumn(title: 'Gross', numeric: true),
                LgColumn(title: 'Net', numeric: true),
              ],
              rows: txns.take(50).map((txn) {
                return DataRow(cells: [
                  DataCell(Text(dateFmt.format(txn.date))),
                  DataCell(Text(txn.shopDomain.replaceAll('.myshopify.com', ''))),
                  DataCell(LgBadge(label: txn.chargeTypeLabel, tone: _typeTone(txn.chargeType))),
                  DataCell(Text(_appName(txn.appId))),
                  DataCell(Text(txn.grossFormatted)),
                  DataCell(Text(txn.netFormatted)),
                ]);
              }).toList(),
            ),
          ),
        ],
      ),
    );
  }

  BadgeTone _typeTone(ChargeType type) => switch (type) {
        ChargeType.recurring => BadgeTone.success,
        ChargeType.usage => BadgeTone.info,
        ChargeType.oneTime => BadgeTone.attention,
        ChargeType.refund => BadgeTone.critical,
      };

  String _appName(String appId) {
    try {
      return mockApps.firstWhere((a) => a.id == appId).name;
    } catch (_) {
      return appId;
    }
  }
}

class _SummaryRow extends StatelessWidget {
  final String label;
  final String value;
  final ThemeData theme;
  final Color? valueColor;
  const _SummaryRow(this.label, this.value, this.theme, {this.valueColor});

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text('$label:', style: theme.textTheme.bodySmall),
        Text(value, style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: valueColor ?? LgColors.textPrimary)),
      ],
    );
  }
}

class _TypeFilter extends StatelessWidget {
  final ChargeType? value;
  final ValueChanged<ChargeType?> onChanged;
  const _TypeFilter({this.value, required this.onChanged});

  @override
  Widget build(BuildContext context) {
    return PopupMenuButton<ChargeType?>(
      onSelected: onChanged,
      itemBuilder: (ctx) => [
        const PopupMenuItem(value: null, child: Text('All Types')),
        ...ChargeType.values.map((t) => PopupMenuItem(value: t, child: Text(t.name.toUpperCase()))),
      ],
      child: Chip(
        label: Text(value?.name.toUpperCase() ?? 'Type'),
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
    return PopupMenuButton<String?>(
      onSelected: onChanged,
      itemBuilder: (ctx) => [
        const PopupMenuItem(value: null, child: Text('All Apps')),
        ...mockApps.map((app) => PopupMenuItem(value: app.id, child: Text(app.name))),
      ],
      child: Chip(
        label: Text(value != null
            ? mockApps.firstWhere((a) => a.id == value, orElse: () => mockApps.first).name
            : 'App'),
        deleteIcon: value != null ? const Icon(Icons.close, size: 14) : null,
        onDeleted: value != null ? () => onChanged(null) : null,
      ),
    );
  }
}
