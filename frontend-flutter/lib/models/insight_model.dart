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

  factory AiInsight.fromJson(Map<String, dynamic> json) {
    return AiInsight(
      id: json['id'].toString(),
      date: DateTime.parse(
          json['date'] as String? ?? DateTime.now().toIso8601String()),
      title: json['title'] as String? ?? '',
      summary: json['summary'] as String? ?? '',
      severity: _parseSeverity(json['severity'] as String? ?? 'INFO'),
    );
  }

  static InsightSeverity _parseSeverity(String s) {
    switch (s.toUpperCase()) {
      case 'WARNING':
        return InsightSeverity.warning;
      case 'CRITICAL':
        return InsightSeverity.critical;
      default:
        return InsightSeverity.info;
    }
  }
}
