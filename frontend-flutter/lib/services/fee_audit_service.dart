import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';

/// One month of the Fee Audit: actual Shopify cut vs the expected (detected-tier) cut.
class FeeAuditRow {
  final String month;
  final int grossCents;
  final int shopifyCutCents;
  final double effectiveFeePct;
  final int expectedCutCents;
  final int feeVarianceCents;
  final bool feeGuardOk;

  const FeeAuditRow({
    required this.month,
    required this.grossCents,
    required this.shopifyCutCents,
    required this.effectiveFeePct,
    required this.expectedCutCents,
    required this.feeVarianceCents,
    required this.feeGuardOk,
  });

  factory FeeAuditRow.fromJson(Map<String, dynamic> json) => FeeAuditRow(
        month: json['month'] as String? ?? '',
        grossCents: (json['gross_cents'] as num?)?.toInt() ?? 0,
        shopifyCutCents: (json['shopify_cut_cents'] as num?)?.toInt() ?? 0,
        effectiveFeePct: (json['effective_fee_pct'] as num?)?.toDouble() ?? 0,
        expectedCutCents: (json['expected_cut_cents'] as num?)?.toInt() ?? 0,
        feeVarianceCents: (json['fee_variance_cents'] as num?)?.toInt() ?? 0,
        feeGuardOk: json['fee_guard_ok'] as bool? ?? true,
      );
}

/// Full Fee Audit report: the tier verdict, headline KPIs, and the per-month table.
class FeeAuditReport {
  final String currency;
  final String configuredTier;
  final double configuredFeePct;
  final double detectedFeePct;
  final bool tierMatches;
  final int totalGrossCents;
  final int totalCutCents;
  final double effectiveFeePct;
  final int flaggedMonths;
  final int monthsAudited;
  final int savingsVsDefaultCents;
  final List<FeeAuditRow> months;

  const FeeAuditReport({
    required this.currency,
    required this.configuredTier,
    required this.configuredFeePct,
    required this.detectedFeePct,
    required this.tierMatches,
    required this.totalGrossCents,
    required this.totalCutCents,
    required this.effectiveFeePct,
    required this.flaggedMonths,
    required this.monthsAudited,
    required this.savingsVsDefaultCents,
    required this.months,
  });

  /// True when Shopify's fees match the app's rate across every audited month.
  bool get allClear => flaggedMonths == 0 && monthsAudited > 0;

  factory FeeAuditReport.fromJson(Map<String, dynamic> json) => FeeAuditReport(
        currency: json['currency'] as String? ?? 'USD',
        configuredTier: json['configured_tier'] as String? ?? '',
        configuredFeePct: (json['configured_fee_pct'] as num?)?.toDouble() ?? 0,
        detectedFeePct: (json['detected_fee_pct'] as num?)?.toDouble() ?? 0,
        tierMatches: json['tier_matches'] as bool? ?? true,
        totalGrossCents: (json['total_gross_cents'] as num?)?.toInt() ?? 0,
        totalCutCents: (json['total_cut_cents'] as num?)?.toInt() ?? 0,
        effectiveFeePct: (json['effective_fee_pct'] as num?)?.toDouble() ?? 0,
        flaggedMonths: (json['flagged_months'] as num?)?.toInt() ?? 0,
        monthsAudited: (json['months_audited'] as num?)?.toInt() ?? 0,
        savingsVsDefaultCents:
            (json['savings_vs_default_cents'] as num?)?.toInt() ?? 0,
        months: (json['months'] as List<dynamic>?)
                ?.map((e) => FeeAuditRow.fromJson(e as Map<String, dynamic>))
                .toList() ??
            const [],
      );

  static FeeAuditReport empty() => const FeeAuditReport(
        currency: 'USD',
        configuredTier: '',
        configuredFeePct: 0,
        detectedFeePct: 0,
        tierMatches: true,
        totalGrossCents: 0,
        totalCutCents: 0,
        effectiveFeePct: 0,
        flaggedMonths: 0,
        monthsAudited: 0,
        savingsVsDefaultCents: 0,
        months: [],
      );
}

class FeeAuditService {
  final ApiClient _client;

  FeeAuditService(this._client);

  Future<FeeAuditReport> fetchReport(String appId,
      {int months = 6, CancelToken? cancelToken}) async {
    final response = await _client.get(
      '/api/v1/apps/$appId/reports/fee-audit',
      queryParameters: {'months': months},
      cancelToken: cancelToken,
    );
    return FeeAuditReport.fromJson(response.data as Map<String, dynamic>);
  }

  Future<Uint8List> fetchCsvBytes(String appId,
      {int months = 6, CancelToken? cancelToken}) async {
    final response = await _client.get<List<int>>(
      '/api/v1/apps/$appId/reports/fee-audit',
      queryParameters: {'months': months, 'format': 'csv'},
      cancelToken: cancelToken,
      options: Options(responseType: ResponseType.bytes),
    );
    return Uint8List.fromList(response.data ?? const <int>[]);
  }
}
