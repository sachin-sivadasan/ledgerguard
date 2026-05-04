import 'package:flutter/material.dart';

enum LgDeviceType { mobile, tablet, desktop }

class LgBreakpoints {
  static const double mobile = 600;
  static const double tablet = 900;

  static LgDeviceType deviceType(BuildContext context) {
    final width = MediaQuery.sizeOf(context).width;
    if (width < mobile) return LgDeviceType.mobile;
    if (width < tablet) return LgDeviceType.tablet;
    return LgDeviceType.desktop;
  }

  static bool isMobile(BuildContext context) =>
      deviceType(context) == LgDeviceType.mobile;

  static bool isTablet(BuildContext context) =>
      deviceType(context) == LgDeviceType.tablet;

  static bool isDesktop(BuildContext context) =>
      deviceType(context) == LgDeviceType.desktop;

  static int metricColumns(BuildContext context) =>
      switch (deviceType(context)) {
        LgDeviceType.mobile => 2,
        LgDeviceType.tablet => 3,
        LgDeviceType.desktop => 4,
      };
}

class LgResponsive extends StatelessWidget {
  final Widget mobile;
  final Widget? tablet;
  final Widget desktop;

  const LgResponsive({
    super.key,
    required this.mobile,
    this.tablet,
    required this.desktop,
  });

  @override
  Widget build(BuildContext context) {
    return switch (LgBreakpoints.deviceType(context)) {
      LgDeviceType.mobile => mobile,
      LgDeviceType.tablet => tablet ?? desktop,
      LgDeviceType.desktop => desktop,
    };
  }
}
