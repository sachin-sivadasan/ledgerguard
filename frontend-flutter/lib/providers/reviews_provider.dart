import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../mock_data/mock_reviews.dart';
import '../models/review_model.dart';
import '../services/reviews_service.dart';

class ReviewsProvider extends ChangeNotifier {
  final ReviewsService _service;

  bool _isLoading = false;
  bool _demoMode = false;
  String? _error;
  bool _isServiceUnavailable = false;
  String? _selectedAppId;
  CancelToken? _cancelToken;

  ReviewsReport? _report;

  ReviewsProvider(this._service);

  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isServiceUnavailable => _isServiceUnavailable;
  String? get selectedAppId => _selectedAppId;
  ReviewsReport? get report => _report;

  /// Wired via DemoModeCoordinator. In demo mode the report shows a mock
  /// dataset (consistent with every other screen) and no API call is made.
  void setDemoMode(bool value) {
    _demoMode = value;
    _cancelToken?.cancel('demo mode change');
    _isLoading = false;
    _error = null;
    _isServiceUnavailable = false;
    _report = value ? _mockReport() : null;
    notifyListeners();
  }

  void setSelectedApp(String? appId) {
    _selectedAppId = appId;
    notifyListeners();
    if (!_demoMode && appId != null) {
      loadReport(appId);
    }
  }

  Future<void> loadReport(String appId) async {
    _cancelToken?.cancel('Superseded');
    _cancelToken = CancelToken();
    _isLoading = true;
    _error = null;
    _isServiceUnavailable = false;
    notifyListeners();
    final token = _cancelToken;
    try {
      _report = await _service.fetchReport(appId, cancelToken: token);
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) {
        // A newer load superseded this one and will manage loading state.
        // But if this token is still the active one (a lone cancel with no
        // successor), fall through so we clear the spinner instead of leaving
        // it stuck forever.
        if (token != _cancelToken) return;
      } else if (e.response?.statusCode == 503) {
        _isServiceUnavailable = true;
        _error = 'Service temporarily unavailable.';
      } else {
        _error = e.message ?? e.toString();
      }
    } catch (e) {
      _error = e.toString();
    }
    // Only the most recent request is allowed to settle loading state, so a
    // superseded request can't stomp a newer in-flight load.
    if (token == _cancelToken) {
      _isLoading = false;
      notifyListeners();
    }
  }

  /// Fetches the CSV export bytes for the currently selected app through the
  /// authenticated ApiClient. Returns null if no app is selected.
  Future<Uint8List?> fetchCsvBytes() {
    final appId = _selectedAppId;
    if (appId == null) return Future.value(null);
    return _service.fetchCsvBytes(appId);
  }

  /// Builds a demo report from [mockReviews] so demo mode mirrors the shape of
  /// the live API response (avg rating, distribution, sentiment, recent feed).
  ReviewsReport _mockReport() {
    final reviews = mockReviews;
    final total = reviews.length;

    final sum = reviews.fold<int>(0, (acc, r) => acc + r.rating);
    final avg = total > 0 ? sum / total : 0.0;

    final counts = <int, int>{for (var star = 1; star <= 5; star++) star: 0};
    for (final r in reviews) {
      counts[r.rating] = (counts[r.rating] ?? 0) + 1;
    }
    final distribution = [
      for (var star = 5; star >= 1; star--)
        RatingBucket(rating: star, count: counts[star] ?? 0),
    ];

    var positive = 0, neutral = 0, negative = 0;
    for (final r in reviews) {
      switch (r.sentiment) {
        case ReviewSentiment.positive:
          positive++;
        case ReviewSentiment.neutral:
          neutral++;
        case ReviewSentiment.negative:
          negative++;
      }
    }

    final recent = [...reviews]..sort((a, b) => b.date.compareTo(a.date));

    return ReviewsReport(
      avgRating: avg,
      totalReviews: total,
      distribution: distribution,
      sentiment: ReviewSentimentBreakdown(
        positive: positive,
        neutral: neutral,
        negative: negative,
      ),
      recent: recent.take(12).toList(), // match backend maxRecentReviews
    );
  }
}
