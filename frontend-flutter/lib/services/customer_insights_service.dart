import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';

/// One MRR bucket over the active customer base.
class RevenueBand {
  final String label;
  final int customers;
  final int mrrCents;
  final double pctOfCustomers;

  const RevenueBand({
    required this.label,
    required this.customers,
    required this.mrrCents,
    required this.pctOfCustomers,
  });

  factory RevenueBand.fromJson(Map<String, dynamic> json) => RevenueBand(
        label: json['label'] as String? ?? '',
        customers: (json['customers'] as num?)?.toInt() ?? 0,
        mrrCents: (json['mrrCents'] as num?)?.toInt() ?? 0,
        pctOfCustomers: (json['pctOfCustomers'] as num?)?.toDouble() ?? 0,
      );
}

/// One risk bucket (SAFE / AT_RISK / CHURNED) across the whole base.
class RiskSegment {
  final String riskState;
  final int customers;
  final int mrrCents;

  const RiskSegment({
    required this.riskState,
    required this.customers,
    required this.mrrCents,
  });

  factory RiskSegment.fromJson(Map<String, dynamic> json) => RiskSegment(
        riskState: json['riskState'] as String? ?? '',
        customers: (json['customers'] as num?)?.toInt() ?? 0,
        mrrCents: (json['mrrCents'] as num?)?.toInt() ?? 0,
      );
}

/// One plan's slice of the active base, split by risk (the crosstab).
class PlanRiskRow {
  final String planName;
  final int customers;
  final int safeCount;
  final int atRiskCount;
  final int mrrCents;
  final int atRiskMrrCents;

  const PlanRiskRow({
    required this.planName,
    required this.customers,
    required this.safeCount,
    required this.atRiskCount,
    required this.mrrCents,
    required this.atRiskMrrCents,
  });

  factory PlanRiskRow.fromJson(Map<String, dynamic> json) => PlanRiskRow(
        planName: json['planName'] as String? ?? '',
        customers: (json['customers'] as num?)?.toInt() ?? 0,
        safeCount: (json['safeCount'] as num?)?.toInt() ?? 0,
        atRiskCount: (json['atRiskCount'] as num?)?.toInt() ?? 0,
        mrrCents: (json['mrrCents'] as num?)?.toInt() ?? 0,
        atRiskMrrCents: (json['atRiskMrrCents'] as num?)?.toInt() ?? 0,
      );
}

/// One of the highest-MRR active customers.
class TopCustomer {
  final String shopName;
  final String planName;
  final int mrrCents;
  final String riskState;

  const TopCustomer({
    required this.shopName,
    required this.planName,
    required this.mrrCents,
    required this.riskState,
  });

  factory TopCustomer.fromJson(Map<String, dynamic> json) => TopCustomer(
        shopName: json['shopName'] as String? ?? '',
        planName: json['planName'] as String? ?? '',
        mrrCents: (json['mrrCents'] as num?)?.toInt() ?? 0,
        riskState: json['riskState'] as String? ?? '',
      );
}

class CustomerInsights {
  final String currency;
  final int totalCustomers;
  final int activeMrrCents;
  final int atRiskCustomers;
  final int atRiskMrrCents;
  final List<RevenueBand> revenueBands;
  final List<RiskSegment> riskSegments;
  final List<PlanRiskRow> planRisk;
  final List<TopCustomer> topCustomers;

  const CustomerInsights({
    required this.currency,
    required this.totalCustomers,
    required this.activeMrrCents,
    required this.atRiskCustomers,
    required this.atRiskMrrCents,
    required this.revenueBands,
    required this.riskSegments,
    required this.planRisk,
    required this.topCustomers,
  });

  factory CustomerInsights.fromJson(Map<String, dynamic> json) {
    List<T> list<T>(String key, T Function(Map<String, dynamic>) f) =>
        (json[key] as List<dynamic>?)
            ?.map((e) => f(e as Map<String, dynamic>))
            .toList() ??
        const [];
    return CustomerInsights(
      currency: json['currency'] as String? ?? 'USD',
      totalCustomers: (json['totalCustomers'] as num?)?.toInt() ?? 0,
      activeMrrCents: (json['activeMrrCents'] as num?)?.toInt() ?? 0,
      atRiskCustomers: (json['atRiskCustomers'] as num?)?.toInt() ?? 0,
      atRiskMrrCents: (json['atRiskMrrCents'] as num?)?.toInt() ?? 0,
      revenueBands: list('revenueBands', RevenueBand.fromJson),
      riskSegments: list('riskSegments', RiskSegment.fromJson),
      planRisk: list('planRisk', PlanRiskRow.fromJson),
      topCustomers: list('topCustomers', TopCustomer.fromJson),
    );
  }

  static CustomerInsights empty() => const CustomerInsights(
        currency: 'USD',
        totalCustomers: 0,
        activeMrrCents: 0,
        atRiskCustomers: 0,
        atRiskMrrCents: 0,
        revenueBands: [],
        riskSegments: [],
        planRisk: [],
        topCustomers: [],
      );
}

class CustomerInsightsService {
  final ApiClient _client;

  CustomerInsightsService(this._client);

  Future<CustomerInsights> fetchReport(String appId,
      {CancelToken? cancelToken}) async {
    final response = await _client.get(
      '/api/v1/apps/$appId/reports/customer-insights',
      cancelToken: cancelToken,
    );
    return CustomerInsights.fromJson(response.data as Map<String, dynamic>);
  }

  Future<Uint8List> fetchCsvBytes(String appId, {CancelToken? cancelToken}) async {
    final response = await _client.get<List<int>>(
      '/api/v1/apps/$appId/reports/customer-insights',
      queryParameters: {'format': 'csv'},
      cancelToken: cancelToken,
      options: Options(responseType: ResponseType.bytes),
    );
    return Uint8List.fromList(response.data ?? const <int>[]);
  }
}
