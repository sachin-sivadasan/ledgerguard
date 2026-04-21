import 'package:flutter/foundation.dart';
import '../mock_data/mock_apps.dart';
import '../mock_data/mock_reviews.dart';
import '../models/app_model.dart';
import '../models/review_model.dart';

class AppsProvider extends ChangeNotifier {
  bool _demoMode = true;

  bool get demoMode => _demoMode;
  void setDemoMode(bool value) {
    _demoMode = value;
    notifyListeners();
  }

  List<ShopifyApp> get apps => _demoMode ? mockApps : [];

  List<AppReview> getReviewsForApp(String appId) =>
      mockReviews.where((r) => r.appId == appId).toList()
        ..sort((a, b) => b.date.compareTo(a.date));

  List<AppReview> get allReviews => mockReviews.toList()
    ..sort((a, b) => b.date.compareTo(a.date));

  double avgRatingForApp(String appId) {
    final reviews = getReviewsForApp(appId);
    if (reviews.isEmpty) return 0;
    return reviews.map((r) => r.rating).reduce((a, b) => a + b) / reviews.length;
  }

  Map<int, int> ratingDistribution(String appId) {
    final reviews = getReviewsForApp(appId);
    final dist = <int, int>{1: 0, 2: 0, 3: 0, 4: 0, 5: 0};
    for (final r in reviews) {
      dist[r.rating] = (dist[r.rating] ?? 0) + 1;
    }
    return dist;
  }
}
