import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../services/mobile_reviews_service.dart';

class MobileReviewsProvider extends ChangeNotifier {
  final MobileReviewsService _service;

  MobileReviewsProvider(this._service);

  bool _isLoading = false;
  bool _isSaving = false;
  bool _demoMode = false;
  String? _error;
  bool _isServiceUnavailable = false;
  String? _selectedAppId;
  CancelToken? _cancelToken;
  MobileReviewsData? _data;

  bool get isLoading => _isLoading;
  bool get isSaving => _isSaving;
  String? get error => _error;
  bool get isServiceUnavailable => _isServiceUnavailable;
  String? get selectedAppId => _selectedAppId;
  MobileReviewsData? get data => _data;

  void setDemoMode(bool value) {
    _demoMode = value;
    _cancelToken?.cancel('demo mode change');
    _isLoading = false;
    _error = null;
    _isServiceUnavailable = false;
    _data = value ? _mock() : null;
    notifyListeners();
  }

  void setSelectedApp(String? appId) {
    _selectedAppId = appId;
    notifyListeners();
    if (!_demoMode && appId != null) load(appId);
  }

  Future<void> load(String appId) async {
    _cancelToken?.cancel('Superseded');
    _cancelToken = CancelToken();
    _isLoading = true;
    _error = null;
    _isServiceUnavailable = false;
    notifyListeners();
    final token = _cancelToken;
    try {
      _data = await _service.fetch(appId, cancelToken: token);
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) {
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
    if (token == _cancelToken) {
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<bool> saveLinks(String appStore, String googlePlay) async {
    final appId = _selectedAppId;
    if (appId == null || _demoMode) return false;
    _isSaving = true;
    _error = null;
    notifyListeners();
    var ok = false;
    try {
      await _service.saveLinks(appId, appStore: appStore, googlePlay: googlePlay);
      ok = true;
    } on DioException catch (e) {
      // Backend error envelope is {"error": {"message": ...}}.
      final data = e.response?.data;
      _error = data is Map && data['error'] is Map
          ? (data['error']['message']?.toString() ?? e.message)
          : e.message;
    } catch (e) {
      _error = e.toString();
    }
    _isSaving = false;
    notifyListeners();
    if (ok) await load(appId);
    return ok;
  }

  MobileReviewsData _mock() => MobileReviewsData(
        iosAppId: '310633997',
        playPackage: 'com.example.app',
        appStore: const StoreBlock(
          linked: true,
          appName: 'Your App',
          iconUrl: '',
          ratingValue: 4.6,
          ratingCount: 18402,
          installs: '',
          storeUrl: '',
          reviewsAvailable: true,
          positive: 2,
          neutral: 1,
          negative: 1,
          reviews: [
            MobileReview(author: 'merchant_ann', rating: 5, title: 'Huge time saver', body: 'Setup took minutes and support is great.', version: '3.2.1'),
            MobileReview(author: 'shopowner_ken', rating: 4, title: 'Solid', body: 'Works well, would love dark mode.', version: '3.2.1'),
            MobileReview(author: 'devkat', rating: 3, title: 'Okay', body: 'Does the job but a bit slow on sync.', version: '3.2.0'),
            MobileReview(author: 'grumpy_cat', rating: 1, title: 'Crashes', body: 'Crashed twice on my Android tablet.', version: '3.1.9'),
          ],
          error: '',
        ),
        googlePlay: const StoreBlock(
          linked: true,
          appName: 'Your App',
          iconUrl: '',
          ratingValue: 4.3,
          ratingCount: 5880,
          installs: '500K+',
          storeUrl: '',
          reviewsAvailable: false,
          positive: 0,
          neutral: 0,
          negative: 0,
          reviews: [],
          error: '',
        ),
      );
}
