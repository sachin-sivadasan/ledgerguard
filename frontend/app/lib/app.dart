import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import 'core/di/injection.dart';
import 'core/theme/app_theme.dart';
import 'domain/repositories/auth_repository.dart';
import 'presentation/blocs/app_selection/app_selection.dart';
import 'presentation/blocs/auth/auth.dart';
import 'presentation/blocs/dashboard/dashboard.dart';
import 'presentation/blocs/insight/insight.dart';
import 'presentation/blocs/partner_integration/partner_integration.dart';
import 'presentation/blocs/preferences/preferences.dart';
import 'presentation/blocs/risk/risk.dart';
import 'presentation/blocs/role/role.dart';
import 'presentation/router/app_router.dart';

/// Main application widget
class LedgerGuardApp extends StatefulWidget {
  const LedgerGuardApp({super.key});

  @override
  State<LedgerGuardApp> createState() => _LedgerGuardAppState();
}

class _LedgerGuardAppState extends State<LedgerGuardApp> {
  late final AuthBloc _authBloc;
  late final RoleBloc _roleBloc;
  late final PartnerIntegrationBloc _partnerIntegrationBloc;
  late final AppSelectionBloc _appSelectionBloc;
  late final DashboardBloc _dashboardBloc;
  late final InsightBloc _insightBloc;
  late final PreferencesBloc _preferencesBloc;
  late final RiskBloc _riskBloc;
  late final AppRouter _appRouter;
  late final AuthRepository _authRepository;

  @override
  void initState() {
    super.initState();
    _authRepository = getIt<AuthRepository>();
    _authBloc = getIt<AuthBloc>();
    _roleBloc = getIt<RoleBloc>();
    _partnerIntegrationBloc = getIt<PartnerIntegrationBloc>();
    _appSelectionBloc = getIt<AppSelectionBloc>();
    _dashboardBloc = getIt<DashboardBloc>();
    _insightBloc = getIt<InsightBloc>();
    _preferencesBloc = getIt<PreferencesBloc>();
    _riskBloc = getIt<RiskBloc>();
    _appRouter = AppRouter(authBloc: _authBloc);

    // Check auth state on startup
    _authBloc.add(const AuthCheckRequested());
  }

  @override
  void dispose() {
    _authBloc.close();
    _roleBloc.close();
    _partnerIntegrationBloc.close();
    _appSelectionBloc.close();
    _dashboardBloc.close();
    _insightBloc.close();
    _preferencesBloc.close();
    _riskBloc.close();
    super.dispose();
  }

  void _onAuthStateChanged(BuildContext context, AuthState state) async {
    if (state is Authenticated) {
      // Fetch user role after successful authentication
      final token = await _authRepository.getIdToken();
      if (token != null) {
        _roleBloc.add(FetchRoleRequested(authToken: token));
      }
    } else if (state is Unauthenticated) {
      // Clear role on sign out
      _roleBloc.add(const ClearRoleRequested());
    }
  }

  @override
  Widget build(BuildContext context) {
    return MultiBlocProvider(
      providers: [
        BlocProvider<AuthBloc>.value(value: _authBloc),
        BlocProvider<RoleBloc>.value(value: _roleBloc),
        BlocProvider<PartnerIntegrationBloc>.value(value: _partnerIntegrationBloc),
        BlocProvider<AppSelectionBloc>.value(value: _appSelectionBloc),
        BlocProvider<DashboardBloc>.value(value: _dashboardBloc),
        BlocProvider<InsightBloc>.value(value: _insightBloc),
        BlocProvider<PreferencesBloc>.value(value: _preferencesBloc),
        BlocProvider<RiskBloc>.value(value: _riskBloc),
      ],
      child: BlocListener<AuthBloc, AuthState>(
        listener: _onAuthStateChanged,
        child: MaterialApp.router(
          title: 'LedgerGuard',
          debugShowCheckedModeBanner: false,
          theme: AppTheme.lightTheme,
          routerConfig: _appRouter.router,
        ),
      ),
    );
  }
}
