import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';
import '../models/paginated_result.dart';
import '../models/transaction_model.dart';

class TransactionService {
  final ApiClient _client;

  TransactionService(this._client);

  Future<PaginatedResult<Transaction>> fetchTransactions(
    String appId, {
    int page = 1,
    int pageSize = 20,
    CancelToken? cancelToken,
  }) async {
    try {
      final response = await _client.get(
        '/api/v1/apps/$appId/transactions',
        queryParameters: {'page': page, 'pageSize': pageSize},
        cancelToken: cancelToken,
      );
      final data = response.data;
      final list = data['transactions'] as List<dynamic>? ?? [];
      return PaginatedResult(
        items: list
            .map((json) => Transaction.fromJson(json as Map<String, dynamic>))
            .toList(),
        total: data['total'] as int? ?? list.length,
        page: data['page'] as int? ?? 1,
        pageSize: data['pageSize'] as int? ?? pageSize,
        totalPages: data['totalPages'] as int? ?? 1,
      );
    } on DioException catch (e) {
      debugPrint('[TransactionService] error: ${e.response?.statusCode}');
      return const PaginatedResult(
          items: [], total: 0, page: 1, pageSize: 20, totalPages: 0);
    }
  }
}
