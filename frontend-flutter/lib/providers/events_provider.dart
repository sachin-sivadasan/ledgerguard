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
  String? _error;
  CancelToken? _cancelToken;

  EventType? _typeFilter;
  String? _appFilter;
  String? _storeFilter;
  TimeRange _timeRange = TimeRange.thisWeek;

  List<AppEvent> _liveEvents = [];

  EventsProvider(this._eventsService);

  bool get demoMode => _demoMode;
  bool get isLoading => _isLoading;
  String? get error => _error;
  EventType? get typeFilter => _typeFilter;
  String? get appFilter => _appFilter;
  String? get storeFilter => _storeFilter;
  TimeRange get timeRange => _timeRange;

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
    notifyListeners();
    try {
      _liveEvents = await _eventsService.fetchEvents(appId,
          cancelToken: _cancelToken);
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) return;
      _error = e.message;
    } catch (e) {
      _error = e.toString();
    }
    _isLoading = false;
    notifyListeners();
  }

  List<AppEvent> get _allEvents =>
      _demoMode ? mockEvents : _liveEvents;

  List<AppEvent> get events {
    var list = _allEvents.toList();

    if (_typeFilter != null) {
      list = list.where((e) => e.type == _typeFilter).toList();
    }
    if (_appFilter != null) {
      list = list.where((e) => e.appId == _appFilter).toList();
    }
    if (_storeFilter != null) {
      list = list
          .where((e) => e.storeDomain.contains(_storeFilter!))
          .toList();
    }

    list.sort((a, b) => b.date.compareTo(a.date));
    return list;
  }

  int get totalEvents => events.length;

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
    return events
        .where((e) => e.type == type && e.date.isAfter(cutoff))
        .length;
  }

  int get installs => _countInRange(EventType.appInstall);
  int get uninstalls => _countInRange(EventType.appUninstall);
  int get churns => _countInRange(EventType.subscriptionCancelled);
  int get billingFailures => _countInRange(EventType.billingFailure);

  void setTimeRange(TimeRange range) {
    _timeRange = range;
    notifyListeners();
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
    notifyListeners();
  }

  void setAppFilter(String? appId) {
    _appFilter = appId;
    notifyListeners();
    if (!_demoMode && appId != null) {
      loadEvents(appId);
    }
  }

  void setStoreFilter(String? store) {
    _storeFilter = store;
    notifyListeners();
  }

  void clearFilters() {
    _typeFilter = null;
    _appFilter = null;
    _storeFilter = null;
    notifyListeners();
  }
}
