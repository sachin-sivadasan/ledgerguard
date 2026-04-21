import 'package:flutter/foundation.dart';

class SettingsProvider extends ChangeNotifier {
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

  void setEmailAlerts(bool v) { _emailAlerts = v; notifyListeners(); }
  void setSlackAlerts(bool v) { _slackAlerts = v; notifyListeners(); }
  void setChurnAlerts(bool v) { _churnAlerts = v; notifyListeners(); }
  void setRevenueAlerts(bool v) { _revenueAlerts = v; notifyListeners(); }
  void setReviewAlerts(bool v) { _reviewAlerts = v; notifyListeners(); }
  void setRiskThresholdDays(int v) { _riskThresholdDays = v; notifyListeners(); }
  void setSyncFrequency(String v) { _syncFrequency = v; notifyListeners(); }
  void setAutoSync(bool v) { _autoSync = v; notifyListeners(); }
  void setWorkspaceName(String v) { _workspaceName = v; notifyListeners(); }
  void setCurrency(String v) { _currency = v; notifyListeners(); }
  void setTimezone(String v) { _timezone = v; notifyListeners(); }
}
