import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';

/// One month of the ledger reconciliation: does net == gross − fee?
class ReconMonth {
  final String month;
  final int grossCents;
  final int feeCents;
  final int netCents;
  final int expectedNetCents;
  final int residualCents;
  final int txCount;
  final bool reconciled;

  const ReconMonth({
    required this.month,
    required this.grossCents,
    required this.feeCents,
    required this.netCents,
    required this.expectedNetCents,
    required this.residualCents,
    required this.txCount,
    required this.reconciled,
  });

  factory ReconMonth.fromJson(Map<String, dynamic> json) => ReconMonth(
        month: json['month'] as String? ?? '',
        grossCents: (json['gross_cents'] as num?)?.toInt() ?? 0,
        feeCents: (json['fee_cents'] as num?)?.toInt() ?? 0,
        netCents: (json['net_cents'] as num?)?.toInt() ?? 0,
        expectedNetCents: (json['expected_net_cents'] as num?)?.toInt() ?? 0,
        residualCents: (json['residual_cents'] as num?)?.toInt() ?? 0,
        txCount: (json['tx_count'] as num?)?.toInt() ?? 0,
        reconciled: json['reconciled'] as bool? ?? true,
      );
}

class ReconReport {
  final String currency;
  final int totalGrossCents;
  final int totalFeeCents;
  final int totalNetCents;
  final int residualCents;
  final bool reconciled;
  final int monthsReconciled;
  final int monthsFlagged;
  final int monthsAudited;
  final List<ReconMonth> months;

  const ReconReport({
    required this.currency,
    required this.totalGrossCents,
    required this.totalFeeCents,
    required this.totalNetCents,
    required this.residualCents,
    required this.reconciled,
    required this.monthsReconciled,
    required this.monthsFlagged,
    required this.monthsAudited,
    required this.months,
  });

  factory ReconReport.fromJson(Map<String, dynamic> json) => ReconReport(
        currency: json['currency'] as String? ?? 'USD',
        totalGrossCents: (json['total_gross_cents'] as num?)?.toInt() ?? 0,
        totalFeeCents: (json['total_fee_cents'] as num?)?.toInt() ?? 0,
        totalNetCents: (json['total_net_cents'] as num?)?.toInt() ?? 0,
        residualCents: (json['residual_cents'] as num?)?.toInt() ?? 0,
        reconciled: json['reconciled'] as bool? ?? true,
        monthsReconciled: (json['months_reconciled'] as num?)?.toInt() ?? 0,
        monthsFlagged: (json['months_flagged'] as num?)?.toInt() ?? 0,
        monthsAudited: (json['months_audited'] as num?)?.toInt() ?? 0,
        months: (json['months'] as List<dynamic>?)
                ?.map((e) => ReconMonth.fromJson(e as Map<String, dynamic>))
                .toList() ??
            const [],
      );

  static ReconReport empty() => const ReconReport(
        currency: 'USD',
        totalGrossCents: 0,
        totalFeeCents: 0,
        totalNetCents: 0,
        residualCents: 0,
        reconciled: true,
        monthsReconciled: 0,
        monthsFlagged: 0,
        monthsAudited: 0,
        months: [],
      );
}

class LedgerReconService {
  final ApiClient _client;

  LedgerReconService(this._client);

  Future<ReconReport> fetchReport(String appId,
      {int months = 6, CancelToken? cancelToken}) async {
    final response = await _client.get(
      '/api/v1/apps/$appId/reports/ledger-reconciliation',
      queryParameters: {'months': months},
      cancelToken: cancelToken,
    );
    return ReconReport.fromJson(response.data as Map<String, dynamic>);
  }

  Future<Uint8List> fetchCsvBytes(String appId,
      {int months = 6, CancelToken? cancelToken}) async {
    final response = await _client.get<List<int>>(
      '/api/v1/apps/$appId/reports/ledger-reconciliation',
      queryParameters: {'months': months, 'format': 'csv'},
      cancelToken: cancelToken,
      options: Options(responseType: ResponseType.bytes),
    );
    return Uint8List.fromList(response.data ?? const <int>[]);
  }
}
