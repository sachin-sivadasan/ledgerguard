import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../providers/apps_provider.dart';
import '../../providers/installs_provider.dart';
import '../../services/installs_service.dart';
import '../../theme/app_colors.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_table.dart';

/// Dedicated, server-paged view of the full Installs events table — the
/// "View all" drill-down from the Installs report. Built for whale-store volume:
/// fetches [kInstallsEventsPageSize] rows per page via the report provider.
class InstallsEventsScreen extends StatefulWidget {
  const InstallsEventsScreen({super.key});

  @override
  State<InstallsEventsScreen> createState() => _InstallsEventsScreenState();
}

class _InstallsEventsScreenState extends State<InstallsEventsScreen> {
  static const _pageSize = kInstallsEventsPageSize;

  int _offset = 0;
  bool _loading = true;
  String? _error;
  InstallsReport? _page;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _load());
  }

  Future<void> _load() async {
    final installs = context.read<InstallsProvider>();
    // Cold deep-link / hard reload: the report page hasn't run, so no app is
    // selected yet. Seed it from AppsProvider so the page has data on its own.
    if (installs.selectedAppId == null) {
      final apps = context.read<AppsProvider>();
      final appId = apps.selectedAppId ??
          (apps.apps.isNotEmpty ? apps.apps.first.id : null);
      if (appId != null) {
        installs.setSelectedApp(appId);
      } else {
        // No app resolvable (apps not loaded yet / none connected). Surface that
        // distinctly — otherwise fetchEventsPage returns empty and the page
        // would lie with "No install events in the selected range."
        setState(() {
          _error = 'No app selected. Open Events from the Installs report.';
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
          await installs.fetchEventsPage(limit: _pageSize, offset: _offset);
      if (!mounted) return;
      setState(() {
        _page = page;
        _loading = false;
      });
    } catch (e) {
      // Don't swallow the cause; distinguish a transient 503 from a real failure
      // (mirrors the CSV-export handler on the Installs report screen).
      debugPrint('installs: events page load failed (offset=$_offset): $e');
      if (!mounted) return;
      final isUnavailable = e is DioException && e.response?.statusCode == 503;
      setState(() {
        _error = isUnavailable
            ? 'Service temporarily unavailable. Please try again shortly.'
            : 'Could not load events. Please try again.';
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
      title: 'Recent install / uninstall events',
      breadcrumb: 'Reports › Growth › Installs',
      subtitle: 'Every install and uninstall event — store, event type, date.',
      backAction: () => context.go('/reports/installs'),
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
    final report = _page ?? InstallsReport.empty();
    final total = report.eventsTotal;
    final rows = report.events;
    final theme = Theme.of(context);
    final secondary =
        theme.textTheme.bodySmall?.copyWith(color: LgColors.textSecondary);
    final to = _offset + rows.length;

    return LgPaginatedTable(
      columns: const [
        LgTableColumn('STORE', flex: 4),
        LgTableColumn('EVENT', flex: 2),
        LgTableColumn('DATE', flex: 2, numeric: true),
      ],
      rows: [
        for (final e in rows)
          [
            Text(
              e.domain.isNotEmpty ? e.domain : '—',
              style: theme.textTheme.titleSmall,
            ),
            _EventChip(event: e.event),
            Text(
              e.dateTime != null
                  ? DateFormat('MMM d, yyyy').format(e.dateTime!)
                  : (e.date.isNotEmpty ? e.date : '—'),
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
      emptyMessage: 'No install events in the selected range.',
    );
  }
}

/// Event pill — green for Install, amber for Uninstall.
class _EventChip extends StatelessWidget {
  final String event;
  const _EventChip({required this.event});

  @override
  Widget build(BuildContext context) {
    final (bg, fg) = switch (event.toLowerCase()) {
      'install' => (LgColors.success.withValues(alpha: 0.14), LgColors.success),
      'uninstall' => (
        LgColors.warning.withValues(alpha: 0.14),
        LgColors.warning,
      ),
      _ => (LgColors.surfaceSecondary, LgColors.textSecondary),
    };
    final label = event.isNotEmpty ? event : '—';
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(10),
      ),
      child: Text(
        label,
        style: TextStyle(fontSize: 11, color: fg, fontWeight: FontWeight.w600),
      ),
    );
  }
}
