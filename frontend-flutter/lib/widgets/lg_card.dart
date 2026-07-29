import 'package:flutter/material.dart';
import '../theme/app_colors.dart';
import '../theme/app_spacing.dart';

/// A framed content card matching the wireframes: white fill (`LgColors.surface`),
/// rx=8, a subtle `LgColors.border` outline. Renders the border EXPLICITLY via a
/// `Container` rather than a Material `Card`: observed on Flutter web (canvaskit,
/// Flutter 3.x), a `Card` at elevation 0 with a `BorderSide` intermittently drops the
/// outline on repaint, so KPI/section cards would render frameless. Re-verify against
/// a `Card`+`side:` if a future SDK fixes that, and delete this workaround if so.
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
