import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/providers/dashboard_provider.dart';

void main() {
  group('DashboardProvider.periodRangeFor', () {
    test('thisWeek is the trailing 7 days including today', () {
      final r = DashboardProvider.periodRangeFor(
          DashboardTimeRange.thisWeek, DateTime(2026, 8, 5, 12));
      expect(r.from, DateTime(2026, 7, 30)); // Aug 5 − 6 days
      expect(r.to, DateTime(2026, 8, 5, 12));
    });

    test('thisMonth starts on the first of the month', () {
      final r = DashboardProvider.periodRangeFor(
          DashboardTimeRange.thisMonth, DateTime(2026, 8, 5));
      expect(r.from, DateTime(2026, 8, 1));
    });

    test('lastMonth rolls back across the January/year boundary', () {
      final r = DashboardProvider.periodRangeFor(
          DashboardTimeRange.lastMonth, DateTime(2026, 1, 15));
      expect(r.from, DateTime(2025, 12, 1));
      expect(r.to, DateTime(2025, 12, 31)); // last day of Dec via DateTime(2026,1,0)
    });

    test('lastMonth.to is the correct last day of a non-leap February', () {
      final r = DashboardProvider.periodRangeFor(
          DashboardTimeRange.lastMonth, DateTime(2025, 3, 10));
      expect(r.to, DateTime(2025, 2, 28));
    });

    test('lastMonth.to handles a leap February', () {
      final r = DashboardProvider.periodRangeFor(
          DashboardTimeRange.lastMonth, DateTime(2028, 3, 10));
      expect(r.to, DateTime(2028, 2, 29));
    });

    test('threeMonths.from rolls back across the year boundary', () {
      final r = DashboardProvider.periodRangeFor(
          DashboardTimeRange.threeMonths, DateTime(2026, 1, 15));
      expect(r.from, DateTime(2025, 11, 1));
    });
  });
}
