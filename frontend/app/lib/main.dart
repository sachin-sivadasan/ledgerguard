import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

import 'app.dart';
import 'core/config/app_config.dart';
import 'core/config/env_config.dart';
import 'core/di/injection.dart';

/// Default entry point - uses dev configuration
/// For production, use: flutter run -t lib/main_prod.dart
void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Initialize environment
  AppConfig.init(EnvConfig.dev);

  // Initialize Firebase
  await _initializeFirebase();

  // Initialize dependencies
  await configureDependencies();

  runApp(const LedgerGuardApp());
}

Future<void> _initializeFirebase() async {
  try {
    // Check if Firebase is already initialized
    if (Firebase.apps.isEmpty) {
      // For web, Firebase must be configured via firebase_options.dart
      // Run: flutterfire configure
      // Until then, we'll skip initialization and show appropriate error states
      if (kDebugMode) {
        debugPrint('Firebase not configured. Run: flutterfire configure');
      }
    }
  } catch (e) {
    if (kDebugMode) {
      debugPrint('Firebase initialization error: $e');
    }
  }
}
