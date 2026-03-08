import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:get_it/get_it.dart';
import 'package:go_router/go_router.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/shopify_app.dart';
import '../blocs/app_selection/app_selection.dart';
import '../blocs/subscription_list/subscription_list.dart';
import 'subscription_list_page.dart';

/// Tab wrapper for subscriptions that resolves appId from selected app.
class SubscriptionsTabPage extends StatelessWidget {
  const SubscriptionsTabPage({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<AppSelectionBloc, AppSelectionState>(
      builder: (context, state) {
        final app = _selectedAppFrom(state);

        if (app == null) {
          return _buildNoAppState(context);
        }

        // Extract numeric ID from GID (e.g., "gid://partners/App/4599915" -> "4599915")
        final parts = app.id.split('/');
        final numericAppId = parts.isNotEmpty ? parts.last : app.id;

        return BlocProvider(
          key: ValueKey(numericAppId),
          create: (_) => GetIt.instance<SubscriptionListBloc>(),
          child: SubscriptionListPage(appId: numericAppId),
        );
      },
    );
  }

  ShopifyApp? _selectedAppFrom(AppSelectionState state) {
    if (state is AppSelectionLoaded) return state.selectedApp;
    if (state is AppSelectionConfirmed) return state.selectedApp;
    if (state is AppSelectionSaving) return state.selectedApp;
    return null;
  }

  Widget _buildNoAppState(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Subscriptions')),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.subscriptions_outlined,
                size: 64,
                color: AppTheme.primary.withOpacity(0.4),
              ),
              const SizedBox(height: 16),
              Text(
                'No App Selected',
                style: Theme.of(context).textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
              ),
              const SizedBox(height: 8),
              Text(
                'Select an app from the Dashboard to view its subscriptions.',
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: Colors.grey[600],
                    ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 24),
              ElevatedButton.icon(
                onPressed: () => context.push('/app-selection'),
                icon: const Icon(Icons.apps),
                label: const Text('Select an App'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
