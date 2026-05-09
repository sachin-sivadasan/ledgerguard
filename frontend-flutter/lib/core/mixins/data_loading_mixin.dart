import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../providers/apps_provider.dart';
import '../../providers/sync_status_provider.dart';
import '../navigation/navigation_refresh_notifier.dart';

/// Mixin that replaces the `_wasDemoMode` and `hasAttemptedLoad` patterns
/// with a single listener-based approach for data loading.
///
/// Detects demo->live transitions, app switches, initial loads,
/// staleness (>2 min since last load), and sync completion via listeners.
mixin DataLoadingMixin<T extends StatefulWidget> on State<T> {
  bool? _lastKnownDemoMode;
  String? _lastLoadedAppId;
  DateTime? _lastLoadedAt;
  bool _wasSyncing = false;
  static const _staleDuration = Duration(minutes: 2);

  /// Subclasses implement this to perform their data load.
  void loadData(String appId);

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) {
        context.read<AppsProvider>().addListener(_onAppsChanged);
        context.read<NavigationRefreshNotifier>().addListener(_onNavigationRefresh);
        context.read<SyncStatusProvider>().addListener(_onSyncStatusChanged);
        _evaluateAndLoad();
      }
    });
  }

  @override
  void dispose() {
    try {
      context.read<AppsProvider>().removeListener(_onAppsChanged);
    } catch (_) {
      // Provider may already be disposed during app shutdown.
    }
    try {
      context.read<NavigationRefreshNotifier>().removeListener(_onNavigationRefresh);
    } catch (_) {}
    try {
      context.read<SyncStatusProvider>().removeListener(_onSyncStatusChanged);
    } catch (_) {}
    super.dispose();
  }

  void _onAppsChanged() {
    if (mounted) _evaluateAndLoad();
  }

  void _onNavigationRefresh() {
    if (mounted) _evaluateAndLoad();
  }

  void _onSyncStatusChanged() {
    if (!mounted) return;
    final syncProvider = context.read<SyncStatusProvider>();
    final currentAppId = _lastLoadedAppId ?? '';
    final currentState = syncProvider.getState(currentAppId);

    // Auto-reload when sync transitions from syncing -> done
    if (!currentState.isSyncing && _wasSyncing) {
      _lastLoadedAt = null;
      _evaluateAndLoad();
    }
    _wasSyncing = currentState.isSyncing;
  }

  void _evaluateAndLoad() {
    final appsProvider = context.read<AppsProvider>();
    final isDemoMode = appsProvider.demoMode;
    final currentAppId =
        appsProvider.apps.isNotEmpty ? appsProvider.apps.first.id : null;

    final wasDemoNowLive = _lastKnownDemoMode == true && !isDemoMode;
    final appChanged =
        currentAppId != null && currentAppId != _lastLoadedAppId;
    final isStale = _lastLoadedAt != null &&
        DateTime.now().difference(_lastLoadedAt!) > _staleDuration;

    _lastKnownDemoMode = isDemoMode;

    if (!isDemoMode && currentAppId != null) {
      if (_lastLoadedAppId == null || wasDemoNowLive || appChanged || isStale) {
        _lastLoadedAppId = currentAppId;
        _lastLoadedAt = DateTime.now();
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (mounted) loadData(currentAppId);
        });
      }
    }

    // Reset on demo mode entry so next live transition re-fires.
    if (isDemoMode) {
      _lastLoadedAppId = null;
      _lastLoadedAt = null;
    }
  }

  /// Call from a "Retry" button in error UI.
  void retryLoad() {
    _lastLoadedAppId = null;
    _lastLoadedAt = null;
    _evaluateAndLoad();
  }

  /// Call from a manual refresh button.
  void refreshData() {
    _lastLoadedAt = null;
    _lastLoadedAppId = null;
    _evaluateAndLoad();
  }
}
