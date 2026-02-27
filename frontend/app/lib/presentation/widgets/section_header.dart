import 'package:flutter/material.dart';

/// Reusable section header widget
class SectionHeader extends StatelessWidget {
  /// The section title
  final String title;

  /// Optional trailing widget (e.g., action button)
  final Widget? trailing;

  /// Text style override
  final TextStyle? style;

  /// Padding around the header
  final EdgeInsetsGeometry padding;

  const SectionHeader({
    super.key,
    required this.title,
    this.trailing,
    this.style,
    this.padding = const EdgeInsets.only(left: 4, bottom: 8),
  });

  @override
  Widget build(BuildContext context) {
    final titleWidget = Padding(
      padding: padding,
      child: Text(
        title,
        style: style ??
            Theme.of(context).textTheme.titleLarge?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
      ),
    );

    if (trailing == null) {
      return titleWidget;
    }

    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Expanded(child: titleWidget),
        trailing!,
      ],
    );
  }
}

/// Smaller section header for sub-sections
class SubSectionHeader extends StatelessWidget {
  /// The section title
  final String title;

  /// Optional trailing widget
  final Widget? trailing;

  const SubSectionHeader({
    super.key,
    required this.title,
    this.trailing,
  });

  @override
  Widget build(BuildContext context) {
    return SectionHeader(
      title: title,
      trailing: trailing,
      style: Theme.of(context).textTheme.titleSmall?.copyWith(
            color: Colors.grey[600],
            fontWeight: FontWeight.w600,
          ),
    );
  }
}
