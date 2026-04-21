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
}
