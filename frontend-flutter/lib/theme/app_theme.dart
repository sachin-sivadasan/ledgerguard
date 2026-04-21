import 'package:flutter/material.dart';
import 'app_colors.dart';

class LgTheme {
  static ThemeData get light {
    final colorScheme = ColorScheme.fromSeed(
      seedColor: LgColors.primary,
      brightness: Brightness.light,
      surface: LgColors.surface,
      onSurface: LgColors.textPrimary,
    );

    return ThemeData(
      useMaterial3: true,
      colorScheme: colorScheme,
      scaffoldBackgroundColor: LgColors.backdrop,
      textTheme: const TextTheme(
        headlineSmall: TextStyle(
            fontSize: 20,
            fontWeight: FontWeight.w600,
            color: LgColors.textPrimary),
        titleMedium: TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.w600,
            color: LgColors.textPrimary),
        titleSmall: TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.w600,
            color: LgColors.textPrimary),
        bodyLarge: TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.w400,
            color: LgColors.textPrimary),
        bodyMedium: TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.w400,
            color: LgColors.textPrimary),
        bodySmall: TextStyle(
            fontSize: 12,
            fontWeight: FontWeight.w400,
            color: LgColors.textSecondary),
      ),
      cardTheme: CardThemeData(
        elevation: 0,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
          side: const BorderSide(color: LgColors.border),
        ),
        color: LgColors.surface,
        margin: EdgeInsets.zero,
      ),
      dividerTheme: const DividerThemeData(
        color: LgColors.border,
        thickness: 1,
        space: 0,
      ),
      appBarTheme: const AppBarTheme(
        backgroundColor: LgColors.surface,
        foregroundColor: LgColors.textPrimary,
        elevation: 0,
        scrolledUnderElevation: 1,
      ),
      inputDecorationTheme: InputDecorationTheme(
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
        isDense: true,
      ),
      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          backgroundColor: LgColors.primary,
          foregroundColor: Colors.white,
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
          shape:
              RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
          textStyle:
              const TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: LgColors.textPrimary,
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
          shape:
              RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
          side: const BorderSide(color: LgColors.border),
          textStyle:
              const TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
        ),
      ),
      tabBarTheme: TabBarThemeData(
        labelColor: LgColors.textPrimary,
        unselectedLabelColor: LgColors.textSecondary,
        indicatorColor: LgColors.primary,
        labelStyle:
            const TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
        unselectedLabelStyle:
            const TextStyle(fontSize: 14, fontWeight: FontWeight.w400),
      ),
      snackBarTheme: SnackBarThemeData(
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      ),
      navigationRailTheme: const NavigationRailThemeData(
        backgroundColor: LgColors.surface,
        selectedIconTheme: IconThemeData(color: LgColors.primary),
        selectedLabelTextStyle:
            TextStyle(color: LgColors.primary, fontWeight: FontWeight.w600),
        unselectedIconTheme: IconThemeData(color: LgColors.textSecondary),
        unselectedLabelTextStyle:
            TextStyle(color: LgColors.textSecondary),
      ),
    );
  }
}
