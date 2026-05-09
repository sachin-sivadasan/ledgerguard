import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';
import '../models/event_model.dart';
import '../models/paginated_result.dart';

class EventsService {
  final ApiClient _client;

  EventsService(this._client);

  Future<PaginatedResult<AppEvent>> fetchEvents(
    String appId, {
    int page = 1,
    int pageSize = 20,
    String? storeDomain,
    String? eventType,
    DateTime? since,
    CancelToken? cancelToken,
  }) async {
    try {
      final params = <String, dynamic>{'page': page, 'pageSize': pageSize};
      if (storeDomain != null && storeDomain.isNotEmpty) {
        params['storeDomain'] = storeDomain;
      }
      if (eventType != null && eventType.isNotEmpty) {
        params['eventType'] = eventType;
      }
      if (since != null) {
        params['since'] = since.toUtc().toIso8601String();
      }
      final response = await _client.get(
        '/api/v1/apps/$appId/events',
        queryParameters: params,
        cancelToken: cancelToken,
      );
      final data = response.data;
      final list = data['events'] as List<dynamic>? ?? [];
      return PaginatedResult(
        items: list
            .map((json) => AppEvent.fromJson(json as Map<String, dynamic>))
            .toList(),
        total: data['total'] as int? ?? list.length,
        page: data['page'] as int? ?? 1,
        pageSize: data['pageSize'] as int? ?? pageSize,
        totalPages: data['totalPages'] as int? ?? 1,
      );
    } on DioException catch (e) {
      debugPrint('[EventsService] error: ${e.response?.statusCode}');
      return const PaginatedResult(
          items: [], total: 0, page: 1, pageSize: 20, totalPages: 0);
    }
  }
}
