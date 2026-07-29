import 'dart:async';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../providers/apps_provider.dart';

/// Shared helper for the report *detail* screens (the paged "View all" pages).
///
/// On a cold deep-link / hard reload directly onto a detail route, `AppsProvider`
/// hasn't delivered apps yet (they load only after org selection), so resolving an
/// app fails transiently and would surface a "No app selected" error. Mixing this
/// in lets the screen [waitForAppsThenReload]: it listens for apps to arrive and
/// then retries via [reloadAfterApps]. If they never arrive within a bounded wait,
/// `onUnavailable` fires so the screen can show the real "no app" state. The
/// listener + timer self-clean on unmount and in [dispose].
mixin ReportDetailAppGate<T extends StatefulWidget> on State<T> {
  AppsProvider? _gateApps;
  VoidCallback? _gateListener;
  Timer? _gateTimeout;

  /// Retry the screen's load once apps have arrived.
  void reloadAfterApps();

  /// Call when no app is resolvable. Waits (up to ~10s) for `AppsProvider` to
  /// deliver at least one app, then retries via [reloadAfterApps]. If none arrive
  /// in time, [onUnavailable] is invoked.
  void waitForAppsThenReload({required VoidCallback onUnavailable}) {
    _cancelAppGate();
    final apps = context.read<AppsProvider>();
    // Race: apps already present — retry now.
    if (apps.apps.isNotEmpty) {
      reloadAfterApps();
      return;
    }
    _gateApps = apps;
    _gateListener = () {
      if (!mounted) {
        _cancelAppGate();
        return;
      }
      if (apps.apps.isNotEmpty) {
        _cancelAppGate();
        reloadAfterApps();
      }
    };
    apps.addListener(_gateListener!);
    _gateTimeout = Timer(const Duration(seconds: 10), () {
      if (!mounted) {
        _cancelAppGate();
        return;
      }
      _cancelAppGate();
      onUnavailable();
    });
  }

  void _cancelAppGate() {
    final listener = _gateListener;
    if (listener != null) {
      _gateApps?.removeListener(listener);
      _gateListener = null;
      _gateApps = null;
    }
    _gateTimeout?.cancel();
    _gateTimeout = null;
  }

  @override
  void dispose() {
    _cancelAppGate();
    super.dispose();
  }
}
