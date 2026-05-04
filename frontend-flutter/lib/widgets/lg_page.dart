import 'package:flutter/material.dart';
import '../theme/app_breakpoints.dart';
import '../theme/app_spacing.dart';

class LgPageAction {
  final String label;
  final VoidCallback onPressed;
  final bool isPrimary;

  const LgPageAction(
      {required this.label, required this.onPressed, this.isPrimary = false});
}

class LgPage extends StatelessWidget {
  final String title;
  final String? subtitle;
  final LgPageAction? primaryAction;
  final List<LgPageAction> secondaryActions;
  final VoidCallback? backAction;
  final Widget child;
  final bool scrollable;

  const LgPage({
    super.key,
    required this.title,
    this.subtitle,
    this.primaryAction,
    this.secondaryActions = const [],
    this.backAction,
    required this.child,
    this.scrollable = true,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final padding = switch (LgBreakpoints.deviceType(context)) {
      LgDeviceType.mobile => LgSpacing.s300,
      LgDeviceType.tablet => LgSpacing.s400,
      LgDeviceType.desktop => LgSpacing.s600,
    };
    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 1200),
        child: Padding(
          padding: EdgeInsets.all(padding),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  if (backAction != null) ...[
                    IconButton(
                      icon: const Icon(Icons.arrow_back),
                      onPressed: backAction,
                      tooltip: 'Back',
                    ),
                    const SizedBox(width: LgSpacing.s200),
                  ],
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(title, style: theme.textTheme.headlineSmall),
                        if (subtitle != null) ...[
                          const SizedBox(height: LgSpacing.s100),
                          Text(subtitle!, style: theme.textTheme.bodySmall),
                        ],
                      ],
                    ),
                  ),
                  ...secondaryActions.map((a) => Padding(
                        padding: const EdgeInsets.only(left: LgSpacing.s200),
                        child: OutlinedButton(
                            onPressed: a.onPressed, child: Text(a.label)),
                      )),
                  if (primaryAction != null) ...[
                    const SizedBox(width: LgSpacing.s200),
                    FilledButton(
                        onPressed: primaryAction!.onPressed,
                        child: Text(primaryAction!.label)),
                  ],
                ],
              ),
              const SizedBox(height: LgSpacing.s600),
              Expanded(child: scrollable ? SingleChildScrollView(child: child) : child),
            ],
          ),
        ),
      ),
    );
  }
}
