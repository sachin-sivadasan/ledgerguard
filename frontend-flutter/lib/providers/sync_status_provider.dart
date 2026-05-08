import 'dart:async';
import 'package:flutter/foundation.dart';
import '../services/sync_status_service.dart';

class AppSyncState {
  final bool isSyncing;
  final String? message;
  final double? progress;
  final String? jobId;
  final DateTime? lastSyncAt;
  final String? error;

  const AppSyncState({
    this.isSyncing = false,
    this.message,
    this.progress,
    this.jobId,
    this.lastSyncAt,
    this.error,
  });

  static const idle = AppSyncState();
}

/// Watches sync job status for connected apps.
///
/// Uses app.id (UUID) directly for both state keys and API calls.
class SyncStatusProvider extends ChangeNotifier {
  final SyncStatusService _service;
  Timer? _pollTimer;
  bool _disposed = false;

  Map<String, AppSyncState> _appStates = {};

  /// List of app IDs (UUIDs) to watch.
  List<String> _watchedAppIds = [];

  SyncStatusProvider(this._service);

  AppSyncState getState(String appId) =>
      _appStates[appId] ?? AppSyncState.idle;

  bool get hasAnySyncing =>
      _appStates.values.any((s) => s.isSyncing);

  /// Start watching sync status for the given app IDs (UUIDs).
  void startWatching(List<String> appIds) {
    _watchedAppIds = appIds;
    _pollTimer?.cancel();
    _poll(); // immediate first poll
    _pollTimer = Timer.periodic(
      const Duration(seconds: 5),
      (_) => _poll(),
    );
  }

  void stopWatching() {
    _pollTimer?.cancel();
    _pollTimer = null;
  }

  Future<void> _poll() async {
    if (_disposed || _watchedAppIds.isEmpty) return;
    try {
      final newStates = <String, AppSyncState>{};

      for (final appId in _watchedAppIds) {
        try {
          final jobs = await _service.getActiveSyncJobs(appId);
          if (jobs.isNotEmpty) {
            final job = jobs.first;
            newStates[appId] = AppSyncState(
              isSyncing: true,
              message: job.message ?? _waveMessage(job.currentWave),
              progress: job.progressPct / 100.0,
              jobId: job.id,
            );
          }
        } catch (e) {
          debugPrint('[SyncStatusProvider] poll error for app $appId: $e');
        }
      }

      // For watched apps with no active job, preserve lastSyncAt or go idle
      for (final appId in _watchedAppIds) {
        if (!newStates.containsKey(appId)) {
          final prev = _appStates[appId];
          newStates[appId] = AppSyncState(
            lastSyncAt: prev?.lastSyncAt,
          );
        }
      }

      _appStates = newStates;
      if (!_disposed) notifyListeners();

      // Slow down polling if nothing is syncing
      _adjustPollRate();
    } catch (e) {
      debugPrint('[SyncStatusProvider] poll error: $e');
    }
  }

  void _adjustPollRate() {
    final anySyncing = _appStates.values.any((s) => s.isSyncing);
    final desiredInterval = anySyncing
        ? const Duration(seconds: 5)
        : const Duration(seconds: 30);

    // Restart timer with new interval
    _pollTimer?.cancel();
    _pollTimer = Timer.periodic(desiredInterval, (_) => _poll());
  }

  String _waveMessage(String? wave) {
    if (wave == null) return 'Syncing...';
    switch (wave) {
      case 'subscription_sync':
        return 'Syncing subscriptions...';
      case 'transaction_sync':
        return 'Syncing transactions...';
      case 'event_sync':
        return 'Syncing events...';
      case 'status_sync':
        return 'Syncing statuses...';
      case 'snapshot':
        return 'Building snapshots...';
      default:
        return 'Syncing...';
    }
  }

  /// Trigger a sync for the given app ID (UUID).
  Future<void> triggerSync(String appId) async {
    try {
      final job = await _service.enqueueSync(appId);
      _appStates[appId] = AppSyncState(
        isSyncing: true,
        message: 'Queued...',
        progress: 0,
        jobId: job.id,
      );
      notifyListeners();
      // Start fast polling
      _adjustPollRate();
    } catch (e) {
      _appStates[appId] = AppSyncState(
        error: e.toString(),
      );
      notifyListeners();
    }
  }

  /// Cancel a running sync for the given app ID (UUID).
  Future<void> cancelSync(String appId) async {
    final state = _appStates[appId];
    if (state?.jobId == null) return;
    try {
      await _service.cancelJob(state!.jobId!);
      _appStates[appId] = const AppSyncState();
      notifyListeners();
    } catch (e) {
      debugPrint('[SyncStatusProvider] cancel error: $e');
    }
  }

  @override
  void dispose() {
    _disposed = true;
    _pollTimer?.cancel();
    super.dispose();
  }
}
