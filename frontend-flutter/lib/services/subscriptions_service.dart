import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';

/// A single per-plan row in the Subscriptions (ARPU / LTV) report.
class SubscriptionsPlan {
  final String planName;
  final int activeSubs;
  final int mrrCents;
  final int arpuCents;
  final int ltvCents;

  /// Plan's share of total active subscriptions, in [0,1].
  final double pctOfSubs;

  SubscriptionsPlan({
    required this.planName,
    required this.activeSubs,
    required this.mrrCents,
    required this.arpuCents,
    required this.ltvCents,
    required this.pctOfSubs,
  });

  factory SubscriptionsPlan.fromJson(Map<String, dynamic> json) {
    return SubscriptionsPlan(
      planName: json['planName'] as String? ?? '',
      activeSubs: (json['activeSubs'] as num?)?.toInt() ?? 0,
      mrrCents: (json['mrrCents'] as num?)?.toInt() ?? 0,
      arpuCents: (json['arpuCents'] as num?)?.toInt() ?? 0,
      ltvCents: (json['ltvCents'] as num?)?.toInt() ?? 0,
      // Share of active subs — clamp defensively to [0,1].
      pctOfSubs: ((json['pctOfSubs'] as num?)?.toDouble() ?? 0).clamp(0.0, 1.0),
    );
  }
}

/// Full Subscriptions (ARPU / LTV) report payload.
class SubscriptionsReport {
  final String currency;
  final int activeSubs;
  final int activeMrrCents;
  final int arpuCents;

  /// Lifetime value in cents. 0 means "undefined" — the churn rate was 0, so the
  /// UI renders "—" rather than a misleading $0.
  final int ltvCents;

  /// Monthly churn rate used for LTV, in [0,1] (same definition as the Churn report).
  final double churnRate;
  final List<SubscriptionsPlan> plans;

  SubscriptionsReport({
    required this.currency,
    required this.activeSubs,
    required this.activeMrrCents,
    required this.arpuCents,
    required this.ltvCents,
    required this.churnRate,
    required this.plans,
  });

  factory SubscriptionsReport.fromJson(Map<String, dynamic> json) {
    return SubscriptionsReport(
      currency: json['currency'] as String? ?? 'USD',
      activeSubs: (json['activeSubs'] as num?)?.toInt() ?? 0,
      activeMrrCents: (json['activeMrrCents'] as num?)?.toInt() ?? 0,
      arpuCents: (json['arpuCents'] as num?)?.toInt() ?? 0,
      ltvCents: (json['ltvCents'] as num?)?.toInt() ?? 0,
      churnRate: ((json['churnRate'] as num?)?.toDouble() ?? 0).clamp(0.0, 1.0),
      plans: (json['plans'] as List<dynamic>?)
              ?.map((e) => SubscriptionsPlan.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  static SubscriptionsReport empty() => SubscriptionsReport(
        currency: 'USD',
        activeSubs: 0,
        activeMrrCents: 0,
        arpuCents: 0,
        ltvCents: 0,
        churnRate: 0,
        plans: const [],
      );
}

class SubscriptionsService {
  final ApiClient _client;

  SubscriptionsService(this._client);

  // Note: no from/to params — this report is point-in-time (current active base +
  // latest-snapshot churn), so the backend ignores a date range by design.
  Future<SubscriptionsReport> fetchReport(
    String appId, {
    CancelToken? cancelToken,
  }) async {
    final response = await _client.get(
      '/api/v1/apps/$appId/reports/subscriptions',
      cancelToken: cancelToken,
    );
    return SubscriptionsReport.fromJson(response.data as Map<String, dynamic>);
  }

  /// Fetches the CSV export of the subscriptions report through the authenticated
  /// [ApiClient] (Firebase Bearer token injected by the Dio interceptor).
  ///
  /// Returns the raw response bytes so the caller can trigger a client-side
  /// download without relying on an external browser navigation (which would
  /// 401 because it carries no auth header).
  Future<Uint8List> fetchCsvBytes(
    String appId, {
    CancelToken? cancelToken,
  }) async {
    final queryParameters = <String, dynamic>{'format': 'csv'};
    final response = await _client.get<List<int>>(
      '/api/v1/apps/$appId/reports/subscriptions',
      queryParameters: queryParameters,
      cancelToken: cancelToken,
      options: Options(responseType: ResponseType.bytes),
    );
    return Uint8List.fromList(response.data ?? const <int>[]);
  }
}
