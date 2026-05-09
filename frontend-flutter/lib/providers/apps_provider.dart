import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../mock_data/mock_apps.dart';
import '../mock_data/mock_reviews.dart';
import '../models/app_model.dart';
import '../models/review_model.dart';
import '../services/app_service.dart';
import '../services/user_preferences_service.dart';

class AppsProvider extends ChangeNotifier {
  final AppService _appService;
  final UserPreferencesService? _prefsService;
  static const _demoPrefKey = 'demo_mode';

  bool _demoMode = false;
  bool _isLoading = false;
  String? _error;

  List<ShopifyApp> _liveApps = [];
  List<AppReview> _liveReviews = [];
  String? _selectedAppId;

  AppsProvider(this._appService, [this._prefsService]) {
    _loadDemoPref();
  }

  String? get selectedAppId => _selectedAppId;

  Future<void> _loadDemoPref() async {
    final prefs = await SharedPreferences.getInstance();
    final saved = prefs.getBool(_demoPrefKey);
    if (saved != null && saved != _demoMode) {
      _demoMode = saved;
      notifyListeners();
    }
  }

  bool get demoMode => _demoMode;
  bool get isLoading => _isLoading;
  String? get error => _error;

  void setDemoMode(bool value) {
    _demoMode = value;
    SharedPreferences.getInstance().then((p) => p.setBool(_demoPrefKey, value));
    notifyListeners();
  }

  List<ShopifyApp> get apps => _demoMode ? mockApps : _liveApps;

  Future<void> loadApps() async {
    debugPrint('[AppsProvider] loadApps called – demoMode=$_demoMode isLoading=$_isLoading');
    if (_demoMode || _isLoading) return;
    _isLoading = true;
    _error = null;
    notifyListeners();
    try {
      _liveApps = await _appService.fetchApps();
      debugPrint('[AppsProvider] loadApps success – ${_liveApps.length} apps loaded');

      // Resolve selected app from backend preference
      if (_liveApps.isNotEmpty && _selectedAppId == null) {
        final savedAppId = await _prefsService?.getDefaultApp();
        if (savedAppId != null &&
            _liveApps.any((a) => a.id == savedAppId)) {
          _selectedAppId = savedAppId;
        } else {
          _selectedAppId = _liveApps.first.id;
          if (_liveApps.length == 1) {
            _prefsService?.setDefaultApp(_selectedAppId!);
          }
        }
      }
    } catch (e) {
      _error = e.toString();
      debugPrint('[AppsProvider] loadApps error – $e');
    }
    _isLoading = false;
    notifyListeners();
  }

  void setSelectedApp(String appId) {
    _selectedAppId = appId;
    _prefsService?.setDefaultApp(appId); // fire-and-forget
    notifyListeners();
  }

  Future<void> loadReviews(String appId) async {
    if (_demoMode) return;
    try {
      _liveReviews = await _appService.fetchReviews(appId);
      notifyListeners();
    } catch (e) {
      _error = e.toString();
      notifyListeners();
    }
  }

  List<AppReview> getReviewsForApp(String appId) {
    if (_demoMode) {
      return mockReviews.where((r) => r.appId == appId).toList()
        ..sort((a, b) => b.date.compareTo(a.date));
    }
    return _liveReviews.where((r) => r.appId == appId).toList()
      ..sort((a, b) => b.date.compareTo(a.date));
  }

  List<AppReview> get allReviews {
    final reviews = _demoMode ? mockReviews.toList() : _liveReviews.toList();
    return reviews..sort((a, b) => b.date.compareTo(a.date));
  }

  double avgRatingForApp(String appId) {
    final reviews = getReviewsForApp(appId);
    if (reviews.isEmpty) return 0;
    return reviews.map((r) => r.rating).reduce((a, b) => a + b) /
        reviews.length;
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
