enum ReviewSentiment { positive, neutral, negative }

class AppReview {
  final String id;
  final String appId;
  final String author;
  final int rating;
  final String text;
  final DateTime date;
  final ReviewSentiment sentiment;
  final String location;
  final String timeUsing;

  const AppReview({
    required this.id,
    required this.appId,
    required this.author,
    required this.rating,
    required this.text,
    required this.date,
    required this.sentiment,
    this.location = '',
    this.timeUsing = '',
  });

  factory AppReview.fromJson(Map<String, dynamic> json) {
    return AppReview(
      id: json['id'].toString(),
      appId: json['app_id'].toString(),
      author: json['author'] as String? ?? '',
      rating: json['rating'] as int? ?? 0,
      text: json['text'] as String? ?? '',
      date: DateTime.parse(
          json['date'] as String? ?? DateTime.now().toIso8601String()),
      sentiment: _parseSentiment(json['sentiment'] as String? ?? 'NEUTRAL'),
      location: json['location'] as String? ?? '',
      timeUsing: json['time_using'] as String? ?? '',
    );
  }

  static ReviewSentiment _parseSentiment(String s) {
    switch (s.toUpperCase()) {
      case 'POSITIVE':
        return ReviewSentiment.positive;
      case 'NEGATIVE':
        return ReviewSentiment.negative;
      default:
        return ReviewSentiment.neutral;
    }
  }
}
