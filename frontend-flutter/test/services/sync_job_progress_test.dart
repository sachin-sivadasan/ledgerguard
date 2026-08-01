import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/services/sync_status_service.dart';

// APPS-3: /sync/jobs returns completed_items/total_items per job (no progress_pct),
// so progress must derive from those; the parent full_sync reports 0/0.
void main() {
  SyncJob job(Map<String, dynamic> extra) => SyncJob.fromJson({
        'id': 'j',
        'app_id': 'a',
        'status': 'processing',
        ...extra,
      });

  test('progress derives from completed_items / total_items', () {
    final j = job({'job_type': 'status_sync', 'completed_items': 2154, 'total_items': 2930});
    expect((j.progress * 100).round(), 74); // ~73.5% → 74
    expect(j.isFullSync, isFalse);
  });

  test('parent full_sync with 0/0 yields 0 progress (no divide-by-zero)', () {
    final j = job({'job_type': 'full_sync', 'completed_items': 0, 'total_items': 0});
    expect(j.progress, 0.0);
    expect(j.isFullSync, isTrue);
  });

  test('falls back to progress_pct when no item counts', () {
    final j = job({'progress_pct': 42});
    expect(j.progress, closeTo(0.42, 1e-9));
  });

  test('progress is clamped to [0,1]', () {
    final over = job({'completed_items': 50, 'total_items': 40});
    expect(over.progress, 1.0);
  });
}
