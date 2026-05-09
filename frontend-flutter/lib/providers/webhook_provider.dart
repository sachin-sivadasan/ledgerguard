import 'package:flutter/foundation.dart';
import '../mock_data/mock_webhooks.dart';
import '../models/webhook_model.dart';
import 'events_provider.dart' show TimeRange;

class WebhookProvider extends ChangeNotifier {
  bool _demoMode = false;
  WebhookSource? _sourceFilter;
  WebhookStatus? _statusFilter;
  String? _selectedAppId;
  TimeRange _timeRange = TimeRange.today;

  bool get demoMode => _demoMode;
  WebhookSource? get sourceFilter => _sourceFilter;
  WebhookStatus? get statusFilter => _statusFilter;
  String? get selectedAppId => _selectedAppId;
  TimeRange get timeRange => _timeRange;

  void setDemoMode(bool value) {
    _demoMode = value;
    notifyListeners();
  }

  List<WebhookEvent> get _allWebhooks =>
      _demoMode ? mockWebhooks : <WebhookEvent>[];

  List<WebhookEvent> get webhooks {
    var list = _allWebhooks.toList();

    if (_sourceFilter != null) {
      list = list.where((e) => e.source == _sourceFilter).toList();
    }
    if (_statusFilter != null) {
      list = list.where((e) => e.status == _statusFilter).toList();
    }
    if (_selectedAppId != null) {
      list = list.where((e) => e.appId == _selectedAppId).toList();
    }

    list.sort((a, b) => b.receivedAt.compareTo(a.receivedAt));
    return list;
  }

  DateTime _startOfRange() {
    final now = DateTime.now();
    return switch (_timeRange) {
      TimeRange.today => DateTime(now.year, now.month, now.day),
      TimeRange.thisWeek => now.subtract(const Duration(days: 7)),
      TimeRange.thisMonth => DateTime(now.year, now.month, 1),
    };
  }

  List<WebhookEvent> get _webhooksInRange {
    final cutoff = _startOfRange();
    return _allWebhooks
        .where((e) => e.receivedAt.isAfter(cutoff))
        .toList();
  }

  int get totalInRange => _webhooksInRange.length;

  int get failedInRange =>
      _webhooksInRange.where((e) => e.status == WebhookStatus.failed).length;

  String get successRate {
    final inRange = _webhooksInRange;
    if (inRange.isEmpty) return '0%';
    final successes =
        inRange.where((e) => e.status == WebhookStatus.success).length;
    return '${(successes / inRange.length * 100).toStringAsFixed(1)}%';
  }

  List<WebhookEvent> get recentWebhooks => webhooks.take(10).toList();

  void setTimeRange(TimeRange range) {
    _timeRange = range;
    notifyListeners();
  }

  void setSelectedApp(String? appId) {
    _selectedAppId = appId;
    notifyListeners();
  }

  void setSourceFilter(WebhookSource? source) {
    _sourceFilter = source;
    notifyListeners();
  }

  void setStatusFilter(WebhookStatus? status) {
    _statusFilter = status;
    notifyListeners();
  }

  void clearFilters() {
    _sourceFilter = null;
    _statusFilter = null;
    _selectedAppId = null;
    notifyListeners();
  }
}
