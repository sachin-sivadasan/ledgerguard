import 'package:flutter/foundation.dart';
import '../mock_data/mock_insights.dart';
import '../models/insight_model.dart';

class ChatMessage {
  final String text;
  final bool isUser;
  final DateTime timestamp;

  ChatMessage({required this.text, required this.isUser, DateTime? timestamp})
      : timestamp = timestamp ?? DateTime.now();
}

class InsightsProvider extends ChangeNotifier {
  final List<ChatMessage> _messages = [];

  List<AiInsight> get insights => mockInsights;
  List<ChatMessage> get messages => _messages;

  void sendMessage(String text) {
    _messages.add(ChatMessage(text: text, isUser: true));
    notifyListeners();

    // Simulate AI response
    Future.delayed(const Duration(milliseconds: 800), () {
      _messages.add(ChatMessage(
        text: _generateResponse(text),
        isUser: false,
      ));
      notifyListeners();
    });
  }

  String _generateResponse(String query) {
    final q = query.toLowerCase();
    if (q.contains('mrr') || q.contains('revenue')) {
      return 'Your current MRR is trending upward at approximately \$5,800. The primary growth driver is InventorySync Pro, contributing 45% of total recurring revenue.';
    }
    if (q.contains('churn') || q.contains('risk')) {
      return 'You have 12 subscriptions at risk: 8 with one missed cycle, 4 with two missed cycles. The re-engagement email sequence has a 42% recovery rate — I recommend starting there.';
    }
    if (q.contains('forecast')) {
      return 'Based on current trends, your expected MRR next month is \$5,980 (range: \$5,083–\$6,877). The optimistic scenario assumes 2 new Enterprise subscribers.';
    }
    return 'I can help you analyze your revenue data, track churn risk, review forecasts, and identify growth opportunities. What would you like to know?';
  }
}
