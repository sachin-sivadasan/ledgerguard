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

  // Initialize dependencies
  await configureDependencies();

  // TODO: Initialize Firebase when firebase_options.dart is generated
  // await Firebase.initializeApp(
  //   options: DefaultFirebaseOptions.currentPlatform,
  // );

  runApp(const LedgerGuardApp());
}
