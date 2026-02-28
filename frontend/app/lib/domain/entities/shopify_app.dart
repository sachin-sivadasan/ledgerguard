import 'package:equatable/equatable.dart';
import 'revenue_share_tier.dart';

/// Represents a Shopify app from the Partner account
class ShopifyApp extends Equatable {
  final String id;
  final String name;
  final String? iconUrl;
  final String? description;
  final int? installCount;
  final RevenueShareTier revenueShareTier;
  final DateTime? createdAt;
  final DateTime? updatedAt;

  const ShopifyApp({
    required this.id,
    required this.name,
    this.iconUrl,
    this.description,
    this.installCount,
    this.revenueShareTier = RevenueShareTier.default20,
    this.createdAt,
    this.updatedAt,
  });

  /// Create a copy with updated fields
  ShopifyApp copyWith({
    String? id,
    String? name,
    String? iconUrl,
    String? description,
    int? installCount,
    RevenueShareTier? revenueShareTier,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) {
    return ShopifyApp(
      id: id ?? this.id,
      name: name ?? this.name,
      iconUrl: iconUrl ?? this.iconUrl,
      description: description ?? this.description,
      installCount: installCount ?? this.installCount,
      revenueShareTier: revenueShareTier ?? this.revenueShareTier,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }

  @override
  List<Object?> get props => [
        id,
        name,
        iconUrl,
        description,
        installCount,
        revenueShareTier,
        createdAt,
        updatedAt,
      ];
}
