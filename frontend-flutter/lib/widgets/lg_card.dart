import 'package:flutter/material.dart';
import '../theme/app_colors.dart';
import '../theme/app_spacing.dart';

/// A framed content card matching the wireframes: white fill, rx=8, a #E5E7EB-style
/// border. Renders the border EXPLICITLY (a Material `Card` at elevation 0 with a
/// BorderSide renders inconsistently on Flutter web), so KPI/section cards always show
/// their framing.
class LgCard extends StatelessWidget {
  final Widget child;
  final EdgeInsetsGeometry? padding;
  final String? title;

  const LgCard({super.key, required this.child, this.padding, this.title});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      decoration: BoxDecoration(
        color: LgColors.surface,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: LgColors.border),
      ),
      padding: padding ?? const EdgeInsets.all(LgSpacing.s400),
      child: title != null
          ? Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(title!, style: theme.textTheme.titleSmall),
                const SizedBox(height: LgSpacing.s300),
                child,
              ],
            )
          : child,
    );
  }
}
