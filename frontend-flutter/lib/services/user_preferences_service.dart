import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';

class DashboardPrefs {
  final List<String> primaryKpis;
  final List<String> secondaryWidgets;

  const DashboardPrefs({
    required this.primaryKpis,
    required this.secondaryWidgets,
  });
}

class NotificationPrefs {
  final bool emailEnabled;
  final bool slackEnabled;
  final bool churnAlertsEnabled;
  final bool revenueAlertsEnabled;
  final bool reviewAlertsEnabled;
  final int riskThresholdDays;

  const NotificationPrefs({
    required this.emailEnabled,
    required this.slackEnabled,
    required this.churnAlertsEnabled,
    required this.revenueAlertsEnabled,
    required this.reviewAlertsEnabled,
    required this.riskThresholdDays,
  });
}

class SyncWorkspacePrefs {
  final bool autoSync;
  final String syncFrequency;
  final String workspaceName;
  final String currency;
  final String timezone;

  const SyncWorkspacePrefs({
    required this.autoSync,
    required this.syncFrequency,
    required this.workspaceName,
    required this.currency,
    required this.timezone,
  });
}

class UserPreferencesService {
  final ApiClient _client;

  UserPreferencesService(this._client);

  Future<String?> getSelectedOrg() async {
    try {
      final response =
          await _client.get('/api/v1/user/preferences/selected-org');
      final id = response.data['selected_org_id'];
      return id is String ? id : null;
    } on DioException catch (e) {
      debugPrint(
          '[UserPreferencesService] getSelectedOrg error: ${e.response?.statusCode}');
      return null;
    }
  }

  Future<void> setSelectedOrg(String orgId) async {
    try {
      await _client.put('/api/v1/user/preferences/selected-org', data: {
        'selected_org_id': orgId,
      });
    } on DioException catch (e) {
      debugPrint(
          '[UserPreferencesService] setSelectedOrg error: ${e.response?.statusCode}');
    }
  }

  Future<String?> getDefaultApp() async {
    try {
      final response =
          await _client.get('/api/v1/user/preferences/default-app');
      final id = response.data['default_app_id'];
      return id is String ? id : null;
    } on DioException catch (e) {
      debugPrint(
          '[UserPreferencesService] getDefaultApp error: ${e.response?.statusCode}');
      return null;
    }
  }

  Future<void> setDefaultApp(String appId) async {
    try {
      await _client.put('/api/v1/user/preferences/default-app', data: {
        'default_app_id': appId,
      });
    } on DioException catch (e) {
      debugPrint(
          '[UserPreferencesService] setDefaultApp error: ${e.response?.statusCode}');
    }
  }

  Future<DashboardPrefs?> getDashboardPreferences() async {
    try {
      final response =
          await _client.get('/api/v1/user/preferences/dashboard');
      final data = response.data;
      return DashboardPrefs(
        primaryKpis: List<String>.from(data['primary_kpis'] ?? []),
        secondaryWidgets: List<String>.from(data['secondary_widgets'] ?? []),
      );
    } on DioException catch (e) {
      debugPrint(
          '[UserPreferencesService] getDashboardPreferences error: ${e.response?.statusCode}');
      return null;
    }
  }

  Future<void> saveDashboardPreferences(
      List<String> primaryKpis, List<String> secondaryWidgets) async {
    try {
      await _client.put('/api/v1/user/preferences/dashboard', data: {
        'primary_kpis': primaryKpis,
        'secondary_widgets': secondaryWidgets,
      });
    } on DioException catch (e) {
      debugPrint(
          '[UserPreferencesService] saveDashboardPreferences error: ${e.response?.statusCode}');
    }
  }

  Future<NotificationPrefs?> getNotificationPreferences() async {
    try {
      final response =
          await _client.get('/api/v1/users/notification-preferences');
      final data = response.data;
      return NotificationPrefs(
        emailEnabled: data['email_enabled'] ?? true,
        slackEnabled: data['slack_enabled'] ?? false,
        churnAlertsEnabled: data['churn_alerts_enabled'] ?? true,
        revenueAlertsEnabled: data['revenue_alerts_enabled'] ?? true,
        reviewAlertsEnabled: data['review_alerts_enabled'] ?? true,
        riskThresholdDays: data['risk_threshold_days'] ?? 30,
      );
    } on DioException catch (e) {
      debugPrint(
          '[UserPreferencesService] getNotificationPreferences error: ${e.response?.statusCode}');
      return null;
    }
  }

  Future<void> saveNotificationPreferences(NotificationPrefs prefs) async {
    try {
      await _client.put('/api/v1/users/notification-preferences', data: {
        'email_enabled': prefs.emailEnabled,
        'slack_enabled': prefs.slackEnabled,
        'churn_alerts_enabled': prefs.churnAlertsEnabled,
        'revenue_alerts_enabled': prefs.revenueAlertsEnabled,
        'review_alerts_enabled': prefs.reviewAlertsEnabled,
        'risk_threshold_days': prefs.riskThresholdDays,
        'critical_alerts_enabled': prefs.churnAlertsEnabled,
        'daily_summary_enabled': prefs.revenueAlertsEnabled,
        'daily_summary_time': '08:00',
        'slack_webhook_url': '',
      });
    } on DioException catch (e) {
      debugPrint(
          '[UserPreferencesService] saveNotificationPreferences error: ${e.response?.statusCode}');
    }
  }

  Future<SyncWorkspacePrefs?> getSyncWorkspacePreferences() async {
    try {
      final response =
          await _client.get('/api/v1/user/preferences/settings');
      final data = response.data;
      return SyncWorkspacePrefs(
        autoSync: data['auto_sync'] ?? true,
        syncFrequency: data['sync_frequency'] ?? 'Every 6 hours',
        workspaceName: data['workspace_name'] ?? 'My Shopify Apps',
        currency: data['currency'] ?? 'USD',
        timezone: data['timezone'] ?? 'America/New_York',
      );
    } on DioException catch (e) {
      debugPrint(
          '[UserPreferencesService] getSyncWorkspacePreferences error: ${e.response?.statusCode}');
      return null;
    }
  }

  Future<void> saveSyncWorkspacePreferences(SyncWorkspacePrefs prefs) async {
    try {
      await _client.put('/api/v1/user/preferences/settings', data: {
        'auto_sync': prefs.autoSync,
        'sync_frequency': prefs.syncFrequency,
        'workspace_name': prefs.workspaceName,
        'currency': prefs.currency,
        'timezone': prefs.timezone,
      });
    } on DioException catch (e) {
      debugPrint(
          '[UserPreferencesService] saveSyncWorkspacePreferences error: ${e.response?.statusCode}');
    }
  }
}
