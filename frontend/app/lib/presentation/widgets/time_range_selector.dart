import 'package:flutter/material.dart';

import '../../domain/entities/time_range.dart';

/// Widget for selecting time range preset
class TimeRangeSelector extends StatelessWidget {
  /// Currently selected time range
  final TimeRange currentRange;

  /// Callback when time range changes
  final ValueChanged<TimeRange> onRangeChanged;

  const TimeRangeSelector({
    super.key,
    required this.currentRange,
    required this.onRangeChanged,
  });

  @override
  Widget build(BuildContext context) {
    return PopupMenuButton<TimeRangePreset>(
      initialValue: currentRange.preset,
      onSelected: (preset) {
        if (preset == TimeRangePreset.custom) {
          _showCustomRangePicker(context);
        } else {
          onRangeChanged(TimeRange.fromPreset(preset));
        }
      },
      offset: const Offset(0, 40),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(8),
      ),
      itemBuilder: (context) => TimeRangePreset.values
          .where((p) => p != TimeRangePreset.custom) // Hide custom for now
          .map((preset) => PopupMenuItem<TimeRangePreset>(
                value: preset,
                child: Row(
                  children: [
                    if (preset == currentRange.preset)
                      Icon(
                        Icons.check,
                        size: 18,
                        color: Theme.of(context).colorScheme.primary,
                      )
                    else
                      const SizedBox(width: 18),
                    const SizedBox(width: 8),
                    Text(preset.displayName),
                  ],
                ),
              ))
          .toList(),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.surfaceContainerHighest,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(
            color: Theme.of(context).colorScheme.outline.withOpacity(0.2),
          ),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.calendar_today,
              size: 16,
              color: Theme.of(context).colorScheme.onSurfaceVariant,
            ),
            const SizedBox(width: 6),
            Text(
              currentRange.preset.displayName,
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w500,
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(width: 4),
            Icon(
              Icons.arrow_drop_down,
              size: 18,
              color: Theme.of(context).colorScheme.onSurfaceVariant,
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _showCustomRangePicker(BuildContext context) async {
    final initialRange = DateTimeRange(
      start: currentRange.start,
      end: currentRange.end,
    );

    final picked = await showDateRangePicker(
      context: context,
      firstDate: DateTime.now().subtract(const Duration(days: 365)),
      lastDate: DateTime.now(),
      initialDateRange: initialRange,
      builder: (context, child) {
        return Theme(
          data: Theme.of(context).copyWith(
            colorScheme: Theme.of(context).colorScheme,
          ),
          child: child!,
        );
      },
    );

    if (picked != null) {
      onRangeChanged(TimeRange.custom(picked.start, picked.end));
    }
  }
}

/// Compact time range chip for smaller spaces
class TimeRangeChip extends StatelessWidget {
  /// Currently selected time range
  final TimeRange currentRange;

  /// Callback when tapped
  final VoidCallback onTap;

  const TimeRangeChip({
    super.key,
    required this.currentRange,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.secondaryContainer,
          borderRadius: BorderRadius.circular(16),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              currentRange.preset.displayName,
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w500,
                color: Theme.of(context).colorScheme.onSecondaryContainer,
              ),
            ),
            const SizedBox(width: 4),
            Icon(
              Icons.expand_more,
              size: 16,
              color: Theme.of(context).colorScheme.onSecondaryContainer,
            ),
          ],
        ),
      ),
    );
  }
}
