import '../core/network/api_client.dart';

class SyncJob {
  final String id;
  final String appId;
  final String status;
  final String? currentWave;
  final String? jobType;
  final int progressPct;
  final int completedItems;
  final int totalItems;
  final String? message;
  final DateTime createdAt;
  final DateTime? completedAt;

  const SyncJob({
    required this.id,
    required this.appId,
    required this.status,
    this.currentWave,
    this.jobType,
    required this.progressPct,
    this.completedItems = 0,
    this.totalItems = 0,
    this.message,
    required this.createdAt,
    this.completedAt,
  });

  factory SyncJob.fromJson(Map<String, dynamic> json) {
    return SyncJob(
      id: json['id'].toString(),
      appId: json['app_id'].toString(),
      status: json['status'] as String? ?? 'unknown',
      currentWave: json['current_wave'] as String?,
      jobType: json['job_type'] as String?,
      progressPct: (json['progress_pct'] as num?)?.toInt() ?? 0,
      completedItems: (json['completed_items'] as num?)?.toInt() ?? 0,
      totalItems: (json['total_items'] as num?)?.toInt() ?? 0,
      message: json['message'] as String?,
      createdAt: DateTime.tryParse(json['created_at']?.toString() ?? '') ??
          DateTime.now(),
      completedAt: json['completed_at'] != null
          ? DateTime.tryParse(json['completed_at'].toString())
          : null,
    );
  }

  /// Fraction complete (0–1). The API sends completed_items/total_items per job
  /// (not progress_pct), so derive from those; fall back to progress_pct.
  double get progress {
    if (totalItems > 0) {
      return (completedItems / totalItems).clamp(0.0, 1.0);
    }
    return (progressPct / 100.0).clamp(0.0, 1.0);
  }

  bool get isFullSync => jobType == 'full_sync';

  bool get isActive =>
      status == 'pending' || status == 'processing' || status == 'queued';
}

/// Which job drives the progress bar vs. which one Cancel targets.
class SyncProgressSelection {
  /// The job whose [SyncJob.progress] represents the sync — the furthest-along
  /// child with real item counts (the parent full_sync reports 0/0).
  final SyncJob lead;

  /// The job to cancel — the parent full_sync when present (cancel cascades to
  /// children), otherwise the lead.
  final SyncJob cancelTarget;

  const SyncProgressSelection(this.lead, this.cancelTarget);

  /// Resolves the lead + cancel target from the active jobs, or null if empty.
  static SyncProgressSelection? from(List<SyncJob> jobs) {
    if (jobs.isEmpty) return null;
    final withItems = jobs.where((j) => j.totalItems > 0).toList();
    final lead = withItems.isNotEmpty
        ? withItems.reduce((a, b) => a.progress >= b.progress ? a : b)
        : jobs.first;
    final cancelTarget =
        jobs.firstWhere((j) => j.isFullSync, orElse: () => lead);
    return SyncProgressSelection(lead, cancelTarget);
  }
}

class SyncStatusService {
  final ApiClient _client;

  SyncStatusService(this._client);

  /// Fetch active sync jobs for a specific app (appUuid = internal UUID).
  Future<List<SyncJob>> getActiveSyncJobs(String appUuid) async {
    final response = await _client.get(
      '/api/v1/sync/jobs',
      queryParameters: {
        'app_id': appUuid,
        'status': 'processing',
        'limit': '5',
      },
    );
    final data = response.data;
    if (data is Map && data['jobs'] is List) {
      return (data['jobs'] as List)
          .map((j) => SyncJob.fromJson(j as Map<String, dynamic>))
          .toList();
    }
    return [];
  }

  /// Trigger a full sync for an app (appUuid = internal UUID).
  Future<SyncJob> enqueueSync(String appUuid) async {
    final response = await _client.post(
      '/api/v1/sync/enqueue/$appUuid',
      queryParameters: {'type': 'full'},
    );
    final data = response.data;
    if (data is Map<String, dynamic>) {
      return SyncJob.fromJson(data);
    }
    throw Exception('Unexpected response from sync enqueue');
  }

  /// Cancel a running sync job.
  Future<void> cancelJob(String jobId) async {
    await _client.post('/api/v1/sync/jobs/$jobId/cancel');
  }
}
