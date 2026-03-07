import '../entities/chat_message.dart';

/// SSE event from the chat API.
class ChatSSEEvent {
  final String type; // "tool_call", "tool_result", "response", "error"
  final Map<String, dynamic> data;

  const ChatSSEEvent({required this.type, required this.data});
}

/// Repository interface for the AI chat endpoint.
abstract class ChatRepository {
  /// Sends messages to the chat API and yields SSE events as they arrive.
  Stream<ChatSSEEvent> sendMessage({
    required List<ChatMessage> messages,
    String? scopedModule,
    String? appId,
  });
}
