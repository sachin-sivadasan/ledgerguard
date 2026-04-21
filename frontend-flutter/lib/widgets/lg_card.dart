import 'package:flutter/material.dart';
import '../theme/app_spacing.dart';

class LgCard extends StatelessWidget {
  final Widget child;
  final EdgeInsetsGeometry? padding;
  final String? title;

  const LgCard({super.key, required this.child, this.padding, this.title});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      child: Padding(
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
      ),
    );
  }
}
