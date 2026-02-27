import '../../domain/entities/daily_insight.dart';
import '../../domain/repositories/insight_repository.dart';

/// Mock implementation of InsightRepository for testing
class MockInsightRepository implements InsightRepository {
  /// Flag to simulate fetch error
  bool shouldFail = false;

  /// Flag to return empty data
  bool returnEmpty = false;

  /// Simulated network delay in milliseconds
  int delayMs = 100;

  /// Custom insight to return
  DailyInsight? customInsight;

  @override
  Future<DailyInsight?> fetchDailyInsight() async {
    await Future.delayed(Duration(milliseconds: delayMs));

    if (shouldFail) {
      throw const InsightException('Network error');
    }

    if (returnEmpty) {
      return null;
    }

    return customInsight ?? _defaultInsight;
  }

  static final DailyInsight _defaultInsight = DailyInsight(
    summary:
        'Your app showed strong performance this week with a 94.2% renewal rate. '
        'Revenue is up 12% compared to last month, driven primarily by new enterprise customers. '
        'Consider focusing on the 45 subscriptions at risk to prevent potential churn.',
    generatedAt: DateTime.now().subtract(const Duration(hours: 2)),
    keyPoints: [
      'Renewal rate increased by 2.1% this week',
      '12 new enterprise subscriptions added',
      '45 subscriptions need attention to prevent churn',
    ],
  );

  void reset() {
    shouldFail = false;
    returnEmpty = false;
    customInsight = null;
  }
}
