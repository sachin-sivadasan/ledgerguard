import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:intl/intl.dart';

import '../core/network/api_client.dart';
import '../models/analytics_model.dart';
import '../models/earning_model.dart';

class EarningsService {
  final ApiClient _client;

  EarningsService(this._client);

  Future<List<EarningPeriod>> fetchEarnings(String appId,
      {CancelToken? cancelToken}) async {
    try {
      final dateFmt = DateFormat('yyyy-MM-dd');
      final now = DateTime.now();
      final start = dateFmt.format(now.subtract(const Duration(days: 365)));
      final end = dateFmt.format(now);
      // Monthly period cards (gross / Shopify cut / net + status) — the daily
      // /earnings feed carries only a net total per day, which can't render the
      // wireframe's per-month breakdown.
      final response = await _client.get(
        '/api/v1/apps/$appId/earnings/periods',
        queryParameters: {'start': start, 'end': end},
        cancelToken: cancelToken,
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

  Future<List<RevenueShareTier>> fetchTiers(
      {CancelToken? cancelToken}) async {
    try {
      final response = await _client.get('/api/v1/tiers',
          cancelToken: cancelToken);
      final list = response.data['tiers'] as List<dynamic>? ?? [];
      return list
          .map((json) =>
              RevenueShareTier.fromJson(json as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      debugPrint('[EarningsService] fetchTiers error: ${e.response?.statusCode}');
      return [];
    }
  }

  Future<FeeBreakdownResponse?> fetchFeeBreakdown(String appId,
      int amountCents,
      {CancelToken? cancelToken}) async {
    try {
      final response = await _client.get(
        '/api/v1/apps/$appId/fees/breakdown',
        queryParameters: {'amount_cents': amountCents},
        cancelToken: cancelToken,
      );
      return FeeBreakdownResponse.fromJson(
          response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      debugPrint(
          '[EarningsService] fetchFeeBreakdown error: ${e.response?.statusCode}');
      return null;
    }
  }

  Future<FeeSummary?> fetchFeeSummary(String appId,
      {String? start, String? end, CancelToken? cancelToken}) async {
    try {
      final params = <String, dynamic>{};
      if (start != null) params['start'] = start;
      if (end != null) params['end'] = end;
      final response = await _client.get(
        '/api/v1/apps/$appId/fees/summary',
        queryParameters: params,
        cancelToken: cancelToken,
      );
      return FeeSummary.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      debugPrint(
          '[EarningsService] fetchFeeSummary error: ${e.response?.statusCode}');
      return null;
    }
  }

  Future<EarningsStatus?> fetchEarningsStatus(String appId,
      {CancelToken? cancelToken}) async {
    try {
      final response = await _client.get(
        '/api/v1/apps/$appId/earnings/status',
        cancelToken: cancelToken,
      );
      return EarningsStatus.fromJson(
          response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      debugPrint(
          '[EarningsService] fetchEarningsStatus error: ${e.response?.statusCode}');
      return null;
    }
  }

  Future<List<ExpenseBreakdown>> fetchMonthlyProfit(String appId,
      {int months = 6, CancelToken? cancelToken}) async {
    try {
      final response = await _client.get(
        '/api/v1/apps/$appId/fees/monthly',
        queryParameters: {'months': months},
        cancelToken: cancelToken,
      );
      final list = response.data['months'] as List<dynamic>? ?? [];
      return list
          .map((json) =>
              ExpenseBreakdown.fromJson(json as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      debugPrint(
          '[EarningsService] fetchMonthlyProfit error: ${e.response?.statusCode}');
      return [];
    }
  }
}
