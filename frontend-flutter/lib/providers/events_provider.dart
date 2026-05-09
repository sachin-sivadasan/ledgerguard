import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../mock_data/mock_events.dart';
import '../models/event_model.dart';
import '../services/events_service.dart';

enum TimeRange { today, thisWeek, thisMonth }

class EventsProvider extends ChangeNotifier {
  final EventsService _eventsService;

  bool _demoMode = false;
  bool _isLoading = false;
  bool _isLoadingMore = false;
  String? _error;
  CancelToken? _cancelToken;
  Timer? _storeSearchDebounce;

  EventType? _typeFilter;
  String? _selectedAppId;
  String? _storeFilter;
  TimeRange _timeRange = TimeRange.thisWeek;

  List<AppEvent> _liveEvents = [];
  int _currentPage = 1;
  int _totalPages = 1;
  int _totalCount = 0;
  static const int _pageSize = 20;

  static const _typeToApiFilter = {
    EventType.appInstall: 'RELATIONSHIP_INSTALLED',
    EventType.appUninstall: 'RELATIONSHIP_UNINSTALLED',
    EventType.appReactivated: 'RELATIONSHIP_REACTIVATED',
    EventType.appDeactivated: 'RELATIONSHIP_DEACTIVATED',
    EventType.subscriptionActivated: 'SUBSCRIPTION_CHARGE_ACCEPTED',
    EventType.subscriptionCancelled: 'SUBSCRIPTION_CHARGE_CANCELED',
    EventType.subscriptionFrozen: 'SUBSCRIPTION_CHARGE_FROZEN',
    EventType.subscriptionUnfrozen: 'SUBSCRIPTION_CHARGE_UNFROZEN',
  };

  EventsProvider(this._eventsService);

  bool get demoMode => _demoMode;
  bool get isLoading => _isLoading;
  bool get isLoadingMore => _isLoadingMore;
  String? get error => _error;
  EventType? get typeFilter => _typeFilter;
  String? get selectedAppId => _selectedAppId;
  String? get storeFilter => _storeFilter;
  TimeRange get timeRange => _timeRange;
  bool get hasMore => _currentPage < _totalPages;
  int get totalCount => _totalCount;

  void setDemoMode(bool value) {
    _demoMode = value;
    notifyListeners();
  }

  Future<void> loadEvents(String appId) async {
    if (_demoMode) return;
    _cancelToken?.cancel('Superseded');
    _cancelToken = CancelToken();
    _isLoading = true;
    _error = null;
    _currentPage = 1;
    _liveEvents = [];
    notifyListeners();
    try {
      final result = await _eventsService.fetchEvents(
        appId,
        page: 1,
        pageSize: _pageSize,
        storeDomain: _storeFilter != null && _storeFilter!.isNotEmpty ? _storeFilter : null,
        eventType: _typeFilter != null ? _typeToApiFilter[_typeFilter!] : null,
        since: _startOfRange(),
        cancelToken: _cancelToken,
      );
      _liveEvents = result.items;
      _totalCount = result.total;
      _totalPages = result.totalPages;
      _currentPage = result.page;
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) return;
      _error = e.message;
    } catch (e) {
      _error = e.toString();
    }
    _isLoading = false;
    notifyListeners();
  }

  Future<void> loadMore() async {
    if (_demoMode || _isLoadingMore || !hasMore || _selectedAppId == null) {
      return;
    }
    _isLoadingMore = true;
    notifyListeners();
    try {
      final result = await _eventsService.fetchEvents(
        _selectedAppId!,
        page: _currentPage + 1,
        pageSize: _pageSize,
        storeDomain: _storeFilter != null && _storeFilter!.isNotEmpty ? _storeFilter : null,
        eventType: _typeFilter != null ? _typeToApiFilter[_typeFilter!] : null,
        since: _startOfRange(),
      );
      _liveEvents.addAll(result.items);
      _currentPage = result.page;
      _totalPages = result.totalPages;
      _totalCount = result.total;
    } catch (e) {
      debugPrint('[EventsProvider] loadMore error: $e');
    }
    _isLoadingMore = false;
    notifyListeners();
  }

  List<AppEvent> get _allEvents =>
      _demoMode ? mockEvents : _liveEvents;

  List<AppEvent> get events {
    var list = _allEvents.toList();

    if (_typeFilter != null) {
      list = list.where((e) => e.type == _typeFilter).toList();
    }
    // Only filter by app in demo mode — live data is already fetched per-app
    if (_selectedAppId != null && _demoMode) {
      list = list.where((e) => e.appId == _selectedAppId).toList();
    }
    // Client-side store filter only in demo mode — live mode uses server-side filter
    if (_storeFilter != null && _demoMode) {
      list = list
          .where((e) => e.storeDomain.contains(_storeFilter!))
          .toList();
    }

    list.sort((a, b) => b.date.compareTo(a.date));
    return list;
  }

  int get totalEvents => _demoMode ? events.length : _totalCount;

  DateTime _startOfRange() {
    final now = DateTime.now();
    return switch (_timeRange) {
      TimeRange.today => DateTime(now.year, now.month, now.day),
      TimeRange.thisWeek => now.subtract(const Duration(days: 7)),
      TimeRange.thisMonth => DateTime(now.year, now.month, 1),
    };
  }

  int _countInRange(EventType type) {
    final cutoff = _startOfRange();
    return _allEvents
        .where((e) => e.type == type && e.date.isAfter(cutoff))
        .length;
  }

  int get installs => _countInRange(EventType.appInstall);
  int get uninstalls => _countInRange(EventType.appUninstall);
  int get churns => _countInRange(EventType.subscriptionCancelled);
  int get billingFailures => _countInRange(EventType.billingFailure);

  void setTimeRange(TimeRange range) {
    _timeRange = range;
    if (!_demoMode && _selectedAppId != null) {
      loadEvents(_selectedAppId!);
    } else {
      notifyListeners();
    }
  }

  List<AppEvent> get recentEvents => events.take(5).toList();

  Map<String, int> get weeklyActivity {
    final weekAgo = DateTime.now().subtract(const Duration(days: 7));
    final source = _allEvents;
    final types = {
      'Installs': EventType.appInstall,
      'Uninstalls': EventType.appUninstall,
      'Churns': EventType.subscriptionCancelled,
      'Bill. Failures': EventType.billingFailure,
      'Upgrades': EventType.planUpgrade,
      'Downgrades': EventType.planDowngrade,
    };
    return {
      for (final entry in types.entries)
        entry.key: source
            .where(
                (e) => e.type == entry.value && e.date.isAfter(weekAgo))
            .length,
    };
  }

  List<AppEvent> eventsForStore(String storeDomain) {
    return _allEvents
        .where((e) => e.storeDomain == storeDomain)
        .toList()
      ..sort((a, b) => b.date.compareTo(a.date));
  }

  Map<EventType, int> get countsByType {
    final counts = <EventType, int>{};
    for (final e in events) {
      counts[e.type] = (counts[e.type] ?? 0) + 1;
    }
    return counts;
  }

  void setTypeFilter(EventType? type) {
    _typeFilter = type;
    if (!_demoMode && _selectedAppId != null) {
      loadEvents(_selectedAppId!);
    }
    notifyListeners();
  }

  void setSelectedApp(String? appId) {
    _selectedAppId = appId;
    notifyListeners();
    if (!_demoMode && appId != null) {
      loadEvents(appId);
    }
  }

  void setStoreFilter(String? store) {
    _storeFilter = store;
    notifyListeners();
    if (!_demoMode && _selectedAppId != null) {
      _storeSearchDebounce?.cancel();
      _storeSearchDebounce = Timer(const Duration(milliseconds: 300), () {
        loadEvents(_selectedAppId!);
      });
    }
  }

  void clearFilters() {
    _typeFilter = null;
    _storeFilter = null;
    if (!_demoMode && _selectedAppId != null) {
      loadEvents(_selectedAppId!);
    }
    notifyListeners();
  }
}
