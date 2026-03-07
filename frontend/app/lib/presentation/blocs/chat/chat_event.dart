import 'package:equatable/equatable.dart';

abstract class ChatEvent extends Equatable {
  const ChatEvent();

  @override
  List<Object?> get props => [];
}

/// User sends a message.
class SendMessageRequested extends ChatEvent {
  final String message;

  const SendMessageRequested(this.message);

  @override
  List<Object?> get props => [message];
}

/// User taps a suggestion chip.
class SuggestionTapped extends ChatEvent {
  final String suggestion;

  const SuggestionTapped(this.suggestion);

  @override
  List<Object?> get props => [suggestion];
}

/// Clear the conversation.
class ClearChatRequested extends ChatEvent {
  const ClearChatRequested();
}
