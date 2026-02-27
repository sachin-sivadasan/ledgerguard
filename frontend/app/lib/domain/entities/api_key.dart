import 'package:equatable/equatable.dart';

/// API Key entity for Revenue API access
class ApiKey extends Equatable {
  final String id;
  final String name;
  final String keyPrefix; // e.g., "lg_live_abc...xyz"
  final DateTime createdAt;
  final DateTime? lastUsedAt;

  const ApiKey({
    required this.id,
    required this.name,
    required this.keyPrefix,
    required this.createdAt,
    this.lastUsedAt,
  });

  /// Create from JSON response
  factory ApiKey.fromJson(Map<String, dynamic> json) {
    return ApiKey(
      id: json['id'] as String,
      name: json['name'] as String,
      keyPrefix: json['key_prefix'] as String,
      createdAt: DateTime.parse(json['created_at'] as String),
      lastUsedAt: json['last_used_at'] != null
          ? DateTime.parse(json['last_used_at'] as String)
          : null,
    );
  }

  /// Format last used time for display
  String get formattedLastUsed {
    if (lastUsedAt == null) return 'Never used';
    final now = DateTime.now();
    final diff = now.difference(lastUsedAt!);

    if (diff.inMinutes < 1) return 'Just now';
    if (diff.inMinutes < 60) return '${diff.inMinutes}m ago';
    if (diff.inHours < 24) return '${diff.inHours}h ago';
    if (diff.inDays < 7) return '${diff.inDays}d ago';

    return '${lastUsedAt!.day}/${lastUsedAt!.month}/${lastUsedAt!.year}';
  }

  /// Format created date for display
  String get formattedCreatedAt {
    return '${createdAt.day}/${createdAt.month}/${createdAt.year}';
  }

  @override
  List<Object?> get props => [id, name, keyPrefix, createdAt, lastUsedAt];
}

/// Result of creating a new API key (contains full key shown once)
class ApiKeyCreationResult {
  final ApiKey apiKey;
  final String fullKey; // Full key shown only once

  const ApiKeyCreationResult({
    required this.apiKey,
    required this.fullKey,
  });

  factory ApiKeyCreationResult.fromJson(Map<String, dynamic> json) {
    return ApiKeyCreationResult(
      apiKey: ApiKey.fromJson(json),
      fullKey: json['key'] as String,
    );
  }
}
