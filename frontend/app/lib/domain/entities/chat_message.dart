import 'package:equatable/equatable.dart';

/// Represents a single message in the chat conversation.
class ChatMessage extends Equatable {
  final String role; // "user", "assistant", "system"
  final String content;
  final DateTime timestamp;
  final bool isLoading;
  final List<String> suggestions;
  final Map<String, dynamic>? subscriptionState;
  final Map<String, dynamic>? metricsState;
  final Map<String, dynamic>? riskState;
  final Map<String, dynamic>? storeHealthState;
  final Map<String, dynamic>? earningsState;

  const ChatMessage({
    required this.role,
    required this.content,
    required this.timestamp,
    this.isLoading = false,
    this.suggestions = const [],
    this.subscriptionState,
    this.metricsState,
    this.riskState,
    this.storeHealthState,
    this.earningsState,
  });

  factory ChatMessage.user(String content) {
    return ChatMessage(
      role: 'user',
      content: content,
      timestamp: DateTime.now(),
    );
  }

  factory ChatMessage.assistant(String content, {
    List<String> suggestions = const [],
    Map<String, dynamic>? subscriptionState,
    Map<String, dynamic>? metricsState,
    Map<String, dynamic>? riskState,
    Map<String, dynamic>? storeHealthState,
    Map<String, dynamic>? earningsState,
  }) {
    return ChatMessage(
      role: 'assistant',
      content: content,
      timestamp: DateTime.now(),
      suggestions: suggestions,
      subscriptionState: subscriptionState,
      metricsState: metricsState,
      riskState: riskState,
      storeHealthState: storeHealthState,
      earningsState: earningsState,
    );
  }

  factory ChatMessage.loading() {
    return ChatMessage(
      role: 'assistant',
      content: '',
      timestamp: DateTime.now(),
      isLoading: true,
    );
  }

  bool get isUser => role == 'user';
  bool get isAssistant => role == 'assistant';
  bool get hasSuggestions => suggestions.isNotEmpty;

  /// Whether this message has any data panel state to display.
  bool get hasDataState =>
      subscriptionState != null ||
      metricsState != null ||
      riskState != null ||
      storeHealthState != null ||
      earningsState != null;

  @override
  List<Object?> get props => [
        role,
        content,
        timestamp,
        isLoading,
        suggestions,
        subscriptionState,
        metricsState,
        riskState,
        storeHealthState,
        earningsState,
      ];
}
