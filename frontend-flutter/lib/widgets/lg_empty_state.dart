import 'package:flutter/material.dart';
import '../theme/app_colors.dart';
import '../theme/app_spacing.dart';
import 'lg_card.dart';

class LgEmptyState extends StatelessWidget {
  final IconData icon;
  final String heading;
  final String description;
  final String? actionLabel;
  final VoidCallback? onAction;

  const LgEmptyState({
    super.key,
    required this.icon,
    required this.heading,
    required this.description,
    this.actionLabel,
    this.onAction,
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
              Icon(icon, size: 48, color: LgColors.textDisabled),
              const SizedBox(height: LgSpacing.s400),
              Text(heading, style: theme.textTheme.titleMedium),
              const SizedBox(height: LgSpacing.s200),
              Text(description,
                  style: theme.textTheme.bodySmall,
                  textAlign: TextAlign.center),
              if (actionLabel != null && onAction != null) ...[
                const SizedBox(height: LgSpacing.s400),
                FilledButton(onPressed: onAction, child: Text(actionLabel!)),
              ],
            ],
          ),
        ),
      ),
    );
  }
}
