import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard/domain/entities/risk_summary.dart';
import 'package:ledgerguard/domain/entities/user_profile.dart';
import 'package:ledgerguard/presentation/widgets/shared.dart';

void main() {
  group('ErrorStateWidget', () {
    testWidgets('renders title and message', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: ErrorStateWidget(
            title: 'Error Title',
            message: 'Error message',
          ),
        ),
      ));

      expect(find.text('Error Title'), findsOneWidget);
      expect(find.text('Error message'), findsOneWidget);
      expect(find.byIcon(Icons.error_outline), findsOneWidget);
    });

    testWidgets('shows retry button when onRetry provided', (tester) async {
      var retryCount = 0;

      await tester.pumpWidget(MaterialApp(
        home: Scaffold(
          body: ErrorStateWidget(
            message: 'Error',
            onRetry: () => retryCount++,
          ),
        ),
      ));

      expect(find.text('Retry'), findsOneWidget);

      await tester.tap(find.text('Retry'));
      expect(retryCount, 1);
    });

    testWidgets('hides retry button when onRetry is null', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: ErrorStateWidget(message: 'Error'),
        ),
      ));

      expect(find.text('Retry'), findsNothing);
    });

    testWidgets('uses custom icon', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: ErrorStateWidget(
            message: 'Error',
            icon: Icons.warning,
          ),
        ),
      ));

      expect(find.byIcon(Icons.warning), findsOneWidget);
    });
  });

  group('EmptyStateWidget', () {
    testWidgets('renders title, message, and icon', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: EmptyStateWidget(
            title: 'No Data',
            message: 'Nothing to show',
            icon: Icons.inbox,
          ),
        ),
      ));

      expect(find.text('No Data'), findsOneWidget);
      expect(find.text('Nothing to show'), findsOneWidget);
      expect(find.byIcon(Icons.inbox), findsOneWidget);
    });

    testWidgets('shows action button when provided', (tester) async {
      var actionCount = 0;

      await tester.pumpWidget(MaterialApp(
        home: Scaffold(
          body: EmptyStateWidget(
            title: 'No Data',
            message: 'Nothing here',
            icon: Icons.inbox,
            actionLabel: 'Add Item',
            onAction: () => actionCount++,
          ),
        ),
      ));

      expect(find.text('Add Item'), findsOneWidget);

      await tester.tap(find.text('Add Item'));
      expect(actionCount, 1);
    });

    testWidgets('hides action button when not provided', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: EmptyStateWidget(
            title: 'No Data',
            message: 'Nothing here',
            icon: Icons.inbox,
          ),
        ),
      ));

      expect(find.byType(ElevatedButton), findsNothing);
    });
  });

  group('SectionHeader', () {
    testWidgets('renders title', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: SectionHeader(title: 'Section Title'),
        ),
      ));

      expect(find.text('Section Title'), findsOneWidget);
    });

    testWidgets('renders trailing widget', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: SectionHeader(
            title: 'Title',
            trailing: Icon(Icons.add),
          ),
        ),
      ));

      expect(find.byIcon(Icons.add), findsOneWidget);
    });
  });

  group('SubSectionHeader', () {
    testWidgets('renders title with smaller style', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: SubSectionHeader(title: 'Sub Section'),
        ),
      ));

      expect(find.text('Sub Section'), findsOneWidget);
    });
  });

  group('StatusBadge', () {
    testWidgets('renders label and color', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: StatusBadge(label: 'Active', color: Colors.green),
        ),
      ));

      expect(find.text('Active'), findsOneWidget);
    });

    testWidgets('renders icon when provided', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: StatusBadge(
            label: 'Pro',
            color: Colors.amber,
            icon: Icons.star,
          ),
        ),
      ));

      expect(find.byIcon(Icons.star), findsOneWidget);
    });
  });

  group('RoleBadge', () {
    testWidgets('renders owner badge', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: RoleBadge(role: UserRole.owner),
        ),
      ));

      expect(find.text('Owner'), findsOneWidget);
    });

    testWidgets('renders admin badge', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: RoleBadge(role: UserRole.admin),
        ),
      ));

      expect(find.text('Admin'), findsOneWidget);
    });
  });

  group('PlanBadge', () {
    testWidgets('renders pro badge with star icon', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: PlanBadge(tier: PlanTier.pro),
        ),
      ));

      expect(find.text('Pro'), findsOneWidget);
      expect(find.byIcon(Icons.star), findsOneWidget);
    });

    testWidgets('renders free badge with star outline', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: PlanBadge(tier: PlanTier.starter),
        ),
      ));

      expect(find.text('Free'), findsOneWidget);
      expect(find.byIcon(Icons.star_border), findsOneWidget);
    });
  });

  group('RiskBadge', () {
    testWidgets('renders safe badge', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: RiskBadge(level: RiskLevel.safe),
        ),
      ));

      expect(find.text('Safe'), findsOneWidget);
    });

    testWidgets('renders one cycle missed badge', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: RiskBadge(level: RiskLevel.oneCycleMissed),
        ),
      ));

      expect(find.text('One Cycle Missed'), findsOneWidget);
    });

    testWidgets('renders two cycles missed badge', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: RiskBadge(level: RiskLevel.twoCyclesMissed),
        ),
      ));

      expect(find.text('Two Cycles Missed'), findsOneWidget);
    });

    testWidgets('renders churned badge', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: RiskBadge(level: RiskLevel.churned),
        ),
      ));

      expect(find.text('Churned'), findsOneWidget);
    });
  });

  group('InfoTile', () {
    testWidgets('renders icon, label, and value', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: InfoTile(
            icon: Icons.email,
            label: 'Email',
            value: 'test@example.com',
          ),
        ),
      ));

      expect(find.byIcon(Icons.email), findsOneWidget);
      expect(find.text('Email'), findsOneWidget);
      expect(find.text('test@example.com'), findsOneWidget);
    });

    testWidgets('renders trailing widget', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: InfoTile(
            icon: Icons.person,
            label: 'Role',
            value: 'Owner',
            trailing: Icon(Icons.check),
          ),
        ),
      ));

      expect(find.byIcon(Icons.check), findsOneWidget);
    });

    testWidgets('handles tap', (tester) async {
      var tapCount = 0;

      await tester.pumpWidget(MaterialApp(
        home: Scaffold(
          body: InfoTile(
            icon: Icons.settings,
            label: 'Settings',
            value: 'Configure',
            onTap: () => tapCount++,
          ),
        ),
      ));

      await tester.tap(find.byType(ListTile));
      expect(tapCount, 1);
    });
  });

  group('NavigationTile', () {
    testWidgets('renders with chevron icon', (tester) async {
      await tester.pumpWidget(MaterialApp(
        home: Scaffold(
          body: NavigationTile(
            icon: Icons.notifications,
            label: 'Notifications',
            onTap: () {},
          ),
        ),
      ));

      expect(find.byIcon(Icons.notifications), findsOneWidget);
      expect(find.text('Notifications'), findsOneWidget);
      expect(find.byIcon(Icons.chevron_right), findsOneWidget);
    });

    testWidgets('handles tap', (tester) async {
      var tapCount = 0;

      await tester.pumpWidget(MaterialApp(
        home: Scaffold(
          body: NavigationTile(
            icon: Icons.settings,
            label: 'Settings',
            onTap: () => tapCount++,
          ),
        ),
      ));

      await tester.tap(find.byType(ListTile));
      expect(tapCount, 1);
    });
  });

  group('LoadingOverlay', () {
    testWidgets('shows child when not loading', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: LoadingOverlay(
            isLoading: false,
            child: Text('Content'),
          ),
        ),
      ));

      expect(find.text('Content'), findsOneWidget);
      expect(find.byType(CircularProgressIndicator), findsNothing);
    });

    testWidgets('shows overlay when loading', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: LoadingOverlay(
            isLoading: true,
            child: Text('Content'),
          ),
        ),
      ));

      expect(find.text('Content'), findsOneWidget);
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });

    testWidgets('shows message when provided', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: LoadingOverlay(
            isLoading: true,
            message: 'Loading...',
            child: Text('Content'),
          ),
        ),
      ));

      expect(find.text('Loading...'), findsOneWidget);
    });
  });

  group('CardSection', () {
    testWidgets('renders title and children', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: CardSection(
            title: 'Section',
            children: [Text('Child 1'), Text('Child 2')],
          ),
        ),
      ));

      expect(find.text('Section'), findsOneWidget);
      expect(find.text('Child 1'), findsOneWidget);
      expect(find.text('Child 2'), findsOneWidget);
    });

    testWidgets('renders without title', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: CardSection(
            children: [Text('Child')],
          ),
        ),
      ));

      expect(find.text('Child'), findsOneWidget);
    });
  });

  group('ContentCard', () {
    testWidgets('renders child', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: ContentCard(
            child: Text('Card Content'),
          ),
        ),
      ));

      expect(find.text('Card Content'), findsOneWidget);
    });
  });
}
