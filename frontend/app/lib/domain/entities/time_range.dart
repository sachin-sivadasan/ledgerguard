import 'package:equatable/equatable.dart';

/// Predefined time range options for dashboard filtering
enum TimeRangePreset {
  thisMonth,
  lastMonth,
  last30Days,
  last90Days,
  custom;

  /// Human-readable display name
  String get displayName {
    switch (this) {
      case TimeRangePreset.thisMonth:
        return 'This Month';
      case TimeRangePreset.lastMonth:
        return 'Last Month';
      case TimeRangePreset.last30Days:
        return 'Last 30 Days';
      case TimeRangePreset.last90Days:
        return 'Last 90 Days';
      case TimeRangePreset.custom:
        return 'Custom';
    }
  }

  /// Backend API value
  String get apiValue {
    switch (this) {
      case TimeRangePreset.thisMonth:
        return 'THIS_MONTH';
      case TimeRangePreset.lastMonth:
        return 'LAST_MONTH';
      case TimeRangePreset.last30Days:
        return 'LAST_30_DAYS';
      case TimeRangePreset.last90Days:
        return 'LAST_90_DAYS';
      case TimeRangePreset.custom:
        return 'CUSTOM';
    }
  }
}

/// Represents a date range for filtering metrics
class TimeRange extends Equatable {
  final DateTime start;
  final DateTime end;
  final TimeRangePreset preset;

  const TimeRange({
    required this.start,
    required this.end,
    required this.preset,
  });

  /// Creates a TimeRange for "This Month" (start of month to today)
  factory TimeRange.thisMonth() {
    final now = DateTime.now();
    final start = DateTime(now.year, now.month, 1);
    return TimeRange(
      start: start,
      end: now,
      preset: TimeRangePreset.thisMonth,
    );
  }

  /// Creates a TimeRange for "Last Month" (full previous month)
  factory TimeRange.lastMonth() {
    final now = DateTime.now();
    final firstOfThisMonth = DateTime(now.year, now.month, 1);
    final lastMonth = firstOfThisMonth.subtract(const Duration(days: 1));
    final start = DateTime(lastMonth.year, lastMonth.month, 1);
    return TimeRange(
      start: start,
      end: lastMonth,
      preset: TimeRangePreset.lastMonth,
    );
  }

  /// Creates a TimeRange for "Last 30 Days"
  factory TimeRange.last30Days() {
    final now = DateTime.now();
    final start = now.subtract(const Duration(days: 29));
    return TimeRange(
      start: start,
      end: now,
      preset: TimeRangePreset.last30Days,
    );
  }

  /// Creates a TimeRange for "Last 90 Days"
  factory TimeRange.last90Days() {
    final now = DateTime.now();
    final start = now.subtract(const Duration(days: 89));
    return TimeRange(
      start: start,
      end: now,
      preset: TimeRangePreset.last90Days,
    );
  }

  /// Creates a custom TimeRange
  factory TimeRange.custom(DateTime start, DateTime end) {
    return TimeRange(
      start: start,
      end: end,
      preset: TimeRangePreset.custom,
    );
  }

  /// Creates a TimeRange from a preset
  factory TimeRange.fromPreset(TimeRangePreset preset) {
    switch (preset) {
      case TimeRangePreset.thisMonth:
        return TimeRange.thisMonth();
      case TimeRangePreset.lastMonth:
        return TimeRange.lastMonth();
      case TimeRangePreset.last30Days:
        return TimeRange.last30Days();
      case TimeRangePreset.last90Days:
        return TimeRange.last90Days();
      case TimeRangePreset.custom:
        // Default to last 30 days for custom
        return TimeRange.last30Days();
    }
  }

  /// Number of days in this range (inclusive)
  int get days => end.difference(start).inDays + 1;

  /// Format start date as YYYY-MM-DD
  String get startFormatted =>
      '${start.year}-${start.month.toString().padLeft(2, '0')}-${start.day.toString().padLeft(2, '0')}';

  /// Format end date as YYYY-MM-DD
  String get endFormatted =>
      '${end.year}-${end.month.toString().padLeft(2, '0')}-${end.day.toString().padLeft(2, '0')}';

  /// Human-readable date range
  String get displayRange {
    final startStr =
        '${_monthAbbrev(start.month)} ${start.day}, ${start.year}';
    final endStr = '${_monthAbbrev(end.month)} ${end.day}, ${end.year}';
    return '$startStr - $endStr';
  }

  String _monthAbbrev(int month) {
    const months = [
      'Jan',
      'Feb',
      'Mar',
      'Apr',
      'May',
      'Jun',
      'Jul',
      'Aug',
      'Sep',
      'Oct',
      'Nov',
      'Dec'
    ];
    return months[month - 1];
  }

  @override
  List<Object?> get props => [start, end, preset];
}
