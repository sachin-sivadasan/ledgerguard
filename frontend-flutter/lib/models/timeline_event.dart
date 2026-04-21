enum TimelineEventType { install, transaction, riskChange, note }

class TimelineEvent {
  final DateTime date;
  final TimelineEventType type;
  final String description;

  const TimelineEvent({
    required this.date,
    required this.type,
    required this.description,
  });
}
