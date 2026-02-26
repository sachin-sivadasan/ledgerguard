import '../entities/risk_summary.dart';

/// Repository interface for risk data
abstract class RiskRepository {
  /// Fetch risk summary for the selected app
  Future<RiskSummary?> fetchRiskSummary();
}

/// Base exception for risk repository errors
class RiskException implements Exception {
  final String message;
  final String code;

  const RiskException(this.message, {this.code = 'unknown'});

  @override
  String toString() => 'RiskException: $message (code: $code)';
}

/// Exception when no app is selected
class NoAppSelectedRiskException extends RiskException {
  const NoAppSelectedRiskException()
      : super('No app selected. Please select an app first.',
            code: 'no-app-selected');
}

/// Exception when user is not authenticated
class UnauthorizedRiskException extends RiskException {
  const UnauthorizedRiskException()
      : super('User must be authenticated to access risk data',
            code: 'unauthorized');
}
