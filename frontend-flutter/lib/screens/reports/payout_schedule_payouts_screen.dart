import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../providers/apps_provider.dart';
import '../../providers/payout_schedule_provider.dart';
import '../../services/payout_schedule_service.dart';
import '../../theme/app_colors.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_table.dart';

/// Dedicated, server-paged view of the full Payout Schedule payouts table — the
/// "View all" drill-down from the Payout Schedule report. Built for whale-store
/// volume: fetches [kPayoutSchedulePageSize] rows per page via the provider.
class PayoutSchedulePayoutsScreen extends StatefulWidget {
  const PayoutSchedulePayoutsScreen({super.key});

  @override
  State<PayoutSchedulePayoutsScreen> createState() =>
      _PayoutSchedulePayoutsScreenState();
}

class _PayoutSchedulePayoutsScreenState
    extends State<PayoutSchedulePayoutsScreen> {
  static const _pageSize = kPayoutSchedulePageSize;

  int _offset = 0;
  bool _loading = true;
  String? _error;
  PayoutScheduleReport? _page;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _load());
  }

  Future<void> _load() async {
    final payouts = context.read<PayoutScheduleProvider>();
    // Cold deep-link / hard reload: the report page hasn't run, so no app is
    // selected yet. Seed it from AppsProvider so the page has data on its own.
    if (payouts.selectedAppId == null) {
      final apps = context.read<AppsProvider>();
      final appId = apps.selectedAppId ??
          (apps.apps.isNotEmpty ? apps.apps.first.id : null);
      if (appId != null) {
        payouts.setSelectedApp(appId);
      } else {
        // No app resolvable (apps not loaded yet / none connected). Surface that
        // distinctly — otherwise fetchPayoutsPage returns empty and the page
        // would lie with "No upcoming payouts in the selected range."
        setState(() {
          _error = 'No app selected. Open Payouts from the Payout Schedule report.';
          _loading = false;
        });
        return;
      }
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final page =
          await payouts.fetchPayoutsPage(limit: _pageSize, offset: _offset);
      if (!mounted) return;
      setState(() {
        _page = page;
        _loading = false;
      });
    } catch (e) {
      // Don't swallow the cause; distinguish a transient 503 from a real failure
      // (mirrors the CSV-export handler on the Payout Schedule report screen).
      debugPrint('payout-schedule: payouts page load failed (offset=$_offset): $e');
      if (!mounted) return;
      final isUnavailable = e is DioException && e.response?.statusCode == 503;
      setState(() {
        _error = isUnavailable
            ? 'Service temporarily unavailable. Please try again shortly.'
            : 'Could not load payouts. Please try again.';
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
      title: 'Upcoming Payouts',
      breadcrumb: 'Reports › Revenue & Billing › Payout Schedule',
      subtitle:
          'Every scheduled payout — amount, charge count, and availability status.',
      backAction: () => context.go('/reports/payout-schedule'),
      // scrollable:false — the table owns its own layout: sticky header, scrolling
      // rows, and a fixed footer pager (LgPaginatedTable).
      scrollable: false,
      child: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_error != null) {
      return LgErrorState(message: _error!, onRetry: _load);
    }
    if (_loading && _page == null) {
      return const Center(child: CircularProgressIndicator());
    }
    final report = _page ?? PayoutScheduleReport.empty();
    final currency = report.currency;
    final total = report.rowsTotal;
    final rows = report.rows;
    final theme = Theme.of(context);
    final secondary =
        theme.textTheme.bodySmall?.copyWith(color: LgColors.textSecondary);
    final to = _offset + rows.length;

    return LgPaginatedTable(
      columns: const [
        LgTableColumn('AVAILABLE DATE', flex: 3),
        LgTableColumn('AMOUNT', flex: 2, numeric: true),
        LgTableColumn('# CHARGES', flex: 2, numeric: true),
        LgTableColumn('STATUS', flex: 2),
      ],
      rows: [
        for (final row in rows)
          [
            Text(
              row.date != null
                  ? DateFormat('MMM d, yyyy').format(row.date!)
                  : '—',
              style: theme.textTheme.titleSmall,
            ),
            Text(_money(row.amountCents, currency),
                style: theme.textTheme.titleSmall),
            Text('${row.chargeCount}', style: secondary),
            _StatusChip(status: row.status),
          ],
      ],
      from: rows.isEmpty ? 0 : _offset + 1,
      to: to,
      total: total,
      canPrev: _offset > 0 && !_loading,
      canNext: to < total && !_loading,
      loading: _loading,
      onPrev: () => _goTo((_offset - _pageSize).clamp(0, _offset)),
      onNext: () => _goTo(_offset + _pageSize),
      emptyMessage: 'No upcoming payouts in the selected range.',
    );
  }
}

/// Status pill — green for Available (ready), amber for Pending (clearing).
class _StatusChip extends StatelessWidget {
  final String status;
  const _StatusChip({required this.status});

  @override
  Widget build(BuildContext context) {
    final (bg, fg) = switch (status.toLowerCase()) {
      'available' => (
        LgColors.success.withValues(alpha: 0.14),
        LgColors.success,
      ),
      'pending' => (LgColors.warning.withValues(alpha: 0.14), LgColors.warning),
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
