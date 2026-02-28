import 'package:equatable/equatable.dart';

import 'subscription.dart';

/// Sort options for subscription list
enum SubscriptionSort {
  riskState,
  price,
  shopName;

  String get apiValue {
    switch (this) {
      case SubscriptionSort.riskState:
        return 'risk_state';
      case SubscriptionSort.price:
        return 'base_price_cents';
      case SubscriptionSort.shopName:
        return 'shop_name';
    }
  }

  String get displayName {
    switch (this) {
      case SubscriptionSort.riskState:
        return 'Risk';
      case SubscriptionSort.price:
        return 'Price';
      case SubscriptionSort.shopName:
        return 'Name';
    }
  }
}

/// Filters for subscription list queries
class SubscriptionFilters extends Equatable {
  final Set<RiskState> riskStates;
  final int? priceMinCents;
  final int? priceMaxCents;
  final BillingInterval? billingInterval;
  final String? searchQuery;
  final SubscriptionSort sort;
  final bool sortAscending;

  const SubscriptionFilters({
    this.riskStates = const {},
    this.priceMinCents,
    this.priceMaxCents,
    this.billingInterval,
    this.searchQuery,
    this.sort = SubscriptionSort.riskState,
    this.sortAscending = true,
  });

  /// Returns true if any filter is active
  bool get hasActiveFilters =>
      riskStates.isNotEmpty ||
      priceMinCents != null ||
      priceMaxCents != null ||
      billingInterval != null ||
      (searchQuery?.isNotEmpty ?? false);

  /// Count of active filters (excluding search)
  int get activeFilterCount {
    int count = 0;
    if (riskStates.isNotEmpty) count++;
    if (priceMinCents != null || priceMaxCents != null) count++;
    if (billingInterval != null) count++;
    return count;
  }

  SubscriptionFilters copyWith({
    Set<RiskState>? riskStates,
    int? priceMinCents,
    int? priceMaxCents,
    bool clearPriceFilter = false,
    BillingInterval? billingInterval,
    bool clearBillingInterval = false,
    String? searchQuery,
    bool clearSearchQuery = false,
    SubscriptionSort? sort,
    bool? sortAscending,
  }) {
    return SubscriptionFilters(
      riskStates: riskStates ?? this.riskStates,
      priceMinCents: clearPriceFilter ? null : (priceMinCents ?? this.priceMinCents),
      priceMaxCents: clearPriceFilter ? null : (priceMaxCents ?? this.priceMaxCents),
      billingInterval: clearBillingInterval ? null : (billingInterval ?? this.billingInterval),
      searchQuery: clearSearchQuery ? null : (searchQuery ?? this.searchQuery),
      sort: sort ?? this.sort,
      sortAscending: sortAscending ?? this.sortAscending,
    );
  }

  /// Build query parameters for API call
  Map<String, String> toQueryParams() {
    final params = <String, String>{};

    if (riskStates.isNotEmpty) {
      params['status'] = riskStates.map((r) => r.apiValue).join(',');
    }

    if (priceMinCents != null) {
      params['priceMin'] = priceMinCents.toString();
    }
    if (priceMaxCents != null) {
      params['priceMax'] = priceMaxCents.toString();
    }

    if (billingInterval != null) {
      params['billingInterval'] = billingInterval == BillingInterval.monthly ? 'MONTHLY' : 'ANNUAL';
    }

    if (searchQuery?.isNotEmpty ?? false) {
      params['search'] = searchQuery!;
    }

    params['sortBy'] = sort.apiValue;
    params['sortOrder'] = sortAscending ? 'asc' : 'desc';

    return params;
  }

  @override
  List<Object?> get props => [riskStates, priceMinCents, priceMaxCents, billingInterval, searchQuery, sort, sortAscending];
}

/// A distinct price point with its count
class PricePoint extends Equatable {
  final int priceCents;
  final int count;

  const PricePoint({
    required this.priceCents,
    required this.count,
  });

  factory PricePoint.fromJson(Map<String, dynamic> json) {
    return PricePoint(
      priceCents: json['priceCents'] as int,
      count: json['count'] as int,
    );
  }

  String get formatted => '\$${(priceCents / 100).toStringAsFixed(2)}';

  @override
  List<Object?> get props => [priceCents, count];
}

/// Price statistics for building filter UI
class PriceStats extends Equatable {
  final int minCents;
  final int maxCents;
  final int avgCents;
  final List<PricePoint> prices;

  const PriceStats({
    required this.minCents,
    required this.maxCents,
    required this.avgCents,
    this.prices = const [],
  });

  factory PriceStats.fromJson(Map<String, dynamic> json) {
    final pricesList = (json['prices'] as List<dynamic>? ?? [])
        .map((p) => PricePoint.fromJson(p as Map<String, dynamic>))
        .toList();

    return PriceStats(
      minCents: json['minCents'] as int,
      maxCents: json['maxCents'] as int,
      avgCents: json['avgCents'] as int,
      prices: pricesList,
    );
  }

  String get formattedMin => '\$${(minCents / 100).toStringAsFixed(2)}';
  String get formattedMax => '\$${(maxCents / 100).toStringAsFixed(2)}';
  String get formattedAvg => '\$${(avgCents / 100).toStringAsFixed(2)}';

  @override
  List<Object?> get props => [minCents, maxCents, avgCents, prices];
}

/// Aggregate subscription statistics
class SubscriptionSummary extends Equatable {
  final int activeCount;
  final int atRiskCount;
  final int churnedCount;
  final int avgPriceCents;
  final int totalCount;

  const SubscriptionSummary({
    required this.activeCount,
    required this.atRiskCount,
    required this.churnedCount,
    required this.avgPriceCents,
    required this.totalCount,
  });

  factory SubscriptionSummary.fromJson(Map<String, dynamic> json) {
    return SubscriptionSummary(
      activeCount: json['activeCount'] as int,
      atRiskCount: json['atRiskCount'] as int,
      churnedCount: json['churnedCount'] as int,
      avgPriceCents: json['avgPriceCents'] as int,
      totalCount: json['totalCount'] as int,
    );
  }

  String get formattedAvgPrice {
    final dollars = avgPriceCents / 100;
    return '\$${dollars.toStringAsFixed(2)}';
  }

  @override
  List<Object?> get props => [activeCount, atRiskCount, churnedCount, avgPriceCents, totalCount];
}

/// Paginated subscription list response (new format)
class PaginatedSubscriptionResponse extends Equatable {
  final List<Subscription> subscriptions;
  final int total;
  final int page;
  final int pageSize;
  final int totalPages;

  const PaginatedSubscriptionResponse({
    required this.subscriptions,
    required this.total,
    required this.page,
    required this.pageSize,
    required this.totalPages,
  });

  bool get hasNextPage => page < totalPages;
  bool get hasPreviousPage => page > 1;

  /// Range text like "1-25 of 847"
  String get rangeText {
    final start = (page - 1) * pageSize + 1;
    final end = start + subscriptions.length - 1;
    return '$start-$end of $total';
  }

  factory PaginatedSubscriptionResponse.fromJson(Map<String, dynamic> json) {
    final subscriptionsList = (json['subscriptions'] as List<dynamic>? ?? [])
        .map((s) => Subscription.fromJson(s as Map<String, dynamic>))
        .toList();

    return PaginatedSubscriptionResponse(
      subscriptions: subscriptionsList,
      total: json['total'] as int? ?? subscriptionsList.length,
      page: json['page'] as int? ?? 1,
      pageSize: json['pageSize'] as int? ?? 25,
      totalPages: json['totalPages'] as int? ?? 1,
    );
  }

  @override
  List<Object?> get props => [subscriptions, total, page, pageSize, totalPages];
}
