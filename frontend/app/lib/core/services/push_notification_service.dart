import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';

import '../../domain/repositories/auth_repository.dart';
import '../config/app_config.dart';
import 'package:dio/dio.dart';

/// Handles FCM token retrieval, permission requests, and device registration with backend.
class PushNotificationService {
  final AuthRepository _authRepository;
  final FirebaseMessaging _messaging;

  PushNotificationService({
    required AuthRepository authRepository,
    FirebaseMessaging? messaging,
  })  : _authRepository = authRepository,
        _messaging = messaging ?? FirebaseMessaging.instance;

  /// Initialize push notifications: request permission, get token, register with backend.
  Future<void> initialize() async {
    try {
      // Request permission
      final settings = await _messaging.requestPermission(
        alert: true,
        badge: true,
        sound: true,
      );

      debugPrint('[FCM] Permission status: ${settings.authorizationStatus}');

      if (settings.authorizationStatus != AuthorizationStatus.authorized &&
          settings.authorizationStatus != AuthorizationStatus.provisional) {
        debugPrint('[FCM] Push notifications not authorized');
        return;
      }

      // Get FCM token
      final token = await _messaging.getToken();
      if (token != null) {
        debugPrint('[FCM] Device token: $token');
        await _registerDeviceToken(token);
      }

      // Listen for token refresh
      _messaging.onTokenRefresh.listen((newToken) {
        debugPrint('[FCM] Token refreshed: $newToken');
        _registerDeviceToken(newToken);
      });

      // Handle foreground messages
      FirebaseMessaging.onMessage.listen((RemoteMessage message) {
        debugPrint('[FCM] Foreground message: ${message.notification?.title} - ${message.notification?.body}');
      });
    } catch (e) {
      debugPrint('[FCM] Initialization error: $e');
    }
  }

  /// Register device token with backend API
  Future<void> _registerDeviceToken(String deviceToken) async {
    try {
      final authToken = await _authRepository.getIdToken();
      if (authToken == null) {
        debugPrint('[FCM] No auth token, skipping device registration');
        return;
      }

      final platform = _getPlatform();
      final dio = Dio(BaseOptions(baseUrl: AppConfig.apiBaseUrl));

      final response = await dio.post(
        '/api/v1/devices',
        data: {
          'device_token': deviceToken,
          'platform': platform,
        },
        options: Options(
          headers: {'Authorization': 'Bearer $authToken'},
        ),
      );

      debugPrint('[FCM] Device registered: ${response.statusCode}');
    } catch (e) {
      debugPrint('[FCM] Device registration failed: $e');
    }
  }

  String _getPlatform() {
    if (defaultTargetPlatform == TargetPlatform.iOS) return 'ios';
    if (defaultTargetPlatform == TargetPlatform.android) return 'android';
    return 'web';
  }
}
