import 'package:flutter/material.dart';
import '../theme/app_breakpoints.dart';
import '../theme/app_colors.dart';
import '../theme/app_spacing.dart';

/// A column definition for [LgTable] / [LgPaginatedTable].
class LgTableColumn {
  final String label;
  final int flex;

  /// Right-aligns the header + cells (for money / counts).
  final bool numeric;
  const LgTableColumn(this.label, {this.flex = 2, this.numeric = false});
}

/// On mobile, tables are data-dense: shrink all cell/header text ~10% and tighten
/// padding so wide tables (many columns) fit without cramping. Applied centrally so
/// every report table (incl. screen-built cells) benefits without per-screen styling.
Widget _maybeDense(BuildContext context, Widget table) {
  if (!LgBreakpoints.isMobile(context)) return table;
  final mq = MediaQuery.of(context);
  return MediaQuery(
    data: mq.copyWith(textScaler: const TextScaler.linear(0.9)),
    child: table,
  );
}

double _hPad(bool dense) => dense ? LgSpacing.s300 : LgSpacing.s400;
double _vPad(bool dense) => dense ? LgSpacing.s200 : LgSpacing.s300;
double _gap(bool dense) => dense ? LgSpacing.s200 : LgSpacing.s400;

/// Shared header row: shaded (rx=8), inter-column gaps (skipped before the first
/// column so a right-aligned numeric header can't abut the next left header —
/// e.g. "NET"+"STATUS" -> "NETSTATUS").
Widget _lgHeaderRow(ThemeData theme, List<LgTableColumn> columns, bool dense) {
  return Container(
    padding:
        EdgeInsets.symmetric(horizontal: _hPad(dense), vertical: _vPad(dense)),
    decoration: BoxDecoration(
      color: LgColors.surfaceSecondary,
      borderRadius: BorderRadius.circular(8),
    ),
    child: Row(
      children: [
        for (var i = 0; i < columns.length; i++) ...[
          if (i > 0) SizedBox(width: _gap(dense)),
          Expanded(
            flex: columns[i].flex,
            child: Text(
              columns[i].label,
              textAlign: columns[i].numeric ? TextAlign.end : TextAlign.start,
              // Keep headers on one line — a narrow numeric column otherwise
              // breaks a single word mid-character ("CHARGE\nS").
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: theme.textTheme.bodySmall?.copyWith(
                color: LgColors.textSecondary,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
        ],
      ],
    ),
  );
}

/// Shared data row: flat, thin bottom divider, cells aligned per column.
Widget _lgDataRow(
    ThemeData theme, List<LgTableColumn> columns, List<Widget> row, bool dense) {
  return Container(
    padding:
        EdgeInsets.symmetric(horizontal: _hPad(dense), vertical: _vPad(dense)),
    decoration: const BoxDecoration(
      border: Border(bottom: BorderSide(color: LgColors.border)),
    ),
    child: Row(
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        for (var i = 0; i < columns.length; i++) ...[
          if (i > 0) SizedBox(width: _gap(dense)),
          Expanded(
            flex: columns[i].flex,
            child: Align(
              alignment: columns[i].numeric
                  ? Alignment.centerRight
                  : Alignment.centerLeft,
              child: row[i],
            ),
          ),
        ],
      ],
    ),
  );
}

/// A data table matching the wireframes: a shaded header row (rx=8) followed by
/// flat data rows. Sizes to its content — for report previews that scroll with
/// the page. For a full, paged table use [LgPaginatedTable].
/// Each row is a list of cell widgets, one per column (same order + count as [columns]).
class LgTable extends StatelessWidget {
  final List<LgTableColumn> columns;
  final List<List<Widget>> rows;
  const LgTable({super.key, required this.columns, required this.rows});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final dense = LgBreakpoints.isMobile(context);
    return _maybeDense(
      context,
      Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _lgHeaderRow(theme, columns, dense),
          for (final row in rows) _lgDataRow(theme, columns, row, dense),
        ],
      ),
    );
  }
}

/// A full-height paginated data table following the standard data-grid layout:
/// a **sticky column header**, a scrolling row body, and a **fixed footer pager**
/// ("Showing X–Y of Z" + Prev/Next) — so the header never leaves and you never
/// scroll to reach the pager. Place it in a bounded-height slot, e.g. inside
/// `LgPage(scrollable: false, child: LgPaginatedTable(...))`.
class LgPaginatedTable extends StatelessWidget {
  final List<LgTableColumn> columns;
  final List<List<Widget>> rows;

  /// 1-based index of the first/last visible row, and the full row count.
  final int from;
  final int to;
  final int total;

  final bool canPrev;
  final bool canNext;

  /// Dims the body while a page is in flight (buttons should also be disabled).
  final bool loading;
  final VoidCallback onPrev;
  final VoidCallback onNext;
  final String emptyMessage;

  const LgPaginatedTable({
    super.key,
    required this.columns,
    required this.rows,
    required this.from,
    required this.to,
    required this.total,
    required this.canPrev,
    required this.canNext,
    required this.onPrev,
    required this.onNext,
    this.loading = false,
    this.emptyMessage = 'No rows in the selected range.',
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final dense = LgBreakpoints.isMobile(context);
    final secondary =
        theme.textTheme.bodySmall?.copyWith(color: LgColors.textSecondary);

    final Widget body;
    if (rows.isEmpty) {
      body = Center(
        child: loading
            ? const CircularProgressIndicator()
            : Text(emptyMessage, style: secondary),
      );
    } else {
      body = Opacity(
        opacity: loading ? 0.5 : 1,
        child: ListView.builder(
          itemCount: rows.length,
          itemBuilder: (_, i) => _lgDataRow(theme, columns, rows[i], dense),
        ),
      );
    }

    return _maybeDense(
      context,
      Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _lgHeaderRow(theme, columns, dense), // sticky
          Expanded(child: body), // scrolls under the header
          // Fixed footer pager.
          Container(
            padding: EdgeInsets.symmetric(
                horizontal: _hPad(dense), vertical: _vPad(dense)),
            decoration: const BoxDecoration(
              border: Border(top: BorderSide(color: LgColors.border)),
            ),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    total == 0 ? '0 rows' : 'Showing $from–$to of $total',
                    style: secondary,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
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
            ),
          ),
        ],
      ),
    );
  }
}
