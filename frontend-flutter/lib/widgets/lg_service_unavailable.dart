import 'package:flutter/material.dart';
import '../theme/app_spacing.dart';
import 'lg_card.dart';

/// Shown when the backend returns 503 (e.g. database down).
class LgServiceUnavailable extends StatelessWidget {
  final VoidCallback? onRetry;
  const LgServiceUnavailable({super.key, this.onRetry});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Center(
      child: LgCard(
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: LgSpacing.s1200),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.cloud_off, size: 48, color: Colors.orange),
              const SizedBox(height: LgSpacing.s400),
              Text('Service Temporarily Unavailable',
                  style: theme.textTheme.titleMedium),
              const SizedBox(height: LgSpacing.s200),
              Text(
                'The backend service is unreachable. Retrying automatically...',
                style: theme.textTheme.bodySmall,
                textAlign: TextAlign.center,
              ),
              if (onRetry != null) ...[
                const SizedBox(height: LgSpacing.s400),
                FilledButton.icon(
                  onPressed: onRetry,
                  icon: const Icon(Icons.refresh, size: 18),
                  label: const Text('Retry Now'),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}
