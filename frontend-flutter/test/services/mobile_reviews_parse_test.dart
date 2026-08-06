import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/services/mobile_reviews_service.dart';

void main() {
  group('MobileReviewsData.fromJson', () {
    test('parses both stores; Google is rating-only', () {
      final d = MobileReviewsData.fromJson({
        'iosAppId': '310633997',
        'playPackage': 'com.whatsapp',
        'appStore': {
          'linked': true,
          'appName': 'WhatsApp',
          'ratingValue': 4.69,
          'ratingCount': 18368633,
          'reviewsAvailable': true,
          'positive': 1,
          'neutral': 1,
          'negative': 1,
          'reviews': [
            {'author': 'a', 'rating': 5, 'title': 'Love', 'body': 'great', 'version': '1.0'},
          ],
        },
        'googlePlay': {
          'linked': true,
          'appName': 'WhatsApp',
          'ratingValue': 4.63,
          'ratingCount': 240297698,
          'installs': '10B+',
          'reviewsAvailable': false,
          'reviews': [],
        },
      });

      expect(d.hasAnyLink, isTrue);
      expect(d.appStore!.ratingCount, 18368633);
      expect(d.appStore!.reviewsAvailable, isTrue);
      expect(d.appStore!.reviews.single.rating, 5);
      expect(d.appStore!.installs, ''); // Apple has no install data
      expect(d.googlePlay!.reviewsAvailable, isFalse);
      expect(d.googlePlay!.ratingValue, 4.63);
      expect(d.googlePlay!.installs, '10B+');
    });

    test('omitted stores are null and hasAnyLink is false', () {
      final d = MobileReviewsData.fromJson({'iosAppId': '', 'playPackage': ''});
      expect(d.appStore, isNull);
      expect(d.googlePlay, isNull);
      expect(d.hasAnyLink, isFalse);
    });
  });
}
