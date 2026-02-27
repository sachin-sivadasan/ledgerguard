import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:get_it/get_it.dart';
import 'package:go_router/go_router.dart';

import '../blocs/auth/auth.dart';
import '../blocs/subscription_detail/subscription_detail.dart';
import '../blocs/subscription_list/subscription_list.dart';
import '../pages/admin/manual_integration_page.dart';
import '../pages/app_selection_page.dart';
import '../pages/dashboard_page.dart';
import '../pages/login_page.dart';
import '../pages/notification_settings_page.dart';
import '../pages/partner_integration_page.dart';
import '../pages/profile_page.dart';
import '../pages/risk_breakdown_page.dart';
import '../pages/signup_page.dart';
import '../pages/placeholder_page.dart';
import '../pages/subscription_detail_page.dart';
import '../pages/subscription_list_page.dart';

/// App routes configuration using GoRouter
class AppRouter {
  final AuthBloc authBloc;

  AppRouter({required this.authBloc});

  late final GoRouter router = GoRouter(
    initialLocation: '/login',
    refreshListenable: GoRouterRefreshStream(authBloc.stream),
    redirect: (context, state) {
      final isAuthenticated = authBloc.state is Authenticated;
      final isAuthRoute = state.matchedLocation == '/login' ||
          state.matchedLocation == '/signup';

      // If not authenticated and not on auth route, redirect to login
      if (!isAuthenticated && !isAuthRoute) {
        return '/login';
      }

      // If authenticated and on auth route, redirect to dashboard
      if (isAuthenticated && isAuthRoute) {
        return '/dashboard';
      }

      // No redirect needed
      return null;
    },
    routes: [
      GoRoute(
        path: '/login',
        name: 'login',
        builder: (context, state) => const LoginPage(),
      ),
      GoRoute(
        path: '/signup',
        name: 'signup',
        builder: (context, state) => const SignupPage(),
      ),
      GoRoute(
        path: '/dashboard',
        name: 'dashboard',
        builder: (context, state) => const DashboardPage(),
      ),
      GoRoute(
        path: '/onboarding',
        name: 'onboarding',
        builder: (context, state) => const PlaceholderPage(title: 'Onboarding'),
      ),
      GoRoute(
        path: '/settings',
        name: 'settings',
        builder: (context, state) => const PlaceholderPage(title: 'Settings'),
      ),
      GoRoute(
        path: '/partner-integration',
        name: 'partner-integration',
        builder: (context, state) => const PartnerIntegrationPage(),
      ),
      GoRoute(
        path: '/app-selection',
        name: 'app-selection',
        builder: (context, state) => const AppSelectionPage(),
      ),
      GoRoute(
        path: '/admin/manual-integration',
        name: 'manual-integration',
        builder: (context, state) => const ManualIntegrationPage(),
      ),
      GoRoute(
        path: '/risk-breakdown',
        name: 'risk-breakdown',
        builder: (context, state) => const RiskBreakdownPage(),
      ),
      GoRoute(
        path: '/settings/notifications',
        name: 'notification-settings',
        builder: (context, state) => const NotificationSettingsPage(),
      ),
      GoRoute(
        path: '/profile',
        name: 'profile',
        builder: (context, state) => const ProfilePage(),
      ),
      GoRoute(
        path: '/apps/:appId/subscriptions',
        name: 'subscription-list',
        builder: (context, state) {
          final appId = state.pathParameters['appId']!;
          return BlocProvider(
            create: (_) => GetIt.instance<SubscriptionListBloc>(),
            child: SubscriptionListPage(appId: appId),
          );
        },
      ),
      GoRoute(
        path: '/apps/:appId/subscriptions/:subscriptionId',
        name: 'subscription-detail',
        builder: (context, state) {
          final appId = state.pathParameters['appId']!;
          final subscriptionId = state.pathParameters['subscriptionId']!;
          return BlocProvider(
            create: (_) => GetIt.instance<SubscriptionDetailBloc>(),
            child: SubscriptionDetailPage(
              appId: appId,
              subscriptionId: subscriptionId,
            ),
          );
        },
      ),
    ],
  );
}

/// Converts a Stream into a Listenable for GoRouter refresh
class GoRouterRefreshStream extends ChangeNotifier {
  GoRouterRefreshStream(Stream<dynamic> stream) {
    notifyListeners();
    _subscription = stream.asBroadcastStream().listen((_) => notifyListeners());
  }

  late final dynamic _subscription;

  @override
  void dispose() {
    _subscription.cancel();
    super.dispose();
  }
}
