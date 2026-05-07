import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';
import '../models/transaction_model.dart';

class TransactionService {
  final ApiClient _client;

  TransactionService(this._client);

  Future<List<Transaction>> fetchTransactions(String appId) async {
    try {
      final response = await _client.get('/api/v1/apps/$appId/transactions');
      final list = response.data['transactions'] as List<dynamic>? ?? [];
      return list
          .map((json) => Transaction.fromJson(json as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      debugPrint('[TransactionService] error: ${e.response?.statusCode}');
      return [];
    }
  }
}
