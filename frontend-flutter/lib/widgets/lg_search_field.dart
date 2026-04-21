import 'package:flutter/material.dart';

class LgSearchField extends StatelessWidget {
  final String hintText;
  final ValueChanged<String> onChanged;
  final String value;

  const LgSearchField({
    super.key,
    this.hintText = 'Search...',
    required this.onChanged,
    this.value = '',
  });

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 300,
      child: TextField(
        controller: TextEditingController(text: value)
          ..selection =
              TextSelection.fromPosition(TextPosition(offset: value.length)),
        onChanged: onChanged,
        decoration: InputDecoration(
          hintText: hintText,
          prefixIcon: const Icon(Icons.search, size: 20),
          suffixIcon: value.isNotEmpty
              ? IconButton(
                  icon: const Icon(Icons.clear, size: 18),
                  onPressed: () => onChanged(''),
                )
              : null,
        ),
      ),
    );
  }
}
