import 'package:flutter/foundation.dart';
import 'package:mixpanel_flutter/mixpanel_flutter.dart';

/// Wraps Mixpanel SDK for event tracking across the app.
///
/// Pass a Mixpanel token at build time via --dart-define:
///   flutter run --dart-define=MIXPANEL_TOKEN=your-token
///
/// When the token is empty (default in dev), all calls are no-ops.
class MixpanelService {
  static const _token = String.fromEnvironment('MIXPANEL_TOKEN');

  Mixpanel? _mixpanel;

  bool get isEnabled => _token.isNotEmpty;

  Future<void> init() async {
    if (!isEnabled) {
      if (kDebugMode) debugPrint('[MixpanelService] No token – tracking disabled');
      return;
    }
    _mixpanel = await Mixpanel.init(_token, trackAutomaticEvents: true);
    if (kDebugMode) debugPrint('[MixpanelService] Initialized');
  }

  /// Identify the current user (call after login/signup).
  void identify(String userId, {String? email}) {
    _mixpanel?.identify(userId);
    if (email != null) {
      _mixpanel?.getPeople().set('\$email', email);
    }
  }

  /// Reset identity on sign-out.
  void reset() => _mixpanel?.reset();

  /// Track a named event with optional properties.
  void track(String event, [Map<String, dynamic>? properties]) {
    _mixpanel?.track(event, properties: properties);
  }

  // ── Convenience helpers ──

  void trackLogin(String method) => track('login', {'method': method});

  void trackSignup(String method) => track('signup', {'method': method});

  void trackLogout() => track('logout');

  void trackPageView(String page) => track('page_view', {'page': page});

  void trackDashboardViewed({String? appId}) =>
      track('dashboard_viewed', {'app_id': appId ?? 'all'});

  void trackAppSelected(String appId, String appName) =>
      track('app_selected', {'app_id': appId, 'app_name': appName});
}
