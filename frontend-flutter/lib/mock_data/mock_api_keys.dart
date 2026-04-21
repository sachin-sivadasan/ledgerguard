import '../models/api_key_model.dart';

final _now = DateTime.now();

final mockApiKeys = <ApiKey>[
  ApiKey(
    id: 'key-1',
    name: 'Production Key',
    keyPrefix: 'lg_live_a1b2...',
    createdAt: _now.subtract(const Duration(days: 90)),
    lastUsedAt: _now.subtract(const Duration(hours: 2)),
    status: ApiKeyStatus.active,
    permissions: ['read', 'write'],
  ),
  ApiKey(
    id: 'key-2',
    name: 'Staging Key',
    keyPrefix: 'lg_test_c3d4...',
    createdAt: _now.subtract(const Duration(days: 60)),
    lastUsedAt: _now.subtract(const Duration(days: 1)),
    status: ApiKeyStatus.active,
    permissions: ['read'],
  ),
  ApiKey(
    id: 'key-3',
    name: 'CI/CD Pipeline',
    keyPrefix: 'lg_live_e5f6...',
    createdAt: _now.subtract(const Duration(days: 30)),
    lastUsedAt: null,
    status: ApiKeyStatus.active,
    permissions: ['read', 'write', 'sync'],
  ),
  ApiKey(
    id: 'key-4',
    name: 'Old Integration',
    keyPrefix: 'lg_live_g7h8...',
    createdAt: _now.subtract(const Duration(days: 180)),
    lastUsedAt: _now.subtract(const Duration(days: 30)),
    status: ApiKeyStatus.revoked,
    permissions: ['read'],
  ),
];
