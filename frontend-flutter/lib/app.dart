import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'providers/apps_provider.dart';
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
import 'screens/reports/churn_screen.dart';
import 'screens/reports/cohorts_screen.dart';
import 'screens/reports/earnings_charges_screen.dart';
import 'screens/reports/earnings_report_screen.dart';
import 'screens/reports/mrr_report_screen.dart';
import 'screens/reports/active_customers_screen.dart';
import 'screens/reports/activation_screen.dart';
import 'screens/reports/reports_screen.dart';
import 'screens/reports/retention_screen.dart';
import 'screens/reports/reviews_screen.dart';
import 'screens/reports/revenue_at_risk_screen.dart';
import 'screens/reports/revenue_mix_screen.dart';
import 'screens/reports/usage_screen.dart';
import 'screens/reports/usage_stores_screen.dart';
import 'screens/reports/usage_trends_screen.dart';
import 'screens/reports/subscriptions_screen.dart';
import 'screens/reports/payout_schedule_screen.dart';
import 'screens/reports/payout_schedule_payouts_screen.dart';
import 'screens/reports/payout_history_screen.dart';
import 'screens/reports/payout_history_payouts_screen.dart';
import 'screens/reports/fee_audit_screen.dart';
import 'screens/reports/ledger_recon_screen.dart';
import 'screens/reports/customer_insights_screen.dart';
import 'screens/reports/mobile_reviews_screen.dart';
import 'screens/reports/installs_screen.dart';
import 'screens/reports/installs_events_screen.dart';
import 'screens/reports/net_new_subs_screen.dart';
import 'screens/reports/net_new_subs_subscriptions_screen.dart';
import 'screens/reports/revenue_at_risk_stores_screen.dart';
import 'screens/reports/churn_stores_screen.dart';
import 'screens/reports/uninstall_context_screen.dart';
import 'screens/risk/risk_screen.dart';
import 'screens/settings/connect_shopify_screen.dart';
import 'screens/settings/dashboard_settings_screen.dart';
import 'screens/settings/plan_labels_screen.dart';
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

  @override
  void initState() {
    super.initState();
    final authProvider = context.read<AuthProvider>()
      ..setMixpanel(context.read<MixpanelService>());

    authProvider.addListener(_onAuthChanged);
    context.read<OrganizationProvider>().addListener(_onOrgChanged);

    // If already authenticated (e.g., persisted session), load orgs immediately
    if (authProvider.isAuthenticated) {
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
        GoRoute(path: '/login', builder: (c, s) => const LoginScreen()),
        GoRoute(path: '/sign-up', builder: (c, s) => const SignUpScreen()),
        GoRoute(
          path: '/forgot-password',
          builder: (c, s) => const ForgotPasswordScreen(),
        ),
        StatefulShellRoute.indexedStack(
          builder: (context, state, navigationShell) =>
              AppShell(navigationShell: navigationShell),
          branches: [
            StatefulShellBranch(
              routes: [
                GoRoute(path: '/', builder: (c, s) => const DashboardScreen()),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: '/subscriptions',
                  builder: (c, s) => const SubscriptionListScreen(),
                  routes: [
                    GoRoute(
                      path: ':id',
                      builder: (c, s) => SubscriptionDetailScreen(
                        subscriptionId: s.pathParameters['id']!,
                      ),
                    ),
                  ],
                ),
              ],
            ),
            StatefulShellBranch(
              routes: [
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
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: '/transactions',
                  builder: (c, s) => const TransactionsScreen(),
                ),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: '/events',
                  builder: (c, s) => const EventsScreen(),
                ),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: '/webhooks',
                  builder: (c, s) => const WebhooksScreen(),
                ),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(path: '/risk', builder: (c, s) => const RiskScreen()),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: '/analytics',
                  builder: (c, s) => const AnalyticsScreen(),
                ),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: '/earnings',
                  builder: (c, s) => const EarningsScreen(),
                ),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(path: '/apps', builder: (c, s) => const AppsScreen()),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: '/api-keys',
                  builder: (c, s) => const ApiKeysScreen(),
                ),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: '/insights',
                  builder: (c, s) => const InsightsScreen(),
                ),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: '/settings',
                  builder: (c, s) => const SettingsScreen(),
                  routes: [
                    GoRoute(
                      path: 'dashboard',
                      builder: (c, s) => const DashboardSettingsScreen(),
                    ),
                    GoRoute(
                      path: 'connect-shopify',
                      builder: (c, s) => const ConnectShopifyScreen(),
                    ),
                    GoRoute(
                      path: 'plan-labels',
                      builder: (c, s) => const PlanLabelsScreen(),
                    ),
                    GoRoute(
                      path: 'team',
                      builder: (c, s) => const TeamScreen(),
                    ),
                    GoRoute(
                      path: 'audit-log',
                      builder: (c, s) => const AuditLogScreen(),
                    ),
                  ],
                ),
              ],
            ),
            // Reports — appended LAST so existing branch indices don't shift.
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: '/reports',
                  builder: (c, s) => const ReportsScreen(),
                  routes: [
                    GoRoute(
                      path: 'revenue-at-risk',
                      builder: (c, s) => const RevenueAtRiskScreen(),
                    ),
                    GoRoute(
                      path: 'revenue-at-risk/stores',
                      builder: (c, s) => const RevenueAtRiskStoresScreen(),
                    ),
                    GoRoute(
                      path: 'churn',
                      builder: (c, s) => const ChurnScreen(),
                    ),
                    GoRoute(
                      path: 'churn/stores',
                      builder: (c, s) => const ChurnStoresScreen(),
                    ),
                    GoRoute(
                      path: 'retention',
                      builder: (c, s) => const RetentionScreen(),
                    ),
                    GoRoute(
                      path: 'usage',
                      builder: (c, s) => const UsageScreen(),
                    ),
                    GoRoute(
                      path: 'usage/stores',
                      builder: (c, s) => const UsageStoresScreen(),
                    ),
                    GoRoute(
                      path: 'usage-trends',
                      builder: (c, s) => const UsageTrendsScreen(),
                    ),
                    GoRoute(
                      path: 'subscriptions',
                      builder: (c, s) => const SubscriptionsScreen(),
                    ),
                    GoRoute(
                      path: 'payout-schedule',
                      builder: (c, s) => const PayoutScheduleScreen(),
                    ),
                    GoRoute(
                      path: 'payout-schedule/payouts',
                      builder: (c, s) => const PayoutSchedulePayoutsScreen(),
                    ),
                    GoRoute(
                      path: 'payout-history',
                      builder: (c, s) => const PayoutHistoryScreen(),
                    ),
                    GoRoute(
                      path: 'payout-history/payouts',
                      builder: (c, s) => const PayoutHistoryPayoutsScreen(),
                    ),
                    GoRoute(
                      path: 'installs',
                      builder: (c, s) => const InstallsScreen(),
                    ),
                    GoRoute(
                      path: 'fee-audit',
                      builder: (c, s) => const FeeAuditScreen(),
                    ),
                    GoRoute(
                      path: 'ledger-reconciliation',
                      builder: (c, s) => const LedgerReconScreen(),
                    ),
                    GoRoute(
                      path: 'customer-insights',
                      builder: (c, s) => const CustomerInsightsScreen(),
                    ),
                    GoRoute(
                      path: 'mobile-reviews',
                      builder: (c, s) => const MobileReviewsScreen(),
                    ),
                    GoRoute(
                      path: 'installs/events',
                      builder: (c, s) => const InstallsEventsScreen(),
                    ),
                    GoRoute(
                      path: 'net-new-subscriptions',
                      builder: (c, s) => const NetNewSubsScreen(),
                    ),
                    GoRoute(
                      path: 'net-new-subscriptions/subscriptions',
                      builder: (c, s) => const NetNewSubsSubscriptionsScreen(),
                    ),
                    GoRoute(
                      path: 'cohorts',
                      builder: (c, s) => const CohortsReportScreen(),
                    ),
                    GoRoute(
                      path: 'reviews',
                      builder: (c, s) => const ReviewsReportScreen(),
                    ),
                    GoRoute(
                      path: 'uninstall-context',
                      builder: (c, s) => const UninstallContextScreen(),
                    ),
                    GoRoute(
                      path: 'earnings',
                      builder: (c, s) => const EarningsReportScreen(),
                    ),
                    GoRoute(
                      path: 'earnings/charges',
                      builder: (c, s) => const EarningsChargesScreen(),
                    ),
                    GoRoute(
                      path: 'mrr',
                      builder: (c, s) => const MrrReportScreen(),
                    ),
                    GoRoute(
                      path: 'active-customers',
                      builder: (c, s) => const ActiveCustomersScreen(),
                    ),
                    GoRoute(
                      path: 'activation',
                      builder: (c, s) => const ActivationScreen(),
                    ),
                    GoRoute(
                      path: 'revenue-mix',
                      builder: (c, s) => const RevenueMixScreen(),
                    ),
                  ],
                ),
              ],
            ),
          ],
        ),
      ],
    );
  }

  void _onAuthChanged() {
    final auth = context.read<AuthProvider>();
    if (auth.isAuthenticated) {
      // Idempotent: provider's own _isLoading guard prevents concurrent calls.
      context.read<OrganizationProvider>().loadOrganizations();
    }
  }

  void _onOrgChanged() {
    final orgProvider = context.read<OrganizationProvider>();
    debugPrint(
      '[App] _onOrgChanged – currentOrg=${orgProvider.currentOrg?.name}',
    );
    if (orgProvider.currentOrg != null && orgProvider.error == null) {
      debugPrint('[App] → calling loadApps()');
      // Idempotent: provider's own _isLoading guard prevents concurrent calls.
      context.read<AppsProvider>().loadApps();
    }
  }

  @override
  void dispose() {
    context.read<AuthProvider>().removeListener(_onAuthChanged);
    context.read<OrganizationProvider>().removeListener(_onOrgChanged);
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
