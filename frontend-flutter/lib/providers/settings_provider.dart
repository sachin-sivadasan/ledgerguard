import 'package:flutter/foundation.dart';

import '../services/user_preferences_service.dart';

class SettingsProvider extends ChangeNotifier {
  final UserPreferencesService? _prefsService;

  SettingsProvider([this._prefsService]);

  // Loading state
  bool _isLoading = false;
  bool get isLoading => _isLoading;

  // Notification preferences
  bool _emailAlerts = true;
  bool _slackAlerts = false;
  bool _churnAlerts = true;
  bool _revenueAlerts = true;
  bool _reviewAlerts = true;
  int _riskThresholdDays = 30;

  // Sync schedule
  String _syncFrequency = 'Every 6 hours';
  bool _autoSync = true;

  // Workspace
  String _workspaceName = 'My Shopify Apps';
  String _currency = 'USD';
  String _timezone = 'America/New_York';

  bool get emailAlerts => _emailAlerts;
  bool get slackAlerts => _slackAlerts;
  bool get churnAlerts => _churnAlerts;
  bool get revenueAlerts => _revenueAlerts;
  bool get reviewAlerts => _reviewAlerts;
  int get riskThresholdDays => _riskThresholdDays;
  String get syncFrequency => _syncFrequency;
  bool get autoSync => _autoSync;
  String get workspaceName => _workspaceName;
  String get currency => _currency;
  String get timezone => _timezone;

  /// Load all preferences from the backend API
  Future<void> loadPreferences() async {
    if (_prefsService == null) return;
    _isLoading = true;
    notifyListeners();

    try {
      final results = await Future.wait([
        _prefsService.getNotificationPreferences(),
        _prefsService.getSyncWorkspacePreferences(),
      ]);

      final notifPrefs = results[0] as NotificationPrefs?;
      if (notifPrefs != null) {
        _emailAlerts = notifPrefs.emailEnabled;
        _slackAlerts = notifPrefs.slackEnabled;
        _churnAlerts = notifPrefs.churnAlertsEnabled;
        _revenueAlerts = notifPrefs.revenueAlertsEnabled;
        _reviewAlerts = notifPrefs.reviewAlertsEnabled;
        _riskThresholdDays = notifPrefs.riskThresholdDays;
      }

      final swPrefs = results[1] as SyncWorkspacePrefs?;
      if (swPrefs != null) {
        _autoSync = swPrefs.autoSync;
        _syncFrequency = swPrefs.syncFrequency;
        _workspaceName = swPrefs.workspaceName;
        _currency = swPrefs.currency;
        _timezone = swPrefs.timezone;
      }
    } catch (e) {
      debugPrint('[SettingsProvider] loadPreferences error: $e');
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }

  // --- Notification setters (optimistic + fire-and-forget save) ---

  void setEmailAlerts(bool v) {
    _emailAlerts = v;
    notifyListeners();
    _saveNotificationPreferences();
  }

  void setSlackAlerts(bool v) {
    _slackAlerts = v;
    notifyListeners();
    _saveNotificationPreferences();
  }

  void setChurnAlerts(bool v) {
    _churnAlerts = v;
    notifyListeners();
    _saveNotificationPreferences();
  }

  void setRevenueAlerts(bool v) {
    _revenueAlerts = v;
    notifyListeners();
    _saveNotificationPreferences();
  }

  void setReviewAlerts(bool v) {
    _reviewAlerts = v;
    notifyListeners();
    _saveNotificationPreferences();
  }

  void setRiskThresholdDays(int v) {
    _riskThresholdDays = v;
    notifyListeners();
    _saveNotificationPreferences();
  }

  // --- Sync/workspace setters (optimistic + fire-and-forget save) ---

  void setSyncFrequency(String v) {
    _syncFrequency = v;
    notifyListeners();
    _saveSyncWorkspacePreferences();
  }

  void setAutoSync(bool v) {
    _autoSync = v;
    notifyListeners();
    _saveSyncWorkspacePreferences();
  }

  void setWorkspaceName(String v) {
    _workspaceName = v;
    notifyListeners();
    _saveSyncWorkspacePreferences();
  }

  void setCurrency(String v) {
    _currency = v;
    notifyListeners();
    _saveSyncWorkspacePreferences();
  }

  void setTimezone(String v) {
    _timezone = v;
    notifyListeners();
    _saveSyncWorkspacePreferences();
  }

  // --- Private save helpers ---

  void _saveNotificationPreferences() {
    _prefsService?.saveNotificationPreferences(NotificationPrefs(
      emailEnabled: _emailAlerts,
      slackEnabled: _slackAlerts,
      churnAlertsEnabled: _churnAlerts,
      revenueAlertsEnabled: _revenueAlerts,
      reviewAlertsEnabled: _reviewAlerts,
      riskThresholdDays: _riskThresholdDays,
    ));
  }

  void _saveSyncWorkspacePreferences() {
    _prefsService?.saveSyncWorkspacePreferences(SyncWorkspacePrefs(
      autoSync: _autoSync,
      syncFrequency: _syncFrequency,
      workspaceName: _workspaceName,
      currency: _currency,
      timezone: _timezone,
    ));
  }
}
