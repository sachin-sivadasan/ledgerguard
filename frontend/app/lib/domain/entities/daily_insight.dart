import 'package:equatable/equatable.dart';

/// AI-generated daily insight for the app
class DailyInsight extends Equatable {
  /// Executive summary text
  final String summary;

  /// When the insight was generated
  final DateTime generatedAt;

  /// Optional key takeaway points
  final List<String> keyPoints;

  const DailyInsight({
    required this.summary,
    required this.generatedAt,
    this.keyPoints = const [],
  });

  /// Create from JSON map
  factory DailyInsight.fromJson(Map<String, dynamic> json) {
    return DailyInsight(
      summary: json['summary'] as String? ?? '',
      generatedAt: json['generated_at'] != null
          ? DateTime.parse(json['generated_at'] as String)
          : DateTime.now(),
      keyPoints: (json['key_points'] as List<dynamic>?)
              ?.map((e) => e as String)
              .toList() ??
          [],
    );
  }

  /// Format generation time as relative or absolute
  String get formattedGeneratedAt {
    final now = DateTime.now();
    final diff = now.difference(generatedAt);

    if (diff.inMinutes < 60) {
      return '${diff.inMinutes} min ago';
    } else if (diff.inHours < 24) {
      return '${diff.inHours} hours ago';
    } else if (diff.inDays == 1) {
      return 'Yesterday';
    } else {
      return '${generatedAt.month}/${generatedAt.day}/${generatedAt.year}';
    }
  }

  bool get hasKeyPoints => keyPoints.isNotEmpty;

  @override
  List<Object?> get props => [summary, generatedAt, keyPoints];
}
