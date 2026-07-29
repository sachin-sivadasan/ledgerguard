import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../providers/apps_provider.dart';
import '../../providers/usage_provider.dart';
import '../../services/usage_service.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_table.dart';

/// Dedicated, server-paged view of the full Usage stores table — the
/// "View all" drill-down from the Usage & One-Time report. Built for
/// whale-store volume: fetches [kUsageStoresPageSize] rows per page via the
/// report provider.
class UsageStoresScreen extends StatefulWidget {
  const UsageStoresScreen({super.key});

  @override
  State<UsageStoresScreen> createState() => _UsageStoresScreenState();
}

class _UsageStoresScreenState extends State<UsageStoresScreen>
    with DataLoadingMixin {
  static const _pageSize = kUsageStoresPageSize;

  int _offset = 0;
  bool _loading = true;
  String? _error;
  UsageReport? _page;

  @override
  void loadData(String appId) {
    final provider = context.read<UsageProvider>();
    if (provider.selectedAppId != appId) provider.setSelectedApp(appId);
    _loadPage();
  }

  Future<void> _loadPage() async {
    final usage = context.read<UsageProvider>();
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final page =
          await usage.fetchStoresPage(limit: _pageSize, offset: _offset);
      if (!mounted) return;
      setState(() {
        _page = page;
        _loading = false;
      });
    } catch (e) {
      // Don't swallow the cause; distinguish a transient 503 from a real failure
      // (mirrors the CSV-export handler on the Usage report screen).
      debugPrint('usage: stores page load failed (offset=$_offset): $e');
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
      title: 'Top Usage Stores (ranked by usage revenue)',
      breadcrumb: 'Reports › Revenue & Billing › Usage & One-Time',
      subtitle:
          'USAGE and ONE-TIME revenue by store — tracked separately from recurring MRR.',
      backAction: () => context.go('/reports/usage'),
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
              description: 'Open Stores from the Usage report.',
            );
    }
    if (_error != null) {
      return LgErrorState(message: _error!, onRetry: _loadPage);
    }
    if (_loading && _page == null) {
      return const Center(child: CircularProgressIndicator());
    }
    final report = _page ?? UsageReport.empty();
    final currency = report.currency;
    final total = report.storesTotal;
    final rows = report.stores;
    final theme = Theme.of(context);
    final to = _offset + rows.length;

    return LgPaginatedTable(
      columns: const [
        LgTableColumn('STORE', flex: 3),
        LgTableColumn('USAGE \$', flex: 2, numeric: true),
        LgTableColumn('ONE-TIME \$', flex: 2, numeric: true),
        LgTableColumn('CHARGES', flex: 1, numeric: true),
      ],
      rows: [
        for (final s in rows)
          [
            _StoreCell(store: s),
            Text(_money(s.usageCents, currency),
                style: theme.textTheme.titleSmall),
            Text(_money(s.oneTimeCents, currency),
                style: theme.textTheme.titleSmall),
            Text('${s.chargeCount}', style: theme.textTheme.titleSmall),
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
      emptyMessage: 'No usage or one-time charges in the selected range.',
    );
  }
}

/// Two-line store cell: name over domain (domain hidden when absent).
class _StoreCell extends StatelessWidget {
  final UsageStore store;
  const _StoreCell({required this.store});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final name = store.shopName.isNotEmpty
        ? store.shopName
        : (store.domain.isNotEmpty ? store.domain : '—');

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(name, style: theme.textTheme.titleSmall),
        if (store.domain.isNotEmpty) ...[
          const SizedBox(height: LgSpacing.s100),
          Text(
            store.domain,
            style: theme.textTheme.bodySmall?.copyWith(
              color: LgColors.textSecondary,
            ),
          ),
        ],
      ],
    );
  }
}

String _money(int cents, String currency) =>
    NumberFormat.simpleCurrency(name: currency).format(cents / 100);
