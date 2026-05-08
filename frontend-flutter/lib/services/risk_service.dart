import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';
import '../models/analytics_model.dart';
import '../models/store_model.dart';

class RiskSummary {
  final RiskDistribution distribution;
  final List<Store> atRiskStores;

  RiskSummary({required this.distribution, required this.atRiskStores});

  factory RiskSummary.fromJson(Map<String, dynamic> json) {
    return RiskSummary(
      distribution: json['distribution'] != null
          ? RiskDistribution.fromJson(
              json['distribution'] as Map<String, dynamic>)
          : const RiskDistribution(
              safe: 0, oneCycle: 0, twoCycle: 0, churned: 0),
      atRiskStores: (json['at_risk_stores'] as List<dynamic>?)
              ?.map((e) => Store.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

class RiskService {
  final ApiClient _client;

  RiskService(this._client);

  Future<RiskSummary> fetchRiskSummary(String appId,
      {CancelToken? cancelToken}) async {
    try {
      final response = await _client.get('/api/v1/apps/$appId/risk/summary',
          cancelToken: cancelToken);
      return RiskSummary.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      debugPrint('[RiskService] error: ${e.response?.statusCode}');
      return RiskSummary(
        distribution: const RiskDistribution(
            safe: 0, oneCycle: 0, twoCycle: 0, churned: 0),
        atRiskStores: [],
      );
    }
  }
}
