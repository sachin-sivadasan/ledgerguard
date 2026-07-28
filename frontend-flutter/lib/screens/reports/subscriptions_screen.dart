import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../providers/apps_provider.dart';
import '../../providers/subscriptions_provider.dart';
import '../../services/subscriptions_service.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_service_unavailable.dart';

class SubscriptionsScreen extends StatefulWidget {
  const SubscriptionsScreen({super.key});

  @override
  State<SubscriptionsScreen> createState() => _SubscriptionsScreenState();
}

class _SubscriptionsScreenState extends State<SubscriptionsScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<SubscriptionsProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<SubscriptionsProvider>();
    final appId = provider.selectedAppId;
    if (appId == null) return;

    try {
      final bytes = await provider.fetchCsvBytes();
      if (bytes == null || bytes.isEmpty) {
        if (!mounted) return;
        messenger.showSnackBar(
          const SnackBar(content: Text('CSV export returned no data.')),
        );
        return;
      }
      final filename =
          'subscriptions-${DateTime.now().toIso8601String().split('T').first}.csv';
      final ok = downloadBytes(bytes, filename, 'text/csv');
      if (!mounted) return;
      if (!ok) {
        messenger.showSnackBar(
          const SnackBar(
            content: Text('CSV export is only available on the web app.'),
          ),
        );
      }
    } catch (e) {
      // Don't swallow the cause — surface a 503 with a matching
      // service-unavailable message, and log everything else so export
      // failures stay diagnosable.
      debugPrint('subscriptions: CSV export failed: $e');
      if (!mounted) return;
      final isUnavailable = e is DioException && e.response?.statusCode == 503;
      messenger.showSnackBar(
        SnackBar(
          content: Text(isUnavailable
              ? 'Service temporarily unavailable. Please try again shortly.'
              : 'Could not export CSV. Please try again.'),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final appsProvider = context.watch<AppsProvider>();
    final hasApps = appsProvider.apps.isNotEmpty;

    if (!hasApps) {
      return LgPage(
        title: 'Subscriptions',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgEmptyState(
          icon: Icons.people_outline,
          heading: 'No data yet',
          description: 'Connect your Shopify app to see subscription metrics.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<SubscriptionsProvider>();

    if (provider.isServiceUnavailable) {
      return LgPage(
        title: 'Subscriptions',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgServiceUnavailable(onRetry: retryLoad),
      );
    }

    if (provider.error != null) {
      return LgPage(
        title: 'Subscriptions',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.report == null) {
      return LgPage(
        title: 'Subscriptions',
        breadcrumb: 'Reports › Revenue & Billing',
        backAction: () => context.go('/reports'),
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final report = provider.report ?? SubscriptionsReport.empty();
    final appsList = appsProvider.apps;
    final showAppFilter = appsList.length > 1;
    final currency = report.currency;
    final hasData = report.activeSubs > 0 || report.plans.isNotEmpty;

    return LgPage(
      title: 'Subscriptions',
      breadcrumb: 'Reports › Revenue & Billing',
      subtitle:
          'Active subscription base by plan — ARPU and lifetime value composition',
      backAction: () => context.go('/reports'),
      onRefresh: refreshData,
      secondaryActions: [
        LgPageAction(label: 'Export CSV', onPressed: _exportCsv),
      ],
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (showAppFilter) ...[
            PopupMenuButton<String>(
              onSelected: provider.setSelectedApp,
              itemBuilder: (_) => appsList
                  .map((app) =>
                      PopupMenuItem(value: app.id, child: Text(app.name)))
                  .toList(),
              child: Chip(
                label: Text(appsList
                    .firstWhere((a) => a.id == provider.selectedAppId,
                        orElse: () => appsList.first)
                    .name),
              ),
            ),
            const SizedBox(height: LgSpacing.s300),
          ],
          if (!hasData)
            const LgEmptyState(
              icon: Icons.people_outline,
              heading: 'No active subscriptions in range',
              description:
                  'Once your app has active recurring subscriptions, ARPU, lifetime value and your per-plan composition will appear here.',
            )
          else ...[
            _HeroRow(report: report, currency: currency),
            const SizedBox(height: LgSpacing.s600),
            _CompositionCard(report: report),
            const SizedBox(height: LgSpacing.s600),
            _PlanTable(report: report, currency: currency),
            const SizedBox(height: LgSpacing.s400),
            Text(
              'RECURRING subscriptions only — USAGE charges are excluded from ARPU / LTV.',
              style: Theme.of(context)
                  .textTheme
                  .bodySmall
                  ?.copyWith(color: LgColors.textSecondary),
            ),
          ],
        ],
      ),
    );
  }
}

// ─── Hero KPI cards ─────────────────────────────────────────────────
class _HeroRow extends StatelessWidget {
  final SubscriptionsReport report;
  final String currency;
  const _HeroRow({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final cards = <Widget>[
      _KpiCard(
        label: 'Active Subscriptions',
        value: '${report.activeSubs}',
        color: LgColors.textPrimary,
        footnote: 'subscriptions in a SAFE (paying) state',
      ),
      _KpiCard(
        label: 'ARPU',
        value: _money(report.arpuCents, currency),
        color: LgColors.success,
        footnote: 'MRR ÷ active subscriptions (monthly)',
      ),
      _KpiCard(
        label: 'LTV',
        // 0 means the churn rate was 0, so lifetime value is undefined — show "—".
        value: report.ltvCents > 0 ? _money(report.ltvCents, currency) : '—',
        color: LgColors.textPrimary,
        footnote: report.ltvCents > 0
            ? 'ARPU ÷ ${(report.churnRate * 100).toStringAsFixed(1)}% monthly churn'
            : 'needs a non-zero churn rate to compute',
      ),
    ];

    if (LgBreakpoints.isMobile(context)) {
      return Column(
        children: [
          for (var i = 0; i < cards.length; i++) ...[
            if (i > 0) const SizedBox(height: LgSpacing.s300),
            cards[i],
          ],
        ],
      );
    }
    return Row(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        for (var i = 0; i < cards.length; i++) ...[
          if (i > 0) const SizedBox(width: LgSpacing.s300),
          Expanded(child: cards[i]),
        ],
      ],
    );
  }
}

class _KpiCard extends StatelessWidget {
  final String label;
  final String value;
  final Color? color;
  final String? footnote;
  const _KpiCard({
    required this.label,
    required this.value,
    this.color,
    this.footnote,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return LgCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label,
              style: theme.textTheme.bodySmall
                  ?.copyWith(color: LgColors.textSecondary)),
          const SizedBox(height: LgSpacing.s200),
          Text(value,
              style: theme.textTheme.headlineMedium
                  ?.copyWith(color: color, fontWeight: FontWeight.w700)),
          if (footnote != null) ...[
            const SizedBox(height: LgSpacing.s100),
            Text(footnote!,
                style: theme.textTheme.bodySmall
                    ?.copyWith(color: LgColors.textSecondary)),
          ],
        ],
      ),
    );
  }
}

// ─── Composition card: subscriptions by plan (horizontal bars) ──────
class _CompositionCard extends StatelessWidget {
  final SubscriptionsReport report;
  const _CompositionCard({required this.report});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    // Bars already arrive sorted by active subs desc; re-sort defensively.
    final plans = [...report.plans]
      ..sort((a, b) => b.activeSubs.compareTo(a.activeSubs));

    return LgCard(
      title: 'Subscriptions by Plan',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          for (final p in plans) ...[
            _PlanBar(plan: p),
            const SizedBox(height: LgSpacing.s300),
          ],
          Text(
            'Bar length = share of active subscriptions.',
            style: theme.textTheme.bodySmall
                ?.copyWith(color: LgColors.textSecondary),
          ),
        ],
      ),
    );
  }
}

class _PlanBar extends StatelessWidget {
  final SubscriptionsPlan plan;
  const _PlanBar({required this.plan});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final name = plan.planName.isNotEmpty ? plan.planName : '(no plan)';
    // Guard against a >1 fraction defensively so the bar never overflows.
    final fraction = plan.pctOfSubs.clamp(0.0, 1.0);

    return Row(
      children: [
        SizedBox(
          width: 96,
          child: Text(name,
              style: theme.textTheme.bodyMedium, overflow: TextOverflow.ellipsis),
        ),
        const SizedBox(width: LgSpacing.s300),
        Expanded(
          child: ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: Stack(
              children: [
                Container(height: 18, color: LgColors.border),
                FractionallySizedBox(
                  widthFactor: fraction,
                  child: Container(height: 18, color: LgColors.primary),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(width: LgSpacing.s300),
        SizedBox(
          width: 44,
          child: Text('${plan.activeSubs}',
              textAlign: TextAlign.end,
              style: theme.textTheme.bodySmall
                  ?.copyWith(color: LgColors.textSecondary)),
        ),
      ],
    );
  }
}

// ─── Plan detail table ──────────────────────────────────────────────
class _PlanTable extends StatelessWidget {
  final SubscriptionsReport report;
  final String currency;
  const _PlanTable({required this.report, required this.currency});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final plans = [...report.plans]
      ..sort((a, b) => b.activeSubs.compareTo(a.activeSubs));

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Plan Detail', style: theme.textTheme.titleMedium),
        const SizedBox(height: LgSpacing.s300),
        // Column header row: PLAN | SUBS | ARPU | LTV
        Padding(
          padding: const EdgeInsets.symmetric(
              horizontal: LgSpacing.s400, vertical: LgSpacing.s200),
          child: Row(
            children: [
              Expanded(
                flex: 3,
                child: Text('PLAN',
                    style: theme.textTheme.bodySmall
                        ?.copyWith(color: LgColors.textSecondary)),
              ),
              _headCell(theme, 'SUBS'),
              _headCell(theme, 'ARPU'),
              _headCell(theme, 'LTV'),
            ],
          ),
        ),
        ...plans.map((p) => Padding(
              padding: const EdgeInsets.only(bottom: LgSpacing.s200),
              child: _PlanRow(plan: p, currency: currency),
            )),
        const SizedBox(height: LgSpacing.s200),
        Text(
          'Per-plan LTV uses the blended (app-level) churn rate.',
          style: theme.textTheme.bodySmall
              ?.copyWith(color: LgColors.textSecondary),
        ),
      ],
    );
  }

  Widget _headCell(ThemeData theme, String label) => Expanded(
        flex: 2,
        child: Text(label,
            textAlign: TextAlign.end,
            style: theme.textTheme.bodySmall
                ?.copyWith(color: LgColors.textSecondary)),
      );
}

class _PlanRow extends StatelessWidget {
  final SubscriptionsPlan plan;
  final String currency;
  const _PlanRow({required this.plan, required this.currency});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final name = plan.planName.isNotEmpty ? plan.planName : '(no plan)';

    return LgCard(
      child: Row(
        children: [
          Expanded(
            flex: 3,
            child: Text(name, style: theme.textTheme.titleSmall),
          ),
          Expanded(
            flex: 2,
            child: Text('${plan.activeSubs}',
                textAlign: TextAlign.end, style: theme.textTheme.titleSmall),
          ),
          Expanded(
            flex: 2,
            child: Text(_money(plan.arpuCents, currency),
                textAlign: TextAlign.end, style: theme.textTheme.titleSmall),
          ),
          Expanded(
            flex: 2,
            child: Text(
              // "—" when LTV is undefined (churn rate 0), matching the hero card.
              plan.ltvCents > 0 ? _money(plan.ltvCents, currency) : '—',
              textAlign: TextAlign.end,
              style: theme.textTheme.titleSmall,
            ),
          ),
        ],
      ),
    );
  }
}

String _money(int cents, String currency) {
  final format = NumberFormat.simpleCurrency(name: currency);
  return format.format(cents / 100);
}
