import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../theme/app_colors.dart';
import '../theme/app_spacing.dart';
import 'lg_card.dart';

class LgOnboardingChecklist extends StatelessWidget {
  const LgOnboardingChecklist({super.key});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return LgCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.rocket_launch_outlined,
                  size: 24, color: LgColors.primary),
              const SizedBox(width: LgSpacing.s300),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('Get Started with LedgerGuard',
                        style: theme.textTheme.titleMedium),
                    const SizedBox(height: LgSpacing.s100),
                    Text(
                      'Complete these steps to see your revenue intelligence.',
                      style: theme.textTheme.bodySmall,
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: LgSpacing.s600),
          _Step(
            number: 1,
            title: 'Connect Shopify Partner Account',
            description: 'Link your Partner API credentials',
          ),
          const SizedBox(height: LgSpacing.s400),
          _Step(
            number: 2,
            title: 'Add Your First App',
            description: 'Select which app to track',
            trailing: TextButton(
              onPressed: () => context.go('/apps'),
              child: const Text('Go to Apps'),
            ),
          ),
          const SizedBox(height: LgSpacing.s400),
          _Step(
            number: 3,
            title: 'Wait for First Sync',
            description: 'Data appears within minutes',
          ),
        ],
      ),
    );
  }
}

class _Step extends StatelessWidget {
  final int number;
  final String title;
  final String description;
  final Widget? trailing;

  const _Step({
    required this.number,
    required this.title,
    required this.description,
    this.trailing,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Row(
      children: [
        Container(
          width: 28,
          height: 28,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            border: Border.all(color: LgColors.border, width: 1.5),
          ),
          child: Center(
            child: Text(
              '$number',
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w600,
                color: LgColors.textSecondary,
              ),
            ),
          ),
        ),
        const SizedBox(width: LgSpacing.s300),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(title, style: theme.textTheme.titleSmall),
              Text(description, style: theme.textTheme.bodySmall),
            ],
          ),
        ),
        ?trailing,
      ],
    );
  }
}
