import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import 'core/di/injection.dart';
import 'core/theme/app_theme.dart';
import 'presentation/blocs/auth/auth.dart';
import 'presentation/router/app_router.dart';

/// Main application widget
class LedgerGuardApp extends StatefulWidget {
  const LedgerGuardApp({super.key});

  @override
  State<LedgerGuardApp> createState() => _LedgerGuardAppState();
}

class _LedgerGuardAppState extends State<LedgerGuardApp> {
  late final AuthBloc _authBloc;
  late final AppRouter _appRouter;

  @override
  void initState() {
    super.initState();
    _authBloc = getIt<AuthBloc>();
    _appRouter = AppRouter(authBloc: _authBloc);

    // Check auth state on startup
    _authBloc.add(const AuthCheckRequested());
  }

  @override
  void dispose() {
    _authBloc.close();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return BlocProvider<AuthBloc>.value(
      value: _authBloc,
      child: MaterialApp.router(
        title: 'LedgerGuard',
        debugShowCheckedModeBanner: false,
        theme: AppTheme.lightTheme,
        routerConfig: _appRouter.router,
      ),
    );
  }
}
