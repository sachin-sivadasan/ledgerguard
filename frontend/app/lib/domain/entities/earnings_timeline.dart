import 'package:equatable/equatable.dart';

/// Mode for earnings display
enum EarningsMode {
  /// Show only total amounts
  combined,

  /// Show subscription and usage breakdown
  split,
}

/// Single day's earnings entry
class EarningsEntry extends Equatable {
  /// Date in YYYY-MM-DD format
  final String date;

  /// Total amount in cents
  final int totalAmountCents;

  /// Subscription amount in cents (only in split mode)
  final int subscriptionAmountCents;

  /// Usage amount in cents (only in split mode)
  final int usageAmountCents;

  const EarningsEntry({
    required this.date,
    required this.totalAmountCents,
    this.subscriptionAmountCents = 0,
    this.usageAmountCents = 0,
  });

  /// Format total as currency string
  String get formattedTotal => _formatCurrency(totalAmountCents);

  /// Format subscription as currency string
  String get formattedSubscription => _formatCurrency(subscriptionAmountCents);

  /// Format usage as currency string
  String get formattedUsage => _formatCurrency(usageAmountCents);

  /// Get day of month (1-31)
  int get dayOfMonth {
    final parts = date.split('-');
    if (parts.length == 3) {
      return int.tryParse(parts[2]) ?? 1;
    }
    return 1;
  }

  String _formatCurrency(int cents) {
    final dollars = cents / 100;
    if (dollars >= 1000) {
      return '\$${(dollars / 1000).toStringAsFixed(1)}K';
    }
    return '\$${dollars.toStringAsFixed(0)}';
  }

  factory EarningsEntry.fromJson(Map<String, dynamic> json) {
    return EarningsEntry(
      date: json['date'] as String,
      totalAmountCents: json['total_amount_cents'] as int,
      subscriptionAmountCents: json['subscription_amount_cents'] as int? ?? 0,
      usageAmountCents: json['usage_amount_cents'] as int? ?? 0,
    );
  }

  @override
  List<Object?> get props => [
        date,
        totalAmountCents,
        subscriptionAmountCents,
        usageAmountCents,
      ];
}

/// Earnings timeline for a date range
class EarningsTimeline extends Equatable {
  /// Start date in YYYY-MM-DD format
  final String startDate;

  /// End date in YYYY-MM-DD format
  final String endDate;

  /// Daily earnings entries sorted by date
  final List<EarningsEntry> earnings;

  const EarningsTimeline({
    required this.startDate,
    required this.endDate,
    required this.earnings,
  });

  /// Get total earnings for the period
  int get totalEarnings =>
      earnings.fold(0, (sum, e) => sum + e.totalAmountCents);

  /// Get total subscription earnings
  int get totalSubscription =>
      earnings.fold(0, (sum, e) => sum + e.subscriptionAmountCents);

  /// Get total usage earnings
  int get totalUsage => earnings.fold(0, (sum, e) => sum + e.usageAmountCents);

  /// Format total earnings as currency
  String get formattedTotal => _formatCurrency(totalEarnings);

  /// Get maximum daily earnings (for chart scaling)
  int get maxDailyTotal {
    if (earnings.isEmpty) return 0;
    return earnings.map((e) => e.totalAmountCents).reduce((a, b) => a > b ? a : b);
  }

  /// Get display date range (e.g., "Jan 1 - Jan 31, 2024")
  String get displayDateRange {
    final start = _parseDate(startDate);
    final end = _parseDate(endDate);
    if (start == null || end == null) return '$startDate - $endDate';

    final startStr = '${_monthAbbrev(start.month)} ${start.day}';
    final endStr = '${_monthAbbrev(end.month)} ${end.day}, ${end.year}';
    return '$startStr - $endStr';
  }

  DateTime? _parseDate(String dateStr) {
    try {
      return DateTime.parse(dateStr);
    } catch (_) {
      return null;
    }
  }

  String _monthAbbrev(int month) {
    const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
    return months[month - 1];
  }

  String _formatCurrency(int cents) {
    final dollars = cents / 100;
    if (dollars >= 1000000) {
      return '\$${(dollars / 1000000).toStringAsFixed(2)}M';
    } else if (dollars >= 1000) {
      return '\$${(dollars / 1000).toStringAsFixed(1)}K';
    }
    return '\$${dollars.toStringAsFixed(2)}';
  }

  factory EarningsTimeline.fromJson(Map<String, dynamic> json) {
    return EarningsTimeline(
      startDate: json['start_date'] as String,
      endDate: json['end_date'] as String,
      earnings: (json['earnings'] as List<dynamic>)
          .map((e) => EarningsEntry.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  @override
  List<Object?> get props => [startDate, endDate, earnings];
}
