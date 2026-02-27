import '../entities/daily_insight.dart';

/// Repository interface for AI insights
abstract class InsightRepository {
  /// Fetch the daily insight for the selected app
  Future<DailyInsight?> fetchDailyInsight();
}

/// Base exception for insight repository errors
class InsightException implements Exception {
  final String message;
  final String code;

  const InsightException(this.message, {this.code = 'unknown'});

  @override
  String toString() => 'InsightException: $message (code: $code)';
}

/// Exception when no app is selected
class NoAppSelectedInsightException extends InsightException {
  const NoAppSelectedInsightException()
      : super('No app selected. Please select an app first.',
            code: 'no-app-selected');
}

/// Exception when user is not authenticated
class UnauthorizedInsightException extends InsightException {
  const UnauthorizedInsightException()
      : super('User must be authenticated to access insights',
            code: 'unauthorized');
}

/// Exception when user doesn't have PRO tier
class ProRequiredInsightException extends InsightException {
  const ProRequiredInsightException()
      : super('PRO plan required to access AI insights',
            code: 'pro-required');
}
