import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:ledgerguard/domain/entities/subscription.dart';
import 'package:ledgerguard/presentation/widgets/subscription_tile.dart';

void main() {
  group('SubscriptionTile', () {
    final testSubscription = Subscription(
      id: 'sub-1',
      shopifyGid: 'gid://shopify/AppSubscription/1',
      myshopifyDomain: 'acme-store.myshopify.com',
      shopName: 'Acme Store',
      planName: 'Pro Plan',
      basePriceCents: 2999,
      billingInterval: BillingInterval.monthly,
      riskState: RiskState.safe,
      status: 'ACTIVE',
      createdAt: DateTime(2024, 1, 15),
    );

    final subscriptionWithoutShopName = Subscription(
      id: 'sub-2',
      shopifyGid: 'gid://shopify/AppSubscription/2',
      myshopifyDomain: 'beta-shop.myshopify.com',
      shopName: null,
      planName: 'Basic Plan',
      basePriceCents: 999,
      billingInterval: BillingInterval.monthly,
      riskState: RiskState.oneCycleMissed,
      status: 'ACTIVE',
      createdAt: DateTime(2024, 2, 10),
    );

    final annualSubscription = Subscription(
      id: 'sub-3',
      shopifyGid: 'gid://shopify/AppSubscription/3',
      myshopifyDomain: 'gamma-goods.myshopify.com',
      shopName: 'Gamma Goods',
      planName: 'Enterprise',
      basePriceCents: 29999,
      billingInterval: BillingInterval.annual,
      riskState: RiskState.churned,
      status: 'CANCELLED',
      createdAt: DateTime(2024, 1, 5),
    );

    Widget buildTestWidget(
      Subscription subscription, {
      VoidCallback? onTap,
      double width = 400,
    }) {
      return MaterialApp(
        home: Scaffold(
          body: SizedBox(
            width: width,
            child: SubscriptionTile(
              subscription: subscription,
              onTap: onTap,
            ),
          ),
        ),
      );
    }

    group('displays subscription info', () {
      testWidgets('shows shop name when available', (tester) async {
        await tester.pumpWidget(buildTestWidget(testSubscription));
        expect(find.text('Acme Store'), findsOneWidget);
      });

      testWidgets('shows formatted domain when shop name not available', (tester) async {
        await tester.pumpWidget(buildTestWidget(subscriptionWithoutShopName));
        expect(find.text('Beta Shop'), findsOneWidget);
      });

      testWidgets('shows plan name and price', (tester) async {
        await tester.pumpWidget(buildTestWidget(testSubscription));
        expect(find.text('Pro Plan \u00B7 \$29.99/mo'), findsOneWidget);
      });

      testWidgets('shows annual price format for annual plans', (tester) async {
        await tester.pumpWidget(buildTestWidget(annualSubscription));
        expect(find.text('Enterprise \u00B7 \$299.99/yr'), findsOneWidget);
      });

      testWidgets('shows initials in avatar', (tester) async {
        await tester.pumpWidget(buildTestWidget(testSubscription));
        expect(find.text('AS'), findsOneWidget);
      });
    });

    group('displays risk badge', () {
      testWidgets('shows safe badge for safe subscription', (tester) async {
        await tester.pumpWidget(buildTestWidget(testSubscription));
        expect(find.text('Safe'), findsOneWidget);
      });

      testWidgets('shows at risk badge for oneCycleMissed', (tester) async {
        await tester.pumpWidget(buildTestWidget(subscriptionWithoutShopName));
        expect(find.text('At Risk'), findsOneWidget);
      });

      testWidgets('shows churned badge for churned subscription', (tester) async {
        await tester.pumpWidget(buildTestWidget(annualSubscription));
        expect(find.text('Churned'), findsOneWidget);
      });
    });

    group('shows chevron icon', () {
      testWidgets('displays chevron right icon', (tester) async {
        await tester.pumpWidget(buildTestWidget(testSubscription));
        expect(find.byIcon(Icons.chevron_right), findsOneWidget);
      });
    });

    group('tap handling', () {
      testWidgets('calls onTap when tapped', (tester) async {
        var tapped = false;
        await tester.pumpWidget(buildTestWidget(
          testSubscription,
          onTap: () => tapped = true,
        ));

        await tester.tap(find.byType(InkWell));
        await tester.pumpAndSettle();

        expect(tapped, isTrue);
      });

      testWidgets('does not crash when onTap is null', (tester) async {
        await tester.pumpWidget(buildTestWidget(testSubscription, onTap: null));

        await tester.tap(find.byType(InkWell));
        await tester.pumpAndSettle();

        // Should not throw
      });
    });

    group('responsive layout', () {
      testWidgets('uses compact layout for narrow width', (tester) async {
        await tester.pumpWidget(buildTestWidget(testSubscription, width: 320));

        // In compact mode, only price is shown (not plan name)
        expect(find.text('\$29.99/mo'), findsOneWidget);
        expect(find.text('Pro Plan \u00B7 \$29.99/mo'), findsNothing);
      });

      testWidgets('uses full layout for wider width', (tester) async {
        await tester.pumpWidget(buildTestWidget(testSubscription, width: 500));

        // In full mode, both plan name and price are shown
        expect(find.text('Pro Plan \u00B7 \$29.99/mo'), findsOneWidget);
      });
    });
  });

  group('SubscriptionTileSkeleton', () {
    testWidgets('renders skeleton loader', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SubscriptionTileSkeleton(),
          ),
        ),
      );

      // Should render without errors
      expect(find.byType(SubscriptionTileSkeleton), findsOneWidget);
    });
  });

  group('initials generation', () {
    Widget buildTileWithDomain(String domain, {String? shopName}) {
      final subscription = Subscription(
        id: 'test',
        shopifyGid: 'gid://shopify/AppSubscription/1',
        myshopifyDomain: domain,
        shopName: shopName,
        planName: 'Plan',
        basePriceCents: 1000,
        billingInterval: BillingInterval.monthly,
        riskState: RiskState.safe,
        status: 'ACTIVE',
        createdAt: DateTime(2024, 1, 1),
      );
      return MaterialApp(
        home: Scaffold(
          body: SizedBox(
            width: 400,
            child: SubscriptionTile(subscription: subscription),
          ),
        ),
      );
    }

    testWidgets('generates two-letter initials from two-word name', (tester) async {
      await tester.pumpWidget(buildTileWithDomain('test.myshopify.com', shopName: 'Acme Store'));
      expect(find.text('AS'), findsOneWidget);
    });

    testWidgets('generates two-letter initials from single word', (tester) async {
      await tester.pumpWidget(buildTileWithDomain('mystore.myshopify.com'));
      expect(find.text('MY'), findsOneWidget);
    });

    testWidgets('handles hyphenated domains', (tester) async {
      await tester.pumpWidget(buildTileWithDomain('my-awesome-store.myshopify.com'));
      expect(find.text('MA'), findsOneWidget);
    });

    testWidgets('handles underscored domains', (tester) async {
      await tester.pumpWidget(buildTileWithDomain('my_cool_store.myshopify.com'));
      expect(find.text('MC'), findsOneWidget);
    });
  });
}
