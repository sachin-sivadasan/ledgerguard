import 'package:equatable/equatable.dart';

import '../../../domain/entities/chat_message.dart';

abstract class ChatState extends Equatable {
  const ChatState();

  @override
  List<Object?> get props => [];
}

class ChatInitial extends ChatState {
  const ChatInitial();
}

class ChatLoaded extends ChatState {
  final List<ChatMessage> messages;
  final bool isStreaming;
  final String? activeToolName;

  const ChatLoaded({
    required this.messages,
    this.isStreaming = false,
    this.activeToolName,
  });

  ChatLoaded copyWith({
    List<ChatMessage>? messages,
    bool? isStreaming,
    String? activeToolName,
  }) {
    return ChatLoaded(
      messages: messages ?? this.messages,
      isStreaming: isStreaming ?? this.isStreaming,
      activeToolName: activeToolName,
    );
  }

  @override
  List<Object?> get props => [messages, isStreaming, activeToolName];
}

class ChatError extends ChatState {
  final String message;
  final List<ChatMessage> previousMessages;

  const ChatError(this.message, {this.previousMessages = const []});

  @override
  List<Object?> get props => [message, previousMessages];
}
