import 'package:flutter/material.dart';
import '../theme/app_colors.dart';

/// Preset time windows for report date-range filtering. Mirrors the mantle-brain
/// analytics preset allowlist (7d / 30d / 90d / 12mo / YTD / all-time; default 30d).
enum DateRangePreset {
  last7Days('7d', 'Last 7 days'),
  last30Days('30d', 'Last 30 days'),
  last90Days('90d', 'Last 90 days'),
  last12Months('12mo', 'Last 12 months'),
  yearToDate('ytd', 'Year to date'),
  allTime('all', 'All time');

  final String id;
  final String label;
  const DateRangePreset(this.id, this.label);

  /// The default when nothing is selected — matches the backend's default window.
  static const DateRangePreset defaultPreset = DateRangePreset.last30Days;
}

/// A resolved [from, to] window as YYYY-MM-DD day strings (the format the report
/// endpoints expect). `from` is null only for all-time (no lower bound sent).
class DateRange {
  final String? from;
  final String? to;
  const DateRange(this.from, this.to);
}

/// Resolves a preset to concrete YYYY-MM-DD bounds relative to [now]. The day counts
/// match the backend's `parseDateRange` convention (to − N days). All-time uses a
/// far-past lower bound so the existing windowed endpoints return everything.
DateRange resolveDateRange(DateRangePreset preset, DateTime now) {
  final to = _fmt(now);
  switch (preset) {
    case DateRangePreset.last7Days:
      return DateRange(_fmt(now.subtract(const Duration(days: 7))), to);
    case DateRangePreset.last30Days:
      return DateRange(_fmt(now.subtract(const Duration(days: 30))), to);
    case DateRangePreset.last90Days:
      return DateRange(_fmt(now.subtract(const Duration(days: 90))), to);
    case DateRangePreset.last12Months:
      return DateRange(_fmt(DateTime(now.year - 1, now.month, now.day)), to);
    case DateRangePreset.yearToDate:
      return DateRange(_fmt(DateTime(now.year, 1, 1)), to);
    case DateRangePreset.allTime:
      return DateRange('2000-01-01', to);
  }
}

String _fmt(DateTime d) =>
    '${d.year.toString().padLeft(4, '0')}-${d.month.toString().padLeft(2, '0')}-${d.day.toString().padLeft(2, '0')}';

/// A chip-styled preset selector matching the wireframes' "Last 30 days ▾" control.
class DateRangeSelector extends StatelessWidget {
  final DateRangePreset value;
  final ValueChanged<DateRangePreset> onChanged;
  const DateRangeSelector({super.key, required this.value, required this.onChanged});

  @override
  Widget build(BuildContext context) {
    return PopupMenuButton<DateRangePreset>(
      initialValue: value,
      tooltip: 'Date range',
      onSelected: onChanged,
      itemBuilder: (_) => DateRangePreset.values
          .map((p) => PopupMenuItem(value: p, child: Text(p.label)))
          .toList(),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: LgColors.border),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(value.label,
                style: const TextStyle(fontSize: 12, color: LgColors.textPrimary)),
            const SizedBox(width: 4),
            const Icon(Icons.expand_more, size: 16, color: LgColors.textSecondary),
          ],
        ),
      ),
    );
  }
}
