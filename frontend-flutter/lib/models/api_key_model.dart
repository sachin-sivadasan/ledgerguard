enum ApiKeyStatus { active, revoked }

class ApiKey {
  final String id;
  final String name;
  final String keyPrefix;
  final DateTime createdAt;
  final DateTime? lastUsedAt;
  final ApiKeyStatus status;
  final List<String> permissions;

  const ApiKey({
    required this.id,
    required this.name,
    required this.keyPrefix,
    required this.createdAt,
    this.lastUsedAt,
    required this.status,
    required this.permissions,
  });

  ApiKey copyWith({ApiKeyStatus? status, DateTime? lastUsedAt}) {
    return ApiKey(
      id: id,
      name: name,
      keyPrefix: keyPrefix,
      createdAt: createdAt,
      lastUsedAt: lastUsedAt ?? this.lastUsedAt,
      status: status ?? this.status,
      permissions: permissions,
    );
  }
}
