import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../core/date_range.dart';
import '../providers/apps_provider.dart';
import '../providers/sync_status_provider.dart';
import '../theme/app_breakpoints.dart';
import '../theme/app_colors.dart';
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
  /// Optional breadcrumb rendered above the title, e.g. "Reports › Growth".
  final String? breadcrumb;
  /// When set, renders a date-range preset selector in the header. Requires
  /// [onDateRangeChanged] to be wired.
  final DateRangePreset? dateRange;
  final ValueChanged<DateRangePreset>? onDateRangeChanged;
  final LgPageAction? primaryAction;
  final List<LgPageAction> secondaryActions;
  final VoidCallback? backAction;
  final VoidCallback? onRefresh;
  final Widget child;
  final bool scrollable;

  const LgPage({
    super.key,
    required this.title,
    this.subtitle,
    this.breadcrumb,
    this.dateRange,
    this.onDateRangeChanged,
    this.primaryAction,
    this.secondaryActions = const [],
    this.backAction,
    this.onRefresh,
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

    // Check if current app is syncing
    final appsProvider = context.watch<AppsProvider>();
    final syncProvider = context.watch<SyncStatusProvider>();
    final activeAppId =
        appsProvider.apps.isNotEmpty ? appsProvider.apps.first.id : null;
    final isSyncing = activeAppId != null &&
        syncProvider.getState(activeAppId).isSyncing;

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
                        if (breadcrumb != null) ...[
                          Text(breadcrumb!,
                              style: theme.textTheme.bodySmall
                                  ?.copyWith(color: LgColors.textSecondary)),
                          const SizedBox(height: LgSpacing.s100),
                        ],
                        Row(
                          children: [
                            Flexible(
                              child: Text(title, style: theme.textTheme.headlineSmall),
                            ),
                            if (isSyncing) ...[
                              const SizedBox(width: LgSpacing.s300),
                              const SizedBox(
                                width: 14,
                                height: 14,
                                child: CircularProgressIndicator(strokeWidth: 2),
                              ),
                              const SizedBox(width: 4),
                              Text(
                                'Syncing...',
                                style: TextStyle(
                                  fontSize: 12,
                                  color: LgColors.textSecondary,
                                ),
                              ),
                            ],
                          ],
                        ),
                        if (subtitle != null) ...[
                          const SizedBox(height: LgSpacing.s100),
                          Text(subtitle!, style: theme.textTheme.bodySmall),
                        ],
                      ],
                    ),
                  ),
                  if (dateRange != null && onDateRangeChanged != null) ...[
                    DateRangeSelector(
                      value: dateRange!,
                      onChanged: onDateRangeChanged!,
                    ),
                    const SizedBox(width: LgSpacing.s200),
                  ],
                  if (onRefresh != null)
                    IconButton(
                      icon: const Icon(Icons.refresh),
                      tooltip: 'Refresh',
                      onPressed: onRefresh,
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
