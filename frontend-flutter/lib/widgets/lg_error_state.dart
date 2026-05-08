import 'package:flutter/material.dart';
import '../theme/app_colors.dart';
import '../theme/app_spacing.dart';
import 'lg_card.dart';

/// Reusable error + retry widget matching the existing [LgEmptyState] visual
/// pattern. Shows an error icon, message, and a retry button.
class LgErrorState extends StatelessWidget {
  final String message;
  final VoidCallback onRetry;

  const LgErrorState({
    super.key,
    required this.message,
    required this.onRetry,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return LgCard(
      child: Center(
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: LgSpacing.s1200),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.error_outline,
                  size: 48, color: LgColors.critical),
              const SizedBox(height: LgSpacing.s400),
              Text('Something went wrong',
                  style: theme.textTheme.titleMedium),
              const SizedBox(height: LgSpacing.s200),
              Text(message,
                  style: theme.textTheme.bodySmall,
                  textAlign: TextAlign.center),
              const SizedBox(height: LgSpacing.s400),
              FilledButton.icon(
                onPressed: onRetry,
                icon: const Icon(Icons.refresh, size: 18),
                label: const Text('Retry'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
