import 'package:flutter/foundation.dart';
import '../mock_data/mock_webhooks.dart';
import '../models/webhook_model.dart';

class WebhookProvider extends ChangeNotifier {
  WebhookSource? _sourceFilter;
  WebhookStatus? _statusFilter;

  WebhookSource? get sourceFilter => _sourceFilter;
  WebhookStatus? get statusFilter => _statusFilter;

  List<WebhookEvent> get webhooks {
    var list = mockWebhooks.toList();

    if (_sourceFilter != null) {
      list = list.where((e) => e.source == _sourceFilter).toList();
    }
    if (_statusFilter != null) {
      list = list.where((e) => e.status == _statusFilter).toList();
    }

    list.sort((a, b) => b.receivedAt.compareTo(a.receivedAt));
    return list;
  }

  int get totalToday {
    final now = DateTime.now();
    return mockWebhooks
        .where((e) =>
            e.receivedAt.year == now.year &&
            e.receivedAt.month == now.month &&
            e.receivedAt.day == now.day)
        .length;
  }

  int get failedToday {
    final now = DateTime.now();
    return mockWebhooks
        .where((e) =>
            e.status == WebhookStatus.failed &&
            e.receivedAt.year == now.year &&
            e.receivedAt.month == now.month &&
            e.receivedAt.day == now.day)
        .length;
  }

  String get successRate7d {
    final weekAgo = DateTime.now().subtract(const Duration(days: 7));
    final recent = mockWebhooks.where((e) => e.receivedAt.isAfter(weekAgo));
    if (recent.isEmpty) return '0%';
    final successes =
        recent.where((e) => e.status == WebhookStatus.success).length;
    return '${(successes / recent.length * 100).toStringAsFixed(1)}%';
  }

  List<WebhookEvent> get recentWebhooks => webhooks.take(10).toList();

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
    notifyListeners();
  }
}
