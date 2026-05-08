import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../providers/apps_provider.dart';

/// Mixin that replaces the `_wasDemoMode` and `hasAttemptedLoad` patterns
/// with a single listener-based approach for data loading.
///
/// Detects demo->live transitions, app switches, and initial loads via
/// an [AppsProvider] listener instead of side effects in `build()`.
mixin DataLoadingMixin<T extends StatefulWidget> on State<T> {
  bool? _lastKnownDemoMode;
  String? _lastLoadedAppId;

  /// Subclasses implement this to perform their data load.
  void loadData(String appId);

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) {
        context.read<AppsProvider>().addListener(_onAppsChanged);
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
    super.dispose();
  }

  void _onAppsChanged() {
    if (mounted) _evaluateAndLoad();
  }

  void _evaluateAndLoad() {
    final appsProvider = context.read<AppsProvider>();
    final isDemoMode = appsProvider.demoMode;
    final currentAppId =
        appsProvider.apps.isNotEmpty ? appsProvider.apps.first.id : null;

    final wasDemoNowLive = _lastKnownDemoMode == true && !isDemoMode;
    final appChanged =
        currentAppId != null && currentAppId != _lastLoadedAppId;

    _lastKnownDemoMode = isDemoMode;

    if (!isDemoMode && currentAppId != null) {
      if (_lastLoadedAppId == null || wasDemoNowLive || appChanged) {
        _lastLoadedAppId = currentAppId;
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (mounted) loadData(currentAppId);
        });
      }
    }

    // Reset on demo mode entry so next live transition re-fires.
    if (isDemoMode) {
      _lastLoadedAppId = null;
    }
  }

  /// Call from a "Retry" button in error UI.
  void retryLoad() {
    _lastLoadedAppId = null;
    _evaluateAndLoad();
  }
}
