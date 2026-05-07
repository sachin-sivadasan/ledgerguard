import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'providers/auth_provider.dart';
import 'providers/organization_provider.dart';
import 'services/mixpanel_service.dart';
import 'screens/analytics/analytics_screen.dart';
import 'screens/api_keys/api_keys_screen.dart';
import 'screens/apps/apps_screen.dart';
import 'screens/auth/forgot_password_screen.dart';
import 'screens/auth/login_screen.dart';
import 'screens/auth/sign_up_screen.dart';
import 'screens/dashboard/dashboard_screen.dart';
import 'screens/earnings/earnings_screen.dart';
import 'screens/insights/insights_screen.dart';
import 'screens/risk/risk_screen.dart';
import 'screens/settings/connect_shopify_screen.dart';
import 'screens/settings/settings_screen.dart';
import 'screens/stores/store_detail_screen.dart';
import 'screens/team/audit_log_screen.dart';
import 'screens/team/team_screen.dart';
import 'screens/stores/store_list_screen.dart';
import 'screens/subscriptions/subscription_detail_screen.dart';
import 'screens/subscriptions/subscription_list_screen.dart';
import 'screens/events/events_screen.dart';
import 'screens/transactions/transactions_screen.dart';
import 'screens/webhooks/webhooks_screen.dart';
import 'shell/app_shell.dart';
import 'theme/app_theme.dart';

const _authRoutes = ['/login', '/sign-up', '/forgot-password'];

class App extends StatefulWidget {
  const App({super.key});

  @override
  State<App> createState() => _AppState();
}

class _AppState extends State<App> {
  late final GoRouter _router;
  bool _orgLoaded = false;

  @override
  void initState() {
    super.initState();
    final authProvider = context.read<AuthProvider>()
      ..setMixpanel(context.read<MixpanelService>());

    authProvider.addListener(_onAuthChanged);

    // If already authenticated (e.g., persisted session), load orgs immediately
    if (authProvider.isAuthenticated) {
      _orgLoaded = true;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        context.read<OrganizationProvider>().loadOrganizations();
      });
    }

    _router = GoRouter(
      initialLocation: '/',
      refreshListenable: authProvider,
      redirect: (context, state) {
        final isAuthenticated = authProvider.isAuthenticated;
        final isAuthRoute = _authRoutes.contains(state.matchedLocation);

        if (!isAuthenticated && !isAuthRoute) return '/login';
        if (isAuthenticated && isAuthRoute) return '/';
        return null;
      },
      routes: [
        GoRoute(
          path: '/login',
          builder: (c, s) => const LoginScreen(),
        ),
        GoRoute(
          path: '/sign-up',
          builder: (c, s) => const SignUpScreen(),
        ),
        GoRoute(
          path: '/forgot-password',
          builder: (c, s) => const ForgotPasswordScreen(),
        ),
        StatefulShellRoute.indexedStack(
          builder: (context, state, navigationShell) =>
              AppShell(navigationShell: navigationShell),
          branches: [
            StatefulShellBranch(routes: [
              GoRoute(
                  path: '/', builder: (c, s) => const DashboardScreen()),
            ]),
            StatefulShellBranch(routes: [
              GoRoute(
                path: '/subscriptions',
                builder: (c, s) => const SubscriptionListScreen(),
                routes: [
                  GoRoute(
                    path: ':id',
                    builder: (c, s) => SubscriptionDetailScreen(
                        subscriptionId: s.pathParameters['id']!),
                  ),
                ],
              ),
            ]),
            StatefulShellBranch(routes: [
              GoRoute(
                path: '/stores',
                builder: (c, s) => const StoreListScreen(),
                routes: [
                  GoRoute(
                    path: ':id',
                    builder: (c, s) =>
                        StoreDetailScreen(storeId: s.pathParameters['id']!),
                  ),
                ],
              ),
            ]),
            StatefulShellBranch(routes: [
              GoRoute(
                  path: '/transactions',
                  builder: (c, s) => const TransactionsScreen()),
            ]),
            StatefulShellBranch(routes: [
              GoRoute(
                  path: '/events',
                  builder: (c, s) => const EventsScreen()),
            ]),
            StatefulShellBranch(routes: [
              GoRoute(
                  path: '/webhooks',
                  builder: (c, s) => const WebhooksScreen()),
            ]),
            StatefulShellBranch(routes: [
              GoRoute(
                  path: '/risk', builder: (c, s) => const RiskScreen()),
            ]),
            StatefulShellBranch(routes: [
              GoRoute(
                  path: '/analytics',
                  builder: (c, s) => const AnalyticsScreen()),
            ]),
            StatefulShellBranch(routes: [
              GoRoute(
                  path: '/earnings',
                  builder: (c, s) => const EarningsScreen()),
            ]),
            StatefulShellBranch(routes: [
              GoRoute(
                  path: '/apps', builder: (c, s) => const AppsScreen()),
            ]),
            StatefulShellBranch(routes: [
              GoRoute(
                  path: '/api-keys',
                  builder: (c, s) => const ApiKeysScreen()),
            ]),
            StatefulShellBranch(routes: [
              GoRoute(
                  path: '/insights',
                  builder: (c, s) => const InsightsScreen()),
            ]),
            StatefulShellBranch(routes: [
              GoRoute(
                  path: '/settings',
                  builder: (c, s) => const SettingsScreen(),
                  routes: [
                    GoRoute(
                      path: 'connect-shopify',
                      builder: (c, s) => const ConnectShopifyScreen(),
                    ),
                    GoRoute(
                      path: 'team',
                      builder: (c, s) => const TeamScreen(),
                    ),
                    GoRoute(
                      path: 'audit-log',
                      builder: (c, s) => const AuditLogScreen(),
                    ),
                  ]),
            ]),
          ],
        ),
      ],
    );
  }

  void _onAuthChanged() {
    final auth = context.read<AuthProvider>();
    if (auth.isAuthenticated && !_orgLoaded) {
      _orgLoaded = true;
      context.read<OrganizationProvider>().loadOrganizations();
    } else if (!auth.isAuthenticated) {
      _orgLoaded = false;
    }
  }

  @override
  void dispose() {
    context.read<AuthProvider>().removeListener(_onAuthChanged);
    _router.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'LedgerGuard',
      debugShowCheckedModeBanner: false,
      theme: LgTheme.light,
      routerConfig: _router,
      scrollBehavior: const MaterialScrollBehavior().copyWith(
        dragDevices: {
          PointerDeviceKind.mouse,
          PointerDeviceKind.touch,
          PointerDeviceKind.trackpad,
        },
      ),
    );
  }
}
