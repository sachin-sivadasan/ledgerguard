import 'package:flutter/material.dart';

import 'core/theme/app_theme.dart';
import 'presentation/router/app_router.dart';

/// Main application widget
class LedgerGuardApp extends StatelessWidget {
  const LedgerGuardApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'LedgerGuard',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.lightTheme,
      routerConfig: AppRouter.router,
    );
  }
}
