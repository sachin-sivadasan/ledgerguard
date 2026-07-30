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
    String? chargeType,
    CancelToken? cancelToken,
  }) async {
    try {
      final params = <String, dynamic>{'page': page, 'pageSize': pageSize};
      if (chargeType != null && chargeType.isNotEmpty) {
        params['chargeType'] = chargeType;
      }
      final response = await _client.get(
        '/api/v1/apps/$appId/transactions',
        queryParameters: params,
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
      // Propagate so the provider can short-circuit cancels and surface real errors,
      // instead of silently rendering an empty list as "0 transactions".
      debugPrint('[TransactionService] error: ${e.response?.statusCode}');
      rethrow;
    }
  }

  /// Server-side Gross/Net/Cut totals over the FULL filtered set (not the loaded page).
  Future<TransactionSummary?> fetchSummary(
    String appId, {
    String? chargeType,
    CancelToken? cancelToken,
  }) async {
    try {
      final params = <String, dynamic>{};
      if (chargeType != null && chargeType.isNotEmpty) {
        params['chargeType'] = chargeType;
      }
      final response = await _client.get(
        '/api/v1/apps/$appId/transactions/summary',
        queryParameters: params,
        cancelToken: cancelToken,
      );
      return TransactionSummary.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      debugPrint('[TransactionService] summary error: ${e.response?.statusCode}');
      return null;
    }
  }
}
