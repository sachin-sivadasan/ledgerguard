import 'package:flutter/material.dart';

/// Reusable info tile widget for displaying labeled information
class InfoTile extends StatelessWidget {
  /// Icon to display
  final IconData icon;

  /// Label text
  final String label;

  /// Value text
  final String value;

  /// Optional trailing widget
  final Widget? trailing;

  /// Optional tap callback
  final VoidCallback? onTap;

  const InfoTile({
    super.key,
    required this.icon,
    required this.label,
    required this.value,
    this.trailing,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: Icon(icon, color: Colors.grey[600]),
      title: Text(
        label,
        style: Theme.of(context).textTheme.bodySmall?.copyWith(
              color: Colors.grey[600],
            ),
      ),
      subtitle: Text(
        value,
        style: Theme.of(context).textTheme.bodyLarge?.copyWith(
              fontWeight: FontWeight.w500,
            ),
      ),
      trailing: trailing,
      onTap: onTap,
    );
  }
}

/// Navigation tile with chevron icon
class NavigationTile extends StatelessWidget {
  /// Icon to display
  final IconData icon;

  /// Label text
  final String label;

  /// Tap callback
  final VoidCallback onTap;

  /// Optional subtitle
  final String? subtitle;

  const NavigationTile({
    super.key,
    required this.icon,
    required this.label,
    required this.onTap,
    this.subtitle,
  });

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: Icon(icon, color: Colors.grey[600]),
      title: Text(label),
      subtitle: subtitle != null ? Text(subtitle!) : null,
      trailing: Icon(Icons.chevron_right, color: Colors.grey[400]),
      onTap: onTap,
    );
  }
}
