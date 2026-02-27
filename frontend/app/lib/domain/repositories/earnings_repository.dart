import '../entities/earnings_timeline.dart';

/// Repository interface for earnings timeline data
abstract class EarningsRepository {
  /// Fetch earnings timeline for a specific month
  /// [year] - The year (e.g., 2024)
  /// [month] - The month (1-12)
  /// [mode] - Display mode (combined or split)
  Future<EarningsTimeline> fetchMonthlyEarnings({
    required int year,
    required int month,
    required EarningsMode mode,
  });
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

/// Invalid month requested
class InvalidMonthException extends EarningsException {
  const InvalidMonthException()
      : super('Invalid month. Month must be between 1 and 12.',
            code: 'invalid-month');
}

/// Future month requested
class FutureMonthException extends EarningsException {
  const FutureMonthException()
      : super('Cannot request future months.', code: 'future-month');
}

/// Unauthorized to access earnings
class UnauthorizedEarningsException extends EarningsException {
  const UnauthorizedEarningsException()
      : super('Please sign in to view earnings.', code: 'unauthorized');
}
