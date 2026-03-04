import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

import 'app.dart';
import 'core/config/app_config.dart';
import 'core/config/env_config.dart';
import 'core/di/injection.dart';
import 'firebase_options.dart';

/// Staging entry point - uses staging configuration (GCP Cloud Run API)
/// Usage: flutter run -t lib/main_staging.dart
/// Build: flutter build web --release -t lib/main_staging.dart
void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Initialize environment
  AppConfig.init(EnvConfig.staging);

  // Initialize Firebase
  await _initializeFirebase();

  // Initialize dependencies
  await configureDependencies();

  runApp(const LedgerGuardApp());
}

Future<void> _initializeFirebase() async {
  try {
    if (Firebase.apps.isEmpty) {
      await Firebase.initializeApp(
        options: DefaultFirebaseOptions.currentPlatform,
      );
    }
  } catch (e) {
    if (kDebugMode) {
      debugPrint('Firebase initialization error: $e');
    }
  }
}
