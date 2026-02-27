import 'package:equatable/equatable.dart';
import 'package:flutter/material.dart';

/// User notification preferences
class NotificationPreferences extends Equatable {
  /// Whether critical alerts are enabled
  final bool criticalAlertsEnabled;

  /// Whether daily summary is enabled
  final bool dailySummaryEnabled;

  /// Time for daily summary (hour and minute)
  final TimeOfDay dailySummaryTime;

  const NotificationPreferences({
    this.criticalAlertsEnabled = true,
    this.dailySummaryEnabled = true,
    this.dailySummaryTime = const TimeOfDay(hour: 9, minute: 0),
  });

  /// Create from JSON map
  factory NotificationPreferences.fromJson(Map<String, dynamic> json) {
    final timeString = json['daily_summary_time'] as String?;
    TimeOfDay time = const TimeOfDay(hour: 9, minute: 0);

    if (timeString != null && timeString.contains(':')) {
      final parts = timeString.split(':');
      if (parts.length >= 2) {
        time = TimeOfDay(
          hour: int.tryParse(parts[0]) ?? 9,
          minute: int.tryParse(parts[1]) ?? 0,
        );
      }
    }

    return NotificationPreferences(
      criticalAlertsEnabled: json['critical_alerts_enabled'] as bool? ?? true,
      dailySummaryEnabled: json['daily_summary_enabled'] as bool? ?? true,
      dailySummaryTime: time,
    );
  }

  /// Convert to JSON map
  Map<String, dynamic> toJson() {
    return {
      'critical_alerts_enabled': criticalAlertsEnabled,
      'daily_summary_enabled': dailySummaryEnabled,
      'daily_summary_time': '${dailySummaryTime.hour.toString().padLeft(2, '0')}:${dailySummaryTime.minute.toString().padLeft(2, '0')}',
    };
  }

  /// Create a copy with updated fields
  NotificationPreferences copyWith({
    bool? criticalAlertsEnabled,
    bool? dailySummaryEnabled,
    TimeOfDay? dailySummaryTime,
  }) {
    return NotificationPreferences(
      criticalAlertsEnabled: criticalAlertsEnabled ?? this.criticalAlertsEnabled,
      dailySummaryEnabled: dailySummaryEnabled ?? this.dailySummaryEnabled,
      dailySummaryTime: dailySummaryTime ?? this.dailySummaryTime,
    );
  }

  /// Format time for display
  String get formattedTime {
    final hour = dailySummaryTime.hourOfPeriod == 0 ? 12 : dailySummaryTime.hourOfPeriod;
    final minute = dailySummaryTime.minute.toString().padLeft(2, '0');
    final period = dailySummaryTime.period == DayPeriod.am ? 'AM' : 'PM';
    return '$hour:$minute $period';
  }

  /// Default preferences
  static const defaultPreferences = NotificationPreferences();

  @override
  List<Object?> get props => [criticalAlertsEnabled, dailySummaryEnabled, dailySummaryTime.hour, dailySummaryTime.minute];
}
