import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/subscription.dart';
import 'package:ledgerguard/domain/entities/subscription_filter.dart';
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

  const testSummary = SubscriptionSummary(
    activeCount: 10,
    atRiskCount: 3,
    churnedCount: 2,
    avgPriceCents: 2999,
    totalCount: 15,
  );

  const testPriceStats = PriceStats(
    minCents: 999,
    maxCents: 9999,
    avgCents: 2999,
  );

  final testPaginatedResponse = PaginatedSubscriptionResponse(
    subscriptions: [testSubscription1, testSubscription2],
    total: 2,
    page: 1,
    pageSize: 25,
    totalPages: 1,
  );

  final testPaginatedResponseWithMore = PaginatedSubscriptionResponse(
    subscriptions: [testSubscription1, testSubscription2],
    total: 100,
    page: 1,
    pageSize: 25,
    totalPages: 4,
  );

  final emptyPaginatedResponse = const PaginatedSubscriptionResponse(
    subscriptions: [],
    total: 0,
    page: 1,
    pageSize: 25,
    totalPages: 0,
  );

  setUp(() {
    mockRepository = MockSubscriptionRepository();
  });

  setUpAll(() {
    registerFallbackValue(RiskState.safe);
    registerFallbackValue(const SubscriptionFilters());
  });

  void setupDefaultMocks() {
    when(() => mockRepository.getSummary(any()))
        .thenAnswer((_) async => testSummary);
    when(() => mockRepository.getPriceStats(any()))
        .thenAnswer((_) async => testPriceStats);
  }

  group('SubscriptionListBloc', () {
    test('initial state is SubscriptionListInitial', () {
      final bloc = SubscriptionListBloc(repository: mockRepository);
      expect(bloc.state, const SubscriptionListInitial());
      bloc.close();
    });

    group('LoadSubscriptionsRequested', () {
      blocTest<SubscriptionListBloc, SubscriptionListState>(
        'emits [Loading, Loaded] when subscriptions are fetched successfully',
        build: () {
          setupDefaultMocks();
          when(() => mockRepository.getSubscriptionsFiltered(
                testAppId,
                filters: any(named: 'filters'),
                page: any(named: 'page'),
                pageSize: any(named: 'pageSize'),
              )).thenAnswer((_) async => testPaginatedResponse);
          return SubscriptionListBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(LoadSubscriptionsRequested(appId: testAppId)),
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
          setupDefaultMocks();
          when(() => mockRepository.getSubscriptionsFiltered(
                testAppId,
                filters: any(named: 'filters'),
                page: any(named: 'page'),
                pageSize: any(named: 'pageSize'),
              )).thenAnswer((_) async => emptyPaginatedResponse);
          return SubscriptionListBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(LoadSubscriptionsRequested(appId: testAppId)),
        expect: () => [
          isA<SubscriptionListLoading>(),
          isA<SubscriptionListEmpty>()
              .having((s) => s.appId, 'appId', testAppId),
        ],
      );

      blocTest<SubscriptionListBloc, SubscriptionListState>(
        'emits [Loading, Error] when fetch fails with SubscriptionException',
        build: () {
          setupDefaultMocks();
          when(() => mockRepository.getSubscriptionsFiltered(
                testAppId,
                filters: any(named: 'filters'),
                page: any(named: 'page'),
                pageSize: any(named: 'pageSize'),
              )).thenThrow(const SubscriptionException('Network error'));
          return SubscriptionListBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(LoadSubscriptionsRequested(appId: testAppId)),
        expect: () => [
          isA<SubscriptionListLoading>(),
          isA<SubscriptionListError>()
              .having((s) => s.message, 'message', 'Network error'),
        ],
      );

      blocTest<SubscriptionListBloc, SubscriptionListState>(
        'emits [Loading, Error] when fetch fails with generic exception',
        build: () {
          setupDefaultMocks();
          when(() => mockRepository.getSubscriptionsFiltered(
                testAppId,
                filters: any(named: 'filters'),
                page: any(named: 'page'),
                pageSize: any(named: 'pageSize'),
              )).thenThrow(Exception('Unknown error'));
          return SubscriptionListBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(LoadSubscriptionsRequested(appId: testAppId)),
        expect: () => [
          isA<SubscriptionListLoading>(),
          isA<SubscriptionListError>()
              .having((s) => s.message, 'message', contains('Failed to load subscriptions')),
        ],
      );

      blocTest<SubscriptionListBloc, SubscriptionListState>(
        'sets hasMore to true when more subscriptions available',
        build: () {
          setupDefaultMocks();
          when(() => mockRepository.getSubscriptionsFiltered(
                testAppId,
                filters: any(named: 'filters'),
                page: any(named: 'page'),
                pageSize: any(named: 'pageSize'),
              )).thenAnswer((_) async => testPaginatedResponseWithMore);
          return SubscriptionListBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(LoadSubscriptionsRequested(appId: testAppId)),
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
          setupDefaultMocks();
          when(() => mockRepository.getSubscriptionsFiltered(
                testAppId,
                filters: any(named: 'filters'),
                page: any(named: 'page'),
                pageSize: any(named: 'pageSize'),
              )).thenAnswer((_) async => PaginatedSubscriptionResponse(
                subscriptions: [testSubscription1],
                total: 1,
                page: 1,
                pageSize: 25,
                totalPages: 1,
              ));
          final bloc = SubscriptionListBloc(repository: mockRepository);
          // Set up current app ID by calling load first
          bloc.add(LoadSubscriptionsRequested(appId: testAppId));
          return bloc;
        },
        act: (bloc) async {
          await Future.delayed(const Duration(milliseconds: 100));
          bloc.add(const FilterByRiskStateRequested(riskState: RiskState.safe));
        },
        skip: 2, // Skip initial Loading and Loaded states
        expect: () => [
          isA<SubscriptionListLoaded>()
              .having((s) => s.isLoading, 'isLoading', true),
          isA<SubscriptionListLoaded>()
              .having((s) => s.filterRiskState, 'filterRiskState', RiskState.safe)
              .having((s) => s.subscriptions.length, 'subscriptions count', 1),
        ],
      );

      blocTest<SubscriptionListBloc, SubscriptionListState>(
        'emits Empty when filter returns no results',
        build: () {
          setupDefaultMocks();
          var callCount = 0;
          when(() => mockRepository.getSubscriptionsFiltered(
                testAppId,
                filters: any(named: 'filters'),
                page: any(named: 'page'),
                pageSize: any(named: 'pageSize'),
              )).thenAnswer((_) async {
            callCount++;
            if (callCount > 1) {
              return emptyPaginatedResponse;
            }
            return testPaginatedResponse;
          });
          final bloc = SubscriptionListBloc(repository: mockRepository);
          bloc.add(LoadSubscriptionsRequested(appId: testAppId));
          return bloc;
        },
        act: (bloc) async {
          await Future.delayed(const Duration(milliseconds: 100));
          bloc.add(const FilterByRiskStateRequested(riskState: RiskState.churned));
        },
        skip: 2,
        expect: () => [
          isA<SubscriptionListLoaded>()
              .having((s) => s.isLoading, 'isLoading', true),
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
          setupDefaultMocks();
          when(() => mockRepository.getSubscriptionsFiltered(
                testAppId,
                filters: any(named: 'filters'),
                page: any(named: 'page'),
                pageSize: any(named: 'pageSize'),
              )).thenAnswer((_) async => testPaginatedResponse);
          final bloc = SubscriptionListBloc(repository: mockRepository);
          bloc.add(LoadSubscriptionsRequested(appId: testAppId));
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
          setupDefaultMocks();
          var callCount = 0;
          when(() => mockRepository.getSubscriptionsFiltered(
                testAppId,
                filters: any(named: 'filters'),
                page: any(named: 'page'),
                pageSize: any(named: 'pageSize'),
              )).thenAnswer((_) async {
            callCount++;
            if (callCount > 1) {
              throw Exception('Refresh failed');
            }
            return testPaginatedResponse;
          });
          final bloc = SubscriptionListBloc(repository: mockRepository);
          bloc.add(LoadSubscriptionsRequested(appId: testAppId));
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
          setupDefaultMocks();
          var callCount = 0;
          when(() => mockRepository.getSubscriptionsFiltered(
                testAppId,
                filters: any(named: 'filters'),
                page: any(named: 'page'),
                pageSize: any(named: 'pageSize'),
              )).thenAnswer((_) async {
            callCount++;
            if (callCount == 1) {
              return testPaginatedResponseWithMore;
            }
            return PaginatedSubscriptionResponse(
              subscriptions: [testSubscription3],
              total: 100,
              page: 2,
              pageSize: 25,
              totalPages: 4,
            );
          });
          final bloc = SubscriptionListBloc(repository: mockRepository);
          bloc.add(LoadSubscriptionsRequested(appId: testAppId));
          return bloc;
        },
        act: (bloc) async {
          await Future.delayed(const Duration(milliseconds: 100));
          bloc.add(const LoadMoreSubscriptionsRequested());
        },
        skip: 2,
        expect: () => [
          isA<SubscriptionListLoaded>()
              .having((s) => s.isLoading, 'isLoading', true),
          isA<SubscriptionListLoaded>()
              .having((s) => s.isLoading, 'isLoading', false),
        ],
      );

      blocTest<SubscriptionListBloc, SubscriptionListState>(
        'does nothing when hasMore is false',
        build: () {
          setupDefaultMocks();
          when(() => mockRepository.getSubscriptionsFiltered(
                testAppId,
                filters: any(named: 'filters'),
                page: any(named: 'page'),
                pageSize: any(named: 'pageSize'),
              )).thenAnswer((_) async => testPaginatedResponse);
          final bloc = SubscriptionListBloc(repository: mockRepository);
          bloc.add(LoadSubscriptionsRequested(appId: testAppId));
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
