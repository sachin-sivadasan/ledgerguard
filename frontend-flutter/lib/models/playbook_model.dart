class RecoveryAction {
  final String label;
  final String description;
  final bool completed;

  const RecoveryAction({
    required this.label,
    required this.description,
    this.completed = false,
  });
}

class RecoveryPlaybook {
  final String id;
  final String name;
  final String description;
  final List<RecoveryAction> steps;
  final double successRate;

  const RecoveryPlaybook({
    required this.id,
    required this.name,
    required this.description,
    required this.steps,
    required this.successRate,
  });
}
