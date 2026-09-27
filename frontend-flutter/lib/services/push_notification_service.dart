import 'dart:async';

import '../core/network/api_client.dart';

/// Registers this device's FCM token with the backend so it can receive push,
/// keeps it fresh on rotation, and unregisters it on logout.
///
/// The FCM plugin is injected as plain callables (`getToken`, `requestPermission`,
/// `onTokenRefresh`) so this service is unit-testable without the platform plugin —
/// `main.dart` supplies the real `FirebaseMessaging.instance` implementations.
class PushNotificationService {
  PushNotificationService({
    required this.api,
    required this.getToken,
    required this.requestPermission,
    required this.platform,
    this.onTokenRefresh,
  });

  final ApiClient api;
  final Future<String?> Function() getToken;
  final Future<bool> Function() requestPermission;
  final Stream<String> Function()? onTokenRefresh;

  /// "ios" | "android" | "web" — must match the backend Platform enum.
  final String platform;

  String? _token;
  StreamSubscription<String>? _refreshSub;

  /// Requests permission, obtains the FCM token, registers it with the backend,
  /// and subscribes to token refreshes. Safe to call when already registered.
  /// No-ops (does not register) if permission is denied.
  Future<void> registerToken() async {
    try {
      if (!await requestPermission()) return; // denied -> no push, app still works
      final token = await getToken();
      if (token == null || token.isEmpty) return;
      await _register(token);
      _refreshSub ??= onTokenRefresh?.call().listen(_register);
    } catch (_) {
      // Push is best-effort and this runs fire-and-forget after login: any platform
      // /plugin failure (e.g. web without a configured service worker + VAPID key)
      // must never surface. main.dart also no-ops the plugin on web.
    }
  }

  /// Unregisters the current token (call before signing out so a shared device
  /// stops receiving the previous user's alerts).
  Future<void> unregister() async {
    await _refreshSub?.cancel();
    _refreshSub = null;
    final token = _token;
    if (token == null) return;
    _token = null;
    try {
      await api.delete('/api/v1/devices', data: {'device_token': token});
    } catch (_) {
      // best-effort: sign-out proceeds regardless of network state
    }
  }

  Future<void> _register(String token) async {
    _token = token;
    try {
      await api.post('/api/v1/devices',
          data: {'device_token': token, 'platform': platform});
    } catch (_) {
      // offline / backend down: stay unregistered; retried on next launch or refresh
    }
  }
}
