import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';
import '../models/review_model.dart';

/// A single rating bucket in the distribution (e.g. 5★ → 88 reviews).
class RatingBucket {
  final int rating;
  final int count;

  RatingBucket({required this.rating, required this.count});

  factory RatingBucket.fromJson(Map<String, dynamic> json) {
    return RatingBucket(
      rating: (json['rating'] as num?)?.toInt() ?? 0,
      count: (json['count'] as num?)?.toInt() ?? 0,
    );
  }
}

/// Sentiment breakdown of reviews (positive / neutral / negative counts).
class ReviewSentimentBreakdown {
  final int positive;
  final int neutral;
  final int negative;

  ReviewSentimentBreakdown({
    required this.positive,
    required this.neutral,
    required this.negative,
  });

  factory ReviewSentimentBreakdown.fromJson(Map<String, dynamic> json) {
    return ReviewSentimentBreakdown(
      positive: (json['positive'] as num?)?.toInt() ?? 0,
      neutral: (json['neutral'] as num?)?.toInt() ?? 0,
      negative: (json['negative'] as num?)?.toInt() ?? 0,
    );
  }

  static ReviewSentimentBreakdown empty() =>
      ReviewSentimentBreakdown(positive: 0, neutral: 0, negative: 0);
}

/// Full Reviews report payload.
class ReviewsReport {
  final double avgRating;
  final int totalReviews;
  final List<RatingBucket> distribution;
  final ReviewSentimentBreakdown sentiment;
  final List<AppReview> recent;

  ReviewsReport({
    required this.avgRating,
    required this.totalReviews,
    required this.distribution,
    required this.sentiment,
    required this.recent,
  });

  factory ReviewsReport.fromJson(Map<String, dynamic> json) {
    return ReviewsReport(
      avgRating: (json['avgRating'] as num?)?.toDouble() ?? 0,
      totalReviews: (json['totalReviews'] as num?)?.toInt() ?? 0,
      distribution: (json['distribution'] as List<dynamic>?)
              ?.map((e) => RatingBucket.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      sentiment: json['sentiment'] is Map<String, dynamic>
          ? ReviewSentimentBreakdown.fromJson(
              json['sentiment'] as Map<String, dynamic>)
          : ReviewSentimentBreakdown.empty(),
      recent: (json['recent'] as List<dynamic>?)
              ?.map((e) => AppReview.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  static ReviewsReport empty() => ReviewsReport(
        avgRating: 0,
        totalReviews: 0,
        distribution: const [],
        sentiment: ReviewSentimentBreakdown.empty(),
        recent: const [],
      );
}

class ReviewsService {
  final ApiClient _client;

  ReviewsService(this._client);

  Future<ReviewsReport> fetchReport(
    String appId, {
    CancelToken? cancelToken,
  }) async {
    final response = await _client.get(
      '/api/v1/apps/$appId/reports/reviews',
      cancelToken: cancelToken,
    );
    return ReviewsReport.fromJson(response.data as Map<String, dynamic>);
  }

  /// Fetches the CSV export of the reviews report through the authenticated
  /// [ApiClient] (Firebase Bearer token injected by the Dio interceptor).
  ///
  /// Returns the raw response bytes so the caller can trigger a client-side
  /// download without relying on an external browser navigation (which would
  /// 401 because it carries no auth header).
  Future<Uint8List> fetchCsvBytes(
    String appId, {
    CancelToken? cancelToken,
  }) async {
    final response = await _client.get<List<int>>(
      '/api/v1/apps/$appId/reports/reviews',
      queryParameters: const {'format': 'csv'},
      cancelToken: cancelToken,
      options: Options(responseType: ResponseType.bytes),
    );
    return Uint8List.fromList(response.data ?? const <int>[]);
  }
}
