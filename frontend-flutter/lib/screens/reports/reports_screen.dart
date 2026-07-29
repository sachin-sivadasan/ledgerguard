import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_page.dart';

/// Reports landing page: category groups with a card grid. Only
/// "Revenue at Risk" is active; the rest are marked "Coming soon".
class ReportsScreen extends StatelessWidget {
  const ReportsScreen({super.key});

  static const _categories = <_ReportCategory>[
    _ReportCategory('Retention & Risk', Icons.shield_outlined, [
      _ReportEntry(
        'Revenue at Risk',
        'At-risk MRR, recoverable revenue, ranked stores',
        route: '/reports/revenue-at-risk',
      ),
      _ReportEntry(
        'Churn',
        'Cancellations and lost MRR over time',
        route: '/reports/churn',
      ),
      _ReportEntry(
        'Retention',
        'How many merchants stay active',
        route: '/reports/retention',
      ),
      _ReportEntry(
        'Retention Cohorts',
        'Retention by signup cohort',
        route: '/reports/cohorts',
      ),
      _ReportEntry(
        'Reviews',
        'App Store review sentiment',
        route: '/reports/reviews',
      ),
      _ReportEntry(
        'Uninstall Context',
        'Inferred pre-uninstall state, tenure, and plan',
        route: '/reports/uninstall-context',
      ),
    ]),
    _ReportCategory('Revenue & Billing', Icons.attach_money_outlined, [
      _ReportEntry(
        'Earnings',
        'Net payouts after Shopify fees',
        route: '/reports/earnings',
      ),
      _ReportEntry(
        'Monthly Recurring Revenue',
        'MRR trend and movements',
        route: '/reports/mrr',
      ),
      _ReportEntry(
        'Revenue Mix',
        'Recurring vs usage vs one-time',
        route: '/reports/revenue-mix',
      ),
      _ReportEntry(
        'Usage & One-Time Charges',
        'Usage and add-on revenue',
        route: '/reports/usage',
      ),
      _ReportEntry(
        'Usage Trends',
        'Week-over-week usage momentum and top usage customers',
        route: '/reports/usage-trends',
      ),
      _ReportEntry(
        'Subscriptions',
        'Active base, ARPU and lifetime value',
        route: '/reports/subscriptions',
      ),
      _ReportEntry(
        'Payout Schedule',
        'Upcoming payout timing',
        route: '/reports/payout-schedule',
      ),
      _ReportEntry(
        'Payout History',
        'Completed payouts by month',
        route: '/reports/payout-history',
      ),
    ]),
    _ReportCategory('Growth', Icons.trending_up_outlined, [
      _ReportEntry(
        'Installs',
        'New installs over time',
        route: '/reports/installs',
      ),
      _ReportEntry('Activation', 'Install-to-paid funnel'),
      _ReportEntry(
        'Net-New Subscriptions',
        'New paying subscriptions',
        route: '/reports/net-new-subscriptions',
      ),
    ]),
    _ReportCategory('Customers', Icons.people_outline, [
      _ReportEntry('Active Customers', 'Active paying merchants'),
      _ReportEntry('Customer Insights', 'Segments and behavior'),
    ]),
    _ReportCategory('Guard', Icons.verified_user_outlined, [
      _ReportEntry('Fee Audit', 'Verify Shopify fee calculations'),
      _ReportEntry('Payout Accuracy', 'Reconcile expected vs actual payouts'),
      _ReportEntry('Ledger Reconciliation', 'Cross-check the rebuilt ledger'),
    ]),
  ];

  @override
  Widget build(BuildContext context) {
    return LgPage(
      title: 'Reports',
      subtitle: 'Export-ready insights for your board, accountant, and team',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          for (final category in _categories) ...[
            _CategoryHeader(category: category),
            const SizedBox(height: LgSpacing.s300),
            _ReportGrid(entries: category.reports),
            const SizedBox(height: LgSpacing.s800),
          ],
        ],
      ),
    );
  }
}

class _CategoryHeader extends StatelessWidget {
  final _ReportCategory category;
  const _CategoryHeader({required this.category});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Row(
      children: [
        Icon(category.icon, size: 20, color: LgColors.primary),
        const SizedBox(width: LgSpacing.s200),
        Text(category.title, style: theme.textTheme.titleMedium),
      ],
    );
  }
}

class _ReportGrid extends StatelessWidget {
  final List<_ReportEntry> entries;
  const _ReportGrid({required this.entries});

  @override
  Widget build(BuildContext context) {
    final columns = switch (LgBreakpoints.deviceType(context)) {
      LgDeviceType.mobile => 1,
      LgDeviceType.tablet => 2,
      LgDeviceType.desktop => 3,
    };
    return LayoutBuilder(
      builder: (context, constraints) {
        const spacing = LgSpacing.s300;
        final totalSpacing = spacing * (columns - 1);
        final cardWidth = (constraints.maxWidth - totalSpacing) / columns;
        return Wrap(
          spacing: spacing,
          runSpacing: spacing,
          children: entries
              .map(
                (e) => SizedBox(
                  width: cardWidth,
                  child: _ReportCard(entry: e),
                ),
              )
              .toList(),
        );
      },
    );
  }
}

class _ReportCard extends StatelessWidget {
  final _ReportEntry entry;
  const _ReportCard({required this.entry});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final active = entry.route != null;

    final card = LgCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: Text(
                  entry.title,
                  style: theme.textTheme.titleSmall?.copyWith(
                    color: active ? null : LgColors.textDisabled,
                  ),
                ),
              ),
              if (active)
                const Icon(Icons.chevron_right, color: LgColors.textSecondary)
              else
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 3,
                  ),
                  decoration: BoxDecoration(
                    color: LgColors.surfaceSecondary,
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: const Text(
                    'Coming soon',
                    style: TextStyle(
                      fontSize: 11,
                      color: LgColors.textSecondary,
                    ),
                  ),
                ),
            ],
          ),
          const SizedBox(height: LgSpacing.s200),
          Text(
            entry.description,
            style: theme.textTheme.bodySmall?.copyWith(
              color: active ? LgColors.textSecondary : LgColors.textDisabled,
            ),
          ),
        ],
      ),
    );

    if (!active) return Opacity(opacity: 0.7, child: card);

    return MouseRegion(
      cursor: SystemMouseCursors.click,
      child: GestureDetector(
        onTap: () => context.go(entry.route!),
        child: card,
      ),
    );
  }
}

class _ReportCategory {
  final String title;
  final IconData icon;
  final List<_ReportEntry> reports;
  const _ReportCategory(this.title, this.icon, this.reports);
}

class _ReportEntry {
  final String title;
  final String description;
  final String? route;
  const _ReportEntry(this.title, this.description, {this.route});
}
