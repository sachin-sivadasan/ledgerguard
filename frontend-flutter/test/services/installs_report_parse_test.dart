import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/services/installs_service.dart';

void main() {
  group('InstallsReport.fromJson — lifecycle + conversion (APPS-1b)', () {
    test('parses lifecycle and conversion objects', () {
      final r = InstallsReport.fromJson({
        'installs': 5,
        'uninstalls': 2,
        'net': 3,
        'trend': [],
        'events': [],
        'lifecycle': {
          'active': 240,
          'installed': 318,
          'uninstalled': 62,
          'reactivated': 14,
          'deactivated': 16,
        },
        'conversion': {'installs': 318, 'paid': 76, 'rate': 0.239},
      });

      expect(r.lifecycle.active, 240);
      expect(r.lifecycle.installed, 318);
      expect(r.lifecycle.uninstalled, 62);
      expect(r.lifecycle.reactivated, 14);
      expect(r.lifecycle.deactivated, 16);
      expect(r.conversion.paid, 76);
      expect(r.conversion.installs, 318);
      expect(r.conversion.ratePercent, '24%'); // 0.239 → 24%
    });

    // Older/absent payloads must degrade to zeroed defaults, not throw.
    test('defaults when lifecycle/conversion missing', () {
      final r = InstallsReport.fromJson({
        'installs': 0,
        'uninstalls': 0,
        'net': 0,
        'trend': [],
        'events': [],
      });
      expect(r.lifecycle.active, 0);
      expect(r.conversion.rate, 0);
      expect(r.conversion.ratePercent, '0%');
    });
  });
}
