enum InsightSeverity { info, warning, critical }

class AiInsight {
  final String id;
  final DateTime date;
  final String title;
  final String summary;
  final InsightSeverity severity;

  const AiInsight({
    required this.id,
    required this.date,
    required this.title,
    required this.summary,
    required this.severity,
  });
}
