import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../providers/apps_provider.dart';
import '../../providers/earnings_report_provider.dart';
import '../../services/earnings_report_service.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_table.dart';

/// Dedicated, server-paged view of the full Earnings charges table — the
/// "View all" drill-down from the Earnings report. Built for whale-store volume:
/// fetches [kEarningsChargesPageSize] rows per page via the report provider.
class EarningsChargesScreen extends StatefulWidget {
  const EarningsChargesScreen({super.key});

  @override
  State<EarningsChargesScreen> createState() => _EarningsChargesScreenState();
}

class _EarningsChargesScreenState extends State<EarningsChargesScreen> {
  static const _pageSize = kEarningsChargesPageSize;

  int _offset = 0;
  bool _loading = true;
  String? _error;
  EarningsReport? _page;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _load());
  }

  Future<void> _load() async {
    final earnings = context.read<EarningsReportProvider>();
    // Cold deep-link / hard reload: the report page hasn't run, so no app is
    // selected yet. Seed it from AppsProvider so the page has data on its own.
    if (earnings.selectedAppId == null) {
      final apps = context.read<AppsProvider>();
      final appId = apps.selectedAppId ??
          (apps.apps.isNotEmpty ? apps.apps.first.id : null);
      if (appId != null) earnings.setSelectedApp(appId);
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final page =
          await earnings.fetchChargesPage(limit: _pageSize, offset: _offset);
      if (!mounted) return;
      setState(() {
        _page = page;
        _loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = 'Could not load charges. Please try again.';
        _loading = false;
      });
    }
  }

  void _goTo(int offset) {
    setState(() => _offset = offset);
    _load();
  }

  @override
  Widget build(BuildContext context) {
    return LgPage(
      title: 'Charges',
      breadcrumb: 'Reports › Revenue & Billing › Earnings',
      subtitle: 'Every charge — net earnings by store, payout status, available date.',
      backAction: () => context.go('/reports/earnings'),
      child: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_error != null) {
      return LgErrorState(message: _error!, onRetry: _load);
    }
    if (_loading && _page == null) {
      return const Center(child: Padding(
        padding: EdgeInsets.all(LgSpacing.s600),
        child: CircularProgressIndicator(),
      ));
    }
    final report = _page ?? EarningsReport.empty();
    final currency = report.currency;
    final total = report.chargesTotal;
    final rows = report.charges;
    final theme = Theme.of(context);
    final secondary =
        theme.textTheme.bodySmall?.copyWith(color: LgColors.textSecondary);

    if (total == 0) {
      return Padding(
        padding: const EdgeInsets.symmetric(vertical: LgSpacing.s400),
        child: Text('No charges in the selected range.', style: secondary),
      );
    }

    final from = _offset + 1;
    final to = _offset + rows.length;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Opacity(
          opacity: _loading ? 0.5 : 1,
          child: LgTable(
            columns: const [
              LgTableColumn('DATE', flex: 2),
              LgTableColumn('STORE', flex: 4),
              LgTableColumn('GROSS', flex: 2, numeric: true),
              LgTableColumn('NET', flex: 2, numeric: true),
              LgTableColumn('STATUS', flex: 2),
              LgTableColumn('AVAILABLE DATE', flex: 3, numeric: true),
            ],
            rows: [
              for (final c in rows)
                [
                  Text(_date(c.date), style: secondary),
                  Text(
                    c.shopName.isNotEmpty
                        ? c.shopName
                        : (c.domain.isNotEmpty ? c.domain : '—'),
                    style: theme.textTheme.titleSmall,
                  ),
                  Text(_money(c.grossCents, currency), style: secondary),
                  Text(_money(c.netCents, currency),
                      style: theme.textTheme.titleSmall),
                  _StatusChip(status: c.status),
                  Text(_date(c.availableDate), style: secondary),
                ],
            ],
          ),
        ),
        const SizedBox(height: LgSpacing.s400),
        _Pager(
          from: from,
          to: to,
          total: total,
          canPrev: _offset > 0 && !_loading,
          canNext: to < total && !_loading,
          onPrev: () => _goTo((_offset - _pageSize).clamp(0, _offset)),
          onNext: () => _goTo(_offset + _pageSize),
        ),
      ],
    );
  }
}

class _Pager extends StatelessWidget {
  final int from;
  final int to;
  final int total;
  final bool canPrev;
  final bool canNext;
  final VoidCallback onPrev;
  final VoidCallback onNext;
  const _Pager({
    required this.from,
    required this.to,
    required this.total,
    required this.canPrev,
    required this.canNext,
    required this.onPrev,
    required this.onNext,
  });

  @override
  Widget build(BuildContext context) {
    final secondary = Theme.of(context)
        .textTheme
        .bodySmall
        ?.copyWith(color: LgColors.textSecondary);
    return Row(
      children: [
        Text('Showing $from–$to of $total', style: secondary),
        const Spacer(),
        OutlinedButton(
          onPressed: canPrev ? onPrev : null,
          child: const Text('← Prev'),
        ),
        const SizedBox(width: LgSpacing.s200),
        OutlinedButton(
          onPressed: canNext ? onNext : null,
          child: const Text('Next →'),
        ),
      ],
    );
  }
}

class _StatusChip extends StatelessWidget {
  final String status;
  const _StatusChip({required this.status});

  @override
  Widget build(BuildContext context) {
    final (bg, fg) = switch (status.toLowerCase()) {
      'pending' => (LgColors.warning.withValues(alpha: 0.14), LgColors.warning),
      'available' => (LgColors.success.withValues(alpha: 0.14), LgColors.success),
      _ => (LgColors.surfaceSecondary, LgColors.textSecondary),
    };
    final label = status.isNotEmpty ? status : '—';
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration:
          BoxDecoration(color: bg, borderRadius: BorderRadius.circular(10)),
      child: Text(label,
          style: TextStyle(fontSize: 11, color: fg, fontWeight: FontWeight.w600)),
    );
  }
}

String _money(int cents, String currency) =>
    NumberFormat.simpleCurrency(name: currency).format(cents / 100);

String _date(DateTime? value) =>
    value == null ? '—' : DateFormat('MMM d, yyyy').format(value);
