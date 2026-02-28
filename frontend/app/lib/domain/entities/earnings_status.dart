import 'package:equatable/equatable.dart';

/// Represents earnings availability for a specific date
class EarningsDateEntry extends Equatable {
  /// Date in YYYY-MM-DD format
  final String date;

  /// Amount in cents
  final int amountCents;

  const EarningsDateEntry({
    required this.date,
    required this.amountCents,
  });

  /// Format as currency string
  String get formattedAmount => _formatCurrency(amountCents);

  /// Parse date to DateTime
  DateTime? get parsedDate {
    try {
      return DateTime.parse(date);
    } catch (_) {
      return null;
    }
  }

  /// Format date for display (e.g., "Jan 15")
  String get displayDate {
    final parsed = parsedDate;
    if (parsed == null) return date;

    const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
                    'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
    return '${months[parsed.month - 1]} ${parsed.day}';
  }

  String _formatCurrency(int cents) {
    final dollars = cents / 100;
    if (dollars >= 1000) {
      return '\$${(dollars / 1000).toStringAsFixed(1)}K';
    }
    return '\$${dollars.toStringAsFixed(2)}';
  }

  factory EarningsDateEntry.fromJson(Map<String, dynamic> json) {
    return EarningsDateEntry(
      date: json['date'] as String,
      amountCents: json['amount_cents'] as int,
    );
  }

  @override
  List<Object?> get props => [date, amountCents];
}

/// Represents the overall earnings availability status
class EarningsStatus extends Equatable {
  /// Total pending earnings in cents (not yet available)
  final int totalPendingCents;

  /// Total available earnings in cents (ready for payout)
  final int totalAvailableCents;

  /// Total paid out earnings in cents (already disbursed)
  final int totalPaidOutCents;

  /// Pending earnings grouped by available date
  final List<EarningsDateEntry> pendingByDate;

  /// Earnings becoming available in the next 30 days
  final List<EarningsDateEntry> upcomingAvailability;

  const EarningsStatus({
    required this.totalPendingCents,
    required this.totalAvailableCents,
    required this.totalPaidOutCents,
    required this.pendingByDate,
    required this.upcomingAvailability,
  });

  /// Total earnings across all statuses
  int get totalCents => totalPendingCents + totalAvailableCents + totalPaidOutCents;

  /// Format total pending as currency
  String get formattedPending => _formatCurrency(totalPendingCents);

  /// Format total available as currency
  String get formattedAvailable => _formatCurrency(totalAvailableCents);

  /// Format total paid out as currency
  String get formattedPaidOut => _formatCurrency(totalPaidOutCents);

  /// Get next available date entry (soonest upcoming)
  EarningsDateEntry? get nextAvailable {
    if (upcomingAvailability.isEmpty) return null;
    return upcomingAvailability.first;
  }

  /// Days until next earnings become available
  int? get daysUntilNextAvailable {
    final next = nextAvailable;
    if (next == null) return null;

    final nextDate = next.parsedDate;
    if (nextDate == null) return null;

    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final diff = nextDate.difference(today).inDays;
    return diff >= 0 ? diff : 0;
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

  factory EarningsStatus.fromJson(Map<String, dynamic> json) {
    return EarningsStatus(
      totalPendingCents: json['total_pending_cents'] as int,
      totalAvailableCents: json['total_available_cents'] as int,
      totalPaidOutCents: json['total_paid_out_cents'] as int,
      pendingByDate: (json['pending_by_date'] as List<dynamic>? ?? [])
          .map((e) => EarningsDateEntry.fromJson(e as Map<String, dynamic>))
          .toList(),
      upcomingAvailability: (json['upcoming_availability'] as List<dynamic>? ?? [])
          .map((e) => EarningsDateEntry.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  /// Empty status for loading/initial state
  static const empty = EarningsStatus(
    totalPendingCents: 0,
    totalAvailableCents: 0,
    totalPaidOutCents: 0,
    pendingByDate: [],
    upcomingAvailability: [],
  );

  @override
  List<Object?> get props => [
        totalPendingCents,
        totalAvailableCents,
        totalPaidOutCents,
        pendingByDate,
        upcomingAvailability,
      ];
}
