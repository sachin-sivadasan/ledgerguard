import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../providers/apps_provider.dart';
import '../../providers/revenue_at_risk_provider.dart';
import '../../services/revenue_at_risk_service.dart';
import '../../theme/app_colors.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_risk_badge.dart';
import '../../widgets/lg_table.dart';

/// Dedicated, server-paged view of the full Revenue at Risk ranked-stores table
/// — the "View all" drill-down from the Revenue at Risk report. Built for
/// whale-store volume: fetches [kRevenueAtRiskStoresPageSize] rows per page via
/// the report provider.
class RevenueAtRiskStoresScreen extends StatefulWidget {
  const RevenueAtRiskStoresScreen({super.key});

  @override
  State<RevenueAtRiskStoresScreen> createState() =>
      _RevenueAtRiskStoresScreenState();
}

class _RevenueAtRiskStoresScreenState extends State<RevenueAtRiskStoresScreen>
    with DataLoadingMixin {
  static const _pageSize = kRevenueAtRiskStoresPageSize;

  int _offset = 0;
  bool _loading = true;
  String? _error;
  RevenueAtRiskReport? _page;

  @override
  void loadData(String appId) {
    final provider = context.read<RevenueAtRiskProvider>();
    if (provider.selectedAppId != appId) provider.setSelectedApp(appId);
    _loadPage();
  }

  Future<void> _loadPage() async {
    final risk = context.read<RevenueAtRiskProvider>();
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final page =
          await risk.fetchStoresPage(limit: _pageSize, offset: _offset);
      if (!mounted) return;
      setState(() {
        _page = page;
        _loading = false;
      });
    } catch (e) {
      // Don't swallow the cause; distinguish a transient 503 from a real failure
      // (mirrors the CSV-export handler on the Revenue at Risk report screen).
      debugPrint('revenue-at-risk: stores page load failed (offset=$_offset): $e');
      if (!mounted) return;
      final isUnavailable = e is DioException && e.response?.statusCode == 503;
      setState(() {
        _error = isUnavailable
            ? 'Service temporarily unavailable. Please try again shortly.'
            : 'Could not load stores. Please try again.';
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
      title: 'Ranked Stores',
      breadcrumb: 'Reports › Retention & Risk › Revenue at Risk',
      subtitle:
          'Every at-risk store — MRR, risk state, days late, and recoverable revenue.',
      backAction: () => context.go('/reports/revenue-at-risk'),
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
              description: 'Open Stores from the Revenue at Risk report.',
            );
    }
    if (_error != null) {
      return LgErrorState(message: _error!, onRetry: _loadPage);
    }
    if (_loading && _page == null) {
      return const Center(child: CircularProgressIndicator());
    }
    final report = _page ?? RevenueAtRiskReport.empty();
    final currency = report.currency;
    final total = report.storesTotal;
    final rows = report.stores;
    final theme = Theme.of(context);
    final to = _offset + rows.length;

    return LgPaginatedTable(
      columns: const [
        LgTableColumn('STORE', flex: 4),
        LgTableColumn('MRR', flex: 2, numeric: true),
        LgTableColumn('RISK', flex: 3),
        LgTableColumn('DAYS LATE', flex: 2, numeric: true),
        LgTableColumn('EXPECTED CHARGE', flex: 2, numeric: true),
        LgTableColumn('RECOVERABLE', flex: 2, numeric: true),
      ],
      rows: [
        for (final s in rows)
          [
            _StoreCell(store: s),
            Text('${_money(s.mrrCents, currency)}/mo',
                style: theme.textTheme.titleSmall),
            LgRiskBadge(riskState: s.riskState),
            Text('${s.daysLate}d', style: theme.textTheme.bodySmall),
            Text(
              s.expectedChargeDate != null
                  ? DateFormat('MMM d').format(s.expectedChargeDate!)
                  : '—',
              style: theme.textTheme.bodySmall,
            ),
            Text(
              _money(s.recoverableCents, currency),
              style: theme.textTheme.titleSmall
                  ?.copyWith(color: LgColors.success),
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
      emptyMessage: 'No at-risk stores in the selected range.',
    );
  }
}

/// Two-line store cell: shop name over plan — links to store detail.
class _StoreCell extends StatelessWidget {
  final RevenueAtRiskStore store;
  const _StoreCell({required this.store});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final name = store.shopName.isNotEmpty
        ? store.shopName
        : store.domain.replaceAll('.myshopify.com', '');

    return MouseRegion(
      cursor: SystemMouseCursors.click,
      child: InkWell(
        onTap: () => context.go('/stores/${store.domain}'),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(name, style: theme.textTheme.titleSmall),
            if (store.planName.isNotEmpty) ...[
              const SizedBox(height: 4),
              Text(
                store.planName,
                style: theme.textTheme.bodySmall
                    ?.copyWith(color: LgColors.textSecondary),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

String _money(int cents, String currency) =>
    NumberFormat.simpleCurrency(name: currency).format(cents / 100);
