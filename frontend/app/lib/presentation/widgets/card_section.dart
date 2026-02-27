import 'package:flutter/material.dart';

/// Reusable card section widget with title
class CardSection extends StatelessWidget {
  /// Section title
  final String? title;

  /// Child widgets
  final List<Widget> children;

  /// Card margin
  final EdgeInsetsGeometry margin;

  /// Card padding
  final EdgeInsetsGeometry padding;

  const CardSection({
    super.key,
    this.title,
    required this.children,
    this.margin = EdgeInsets.zero,
    this.padding = EdgeInsets.zero,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (title != null) ...[
          Padding(
            padding: const EdgeInsets.only(left: 4, bottom: 8),
            child: Text(
              title!,
              style: Theme.of(context).textTheme.titleSmall?.copyWith(
                    color: Colors.grey[600],
                    fontWeight: FontWeight.w600,
                  ),
            ),
          ),
        ],
        Card(
          elevation: 0,
          margin: margin,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
            side: BorderSide(color: Colors.grey[200]!),
          ),
          child: Padding(
            padding: padding,
            child: Column(children: children),
          ),
        ),
      ],
    );
  }
}

/// Content card with shadow and rounded corners
class ContentCard extends StatelessWidget {
  /// Child widget
  final Widget child;

  /// Card padding
  final EdgeInsetsGeometry padding;

  /// Card margin
  final EdgeInsetsGeometry margin;

  const ContentCard({
    super.key,
    required this.child,
    this.padding = const EdgeInsets.all(20),
    this.margin = EdgeInsets.zero,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: margin,
      padding: padding,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: child,
    );
  }
}
