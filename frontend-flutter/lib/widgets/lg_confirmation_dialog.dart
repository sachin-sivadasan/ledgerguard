import 'package:flutter/material.dart';

class LgConfirmationDialog extends StatelessWidget {
  final String title;
  final String message;
  final String confirmLabel;
  final String cancelLabel;
  final VoidCallback onConfirm;
  final bool destructive;

  const LgConfirmationDialog({
    super.key,
    required this.title,
    required this.message,
    this.confirmLabel = 'Confirm',
    this.cancelLabel = 'Cancel',
    required this.onConfirm,
    this.destructive = false,
  });

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(title),
      content: Text(message),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: Text(cancelLabel),
        ),
        FilledButton(
          onPressed: () {
            Navigator.of(context).pop();
            onConfirm();
          },
          style: destructive
              ? FilledButton.styleFrom(
                  backgroundColor: Theme.of(context).colorScheme.error)
              : null,
          child: Text(confirmLabel),
        ),
      ],
    );
  }
}
