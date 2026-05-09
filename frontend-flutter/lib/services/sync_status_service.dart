import '../core/network/api_client.dart';

class SyncJob {
  final String id;
  final String appId;
  final String status;
  final String? currentWave;
  final int progressPct;
  final String? message;
  final DateTime createdAt;
  final DateTime? completedAt;

  const SyncJob({
    required this.id,
    required this.appId,
    required this.status,
    this.currentWave,
    required this.progressPct,
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
      progressPct: json['progress_pct'] as int? ?? 0,
      message: json['message'] as String?,
      createdAt: DateTime.tryParse(json['created_at']?.toString() ?? '') ??
          DateTime.now(),
      completedAt: json['completed_at'] != null
          ? DateTime.tryParse(json['completed_at'].toString())
          : null,
    );
  }

  bool get isActive =>
      status == 'pending' || status == 'processing' || status == 'queued';
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
