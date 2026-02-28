import 'package:flutter/material.dart';
import 'package:get_it/get_it.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/revenue_share_tier.dart';
import '../../domain/entities/shopify_app.dart';
import '../../domain/repositories/app_repository.dart';
import 'tier_selector.dart';

/// Card displaying fee insights and savings based on revenue share tier
class FeeInsightsCard extends StatefulWidget {
  /// Total gross revenue in cents for fee calculation
  final int totalGrossCents;

  /// Whether to show in compact mode
  final bool compact;

  const FeeInsightsCard({
    super.key,
    required this.totalGrossCents,
    this.compact = false,
  });

  @override
  State<FeeInsightsCard> createState() => _FeeInsightsCardState();
}

class _FeeInsightsCardState extends State<FeeInsightsCard> {
  final AppRepository _appRepository = GetIt.instance<AppRepository>();

  ShopifyApp? _selectedApp;
  FeeSummary? _feeSummary;
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final app = await _appRepository.getSelectedApp();
      FeeSummary? summary;

      if (app != null) {
        try {
          summary = await _appRepository.getFeeSummary(app.id);
        } catch (_) {
          // Fee summary might not be available, continue with local calculation
        }
      }

      setState(() {
        _selectedApp = app;
        _feeSummary = summary;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return _buildLoadingCard();
    }

    if (_error != null || _selectedApp == null) {
      return const SizedBox.shrink();
    }

    final tier = _selectedApp!.revenueShareTier;
    final grossCents = _feeSummary?.totalGrossCents ?? widget.totalGrossCents;

    // Calculate fee breakdown using local calculation
    final breakdown = FeeBreakdown.calculate(
      grossAmountCents: grossCents,
      tier: tier,
    );

    // Calculate what fees would be with default 20% tier
    final defaultBreakdown = FeeBreakdown.calculate(
      grossAmountCents: grossCents,
      tier: RevenueShareTier.default20,
    );

    final savingsCents = defaultBreakdown.totalFeesCents - breakdown.totalFeesCents;
    final savingsPercent = defaultBreakdown.totalFeesCents > 0
        ? (savingsCents / defaultBreakdown.totalFeesCents) * 100
        : 0.0;

    if (widget.compact) {
      return _buildCompactCard(tier, breakdown, savingsCents);
    }

    return _buildFullCard(tier, breakdown, defaultBreakdown, savingsCents, savingsPercent);
  }

  Widget _buildLoadingCard() {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: Colors.grey.shade200),
      ),
      child: const SizedBox(
        height: 120,
        child: Center(child: CircularProgressIndicator()),
      ),
    );
  }

  Widget _buildCompactCard(
    RevenueShareTier tier,
    FeeBreakdown breakdown,
    int savingsCents,
  ) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(
          color: Color(tier.badgeColor).withOpacity(0.3),
        ),
      ),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: Color(tier.badgeColor).withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Icon(
                Icons.percent,
                color: Color(tier.badgeColor),
                size: 20,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Text(
                        'Fees: ',
                        style: TextStyle(
                          color: Colors.grey[600],
                          fontSize: 12,
                        ),
                      ),
                      Text(
                        '\$${(breakdown.totalFeesCents / 100).toStringAsFixed(2)}',
                        style: const TextStyle(
                          fontWeight: FontWeight.bold,
                          fontSize: 14,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 2),
                  Text(
                    tier.displayName,
                    style: TextStyle(
                      color: Colors.grey[500],
                      fontSize: 11,
                    ),
                  ),
                ],
              ),
            ),
            if (savingsCents > 0)
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: AppTheme.success.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(
                      Icons.savings_outlined,
                      color: AppTheme.success,
                      size: 14,
                    ),
                    const SizedBox(width: 4),
                    Text(
                      '\$${(savingsCents / 100).toStringAsFixed(0)}',
                      style: const TextStyle(
                        color: AppTheme.success,
                        fontWeight: FontWeight.bold,
                        fontSize: 12,
                      ),
                    ),
                  ],
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildFullCard(
    RevenueShareTier tier,
    FeeBreakdown breakdown,
    FeeBreakdown defaultBreakdown,
    int savingsCents,
    double savingsPercent,
  ) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: Colors.grey.shade200),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: Color(tier.badgeColor).withOpacity(0.1),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: Icon(
                    Icons.receipt_long,
                    color: Color(tier.badgeColor),
                    size: 20,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Fee Insights',
                        style: TextStyle(
                          fontWeight: FontWeight.bold,
                          fontSize: 16,
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        'Based on ${tier.displayName}',
                        style: TextStyle(
                          color: Colors.grey[600],
                          fontSize: 12,
                        ),
                      ),
                    ],
                  ),
                ),
                TierIndicator(tier: tier, compact: true),
              ],
            ),

            const SizedBox(height: 16),
            const Divider(height: 1),
            const SizedBox(height: 16),

            // Fee Breakdown
            _buildFeeRow(
              'Gross Revenue',
              '\$${(breakdown.grossAmountCents / 100).toStringAsFixed(2)}',
              icon: Icons.attach_money,
            ),
            const SizedBox(height: 8),
            _buildFeeRow(
              'Revenue Share (${breakdown.revenueSharePercent.toStringAsFixed(0)}%)',
              '-\$${(breakdown.revenueShareCents / 100).toStringAsFixed(2)}',
              color: breakdown.revenueShareCents > 0 ? Colors.orange : AppTheme.success,
            ),
            _buildFeeRow(
              'Processing Fee (2.9%)',
              '-\$${(breakdown.processingFeeCents / 100).toStringAsFixed(2)}',
              color: Colors.orange,
            ),
            _buildFeeRow(
              'Tax on Fees',
              '-\$${(breakdown.taxOnFeesCents / 100).toStringAsFixed(2)}',
              color: Colors.orange,
            ),
            const Divider(height: 16),
            _buildFeeRow(
              'Total Fees',
              '-\$${(breakdown.totalFeesCents / 100).toStringAsFixed(2)}',
              color: AppTheme.danger,
              isBold: true,
            ),
            const SizedBox(height: 4),
            _buildFeeRow(
              'Net Revenue',
              '\$${(breakdown.netAmountCents / 100).toStringAsFixed(2)}',
              color: AppTheme.success,
              isBold: true,
            ),

            // Savings Section
            if (savingsCents > 0) ...[
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: AppTheme.success.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(
                    color: AppTheme.success.withOpacity(0.3),
                  ),
                ),
                child: Row(
                  children: [
                    const Icon(
                      Icons.savings,
                      color: AppTheme.success,
                      size: 24,
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Text(
                            'Tier Savings',
                            style: TextStyle(
                              fontWeight: FontWeight.w600,
                              fontSize: 13,
                            ),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            'vs Default 20% tier',
                            style: TextStyle(
                              color: Colors.grey[600],
                              fontSize: 11,
                            ),
                          ),
                        ],
                      ),
                    ),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Text(
                          '\$${(savingsCents / 100).toStringAsFixed(2)}',
                          style: const TextStyle(
                            color: AppTheme.success,
                            fontWeight: FontWeight.bold,
                            fontSize: 18,
                          ),
                        ),
                        Text(
                          '${savingsPercent.toStringAsFixed(1)}% less',
                          style: const TextStyle(
                            color: AppTheme.success,
                            fontSize: 11,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ],

            // Info about tier
            if (tier == RevenueShareTier.default20) ...[
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.amber.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(
                    color: Colors.amber.withOpacity(0.3),
                  ),
                ),
                child: Row(
                  children: [
                    const Icon(
                      Icons.lightbulb_outline,
                      color: Colors.amber,
                      size: 20,
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        'Consider applying for Shopify\'s Reduced Revenue Share Plan to save on fees.',
                        style: TextStyle(
                          color: Colors.grey[700],
                          fontSize: 12,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildFeeRow(
    String label,
    String amount, {
    Color? color,
    IconData? icon,
    bool isBold = false,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        children: [
          if (icon != null) ...[
            Icon(icon, size: 16, color: Colors.grey[600]),
            const SizedBox(width: 8),
          ],
          Expanded(
            child: Text(
              label,
              style: TextStyle(
                fontSize: 13,
                fontWeight: isBold ? FontWeight.w600 : FontWeight.normal,
                color: Colors.grey[700],
              ),
            ),
          ),
          Text(
            amount,
            style: TextStyle(
              fontSize: isBold ? 15 : 13,
              fontWeight: isBold ? FontWeight.bold : FontWeight.w500,
              color: color ?? Colors.grey[800],
            ),
          ),
        ],
      ),
    );
  }
}

/// Compact fee summary for KPI section
class FeeKpiCard extends StatelessWidget {
  final int totalGrossCents;
  final RevenueShareTier tier;

  const FeeKpiCard({
    super.key,
    required this.totalGrossCents,
    required this.tier,
  });

  @override
  Widget build(BuildContext context) {
    final breakdown = FeeBreakdown.calculate(
      grossAmountCents: totalGrossCents,
      tier: tier,
    );

    final effectiveFeePercent = totalGrossCents > 0
        ? (breakdown.totalFeesCents / totalGrossCents) * 100
        : 0.0;

    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: Colors.grey.shade200),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: Colors.purple.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: const Icon(
                    Icons.receipt_long,
                    color: Colors.purple,
                    size: 20,
                  ),
                ),
                const Spacer(),
                TierIndicator(tier: tier, compact: true),
              ],
            ),
            const SizedBox(height: 12),
            Text(
              '\$${(breakdown.totalFeesCents / 100).toStringAsFixed(2)}',
              style: const TextStyle(
                fontSize: 24,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 4),
            Text(
              'Total Fees (${effectiveFeePercent.toStringAsFixed(1)}%)',
              style: TextStyle(
                color: Colors.grey[600],
                fontSize: 12,
              ),
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Text(
                  'Net: ',
                  style: TextStyle(
                    color: Colors.grey[500],
                    fontSize: 12,
                  ),
                ),
                Text(
                  '\$${(breakdown.netAmountCents / 100).toStringAsFixed(2)}',
                  style: const TextStyle(
                    color: AppTheme.success,
                    fontWeight: FontWeight.w600,
                    fontSize: 12,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
