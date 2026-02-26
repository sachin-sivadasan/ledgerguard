import 'package:go_router/go_router.dart';

import '../pages/placeholder_page.dart';

/// App routes configuration using GoRouter
class AppRouter {
  static final GoRouter router = GoRouter(
    initialLocation: '/',
    routes: [
      GoRoute(
        path: '/',
        name: 'home',
        builder: (context, state) => const PlaceholderPage(title: 'Home'),
      ),
      GoRoute(
        path: '/login',
        name: 'login',
        builder: (context, state) => const PlaceholderPage(title: 'Login'),
      ),
      GoRoute(
        path: '/signup',
        name: 'signup',
        builder: (context, state) => const PlaceholderPage(title: 'Sign Up'),
      ),
      GoRoute(
        path: '/dashboard',
        name: 'dashboard',
        builder: (context, state) => const PlaceholderPage(title: 'Dashboard'),
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
    ],
  );
}
