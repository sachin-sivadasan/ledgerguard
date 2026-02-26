import '../../domain/entities/risk_summary.dart';
import '../../domain/repositories/risk_repository.dart';

/// Mock implementation of RiskRepository for testing
class MockRiskRepository implements RiskRepository {
  /// Flag to simulate fetch error
  bool shouldFail = false;

  /// Flag to return empty data
  bool returnEmpty = false;

  /// Simulated network delay in milliseconds
  int delayMs = 100;

  /// Custom risk summary to return
  RiskSummary? customSummary;

  @override
  Future<RiskSummary?> fetchRiskSummary() async {
    await Future.delayed(Duration(milliseconds: delayMs));

    if (shouldFail) {
      throw const RiskException('Network error');
    }

    if (returnEmpty) {
      return null;
    }

    return customSummary ?? _defaultSummary;
  }

  static const RiskSummary _defaultSummary = RiskSummary(
    safeCount: 842,
    oneCycleMissedCount: 45,
    twoCyclesMissedCount: 18,
    churnedCount: 12,
    revenueAtRiskCents: 1850000,
  );

  void reset() {
    shouldFail = false;
    returnEmpty = false;
    customSummary = null;
  }
}
