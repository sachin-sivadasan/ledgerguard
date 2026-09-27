/// Maps an FCM message `data` payload (from the backend, see SendCriticalAlert /
/// SendDailySummary) to a go_router path. Anything missing, malformed, or unknown
/// falls back to the Dashboard so a notification tap can never crash or dead-end.
String routeForData(Map<String, dynamic> data) {
  switch (data['type']) {
    case 'risk_alert':
      final id = data['subscription_id'];
      if (id is String && id.isNotEmpty) {
        // Safe to route directly: tokens are per-user (unregistered on logout) and the
        // subscription screen enforces access server-side, so this can't expose cross-org data.
        return '/subscriptions/$id';
      }
      return '/'; // risk alert without a subscription id -> dashboard
    case 'daily_summary':
    default:
      return '/'; // daily summary, unknown, or missing type -> dashboard
  }
}
