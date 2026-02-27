import 'package:bloc_test/bloc_test.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/notification_preferences.dart';
import 'package:ledgerguard/domain/repositories/notification_preferences_repository.dart';
import 'package:ledgerguard/presentation/blocs/notification_preferences/notification_preferences.dart';

class MockNotificationPreferencesRepository extends Mock
    implements NotificationPreferencesRepository {}

class FakeNotificationPreferences extends Fake implements NotificationPreferences {}

void main() {
  setUpAll(() {
    registerFallbackValue(FakeNotificationPreferences());
  });
  late MockNotificationPreferencesRepository mockRepository;

  const testPreferences = NotificationPreferences(
    criticalAlertsEnabled: true,
    dailySummaryEnabled: true,
    dailySummaryTime: TimeOfDay(hour: 9, minute: 0),
  );

  const updatedPreferences = NotificationPreferences(
    criticalAlertsEnabled: false,
    dailySummaryEnabled: true,
    dailySummaryTime: TimeOfDay(hour: 9, minute: 0),
  );

  setUp(() {
    mockRepository = MockNotificationPreferencesRepository();
  });

  group('NotificationPreferencesBloc', () {
    test('initial state is NotificationPreferencesInitial', () {
      final bloc = NotificationPreferencesBloc(repository: mockRepository);
      expect(bloc.state, const NotificationPreferencesInitial());
      bloc.close();
    });

    group('LoadNotificationPreferencesRequested', () {
      blocTest<NotificationPreferencesBloc, NotificationPreferencesState>(
        'emits [Loading, Loaded] when fetch succeeds',
        build: () {
          when(() => mockRepository.fetchPreferences())
              .thenAnswer((_) async => testPreferences);
          return NotificationPreferencesBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const LoadNotificationPreferencesRequested()),
        expect: () => [
          isA<NotificationPreferencesLoading>(),
          isA<NotificationPreferencesLoaded>()
              .having((s) => s.preferences, 'preferences', testPreferences),
        ],
      );

      blocTest<NotificationPreferencesBloc, NotificationPreferencesState>(
        'emits [Loading, Error] when fetch fails',
        build: () {
          when(() => mockRepository.fetchPreferences())
              .thenThrow(const LoadNotificationPreferencesException('Network error'));
          return NotificationPreferencesBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const LoadNotificationPreferencesRequested()),
        expect: () => [
          isA<NotificationPreferencesLoading>(),
          isA<NotificationPreferencesError>()
              .having((s) => s.message, 'message', 'Network error'),
        ],
      );

      blocTest<NotificationPreferencesBloc, NotificationPreferencesState>(
        'emits [Loading, Error] when generic exception is thrown',
        build: () {
          when(() => mockRepository.fetchPreferences())
              .thenThrow(Exception('Unexpected error'));
          return NotificationPreferencesBloc(repository: mockRepository);
        },
        act: (bloc) => bloc.add(const LoadNotificationPreferencesRequested()),
        expect: () => [
          isA<NotificationPreferencesLoading>(),
          isA<NotificationPreferencesError>()
              .having((s) => s.message, 'message', contains('Failed to load')),
        ],
      );
    });

    group('ToggleCriticalAlertsRequested', () {
      blocTest<NotificationPreferencesBloc, NotificationPreferencesState>(
        'updates preferences with new value',
        build: () {
          when(() => mockRepository.fetchPreferences())
              .thenAnswer((_) async => testPreferences);
          return NotificationPreferencesBloc(repository: mockRepository);
        },
        seed: () => NotificationPreferencesLoaded(preferences: testPreferences),
        act: (bloc) => bloc.add(const ToggleCriticalAlertsRequested(enabled: false)),
        expect: () => [
          isA<NotificationPreferencesLoaded>()
              .having((s) => s.preferences.criticalAlertsEnabled, 'enabled', false)
              .having((s) => s.hasUnsavedChanges, 'hasUnsavedChanges', true),
        ],
      );

      blocTest<NotificationPreferencesBloc, NotificationPreferencesState>(
        'does nothing when not in loaded state',
        build: () => NotificationPreferencesBloc(repository: mockRepository),
        act: (bloc) => bloc.add(const ToggleCriticalAlertsRequested(enabled: false)),
        expect: () => [],
      );
    });

    group('ToggleDailySummaryRequested', () {
      blocTest<NotificationPreferencesBloc, NotificationPreferencesState>(
        'updates preferences with new value',
        build: () {
          when(() => mockRepository.fetchPreferences())
              .thenAnswer((_) async => testPreferences);
          return NotificationPreferencesBloc(repository: mockRepository);
        },
        seed: () => NotificationPreferencesLoaded(preferences: testPreferences),
        act: (bloc) => bloc.add(const ToggleDailySummaryRequested(enabled: false)),
        expect: () => [
          isA<NotificationPreferencesLoaded>()
              .having((s) => s.preferences.dailySummaryEnabled, 'enabled', false)
              .having((s) => s.hasUnsavedChanges, 'hasUnsavedChanges', true),
        ],
      );
    });

    group('UpdateDailySummaryTimeRequested', () {
      blocTest<NotificationPreferencesBloc, NotificationPreferencesState>(
        'updates preferences with new time',
        build: () {
          when(() => mockRepository.fetchPreferences())
              .thenAnswer((_) async => testPreferences);
          return NotificationPreferencesBloc(repository: mockRepository);
        },
        seed: () => NotificationPreferencesLoaded(preferences: testPreferences),
        act: (bloc) => bloc.add(
            const UpdateDailySummaryTimeRequested(time: TimeOfDay(hour: 14, minute: 30))),
        expect: () => [
          isA<NotificationPreferencesLoaded>()
              .having((s) => s.preferences.dailySummaryTime.hour, 'hour', 14)
              .having((s) => s.preferences.dailySummaryTime.minute, 'minute', 30)
              .having((s) => s.hasUnsavedChanges, 'hasUnsavedChanges', true),
        ],
      );
    });

    group('SaveNotificationPreferencesRequested', () {
      blocTest<NotificationPreferencesBloc, NotificationPreferencesState>(
        'emits [Loaded(saving), Saved, Loaded] when save succeeds',
        build: () {
          when(() => mockRepository.fetchPreferences())
              .thenAnswer((_) async => testPreferences);
          when(() => mockRepository.savePreferences(any()))
              .thenAnswer((_) async {});
          return NotificationPreferencesBloc(repository: mockRepository);
        },
        seed: () => NotificationPreferencesLoaded(
          preferences: updatedPreferences,
          hasUnsavedChanges: true,
        ),
        act: (bloc) => bloc.add(const SaveNotificationPreferencesRequested()),
        expect: () => [
          isA<NotificationPreferencesLoaded>()
              .having((s) => s.isSaving, 'isSaving', true),
          isA<NotificationPreferencesSaved>(),
          isA<NotificationPreferencesLoaded>()
              .having((s) => s.hasUnsavedChanges, 'hasUnsavedChanges', false),
        ],
      );

      blocTest<NotificationPreferencesBloc, NotificationPreferencesState>(
        'emits [Loaded(saving), Error, Loaded] when save fails',
        build: () {
          when(() => mockRepository.fetchPreferences())
              .thenAnswer((_) async => testPreferences);
          when(() => mockRepository.savePreferences(any()))
              .thenThrow(const SaveNotificationPreferencesException('Save failed'));
          return NotificationPreferencesBloc(repository: mockRepository);
        },
        seed: () => NotificationPreferencesLoaded(
          preferences: updatedPreferences,
          hasUnsavedChanges: true,
        ),
        act: (bloc) => bloc.add(const SaveNotificationPreferencesRequested()),
        expect: () => [
          isA<NotificationPreferencesLoaded>()
              .having((s) => s.isSaving, 'isSaving', true),
          isA<NotificationPreferencesError>()
              .having((s) => s.message, 'message', 'Save failed'),
          isA<NotificationPreferencesLoaded>()
              .having((s) => s.isSaving, 'isSaving', false),
        ],
      );

      blocTest<NotificationPreferencesBloc, NotificationPreferencesState>(
        'does nothing when not in loaded state',
        build: () => NotificationPreferencesBloc(repository: mockRepository),
        act: (bloc) => bloc.add(const SaveNotificationPreferencesRequested()),
        expect: () => [],
      );
    });
  });

  group('NotificationPreferences', () {
    test('fromJson parses correctly', () {
      final json = {
        'critical_alerts_enabled': true,
        'daily_summary_enabled': false,
        'daily_summary_time': '14:30',
      };

      final prefs = NotificationPreferences.fromJson(json);
      expect(prefs.criticalAlertsEnabled, true);
      expect(prefs.dailySummaryEnabled, false);
      expect(prefs.dailySummaryTime.hour, 14);
      expect(prefs.dailySummaryTime.minute, 30);
    });

    test('fromJson handles missing values with defaults', () {
      final json = <String, dynamic>{};

      final prefs = NotificationPreferences.fromJson(json);
      expect(prefs.criticalAlertsEnabled, true);
      expect(prefs.dailySummaryEnabled, true);
      expect(prefs.dailySummaryTime.hour, 9);
      expect(prefs.dailySummaryTime.minute, 0);
    });

    test('toJson serializes correctly', () {
      const prefs = NotificationPreferences(
        criticalAlertsEnabled: false,
        dailySummaryEnabled: true,
        dailySummaryTime: TimeOfDay(hour: 8, minute: 15),
      );

      final json = prefs.toJson();
      expect(json['critical_alerts_enabled'], false);
      expect(json['daily_summary_enabled'], true);
      expect(json['daily_summary_time'], '08:15');
    });

    test('copyWith creates copy with updated values', () {
      const original = NotificationPreferences(
        criticalAlertsEnabled: true,
        dailySummaryEnabled: true,
        dailySummaryTime: TimeOfDay(hour: 9, minute: 0),
      );

      final copy = original.copyWith(criticalAlertsEnabled: false);
      expect(copy.criticalAlertsEnabled, false);
      expect(copy.dailySummaryEnabled, true);
      expect(copy.dailySummaryTime.hour, 9);
    });

    test('formattedTime returns correct format', () {
      const morning = NotificationPreferences(
        dailySummaryTime: TimeOfDay(hour: 9, minute: 30),
      );
      expect(morning.formattedTime, '9:30 AM');

      const afternoon = NotificationPreferences(
        dailySummaryTime: TimeOfDay(hour: 14, minute: 0),
      );
      expect(afternoon.formattedTime, '2:00 PM');

      const noon = NotificationPreferences(
        dailySummaryTime: TimeOfDay(hour: 12, minute: 0),
      );
      expect(noon.formattedTime, '12:00 PM');

      const midnight = NotificationPreferences(
        dailySummaryTime: TimeOfDay(hour: 0, minute: 0),
      );
      expect(midnight.formattedTime, '12:00 AM');
    });
  });

  group('NotificationPreferencesException', () {
    test('UnauthorizedNotificationPreferencesException has correct message', () {
      const exception = UnauthorizedNotificationPreferencesException();
      expect(exception.message, contains('authenticated'));
      expect(exception.code, 'unauthorized');
    });

    test('SaveNotificationPreferencesException has correct message', () {
      const exception = SaveNotificationPreferencesException('Custom message');
      expect(exception.message, 'Custom message');
      expect(exception.code, 'save-failed');
    });

    test('LoadNotificationPreferencesException has correct message', () {
      const exception = LoadNotificationPreferencesException();
      expect(exception.message, contains('Failed to load'));
      expect(exception.code, 'load-failed');
    });
  });
}
