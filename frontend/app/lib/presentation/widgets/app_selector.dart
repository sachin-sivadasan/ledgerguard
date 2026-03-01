import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../domain/entities/shopify_app.dart';
import '../blocs/app_selection/app_selection.dart';

/// Widget for selecting between multiple tracked apps
class AppSelector extends StatelessWidget {
  final VoidCallback? onAppChanged;

  const AppSelector({
    super.key,
    this.onAppChanged,
  });

  @override
  Widget build(BuildContext context) {
    return BlocConsumer<AppSelectionBloc, AppSelectionState>(
      listenWhen: (previous, current) {
        // Listen for app changes to trigger refresh
        if (previous is AppSelectionLoaded && current is AppSelectionLoaded) {
          return previous.selectedApp?.id != current.selectedApp?.id;
        }
        if (previous is AppSelectionConfirmed && current is AppSelectionLoaded) {
          return previous.selectedApp.id != current.selectedApp?.id;
        }
        return false;
      },
      listener: (context, state) {
        onAppChanged?.call();
      },
      builder: (context, state) {
        if (state is AppSelectionLoading) {
          return const SizedBox(
            width: 120,
            child: Center(
              child: SizedBox(
                width: 16,
                height: 16,
                child: CircularProgressIndicator(strokeWidth: 2),
              ),
            ),
          );
        }

        if (state is AppSelectionLoaded) {
          return _buildSelector(context, state.apps, state.selectedApp);
        }

        if (state is AppSelectionConfirmed) {
          // Load full app list to enable switching
          context.read<AppSelectionBloc>().add(const FetchAppsRequested());
          return _buildSingleAppChip(context, state.selectedApp);
        }

        if (state is AppSelectionError) {
          return TextButton.icon(
            onPressed: () {
              context.read<AppSelectionBloc>().add(const FetchAppsRequested());
            },
            icon: const Icon(Icons.refresh, size: 16),
            label: const Text('Retry'),
            style: TextButton.styleFrom(
              foregroundColor: Colors.white70,
            ),
          );
        }

        // Initial state - load apps
        return const SizedBox.shrink();
      },
    );
  }

  Widget _buildSelector(
    BuildContext context,
    List<ShopifyApp> apps,
    ShopifyApp? selectedApp,
  ) {
    if (apps.isEmpty) {
      return const SizedBox.shrink();
    }

    if (apps.length == 1) {
      return _buildSingleAppChip(context, apps.first);
    }

    return PopupMenuButton<ShopifyApp>(
      tooltip: 'Switch app',
      onSelected: (app) {
        context.read<AppSelectionBloc>().add(AppSelected(app));
        context.read<AppSelectionBloc>().add(const ConfirmSelectionRequested());
      },
      offset: const Offset(0, 40),
      child: Container(
        constraints: const BoxConstraints(maxWidth: 180),
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
        decoration: BoxDecoration(
          color: Colors.white.withOpacity(0.15),
          borderRadius: BorderRadius.circular(20),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.apps, size: 16, color: Colors.white),
            const SizedBox(width: 8),
            Flexible(
              child: Text(
                selectedApp?.name ?? 'Select App',
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 13,
                  fontWeight: FontWeight.w500,
                ),
                overflow: TextOverflow.ellipsis,
              ),
            ),
            const SizedBox(width: 4),
            const Icon(Icons.arrow_drop_down, size: 18, color: Colors.white70),
          ],
        ),
      ),
      itemBuilder: (context) => apps.map((app) {
        final isSelected = selectedApp?.id == app.id;
        return PopupMenuItem<ShopifyApp>(
          value: app,
          child: Row(
            children: [
              Icon(
                isSelected ? Icons.check_circle : Icons.circle_outlined,
                size: 18,
                color: isSelected ? Theme.of(context).primaryColor : Colors.grey,
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  app.name,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),
        );
      }).toList(),
    );
  }

  Widget _buildSingleAppChip(BuildContext context, ShopifyApp app) {
    return Container(
      constraints: const BoxConstraints(maxWidth: 160),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.15),
        borderRadius: BorderRadius.circular(20),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(Icons.apps, size: 16, color: Colors.white),
          const SizedBox(width: 8),
          Flexible(
            child: Text(
              app.name,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 13,
                fontWeight: FontWeight.w500,
              ),
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ],
      ),
    );
  }
}
