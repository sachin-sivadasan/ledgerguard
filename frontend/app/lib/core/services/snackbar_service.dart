import 'package:flutter/material.dart';

import '../theme/app_theme.dart';

/// Global snackbar service for showing notifications
class SnackbarService {
  static final SnackbarService _instance = SnackbarService._internal();
  factory SnackbarService() => _instance;
  SnackbarService._internal();

  GlobalKey<ScaffoldMessengerState>? _messengerKey;

  /// Initialize with scaffold messenger key
  void init(GlobalKey<ScaffoldMessengerState> key) {
    _messengerKey = key;
  }

  /// Get the current messenger state
  ScaffoldMessengerState? get _messenger => _messengerKey?.currentState;

  /// Show an error snackbar
  void showError(String message, {Duration? duration}) {
    _showSnackbar(
      message: message,
      icon: Icons.error_outline,
      backgroundColor: AppTheme.danger,
      duration: duration ?? const Duration(seconds: 4),
    );
  }

  /// Show a success snackbar
  void showSuccess(String message, {Duration? duration}) {
    _showSnackbar(
      message: message,
      icon: Icons.check_circle_outline,
      backgroundColor: AppTheme.success,
      duration: duration ?? const Duration(seconds: 3),
    );
  }

  /// Show an info snackbar
  void showInfo(String message, {Duration? duration}) {
    _showSnackbar(
      message: message,
      icon: Icons.info_outline,
      backgroundColor: AppTheme.primary,
      duration: duration ?? const Duration(seconds: 3),
    );
  }

  /// Show a warning snackbar
  void showWarning(String message, {Duration? duration}) {
    _showSnackbar(
      message: message,
      icon: Icons.warning_amber_outlined,
      backgroundColor: AppTheme.warning,
      duration: duration ?? const Duration(seconds: 4),
    );
  }

  void _showSnackbar({
    required String message,
    required IconData icon,
    required Color backgroundColor,
    required Duration duration,
  }) {
    final messenger = _messenger;
    if (messenger == null) return;

    messenger.hideCurrentSnackBar();
    messenger.showSnackBar(
      SnackBar(
        content: Row(
          children: [
            Icon(icon, color: Colors.white, size: 20),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                message,
                style: const TextStyle(color: Colors.white),
              ),
            ),
          ],
        ),
        backgroundColor: backgroundColor,
        behavior: SnackBarBehavior.floating,
        duration: duration,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
        ),
        margin: const EdgeInsets.all(16),
      ),
    );
  }

  /// Hide current snackbar
  void hide() {
    _messenger?.hideCurrentSnackBar();
  }
}
