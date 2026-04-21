import 'package:flutter/material.dart';
import '../theme/app_colors.dart';

enum BadgeTone { success, warning, critical, info, attention, newTone, defaultTone }

class LgBadge extends StatelessWidget {
  final String label;
  final BadgeTone tone;

  const LgBadge({super.key, required this.label, this.tone = BadgeTone.defaultTone});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: _bgColor,
        borderRadius: BorderRadius.circular(10),
      ),
      child: Text(
        label,
        style: TextStyle(fontSize: 12, fontWeight: FontWeight.w500, color: _textColor),
      ),
    );
  }

  Color get _bgColor => switch (tone) {
        BadgeTone.success => LgColors.successBg,
        BadgeTone.warning => LgColors.warningBg,
        BadgeTone.critical => LgColors.criticalBg,
        BadgeTone.info => LgColors.infoBg,
        BadgeTone.attention => LgColors.attentionBg,
        BadgeTone.newTone => LgColors.newBg,
        BadgeTone.defaultTone => LgColors.surfaceSecondary,
      };

  Color get _textColor => switch (tone) {
        BadgeTone.success => LgColors.success,
        BadgeTone.warning => LgColors.warning,
        BadgeTone.critical => LgColors.critical,
        BadgeTone.info => LgColors.info,
        BadgeTone.attention => const Color(0xFF916A00),
        BadgeTone.newTone => LgColors.primary,
        BadgeTone.defaultTone => LgColors.textSecondary,
      };
}
