import 'dart:async';

import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../domain/entities/chat_message.dart';
import '../../../domain/repositories/chat_repository.dart';
import 'chat_event.dart';
import 'chat_state.dart';

class ChatBloc extends Bloc<ChatEvent, ChatState> {
  final ChatRepository _repository;
  final String? appId;

  ChatBloc({
    required ChatRepository repository,
    this.appId,
  })  : _repository = repository,
        super(const ChatInitial()) {
    on<SendMessageRequested>(_onSendMessage);
    on<SuggestionTapped>(_onSuggestionTapped);
    on<ClearChatRequested>(_onClearChat);
  }

  Future<void> _onSendMessage(
    SendMessageRequested event,
    Emitter<ChatState> emit,
  ) async {
    final currentMessages = state is ChatLoaded
        ? (state as ChatLoaded).messages
        : state is ChatError
            ? (state as ChatError).previousMessages
            : <ChatMessage>[];

    final userMsg = ChatMessage.user(event.message);
    final loadingMsg = ChatMessage.loading();

    final updatedMessages = [...currentMessages, userMsg, loadingMsg];
    emit(ChatLoaded(messages: updatedMessages, isStreaming: true));

    final sseStream = _repository.sendMessage(
      messages: [...currentMessages, userMsg],
      appId: appId,
    );

    await emit.forEach<ChatSSEEvent>(
      sseStream,
      onData: (sseEvent) {
        final current = state;
        if (current is! ChatLoaded) return current;

        switch (sseEvent.type) {
          case 'tool_call':
            final toolName = sseEvent.data['tool_name'] as String? ?? '';
            return current.copyWith(activeToolName: toolName);

          case 'tool_result':
            return current.copyWith(activeToolName: null);

          case 'response':
            final msg = sseEvent.data['message'] as String? ?? '';
            final suggestions =
                (sseEvent.data['suggestions'] as List<dynamic>?)
                        ?.map((e) => e as String)
                        .toList() ??
                    [];

            final responseMsg = ChatMessage.assistant(
              msg,
              suggestions: suggestions,
              subscriptionState:
                  _castMap(sseEvent.data['subscription_state']),
              metricsState: _castMap(sseEvent.data['metrics_state']),
              riskState: _castMap(sseEvent.data['risk_state']),
              storeHealthState:
                  _castMap(sseEvent.data['store_health_state']),
              earningsState: _castMap(sseEvent.data['earnings_state']),
            );

            // Replace loading message with response
            final msgs = current.messages
                .where((m) => !m.isLoading)
                .toList()
              ..add(responseMsg);
            return ChatLoaded(messages: msgs, isStreaming: false);

          case 'error':
            final errorMsg =
                sseEvent.data['message'] as String? ?? 'Unknown error';
            final msgs =
                current.messages.where((m) => !m.isLoading).toList();
            return ChatError(errorMsg, previousMessages: msgs);

          default:
            return current;
        }
      },
    );
  }

  Future<void> _onSuggestionTapped(
    SuggestionTapped event,
    Emitter<ChatState> emit,
  ) async {
    add(SendMessageRequested(event.suggestion));
  }

  void _onClearChat(ClearChatRequested event, Emitter<ChatState> emit) {
    emit(const ChatInitial());
  }

  Map<String, dynamic>? _castMap(dynamic value) {
    if (value is Map<String, dynamic>) return value;
    if (value is Map) return value.cast<String, dynamic>();
    return null;
  }
}
