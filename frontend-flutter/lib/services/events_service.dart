import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';
import '../models/event_model.dart';

class EventsService {
  final ApiClient _client;

  EventsService(this._client);

  Future<List<AppEvent>> fetchEvents(String appId,
      {CancelToken? cancelToken}) async {
    try {
      final response = await _client.get('/api/v1/apps/$appId/events',
          cancelToken: cancelToken);
      final list = response.data['events'] as List<dynamic>? ?? [];
      return list
          .map((json) => AppEvent.fromJson(json as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      debugPrint('[EventsService] error: ${e.response?.statusCode}');
      return [];
    }
  }
}
