import 'package:flutter/material.dart';
import '../../domain/entities/revenue_share_tier.dart';

/// Widget to display and select revenue share tier
class TierSelector extends StatelessWidget {
  final RevenueShareTier currentTier;
  final ValueChanged<RevenueShareTier>? onTierChanged;
  final bool isLoading;
  final bool readOnly;

  const TierSelector({
    super.key,
    required this.currentTier,
    this.onTierChanged,
    this.isLoading = false,
    this.readOnly = false,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(
          color: Color(currentTier.badgeColor).withOpacity(0.3),
        ),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  Icons.percent,
                  size: 20,
                  color: Color(currentTier.badgeColor),
                ),
                const SizedBox(width: 8),
                const Text(
                  'Revenue Share Tier',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const Spacer(),
                _TierBadge(tier: currentTier),
              ],
            ),
            const SizedBox(height: 12),
            Text(
              currentTier.description,
              style: TextStyle(
                color: Colors.grey[600],
                fontSize: 13,
              ),
            ),
            const SizedBox(height: 16),
            if (!readOnly) ...[
              const Divider(),
              const SizedBox(height: 12),
              const Text(
                'Change Tier',
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                ),
              ),
              const SizedBox(height: 8),
              _TierDropdown(
                currentTier: currentTier,
                onChanged: isLoading ? null : onTierChanged,
                isLoading: isLoading,
              ),
            ],
          ],
        ),
      ),
    );
  }
}

/// Compact tier badge
class _TierBadge extends StatelessWidget {
  final RevenueShareTier tier;

  const _TierBadge({required this.tier});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: Color(tier.badgeColor).withOpacity(0.15),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: Color(tier.badgeColor).withOpacity(0.5),
        ),
      ),
      child: Text(
        '${tier.revenueSharePercent.toStringAsFixed(0)}% + 2.9%',
        style: TextStyle(
          color: Color(tier.badgeColor),
          fontSize: 12,
          fontWeight: FontWeight.bold,
        ),
      ),
    );
  }
}

/// Dropdown for selecting tier
class _TierDropdown extends StatelessWidget {
  final RevenueShareTier currentTier;
  final ValueChanged<RevenueShareTier>? onChanged;
  final bool isLoading;

  const _TierDropdown({
    required this.currentTier,
    this.onChanged,
    this.isLoading = false,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        border: Border.all(color: Colors.grey.shade300),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Stack(
        children: [
          DropdownButtonHideUnderline(
            child: DropdownButton<RevenueShareTier>(
              value: currentTier,
              isExpanded: true,
              padding: const EdgeInsets.symmetric(horizontal: 12),
              borderRadius: BorderRadius.circular(8),
              items: RevenueShareTier.values.map((tier) {
                return DropdownMenuItem(
                  value: tier,
                  child: _TierOption(
                    tier: tier,
                    isSelected: tier == currentTier,
                  ),
                );
              }).toList(),
              onChanged: onChanged == null
                  ? null
                  : (tier) {
                      if (tier != null && tier != currentTier) {
                        onChanged!(tier);
                      }
                    },
            ),
          ),
          if (isLoading)
            Positioned.fill(
              child: Container(
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.7),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Center(
                  child: SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  ),
                ),
              ),
            ),
        ],
      ),
    );
  }
}

/// Tier option in dropdown
class _TierOption extends StatelessWidget {
  final RevenueShareTier tier;
  final bool isSelected;

  const _TierOption({
    required this.tier,
    required this.isSelected,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        children: [
          Container(
            width: 8,
            height: 8,
            decoration: BoxDecoration(
              color: Color(tier.badgeColor),
              shape: BoxShape.circle,
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  tier.displayName,
                  style: TextStyle(
                    fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
                  ),
                ),
                if (!isSelected)
                  Text(
                    tier.description,
                    style: TextStyle(
                      fontSize: 11,
                      color: Colors.grey[500],
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
              ],
            ),
          ),
          if (isSelected)
            Icon(
              Icons.check_circle,
              color: Color(tier.badgeColor),
              size: 18,
            ),
        ],
      ),
    );
  }
}

/// Compact tier indicator for lists
class TierIndicator extends StatelessWidget {
  final RevenueShareTier tier;
  final bool compact;

  const TierIndicator({
    super.key,
    required this.tier,
    this.compact = false,
  });

  @override
  Widget build(BuildContext context) {
    if (compact) {
      return Container(
        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
        decoration: BoxDecoration(
          color: Color(tier.badgeColor).withOpacity(0.15),
          borderRadius: BorderRadius.circular(4),
        ),
        child: Text(
          '${tier.revenueSharePercent.toStringAsFixed(0)}%',
          style: TextStyle(
            color: Color(tier.badgeColor),
            fontSize: 10,
            fontWeight: FontWeight.bold,
          ),
        ),
      );
    }

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 6,
          height: 6,
          decoration: BoxDecoration(
            color: Color(tier.badgeColor),
            shape: BoxShape.circle,
          ),
        ),
        const SizedBox(width: 6),
        Text(
          tier.displayName,
          style: TextStyle(
            fontSize: 12,
            color: Colors.grey[600],
          ),
        ),
      ],
    );
  }
}
