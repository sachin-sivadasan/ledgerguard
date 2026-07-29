import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../providers/apps_provider.dart';
import '../../providers/payout_history_provider.dart';
import '../../services/payout_history_service.dart';
import '../../theme/app_colors.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_table.dart';

/// Dedicated, server-paged view of the full Payout History log — the "View all"
/// drill-down from the Payout History report. Built for whale-store volume:
/// fetches [kPayoutHistoryPageSize] payout periods per page via the report provider.
class PayoutHistoryPayoutsScreen extends StatefulWidget {
  const PayoutHistoryPayoutsScreen({super.key});

  @override
  State<PayoutHistoryPayoutsScreen> createState() =>
      _PayoutHistoryPayoutsScreenState();
}

class _PayoutHistoryPayoutsScreenState
    extends State<PayoutHistoryPayoutsScreen> with DataLoadingMixin {
  static const _pageSize = kPayoutHistoryPageSize;

  int _offset = 0;
  bool _loading = true;
  String? _error;
  PayoutHistoryReport? _page;

  @override
  void loadData(String appId) {
    final provider = context.read<PayoutHistoryProvider>();
    if (provider.selectedAppId != appId) provider.setSelectedApp(appId);
    _loadPage();
  }

  Future<void> _loadPage() async {
    final payouts = context.read<PayoutHistoryProvider>();
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
      // (mirrors the CSV-export handler on the Payout History report screen).
      debugPrint('payout-history: payouts page load failed (offset=$_offset): $e');
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
    _loadPage();
  }

  @override
  Widget build(BuildContext context) {
    final apps = context.watch<AppsProvider>();
    return LgPage(
      title: 'Payout Log',
      breadcrumb: 'Reports › Revenue & Billing › Payout History',
      subtitle:
          'Every completed payout period — paid earnings by charge month, with available date.',
      backAction: () => context.go('/reports/payout-history'),
      // scrollable:false — the table owns its own layout: sticky header, scrolling
      // rows, and a fixed footer pager (LgPaginatedTable).
      scrollable: false,
      child: _buildBody(apps),
    );
  }

  Widget _buildBody(AppsProvider apps) {
    if (apps.apps.isEmpty) {
      return apps.isLoading
          ? const Center(child: CircularProgressIndicator())
          : const LgEmptyState(
              icon: Icons.apps_outlined,
              heading: 'No app selected',
              description: 'Open Payouts from the Payout History report.',
            );
    }
    if (_error != null) {
      return LgErrorState(message: _error!, onRetry: _loadPage);
    }
    if (_loading && _page == null) {
      return const Center(child: CircularProgressIndicator());
    }
    final report = _page ?? PayoutHistoryReport.empty();
    final currency = report.currency;
    final total = report.rowsTotal;
    final rows = report.rows;
    final theme = Theme.of(context);
    final secondary =
        theme.textTheme.bodySmall?.copyWith(color: LgColors.textSecondary);
    final to = _offset + rows.length;

    return LgPaginatedTable(
      columns: const [
        LgTableColumn('PERIOD', flex: 3),
        LgTableColumn('CHARGES', flex: 2, numeric: true),
        LgTableColumn('AMOUNT', flex: 2, numeric: true),
        LgTableColumn('AVAILABLE DATE', flex: 2, numeric: true),
      ],
      rows: [
        for (final row in rows)
          [
            Text(
              row.periodDate != null
                  ? DateFormat('MMM yyyy').format(row.periodDate!)
                  : row.period,
              style: theme.textTheme.titleSmall,
            ),
            Text('${row.chargeCount}', style: secondary),
            Text(
              _money(row.amountCents, currency),
              style:
                  theme.textTheme.titleSmall?.copyWith(color: LgColors.success),
            ),
            Text(
              row.availableDateTime != null
                  ? DateFormat('MMM d, yyyy').format(row.availableDateTime!)
                  : '—',
              style: secondary,
            ),
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
      emptyMessage: 'No payouts in the selected range.',
    );
  }
}

String _money(int cents, String currency) =>
    NumberFormat.simpleCurrency(name: currency).format(cents / 100);
