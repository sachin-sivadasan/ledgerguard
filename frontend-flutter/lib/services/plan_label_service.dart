import 'package:dio/dio.dart';

import '../core/network/api_client.dart';

/// One price tier detected among an app's un-named subscriptions, with its derived
/// pseudo-label, any saved custom label, and how many active customers are on it.
class PlanTier {
  final String billingInterval;
  final int priceCents;
  final String key;
  final String pseudoLabel;
  final String label; // developer-assigned name, "" when unset
  final int customers;

  const PlanTier({
    required this.billingInterval,
    required this.priceCents,
    required this.key,
    required this.pseudoLabel,
    required this.label,
    required this.customers,
  });

  factory PlanTier.fromJson(Map<String, dynamic> json) => PlanTier(
        billingInterval: json['billingInterval'] as String? ?? '',
        priceCents: (json['priceCents'] as num?)?.toInt() ?? 0,
        key: json['key'] as String? ?? '',
        pseudoLabel: json['pseudoLabel'] as String? ?? '',
        label: json['label'] as String? ?? '',
        customers: (json['customers'] as num?)?.toInt() ?? 0,
      );
}

/// The nameable tiers plus a count of minor tiers (prorations/one-offs) the backend hid.
class PlanTiersResult {
  final List<PlanTier> tiers;
  final int hiddenTiers;
  const PlanTiersResult({required this.tiers, required this.hiddenTiers});
}

class PlanLabelService {
  final ApiClient _client;

  PlanLabelService(this._client);

  Future<PlanTiersResult> fetchTiers(String appId, {CancelToken? cancelToken}) async {
    final response = await _client.get(
      '/api/v1/apps/$appId/plan-labels',
      cancelToken: cancelToken,
    );
    final data = response.data as Map<String, dynamic>;
    final tiers = data['tiers'] as List<dynamic>?;
    return PlanTiersResult(
      tiers: (tiers ?? const [])
          .map((e) => PlanTier.fromJson(e as Map<String, dynamic>))
          .toList(),
      hiddenTiers: (data['hiddenTiers'] as num?)?.toInt() ?? 0,
    );
  }

  /// Saves the full label set (each entry echoes a tier's interval + price with the
  /// developer's label; the backend drops blank labels and replaces the app's set).
  Future<void> saveTiers(String appId, List<PlanTier> tiers,
      {CancelToken? cancelToken}) async {
    await _client.put(
      '/api/v1/apps/$appId/plan-labels',
      data: {
        'labels': tiers
            .map((t) => {
                  'billingInterval': t.billingInterval,
                  'priceCents': t.priceCents,
                  'label': t.label,
                })
            .toList(),
      },
      cancelToken: cancelToken,
    );
  }
}
