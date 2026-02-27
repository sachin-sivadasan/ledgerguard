import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/subscription.dart';
import 'package:ledgerguard/domain/repositories/subscription_repository.dart';
import 'package:ledgerguard/presentation/blocs/subscription_list/subscription_list.dart';

class MockSubscriptionRepository extends Mock implements SubscriptionRepository {}

void main() {
  late MockSubscriptionRepository mockRepository;

  const testAppId = 'app-123';

  final testSubscription1 = Subscription(
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

  final testSubscription2 = Subscription(
    id: 'sub-2',
    shopifyGid: 'gid://shopify/AppSubscription/2',
    myshopifyDomain: 'beta-shop.myshopify.com',
    shopName: 'Beta Shop',
    planName: 'Basic Plan',
    basePriceCents: 999,
    billingInterval: BillingInterval.monthly,
    riskState: RiskState.oneCycleMissed,
    status: 'ACTIVE',
    createdAt: DateTime(2024, 2, 10),
  );

  final testSubscription3 = Subscription(
    id: 'sub-3',
    shopifyGid: 'gid://shopify/AppSubscription/3',
    myshopifyDomain: 'gamma-goods.myshopify.com',
    shopName: 'Gamma Goods',
    planName: 'Enterprise',
    basePriceCents: 9999,
    billingInterval: BillingInterval.annual,
    riskState: RiskState.churned,
    status: 'CANCELLED',
    createdAt: DateTime(2024, 1, 5),
  );

  final testResponse = SubscriptionListResponse(
    subscriptions: [testSubscription1, testSubscription2],
    total: 2,
    limit: 50,
    offset: 0,
  );

  final testResponseWithMore = SubscriptionListResponse(
    subscriptions: [testSubscription1, testSubscription2],
    total: 100,
    limit: 50,
    offset: 0,
  );

  final emptyResponse = const SubscriptionListResponse(
    subscriptions: [],
    total: 0,
    limit: 50,
    offset: 0,
  );

  setUp(() {
    mockRepository = MockSubscriptionRepository();
  });

  setUpAll(() {
    registerFallbackValue(RiskState.safe);
  });

  group('SubscriptionListBloc', () {
    test('initial state is SubscriptionListInitial', () {
      final bloc = SubscriptionListBloc(repository: mockRepository);
      expect(bloc.state, const SubscriptionListInitial());
      bloc.close();
    });

    group('FetchSubscriptionsRequested', () {
      blocTest<SubscriptionListBloc, SubscriptionListState>(
        'emits [Loading, Loaded] when subscriptions are fetched successfully',
        build: () {
          when(() => mockRepository.getSubscriptions(
                testAppId,
                riskState: any(named: 'riskState'),
                limit: any(named: 'limit'),
                offset: any(named: 'offset'),
              )).thenAnswer((_) async => testResponse);
          return SubscriptionListBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const FetchSubscriptionsRequested(appId: testAppId)),
        expect: () => [
          isA<SubscriptionListLoading>(),
          isA<SubscriptionListLoaded>()
              .having((s) => s.subscriptions.length, 'subscriptions count', 2)
              .having((s) => s.total, 'total', 2)
              .having((s) => s.appId, 'appId', testAppId),
        ],
      );

      blocTest<SubscriptionListBloc, SubscriptionListState>(
        'emits [Loading, Empty] when no subscriptions found',
        build: () {
          when(() => mockRepository.getSubscriptions(
                testAppId,
                riskState: any(named: 'riskState'),
                limit: any(named: 'limit'),
                offset: any(named: 'offset'),
              )).thenAnswer((_) async => emptyResponse);
          return SubscriptionListBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const FetchSubscriptionsRequested(appId: testAppId)),
        expect: () => [
          isA<SubscriptionListLoading>(),
          isA<SubscriptionListEmpty>()
              .having((s) => s.appId, 'appId', testAppId),
        ],
      );

      blocTest<SubscriptionListBloc, SubscriptionListState>(
        'emits [Loading, Error] when fetch fails with SubscriptionException',
        build: () {
          when(() => mockRepository.getSubscriptions(
                testAppId,
                riskState: any(named: 'riskState'),
                limit: any(named: 'limit'),
                offset: any(named: 'offset'),
              )).thenThrow(const SubscriptionException('Network error'));
          return SubscriptionListBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const FetchSubscriptionsRequested(appId: testAppId)),
        expect: () => [
          isA<SubscriptionListLoading>(),
          isA<SubscriptionListError>()
              .having((s) => s.message, 'message', 'Network error'),
        ],
      );

      blocTest<SubscriptionListBloc, SubscriptionListState>(
        'emits [Loading, Error] when fetch fails with generic exception',
        build: () {
          when(() => mockRepository.getSubscriptions(
                testAppId,
                riskState: any(named: 'riskState'),
                limit: any(named: 'limit'),
                offset: any(named: 'offset'),
              )).thenThrow(Exception('Unknown error'));
          return SubscriptionListBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const FetchSubscriptionsRequested(appId: testAppId)),
        expect: () => [
          isA<SubscriptionListLoading>(),
          isA<SubscriptionListError>()
              .having((s) => s.message, 'message', contains('Failed to load subscriptions')),
        ],
      );

      blocTest<SubscriptionListBloc, SubscriptionListState>(
        'sets hasMore to true when more subscriptions available',
        build: () {
          when(() => mockRepository.getSubscriptions(
                testAppId,
                riskState: any(named: 'riskState'),
                limit: any(named: 'limit'),
                offset: any(named: 'offset'),
              )).thenAnswer((_) async => testResponseWithMore);
          return SubscriptionListBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const FetchSubscriptionsRequested(appId: testAppId)),
        expect: () => [
          isA<SubscriptionListLoading>(),
          isA<SubscriptionListLoaded>()
              .having((s) => s.hasMore, 'hasMore', true)
              .having((s) => s.total, 'total', 100),
        ],
      );
    });

    group('FilterByRiskStateRequested', () {
      blocTest<SubscriptionListBloc, SubscriptionListState>(
        'filters subscriptions by risk state',
        build: () {
          when(() => mockRepository.getSubscriptions(
                testAppId,
                riskState: any(named: 'riskState'),
                limit: any(named: 'limit'),
                offset: any(named: 'offset'),
              )).thenAnswer((_) async => SubscriptionListResponse(
                subscriptions: [testSubscription1],
                total: 1,
                limit: 50,
                offset: 0,
              ));
          final bloc = SubscriptionListBloc(repository: mockRepository);
          // Set up current app ID by calling fetch first
          bloc.add(const FetchSubscriptionsRequested(appId: testAppId));
          return bloc;
        },
        act: (bloc) async {
          await Future.delayed(const Duration(milliseconds: 100));
          bloc.add(const FilterByRiskStateRequested(riskState: RiskState.safe));
        },
        skip: 2, // Skip initial Loading and Loaded states
        expect: () => [
          isA<SubscriptionListLoaded>()
              .having((s) => s.isRefreshing, 'isRefreshing', true),
          isA<SubscriptionListLoaded>()
              .having((s) => s.filterRiskState, 'filterRiskState', RiskState.safe)
              .having((s) => s.subscriptions.length, 'subscriptions count', 1),
        ],
      );

      blocTest<SubscriptionListBloc, SubscriptionListState>(
        'emits Empty when filter returns no results',
        build: () {
          when(() => mockRepository.getSubscriptions(
                testAppId,
                riskState: any(named: 'riskState'),
                limit: any(named: 'limit'),
                offset: any(named: 'offset'),
              )).thenAnswer((invocation) async {
            final riskState = invocation.namedArguments[#riskState];
            if (riskState == RiskState.churned) {
              return emptyResponse;
            }
            return testResponse;
          });
          final bloc = SubscriptionListBloc(repository: mockRepository);
          bloc.add(const FetchSubscriptionsRequested(appId: testAppId));
          return bloc;
        },
        act: (bloc) async {
          await Future.delayed(const Duration(milliseconds: 100));
          bloc.add(const FilterByRiskStateRequested(riskState: RiskState.churned));
        },
        skip: 2,
        expect: () => [
          isA<SubscriptionListLoaded>()
              .having((s) => s.isRefreshing, 'isRefreshing', true),
          isA<SubscriptionListEmpty>(),
        ],
      );

      blocTest<SubscriptionListBloc, SubscriptionListState>(
        'does nothing when no appId set',
        build: () {
          return SubscriptionListBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const FilterByRiskStateRequested(riskState: RiskState.safe)),
        expect: () => [],
      );
    });

    group('RefreshSubscriptionsRequested', () {
      blocTest<SubscriptionListBloc, SubscriptionListState>(
        'refreshes subscriptions when in loaded state',
        build: () {
          when(() => mockRepository.getSubscriptions(
                testAppId,
                riskState: any(named: 'riskState'),
                limit: any(named: 'limit'),
                offset: any(named: 'offset'),
              )).thenAnswer((_) async => testResponse);
          final bloc = SubscriptionListBloc(repository: mockRepository);
          bloc.add(const FetchSubscriptionsRequested(appId: testAppId));
          return bloc;
        },
        act: (bloc) async {
          await Future.delayed(const Duration(milliseconds: 100));
          bloc.add(const RefreshSubscriptionsRequested());
        },
        skip: 2,
        expect: () => [
          isA<SubscriptionListLoaded>()
              .having((s) => s.isRefreshing, 'isRefreshing', true),
          isA<SubscriptionListLoaded>()
              .having((s) => s.isRefreshing, 'isRefreshing', false),
        ],
      );

      blocTest<SubscriptionListBloc, SubscriptionListState>(
        'keeps current state when refresh fails',
        build: () {
          var callCount = 0;
          when(() => mockRepository.getSubscriptions(
                testAppId,
                riskState: any(named: 'riskState'),
                limit: any(named: 'limit'),
                offset: any(named: 'offset'),
              )).thenAnswer((_) async {
            callCount++;
            if (callCount > 1) {
              throw Exception('Refresh failed');
            }
            return testResponse;
          });
          final bloc = SubscriptionListBloc(repository: mockRepository);
          bloc.add(const FetchSubscriptionsRequested(appId: testAppId));
          return bloc;
        },
        act: (bloc) async {
          await Future.delayed(const Duration(milliseconds: 100));
          bloc.add(const RefreshSubscriptionsRequested());
        },
        skip: 2,
        expect: () => [
          isA<SubscriptionListLoaded>()
              .having((s) => s.isRefreshing, 'isRefreshing', true),
          isA<SubscriptionListLoaded>()
              .having((s) => s.isRefreshing, 'isRefreshing', false)
              .having((s) => s.subscriptions.length, 'subscriptions count', 2),
        ],
      );
    });

    group('LoadMoreSubscriptionsRequested', () {
      blocTest<SubscriptionListBloc, SubscriptionListState>(
        'loads more subscriptions when hasMore is true',
        build: () {
          var callCount = 0;
          when(() => mockRepository.getSubscriptions(
                testAppId,
                riskState: any(named: 'riskState'),
                limit: any(named: 'limit'),
                offset: any(named: 'offset'),
              )).thenAnswer((_) async {
            callCount++;
            if (callCount == 1) {
              return testResponseWithMore;
            }
            return SubscriptionListResponse(
              subscriptions: [testSubscription3],
              total: 100,
              limit: 50,
              offset: 50,
            );
          });
          final bloc = SubscriptionListBloc(repository: mockRepository);
          bloc.add(const FetchSubscriptionsRequested(appId: testAppId));
          return bloc;
        },
        act: (bloc) async {
          await Future.delayed(const Duration(milliseconds: 100));
          bloc.add(const LoadMoreSubscriptionsRequested());
        },
        skip: 2,
        expect: () => [
          isA<SubscriptionListLoaded>()
              .having((s) => s.isLoadingMore, 'isLoadingMore', true),
          isA<SubscriptionListLoaded>()
              .having((s) => s.isLoadingMore, 'isLoadingMore', false)
              .having((s) => s.subscriptions.length, 'subscriptions count', 3),
        ],
      );

      blocTest<SubscriptionListBloc, SubscriptionListState>(
        'does nothing when hasMore is false',
        build: () {
          when(() => mockRepository.getSubscriptions(
                testAppId,
                riskState: any(named: 'riskState'),
                limit: any(named: 'limit'),
                offset: any(named: 'offset'),
              )).thenAnswer((_) async => testResponse);
          final bloc = SubscriptionListBloc(repository: mockRepository);
          bloc.add(const FetchSubscriptionsRequested(appId: testAppId));
          return bloc;
        },
        act: (bloc) async {
          await Future.delayed(const Duration(milliseconds: 100));
          bloc.add(const LoadMoreSubscriptionsRequested());
        },
        skip: 2,
        expect: () => [],
      );
    });
  });
}
