import 'dart:async';
import 'dart:convert';

import 'package:dio/dio.dart';

import '../../core/config/app_config.dart';
import '../../domain/entities/chat_message.dart';
import '../../domain/repositories/auth_repository.dart';
import '../../domain/repositories/chat_repository.dart';

/// SSE-based chat repository that streams events from POST /api/v1/chat.
class ApiChatRepository implements ChatRepository {
  final Dio _dio;
  final AuthRepository _authRepository;

  ApiChatRepository({
    Dio? dio,
    required AuthRepository authRepository,
  })  : _dio = dio ?? Dio(BaseOptions(baseUrl: AppConfig.apiBaseUrl)),
        _authRepository = authRepository;

  @override
  Stream<ChatSSEEvent> sendMessage({
    required List<ChatMessage> messages,
    String? scopedModule,
    String? appId,
  }) async* {
    final token = await _authRepository.getIdToken();
    if (token == null) {
      yield const ChatSSEEvent(
        type: 'error',
        data: {'message': 'Not authenticated'},
      );
      return;
    }

    final body = <String, dynamic>{
      'messages': messages
          .where((m) => !m.isLoading)
          .map((m) => {'role': m.role, 'content': m.content})
          .toList(),
    };
    if (scopedModule != null) body['scoped_module'] = scopedModule;
    if (appId != null) body['app_id'] = appId;

    try {
      final response = await _dio.post<ResponseBody>(
        '/api/v1/chat',
        data: body,
        options: Options(
          headers: {
            'Authorization': 'Bearer $token',
            'Content-Type': 'application/json',
            'Accept': 'text/event-stream',
          },
          responseType: ResponseType.stream,
        ),
      );

      final stream = response.data?.stream;
      if (stream == null) {
        yield const ChatSSEEvent(
          type: 'error',
          data: {'message': 'No response stream'},
        );
        return;
      }

      // Parse SSE lines from the byte stream
      String buffer = '';
      await for (final chunk in stream) {
        buffer += utf8.decode(chunk);

        // SSE events are separated by double newlines
        while (buffer.contains('\n\n')) {
          final idx = buffer.indexOf('\n\n');
          final eventBlock = buffer.substring(0, idx);
          buffer = buffer.substring(idx + 2);

          for (final line in eventBlock.split('\n')) {
            if (line.startsWith('data: ')) {
              final jsonStr = line.substring(6);
              try {
                final parsed = json.decode(jsonStr) as Map<String, dynamic>;
                final type = parsed['type'] as String? ?? 'unknown';
                final data = parsed['data'] as Map<String, dynamic>? ?? {};
                yield ChatSSEEvent(type: type, data: data);
              } catch (_) {
                // Skip unparseable SSE data
              }
            }
          }
        }
      }
    } on DioException catch (e) {
      final message = e.response?.statusCode == 401
          ? 'Session expired. Please sign in again.'
          : e.response?.statusCode == 503
              ? 'AI service is not available. Please try later.'
              : 'Failed to connect to chat: ${e.message}';
      yield ChatSSEEvent(type: 'error', data: {'message': message});
    }
  }
}
