import 'package:dio/dio.dart';

import '../core/network/api_client.dart';

class MobileReview {
  final String author;
  final int rating;
  final String title;
  final String body;
  final String version;

  const MobileReview({
    required this.author,
    required this.rating,
    required this.title,
    required this.body,
    required this.version,
  });

  factory MobileReview.fromJson(Map<String, dynamic> j) => MobileReview(
        author: j['author'] as String? ?? '',
        rating: (j['rating'] as num?)?.toInt() ?? 0,
        title: j['title'] as String? ?? '',
        body: j['body'] as String? ?? '',
        version: j['version'] as String? ?? '',
      );
}

/// One store's ratings + (Apple-only) reviews.
class StoreBlock {
  final bool linked;
  final String appName;
  final String iconUrl;
  final double ratingValue;
  final int ratingCount;
  final String storeUrl;
  final bool reviewsAvailable;
  final List<MobileReview> reviews;
  final int positive;
  final int neutral;
  final int negative;
  final String error;

  const StoreBlock({
    required this.linked,
    required this.appName,
    required this.iconUrl,
    required this.ratingValue,
    required this.ratingCount,
    required this.storeUrl,
    required this.reviewsAvailable,
    required this.reviews,
    required this.positive,
    required this.neutral,
    required this.negative,
    required this.error,
  });

  factory StoreBlock.fromJson(Map<String, dynamic> j) => StoreBlock(
        linked: j['linked'] as bool? ?? false,
        appName: j['appName'] as String? ?? '',
        iconUrl: j['iconUrl'] as String? ?? '',
        ratingValue: (j['ratingValue'] as num?)?.toDouble() ?? 0,
        ratingCount: (j['ratingCount'] as num?)?.toInt() ?? 0,
        storeUrl: j['storeUrl'] as String? ?? '',
        reviewsAvailable: j['reviewsAvailable'] as bool? ?? false,
        reviews: (j['reviews'] as List<dynamic>?)
                ?.map((e) => MobileReview.fromJson(e as Map<String, dynamic>))
                .toList() ??
            const [],
        positive: (j['positive'] as num?)?.toInt() ?? 0,
        neutral: (j['neutral'] as num?)?.toInt() ?? 0,
        negative: (j['negative'] as num?)?.toInt() ?? 0,
        error: j['error'] as String? ?? '',
      );
}

class MobileReviewsData {
  final String iosAppId;
  final String playPackage;
  final StoreBlock? appStore;
  final StoreBlock? googlePlay;

  const MobileReviewsData({
    required this.iosAppId,
    required this.playPackage,
    required this.appStore,
    required this.googlePlay,
  });

  bool get hasAnyLink => iosAppId.isNotEmpty || playPackage.isNotEmpty;

  factory MobileReviewsData.fromJson(Map<String, dynamic> j) => MobileReviewsData(
        iosAppId: j['iosAppId'] as String? ?? '',
        playPackage: j['playPackage'] as String? ?? '',
        appStore: j['appStore'] == null
            ? null
            : StoreBlock.fromJson(j['appStore'] as Map<String, dynamic>),
        googlePlay: j['googlePlay'] == null
            ? null
            : StoreBlock.fromJson(j['googlePlay'] as Map<String, dynamic>),
      );

  static MobileReviewsData empty() => const MobileReviewsData(
      iosAppId: '', playPackage: '', appStore: null, googlePlay: null);
}

class MobileReviewsService {
  final ApiClient _client;

  MobileReviewsService(this._client);

  Future<MobileReviewsData> fetch(String appId, {CancelToken? cancelToken}) async {
    final r = await _client.get('/api/v1/apps/$appId/mobile/reviews',
        cancelToken: cancelToken);
    return MobileReviewsData.fromJson(r.data as Map<String, dynamic>);
  }

  /// Saves the store links (accepts raw ids or full store URLs).
  Future<void> saveLinks(String appId,
      {required String appStore, required String googlePlay}) async {
    await _client.put('/api/v1/apps/$appId/mobile/links',
        data: {'appStore': appStore, 'googlePlay': googlePlay});
  }
}
