import 'dart:async';

import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/chat_message.dart';
import 'package:ledgerguard/domain/entities/shopify_app.dart';
import 'package:ledgerguard/domain/repositories/app_repository.dart';
import 'package:ledgerguard/domain/repositories/chat_repository.dart';
import 'package:ledgerguard/presentation/blocs/chat/chat.dart';

class MockChatRepository extends Mock implements ChatRepository {}

class MockAppRepository extends Mock implements AppRepository {}

ChatBloc _buildBloc(MockChatRepository mockRepo, MockAppRepository mockAppRepo) {
  return ChatBloc(repository: mockRepo, appRepository: mockAppRepo);
}

void main() {
  late MockChatRepository mockRepo;
  late MockAppRepository mockAppRepo;

  setUp(() {
    mockRepo = MockChatRepository();
    mockAppRepo = MockAppRepository();
    when(() => mockAppRepo.getSelectedApp()).thenAnswer(
      (_) async => const ShopifyApp(id: 'app-123', name: 'Test App'),
    );
  });

  group('ChatBloc', () {
    blocTest<ChatBloc, ChatState>(
      'emits [ChatLoaded(streaming), ChatLoaded(response)] on simple message',
      build: () {
        when(() => mockRepo.sendMessage(
              messages: any(named: 'messages'),
              appId: any(named: 'appId'),
              scopedModule: any(named: 'scopedModule'),
            )).thenAnswer((_) => Stream.fromIterable([
              const ChatSSEEvent(
                type: 'response',
                data: {
                  'message': 'Your MRR is \$1,500',
                  'suggestions': ['Show trend'],
                },
              ),
            ]));

        return _buildBloc(mockRepo, mockAppRepo);
      },
      act: (bloc) => bloc.add(const SendMessageRequested('What is my MRR?')),
      expect: () => [
        isA<ChatLoaded>()
            .having((s) => s.isStreaming, 'isStreaming', true)
            .having((s) => s.messages.length, 'messages count', 2),
        isA<ChatLoaded>()
            .having((s) => s.isStreaming, 'isStreaming', false)
            .having((s) => s.messages.length, 'messages count', 2)
            .having(
              (s) => s.messages.last.content,
              'response content',
              'Your MRR is \$1,500',
            )
            .having(
              (s) => s.messages.last.suggestions,
              'suggestions',
              ['Show trend'],
            ),
      ],
      verify: (_) {
        // Verify app ID was resolved and passed
        verify(() => mockRepo.sendMessage(
              messages: any(named: 'messages'),
              appId: 'app-123',
              scopedModule: any(named: 'scopedModule'),
            )).called(1);
      },
    );

    blocTest<ChatBloc, ChatState>(
      'emits tool_call active tool name during streaming',
      build: () {
        when(() => mockRepo.sendMessage(
              messages: any(named: 'messages'),
              appId: any(named: 'appId'),
              scopedModule: any(named: 'scopedModule'),
            )).thenAnswer((_) => Stream.fromIterable([
              const ChatSSEEvent(
                type: 'tool_call',
                data: {'tool_name': 'risk__get_risk_summary'},
              ),
              const ChatSSEEvent(
                type: 'tool_result',
                data: {'tool_name': 'risk__get_risk_summary', 'content': '{}'},
              ),
              const ChatSSEEvent(
                type: 'response',
                data: {'message': 'Done'},
              ),
            ]));

        return _buildBloc(mockRepo, mockAppRepo);
      },
      act: (bloc) => bloc.add(const SendMessageRequested('Show risk')),
      expect: () => [
        isA<ChatLoaded>().having((s) => s.isStreaming, 'isStreaming', true),
        isA<ChatLoaded>().having(
            (s) => s.activeToolName, 'activeToolName', 'risk__get_risk_summary'),
        isA<ChatLoaded>().having((s) => s.activeToolName, 'activeToolName', null),
        isA<ChatLoaded>()
            .having((s) => s.isStreaming, 'isStreaming', false)
            .having((s) => s.messages.last.content, 'content', 'Done'),
      ],
    );

    blocTest<ChatBloc, ChatState>(
      'emits ChatError on error event',
      build: () {
        when(() => mockRepo.sendMessage(
              messages: any(named: 'messages'),
              appId: any(named: 'appId'),
              scopedModule: any(named: 'scopedModule'),
            )).thenAnswer((_) => Stream.fromIterable([
              const ChatSSEEvent(
                type: 'error',
                data: {'message': 'AI unavailable'},
              ),
            ]));

        return _buildBloc(mockRepo, mockAppRepo);
      },
      act: (bloc) => bloc.add(const SendMessageRequested('hello')),
      expect: () => [
        isA<ChatLoaded>().having((s) => s.isStreaming, 'isStreaming', true),
        isA<ChatError>().having((s) => s.message, 'message', 'AI unavailable'),
      ],
    );

    blocTest<ChatBloc, ChatState>(
      'ClearChatRequested resets to ChatInitial',
      build: () => _buildBloc(mockRepo, mockAppRepo),
      seed: () => ChatLoaded(messages: [ChatMessage.user('hi')]),
      act: (bloc) => bloc.add(const ClearChatRequested()),
      expect: () => [isA<ChatInitial>()],
    );

    blocTest<ChatBloc, ChatState>(
      'retry after error deduplicates trailing user message',
      build: () {
        when(() => mockRepo.sendMessage(
              messages: any(named: 'messages'),
              appId: any(named: 'appId'),
              scopedModule: any(named: 'scopedModule'),
            )).thenAnswer((_) => Stream.fromIterable([
              const ChatSSEEvent(
                type: 'response',
                data: {'message': 'Your MRR is \$1,500'},
              ),
            ]));

        return _buildBloc(mockRepo, mockAppRepo);
      },
      seed: () => ChatError(
        'AI unavailable',
        previousMessages: [ChatMessage.user('What is my MRR?')],
      ),
      act: (bloc) =>
          bloc.add(const SendMessageRequested('What is my MRR?')),
      verify: (bloc) {
        final captured = verify(() => mockRepo.sendMessage(
              messages: captureAny(named: 'messages'),
              appId: any(named: 'appId'),
              scopedModule: any(named: 'scopedModule'),
            )).captured;
        final sentMessages = captured.first as List<ChatMessage>;
        final userMessages =
            sentMessages.where((m) => m.isUser).toList();
        expect(userMessages.length, 1,
            reason: 'should deduplicate on retry');
      },
    );

    blocTest<ChatBloc, ChatState>(
      'response with data state populates state fields',
      build: () {
        when(() => mockRepo.sendMessage(
              messages: any(named: 'messages'),
              appId: any(named: 'appId'),
              scopedModule: any(named: 'scopedModule'),
            )).thenAnswer((_) => Stream.fromIterable([
              const ChatSSEEvent(
                type: 'response',
                data: {
                  'message': 'Risk summary',
                  'risk_state': {'safe': 40, 'churned': 2},
                },
              ),
            ]));

        return _buildBloc(mockRepo, mockAppRepo);
      },
      act: (bloc) => bloc.add(const SendMessageRequested('Show risk')),
      expect: () => [
        isA<ChatLoaded>(),
        isA<ChatLoaded>().having(
          (s) => s.messages.last.riskState,
          'riskState',
          {'safe': 40, 'churned': 2},
        ),
      ],
    );

    blocTest<ChatBloc, ChatState>(
      'sends null appId when no app is selected',
      build: () {
        when(() => mockAppRepo.getSelectedApp()).thenAnswer((_) async => null);
        when(() => mockRepo.sendMessage(
              messages: any(named: 'messages'),
              appId: any(named: 'appId'),
              scopedModule: any(named: 'scopedModule'),
            )).thenAnswer((_) => Stream.fromIterable([
              const ChatSSEEvent(
                type: 'response',
                data: {'message': 'Please select an app'},
              ),
            ]));

        return _buildBloc(mockRepo, mockAppRepo);
      },
      act: (bloc) => bloc.add(const SendMessageRequested('hello')),
      verify: (_) {
        verify(() => mockRepo.sendMessage(
              messages: any(named: 'messages'),
              appId: null,
              scopedModule: any(named: 'scopedModule'),
            )).called(1);
      },
    );
  });
}
