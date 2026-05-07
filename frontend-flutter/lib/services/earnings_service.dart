import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:intl/intl.dart';

import '../core/network/api_client.dart';
import '../models/earning_model.dart';

class EarningsService {
  final ApiClient _client;

  EarningsService(this._client);

  Future<List<EarningPeriod>> fetchEarnings(String appId) async {
    try {
      final dateFmt = DateFormat('yyyy-MM-dd');
      final now = DateTime.now();
      final start = dateFmt.format(now.subtract(const Duration(days: 365)));
      final end = dateFmt.format(now);
      final response = await _client.get(
        '/api/v1/apps/$appId/earnings',
        queryParameters: {'start': start, 'end': end},
      );
      final list = response.data['earnings'] as List<dynamic>? ?? [];
      return list
          .map((json) => EarningPeriod.fromJson(json as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      debugPrint('[EarningsService] error: ${e.response?.statusCode}');
      return [];
    }
  }
}
