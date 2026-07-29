import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../providers/apps_provider.dart';
import '../../providers/churn_provider.dart';
import '../../services/churn_service.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_table.dart';

/// Dedicated, server-paged view of the full Churn churned-stores table — the
/// "View all" drill-down from the Churn report. Built for whale-store volume:
/// fetches [kChurnStoresPageSize] rows per page via the churn provider.
class ChurnStoresScreen extends StatefulWidget {
  const ChurnStoresScreen({super.key});

  @override
  State<ChurnStoresScreen> createState() => _ChurnStoresScreenState();
}

class _ChurnStoresScreenState extends State<ChurnStoresScreen> {
  static const _pageSize = kChurnStoresPageSize;

  int _offset = 0;
  bool _loading = true;
  String? _error;
  ChurnReport? _page;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _load());
  }

  Future<void> _load() async {
    final churn = context.read<ChurnProvider>();
    // Cold deep-link / hard reload: the report page hasn't run, so no app is
    // selected yet. Seed it from AppsProvider so the page has data on its own.
    if (churn.selectedAppId == null) {
      final apps = context.read<AppsProvider>();
      final appId = apps.selectedAppId ??
          (apps.apps.isNotEmpty ? apps.apps.first.id : null);
      if (appId != null) {
        churn.setSelectedApp(appId);
      } else {
        // No app resolvable (apps not loaded yet / none connected). Surface that
        // distinctly — otherwise fetchStoresPage returns empty and the page
        // would lie with "No churned stores in the selected range."
        setState(() {
          _error = 'No app selected. Open Stores from the Churn report.';
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
          await churn.fetchStoresPage(limit: _pageSize, offset: _offset);
      if (!mounted) return;
      setState(() {
        _page = page;
        _loading = false;
      });
    } catch (e) {
      // Don't swallow the cause; distinguish a transient 503 from a real failure
      // (mirrors the CSV-export handler on the Churn report screen).
      debugPrint('churn: stores page load failed (offset=$_offset): $e');
      if (!mounted) return;
      final isUnavailable = e is DioException && e.response?.statusCode == 503;
      setState(() {
        _error = isUnavailable
            ? 'Service temporarily unavailable. Please try again shortly.'
            : 'Could not load churned stores. Please try again.';
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
      title: 'Churned Stores (ranked by MRR lost)',
      breadcrumb: 'Reports › Retention & Risk › Churn',
      subtitle:
          'Every churned store — MRR lost, churned date, and tenure, ranked by MRR lost.',
      backAction: () => context.go('/reports/churn'),
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
    final report = _page ?? ChurnReport.empty();
    final currency = report.currency;
    final total = report.storesTotal;
    final rows = report.stores;
    final theme = Theme.of(context);
    final to = _offset + rows.length;

    return LgPaginatedTable(
      columns: const [
        LgTableColumn('STORE', flex: 4),
        LgTableColumn('MRR LOST', flex: 2, numeric: true),
        LgTableColumn('CHURNED DATE', flex: 2, numeric: true),
        LgTableColumn('TENURE', flex: 2, numeric: true),
      ],
      rows: [
        for (final s in rows)
          [
            _StoreCell(store: s),
            Text(
              '-${_money(s.mrrLostCents, currency)}',
              style: theme.textTheme.titleSmall?.copyWith(
                color: LgColors.critical,
              ),
            ),
            Text(
              s.churnedDate != null
                  ? DateFormat('MMM d').format(s.churnedDate!)
                  : '—',
              style: theme.textTheme.bodySmall,
            ),
            Text(
              _tenure(s.tenureDays),
              style: theme.textTheme.bodySmall?.copyWith(
                color: LgColors.textSecondary,
              ),
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
      emptyMessage: 'No churned stores in the selected range.',
    );
  }
}

/// Two-line store cell: shop name over plan (or domain) — links to store detail.
class _StoreCell extends StatelessWidget {
  final ChurnStore store;
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
            const SizedBox(height: LgSpacing.s100),
            Text(
              store.planName.isNotEmpty ? store.planName : store.domain,
              style: theme.textTheme.bodySmall?.copyWith(
                color: LgColors.textSecondary,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

String _money(int cents, String currency) =>
    NumberFormat.simpleCurrency(name: currency).format(cents / 100);

String _tenure(int days) {
  if (days <= 0) return '—';
  final months = days / 30.0;
  return '${months.toStringAsFixed(1)} mo';
}
