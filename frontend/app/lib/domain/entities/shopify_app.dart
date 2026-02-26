import 'package:equatable/equatable.dart';

/// Represents a Shopify app from the Partner account
class ShopifyApp extends Equatable {
  final String id;
  final String name;
  final String? iconUrl;
  final String? description;
  final int? installCount;

  const ShopifyApp({
    required this.id,
    required this.name,
    this.iconUrl,
    this.description,
    this.installCount,
  });

  @override
  List<Object?> get props => [id, name, iconUrl, description, installCount];
}
