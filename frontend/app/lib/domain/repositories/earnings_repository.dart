import '../entities/earnings_status.dart';
import '../entities/earnings_timeline.dart';

/// Repository interface for earnings timeline data
abstract class EarningsRepository {
  /// Fetch earnings timeline for a date range
  /// [startDate] - Start date of the range
  /// [endDate] - End date of the range
  /// [mode] - Display mode (combined or split)
  Future<EarningsTimeline> fetchEarnings({
    required DateTime startDate,
    required DateTime endDate,
    required EarningsMode mode,
  });

  /// Fetch earnings availability status
  /// Returns pending, available, and paid out earnings totals
  Future<EarningsStatus> fetchEarningsStatus();
}

/// Exception for earnings-related errors
class EarningsException implements Exception {
  final String message;
  final String? code;

  const EarningsException(this.message, {this.code});

  @override
  String toString() => message;
}

/// No app selected - cannot fetch earnings
class NoAppSelectedEarningsException extends EarningsException {
  const NoAppSelectedEarningsException()
      : super('No app selected. Please select an app first.',
            code: 'no-app-selected');
}

/// Invalid date range requested
class InvalidDateRangeException extends EarningsException {
  const InvalidDateRangeException()
      : super('Invalid date range. Start date must be before end date.',
            code: 'invalid-date-range');
}

/// Unauthorized to access earnings
class UnauthorizedEarningsException extends EarningsException {
  const UnauthorizedEarningsException()
      : super('Please sign in to view earnings.', code: 'unauthorized');
}
