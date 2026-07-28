import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/core/date_range.dart';

void main() {
  final now = DateTime(2026, 7, 15);

  test('default preset is last 30 days', () {
    expect(DateRangePreset.defaultPreset, DateRangePreset.last30Days);
  });

  test('rolling presets subtract N days from now (matches backend to − N)', () {
    expect(resolveDateRange(DateRangePreset.last7Days, now).from, '2026-07-08');
    expect(resolveDateRange(DateRangePreset.last30Days, now).from, '2026-06-15');
    expect(resolveDateRange(DateRangePreset.last90Days, now).from, '2026-04-16');
    // `to` is always today.
    expect(resolveDateRange(DateRangePreset.last30Days, now).to, '2026-07-15');
  });

  test('last 12 months = same day one year back', () {
    expect(resolveDateRange(DateRangePreset.last12Months, now).from, '2025-07-15');
  });

  test('year to date starts Jan 1', () {
    expect(resolveDateRange(DateRangePreset.yearToDate, now).from, '2026-01-01');
  });

  test('all time uses a far-past lower bound (still sends a from)', () {
    final r = resolveDateRange(DateRangePreset.allTime, now);
    expect(r.from, '2000-01-01');
    expect(r.to, '2026-07-15');
  });

  test('dates are zero-padded YYYY-MM-DD', () {
    // Single-digit month/day must pad.
    final r = resolveDateRange(DateRangePreset.yearToDate, DateTime(2026, 3, 5));
    expect(r.from, '2026-01-01');
    expect(r.to, '2026-03-05');
  });
}
