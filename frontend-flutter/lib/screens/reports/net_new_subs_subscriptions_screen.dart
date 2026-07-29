import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/report_detail_app_gate.dart';
import '../../providers/apps_provider.dart';
import '../../providers/net_new_subs_provider.dart';
import '../../services/net_new_subs_service.dart';
import '../../theme/app_colors.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_table.dart';

/// Dedicated, server-paged view of the full new-subscriptions table — the
/// "View all" drill-down from the Net-New Subscriptions report. Built for
/// whale-store volume: fetches [kNetNewSubsPageSize] rows per page via the
/// report provider.
class NetNewSubsSubscriptionsScreen extends StatefulWidget {
  const NetNewSubsSubscriptionsScreen({super.key});

  @override
  State<NetNewSubsSubscriptionsScreen> createState() =>
      _NetNewSubsSubscriptionsScreenState();
}

class _NetNewSubsSubscriptionsScreenState
    extends State<NetNewSubsSubscriptionsScreen> with ReportDetailAppGate {
  static const _pageSize = kNetNewSubsPageSize;

  @override
  void reloadAfterApps() => _load();

  int _offset = 0;
  bool _loading = true;
  String? _error;
  NetNewSubsReport? _page;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _load());
  }

  Future<void> _load() async {
    final subs = context.read<NetNewSubsProvider>();
    // Cold deep-link / hard reload: the report page hasn't run, so no app is
    // selected yet. Seed it from AppsProvider so the page has data on its own.
    if (subs.selectedAppId == null) {
      final apps = context.read<AppsProvider>();
      final appId = apps.selectedAppId ??
          (apps.apps.isNotEmpty ? apps.apps.first.id : null);
      if (appId != null) {
        subs.setSelectedApp(appId);
      } else {
        // Cold deep-link / hard reload: apps haven't arrived yet (they load after
        // org selection). Wait for them, then retry; error only if none arrive.
        setState(() {
          _loading = true;
          _error = null;
        });
        waitForAppsThenReload(onUnavailable: () {
          if (!mounted) return;
          setState(() {
            _error =
                'No app selected. Open Subscriptions from the Net-New Subscriptions report.';
            _loading = false;
          });
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
          await subs.fetchSubscriptionsPage(limit: _pageSize, offset: _offset);
      if (!mounted) return;
      setState(() {
        _page = page;
        _loading = false;
      });
    } catch (e) {
      // Don't swallow the cause; distinguish a transient 503 from a real failure
      // (mirrors the CSV-export handler on the Net-New Subscriptions screen).
      debugPrint(
          'net-new-subs: subscriptions page load failed (offset=$_offset): $e');
      if (!mounted) return;
      final isUnavailable = e is DioException && e.response?.statusCode == 503;
      setState(() {
        _error = isUnavailable
            ? 'Service temporarily unavailable. Please try again shortly.'
            : 'Could not load subscriptions. Please try again.';
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
      title: 'Recent new subscriptions',
      breadcrumb: 'Reports › Growth › Net-New Subscriptions',
      subtitle:
          'Every new subscription — store, plan, MRR, and start date over the period.',
      backAction: () => context.go('/reports/net-new-subscriptions'),
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
    final report = _page ?? NetNewSubsReport.empty();
    final currency = report.currency;
    final total = report.newStoresTotal;
    final rows = report.newStores;
    final theme = Theme.of(context);
    final secondary =
        theme.textTheme.bodySmall?.copyWith(color: LgColors.textSecondary);
    final to = _offset + rows.length;

    return LgPaginatedTable(
      columns: const [
        LgTableColumn('STORE', flex: 4),
        LgTableColumn('PLAN', flex: 2),
        LgTableColumn('MRR', flex: 2, numeric: true),
        LgTableColumn('STARTED', flex: 2, numeric: true),
      ],
      rows: [
        for (final s in rows)
          [
            Text(
              s.domain.isNotEmpty
                  ? s.domain
                  : (s.shopName.isNotEmpty ? s.shopName : '—'),
              style: theme.textTheme.titleSmall,
            ),
            Text(
              s.planName.isNotEmpty ? s.planName : '—',
              style: theme.textTheme.bodyMedium,
            ),
            Text(
              _money(s.mrrCents, currency),
              style: theme.textTheme.titleSmall?.copyWith(
                color: LgColors.success,
              ),
            ),
            Text(_started(s), style: secondary),
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
      emptyMessage: 'No new subscriptions in the selected range.',
    );
  }
}

String _money(int cents, String currency) =>
    NumberFormat.simpleCurrency(name: currency).format(cents / 100);

String _started(NewSubRow s) => s.startedDate != null
    ? DateFormat('MMM d, yyyy').format(s.startedDate!)
    : (s.started.isNotEmpty ? s.started : '—');
