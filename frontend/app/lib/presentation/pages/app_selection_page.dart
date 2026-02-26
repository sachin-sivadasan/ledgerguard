import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/shopify_app.dart';
import '../blocs/app_selection/app_selection.dart';

/// Page for selecting which Shopify app to track
class AppSelectionPage extends StatefulWidget {
  const AppSelectionPage({super.key});

  @override
  State<AppSelectionPage> createState() => _AppSelectionPageState();
}

class _AppSelectionPageState extends State<AppSelectionPage> {
  @override
  void initState() {
    super.initState();
    context.read<AppSelectionBloc>().add(const FetchAppsRequested());
  }

  void _onAppSelected(ShopifyApp app) {
    context.read<AppSelectionBloc>().add(AppSelected(app));
  }

  void _onConfirm() {
    context.read<AppSelectionBloc>().add(const ConfirmSelectionRequested());
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Select App'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.go('/partner-integration'),
        ),
      ),
      body: BlocConsumer<AppSelectionBloc, AppSelectionState>(
        listener: (context, state) {
          if (state is AppSelectionConfirmed) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text('Selected: ${state.selectedApp.name}'),
                backgroundColor: AppTheme.success,
              ),
            );
            context.go('/dashboard');
          }
        },
        builder: (context, state) {
          return Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Header
                Text(
                  'Choose Your App',
                  style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                ),
                const SizedBox(height: 8),
                Text(
                  'Select the app you want to track revenue for.',
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: Colors.grey[600],
                      ),
                ),
                const SizedBox(height: 24),

                // Content based on state
                Expanded(
                  child: _buildContent(context, state),
                ),

                // Confirm button
                if (state is AppSelectionLoaded && state.hasSelection) ...[
                  const SizedBox(height: 16),
                  SizedBox(
                    width: double.infinity,
                    height: 48,
                    child: ElevatedButton(
                      onPressed: _onConfirm,
                      child: const Text('Confirm Selection'),
                    ),
                  ),
                ],

                if (state is AppSelectionSaving) ...[
                  const SizedBox(height: 16),
                  SizedBox(
                    width: double.infinity,
                    height: 48,
                    child: ElevatedButton(
                      onPressed: null,
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const SizedBox(
                            height: 20,
                            width: 20,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: Colors.white,
                            ),
                          ),
                          const SizedBox(width: 12),
                          const Text('Saving...'),
                        ],
                      ),
                    ),
                  ),
                ],
              ],
            ),
          );
        },
      ),
    );
  }

  Widget _buildContent(BuildContext context, AppSelectionState state) {
    if (state is AppSelectionLoading) {
      return const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            CircularProgressIndicator(),
            SizedBox(height: 16),
            Text('Loading apps...'),
          ],
        ),
      );
    }

    if (state is AppSelectionError) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.error_outline,
              size: 64,
              color: AppTheme.danger,
            ),
            const SizedBox(height: 16),
            Text(
              state.message,
              style: TextStyle(color: AppTheme.danger),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: () {
                context.read<AppSelectionBloc>().add(const FetchAppsRequested());
              },
              icon: const Icon(Icons.refresh),
              label: const Text('Retry'),
            ),
          ],
        ),
      );
    }

    if (state is AppSelectionLoaded || state is AppSelectionSaving) {
      final apps = state is AppSelectionLoaded
          ? state.apps
          : (state as AppSelectionSaving).apps;
      final selectedApp = state is AppSelectionLoaded
          ? state.selectedApp
          : (state as AppSelectionSaving).selectedApp;

      return ListView.separated(
        itemCount: apps.length,
        separatorBuilder: (context, index) => const SizedBox(height: 12),
        itemBuilder: (context, index) {
          final app = apps[index];
          final isSelected = selectedApp?.id == app.id;

          return _AppListTile(
            app: app,
            isSelected: isSelected,
            onTap: state is AppSelectionSaving
                ? null
                : () => _onAppSelected(app),
          );
        },
      );
    }

    return const SizedBox.shrink();
  }
}

class _AppListTile extends StatelessWidget {
  final ShopifyApp app;
  final bool isSelected;
  final VoidCallback? onTap;

  const _AppListTile({
    required this.app,
    required this.isSelected,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Material(
      color: isSelected
          ? AppTheme.primary.withOpacity(0.1)
          : Colors.grey[50],
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color: isSelected ? AppTheme.primary : Colors.grey[300]!,
              width: isSelected ? 2 : 1,
            ),
          ),
          child: Row(
            children: [
              // Radio indicator
              Container(
                width: 24,
                height: 24,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  border: Border.all(
                    color: isSelected ? AppTheme.primary : Colors.grey[400]!,
                    width: 2,
                  ),
                ),
                child: isSelected
                    ? Center(
                        child: Container(
                          width: 12,
                          height: 12,
                          decoration: const BoxDecoration(
                            shape: BoxShape.circle,
                            color: AppTheme.primary,
                          ),
                        ),
                      )
                    : null,
              ),
              const SizedBox(width: 16),

              // App icon placeholder
              Container(
                width: 48,
                height: 48,
                decoration: BoxDecoration(
                  color: AppTheme.primary.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(
                  Icons.apps,
                  color: AppTheme.primary,
                ),
              ),
              const SizedBox(width: 16),

              // App info
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      app.name,
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.w600,
                          ),
                    ),
                    if (app.description != null) ...[
                      const SizedBox(height: 4),
                      Text(
                        app.description!,
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: Colors.grey[600],
                            ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ],
                    if (app.installCount != null) ...[
                      const SizedBox(height: 4),
                      Text(
                        '${app.installCount} installs',
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: Colors.grey[500],
                            ),
                      ),
                    ],
                  ],
                ),
              ),

              // Selected indicator
              if (isSelected)
                const Icon(
                  Icons.check_circle,
                  color: AppTheme.primary,
                ),
            ],
          ),
        ),
      ),
    );
  }
}
