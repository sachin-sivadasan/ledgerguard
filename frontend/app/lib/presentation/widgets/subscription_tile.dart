import 'package:flutter/material.dart';

import '../../domain/entities/subscription.dart';
import 'risk_badge.dart';

/// List tile widget for displaying a subscription
class SubscriptionTile extends StatelessWidget {
  final Subscription subscription;
  final VoidCallback? onTap;

  const SubscriptionTile({
    super.key,
    required this.subscription,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final isCompact = constraints.maxWidth < 360;
        final avatarSize = isCompact ? 40.0 : 48.0;
        final padding = isCompact ? 12.0 : 16.0;

        return Material(
          color: Colors.transparent,
          child: InkWell(
            onTap: onTap,
            borderRadius: BorderRadius.circular(12),
            child: Container(
              padding: EdgeInsets.all(padding),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: Colors.grey[200]!),
              ),
              child: Row(
                children: [
                  // Store avatar
                  Container(
                    width: avatarSize,
                    height: avatarSize,
                    decoration: BoxDecoration(
                      color: Colors.blue.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(isCompact ? 8 : 10),
                    ),
                    child: Center(
                      child: Text(
                        _getInitials(subscription.displayName),
                        style: TextStyle(
                          color: Colors.blue,
                          fontWeight: FontWeight.bold,
                          fontSize: isCompact ? 13 : 16,
                        ),
                      ),
                    ),
                  ),
                  SizedBox(width: isCompact ? 10 : 16),
                  // Store and plan info
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          _formatDisplayName(subscription),
                          style: Theme.of(context).textTheme.titleSmall?.copyWith(
                                fontWeight: FontWeight.w600,
                                fontSize: isCompact ? 13 : null,
                              ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                        const SizedBox(height: 3),
                        Text(
                          isCompact
                              ? subscription.formattedPrice
                              : '${subscription.planName} · ${subscription.formattedPrice}',
                          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                color: Colors.grey[600],
                                fontSize: isCompact ? 11 : null,
                              ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ],
                    ),
                  ),
                  SizedBox(width: isCompact ? 8 : 12),
                  // Risk badge
                  RiskBadge(
                    riskState: subscription.riskState,
                    isCompact: true,
                  ),
                  SizedBox(width: isCompact ? 4 : 8),
                  // Chevron
                  Icon(
                    Icons.chevron_right,
                    color: Colors.grey[400],
                    size: isCompact ? 18 : 20,
                  ),
                ],
              ),
            ),
          ),
        );
      },
    );
  }

  String _getInitials(String name) {
    if (name.isEmpty) return '??';

    // Handle both shop names and domains
    final cleanName = name.replaceAll('.myshopify.com', '').trim();
    if (cleanName.isEmpty) return '??';

    // Filter out empty parts
    final parts = cleanName
        .split(RegExp(r'[-_\s]+'))
        .where((p) => p.isNotEmpty)
        .toList();

    if (parts.isEmpty) return '??';

    if (parts.length >= 2) {
      final first = parts[0];
      final second = parts[1];
      if (first.isNotEmpty && second.isNotEmpty) {
        return '${first[0]}${second[0]}'.toUpperCase();
      }
    }

    // Use first part or cleanName
    final firstPart = parts.isNotEmpty ? parts[0] : cleanName;
    if (firstPart.length >= 2) {
      return firstPart.substring(0, 2).toUpperCase();
    }
    if (firstPart.isNotEmpty) {
      return firstPart[0].toUpperCase();
    }
    return '??';
  }

  String _formatDisplayName(Subscription sub) {
    // Use shop name if available, otherwise format domain
    if (sub.shopName?.isNotEmpty == true) {
      return sub.shopName!;
    }

    // Remove .myshopify.com and capitalize
    final cleanDomain = sub.myshopifyDomain.replaceAll('.myshopify.com', '');
    if (cleanDomain.isEmpty) return sub.myshopifyDomain;

    final parts = cleanDomain
        .split(RegExp(r'[-_]+'))
        .where((word) => word.isNotEmpty)
        .toList();

    if (parts.isEmpty) return cleanDomain;

    return parts
        .map((word) => word.isNotEmpty
            ? '${word[0].toUpperCase()}${word.substring(1)}'
            : '')
        .where((w) => w.isNotEmpty)
        .join(' ');
  }
}

/// Skeleton loader for subscription tile
class SubscriptionTileSkeleton extends StatelessWidget {
  const SubscriptionTileSkeleton({super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey[200]!),
      ),
      child: Row(
        children: [
          Container(
            width: 48,
            height: 48,
            decoration: BoxDecoration(
              color: Colors.grey[200],
              borderRadius: BorderRadius.circular(10),
            ),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  height: 16,
                  width: 140,
                  decoration: BoxDecoration(
                    color: Colors.grey[200],
                    borderRadius: BorderRadius.circular(4),
                  ),
                ),
                const SizedBox(height: 8),
                Container(
                  height: 12,
                  width: 100,
                  decoration: BoxDecoration(
                    color: Colors.grey[200],
                    borderRadius: BorderRadius.circular(4),
                  ),
                ),
              ],
            ),
          ),
          Container(
            height: 24,
            width: 60,
            decoration: BoxDecoration(
              color: Colors.grey[200],
              borderRadius: BorderRadius.circular(12),
            ),
          ),
        ],
      ),
    );
  }
}
