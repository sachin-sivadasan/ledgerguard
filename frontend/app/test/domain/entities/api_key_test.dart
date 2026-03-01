import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard/domain/entities/api_key.dart';

void main() {
  group('ApiKeyCreationResult', () {
    test('fromJson parses backend response correctly', () {
      // Exact format from backend Create endpoint
      final json = {
        'api_key': {
          'id': '89fe9fa7-1234-5678-9abc-def012345678',
          'name': 'Test Key',
          'key_prefix': 'lgk_test1234...',
          'created_at': '2024-01-15T10:30:00Z',
          'last_used_at': null,
        },
        'full_key': 'lgk_live_abcdefghijklmnopqrstuvwxyz1234567890_full_secret',
      };

      final result = ApiKeyCreationResult.fromJson(json);

      expect(
        result.fullKey,
        equals('lgk_live_abcdefghijklmnopqrstuvwxyz1234567890_full_secret'),
      );
      expect(result.apiKey.id, equals('89fe9fa7-1234-5678-9abc-def012345678'));
      expect(result.apiKey.name, equals('Test Key'));
      expect(result.apiKey.keyPrefix, equals('lgk_test1234...'));
    });

    test('fullKey is different from keyPrefix', () {
      final json = {
        'api_key': {
          'id': 'test-id',
          'name': 'My Key',
          'key_prefix': 'lgk_abc12345...',
          'created_at': '2024-01-15T10:30:00Z',
        },
        'full_key': 'lgk_abc1234567890abcdefghij_complete_key',
      };

      final result = ApiKeyCreationResult.fromJson(json);

      expect(result.fullKey, isNot(contains('...')));
      expect(result.apiKey.keyPrefix, contains('...'));
    });
  });

  group('ApiKey', () {
    test('fromJson parses correctly', () {
      final json = {
        'id': 'key-123',
        'name': 'Production',
        'key_prefix': 'lgk_prod1234...',
        'created_at': '2024-02-20T15:45:00Z',
        'last_used_at': '2024-02-25T10:00:00Z',
      };

      final apiKey = ApiKey.fromJson(json);

      expect(apiKey.id, equals('key-123'));
      expect(apiKey.name, equals('Production'));
      expect(apiKey.keyPrefix, equals('lgk_prod1234...'));
      expect(apiKey.createdAt, equals(DateTime.utc(2024, 2, 20, 15, 45)));
      expect(apiKey.lastUsedAt, equals(DateTime.utc(2024, 2, 25, 10)));
    });
  });
}
