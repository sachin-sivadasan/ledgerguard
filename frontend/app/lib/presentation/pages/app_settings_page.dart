import 'package:flutter/material.dart';
import 'package:get_it/get_it.dart';
import 'package:go_router/go_router.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/revenue_share_tier.dart';
import '../../domain/entities/shopify_app.dart';
import '../../domain/repositories/app_repository.dart';
import '../widgets/shared.dart';
import '../widgets/tier_selector.dart';

/// Page for managing app-specific settings including revenue share tier
class AppSettingsPage extends StatefulWidget {
  const AppSettingsPage({super.key});

  @override
  State<AppSettingsPage> createState() => _AppSettingsPageState();
}

class _AppSettingsPageState extends State<AppSettingsPage> {
  final AppRepository _appRepository = GetIt.instance<AppRepository>();

  ShopifyApp? _selectedApp;
  bool _isLoading = true;
  bool _isSavingTier = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadSelectedApp();
  }

  Future<void> _loadSelectedApp() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final app = await _appRepository.getSelectedApp();
      setState(() {
        _selectedApp = app;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _isLoading = false;
      });
    }
  }

  Future<void> _onTierChanged(RevenueShareTier tier) async {
    if (_selectedApp == null) return;

    setState(() => _isSavingTier = true);

    try {
      await _appRepository.updateAppTier(
        _selectedApp!.id,
        tier,
      );

      setState(() {
        _selectedApp = _selectedApp!.copyWith(revenueShareTier: tier);
        _isSavingTier = false;
      });

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Row(
              children: [
                const Icon(Icons.check_circle, color: Colors.white),
                const SizedBox(width: 8),
                Text('Tier updated to ${tier.displayName}'),
              ],
            ),
            backgroundColor: AppTheme.success,
            behavior: SnackBarBehavior.floating,
          ),
        );
      }
    } catch (e) {
      setState(() => _isSavingTier = false);

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Row(
              children: [
                const Icon(Icons.error_outline, color: Colors.white),
                const SizedBox(width: 8),
                Text('Failed to update tier: $e'),
              ],
            ),
            backgroundColor: AppTheme.danger,
            behavior: SnackBarBehavior.floating,
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey[50],
      appBar: AppBar(
        title: const Text('App Settings'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.pop(),
        ),
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_error != null) {
      return ErrorStateWidget(
        title: 'Failed to load app',
        message: _error!,
        onRetry: _loadSelectedApp,
      );
    }

    if (_selectedApp == null) {
      return EmptyStateWidget(
        title: 'No App Selected',
        message: 'Please select an app to manage its settings.',
        icon: Icons.apps_outlined,
        actionLabel: 'Select App',
        onAction: () => context.go('/app-selection'),
      );
    }

    return RefreshIndicator(
      onRefresh: _loadSelectedApp,
      child: SingleChildScrollView(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // App Info Card
            _buildAppInfoCard(),
            const SizedBox(height: 24),

            // Revenue Share Tier Section
            _buildSectionHeader('Revenue Share Tier', Icons.percent),
            const SizedBox(height: 12),
            TierSelector(
              currentTier: _selectedApp!.revenueShareTier,
              onTierChanged: _onTierChanged,
              isLoading: _isSavingTier,
            ),
            const SizedBox(height: 24),

            // Fee Calculator Section
            _buildSectionHeader('Fee Calculator', Icons.calculate_outlined),
            const SizedBox(height: 12),
            _FeeCalculatorCard(tier: _selectedApp!.revenueShareTier),
            const SizedBox(height: 24),

            // Tier Comparison Section
            _buildSectionHeader('Tier Comparison', Icons.compare_arrows),
            const SizedBox(height: 12),
            _TierComparisonCard(currentTier: _selectedApp!.revenueShareTier),
          ],
        ),
      ),
    );
  }

  Widget _buildAppInfoCard() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            AppTheme.primary,
            AppTheme.primary.withBlue(180),
          ],
        ),
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: AppTheme.primary.withOpacity(0.3),
            blurRadius: 12,
            offset: const Offset(0, 6),
          ),
        ],
      ),
      child: Row(
        children: [
          Container(
            width: 56,
            height: 56,
            decoration: BoxDecoration(
              color: Colors.white.withOpacity(0.2),
              borderRadius: BorderRadius.circular(12),
            ),
            child: const Icon(
              Icons.apps,
              color: Colors.white,
              size: 28,
            ),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  _selectedApp!.name,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                if (_selectedApp!.description != null &&
                    _selectedApp!.description!.isNotEmpty) ...[
                  const SizedBox(height: 4),
                  Text(
                    _selectedApp!.description!,
                    style: TextStyle(
                      color: Colors.white.withOpacity(0.8),
                      fontSize: 13,
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ],
                if (_selectedApp!.installCount != null) ...[
                  const SizedBox(height: 8),
                  Container(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                    decoration: BoxDecoration(
                      color: Colors.white.withOpacity(0.2),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Text(
                      '${_selectedApp!.installCount} installs',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 12,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSectionHeader(String title, IconData icon) {
    return Row(
      children: [
        Icon(icon, size: 20, color: Colors.grey[600]),
        const SizedBox(width: 8),
        Text(
          title,
          style: TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.w600,
            color: Colors.grey[700],
            letterSpacing: 0.5,
          ),
        ),
      ],
    );
  }
}

/// Card showing fee calculation for a sample amount
class _FeeCalculatorCard extends StatefulWidget {
  final RevenueShareTier tier;

  const _FeeCalculatorCard({required this.tier});

  @override
  State<_FeeCalculatorCard> createState() => _FeeCalculatorCardState();
}

class _FeeCalculatorCardState extends State<_FeeCalculatorCard> {
  double _amount = 49.00;

  @override
  Widget build(BuildContext context) {
    final amountCents = (_amount * 100).round();
    final breakdown = FeeBreakdown.calculate(
      grossAmountCents: amountCents,
      tier: widget.tier,
    );

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
            // Amount slider
            Text(
              'Transaction Amount',
              style: TextStyle(
                color: Colors.grey[600],
                fontSize: 13,
              ),
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Expanded(
                  child: Slider(
                    value: _amount,
                    min: 10,
                    max: 500,
                    divisions: 49,
                    onChanged: (value) => setState(() => _amount = value),
                  ),
                ),
                Container(
                  width: 80,
                  alignment: Alignment.centerRight,
                  child: Text(
                    '\$${_amount.toStringAsFixed(0)}',
                    style: const TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ],
            ),
            const Divider(height: 24),

            // Fee breakdown
            _buildFeeRow(
              'Gross Amount',
              '\$${(breakdown.grossAmountCents / 100).toStringAsFixed(2)}',
              isHeader: true,
            ),
            const SizedBox(height: 8),
            _buildFeeRow(
              'Revenue Share (${breakdown.revenueSharePercent.toStringAsFixed(0)}%)',
              '-\$${(breakdown.revenueShareCents / 100).toStringAsFixed(2)}',
              color: breakdown.revenueShareCents > 0
                  ? Colors.orange
                  : Colors.green,
            ),
            _buildFeeRow(
              'Processing Fee (${breakdown.processingFeePercent.toStringAsFixed(1)}%)',
              '-\$${(breakdown.processingFeeCents / 100).toStringAsFixed(2)}',
              color: Colors.orange,
            ),
            _buildFeeRow(
              'Tax on Fees (8%)',
              '-\$${(breakdown.taxOnFeesCents / 100).toStringAsFixed(2)}',
              color: Colors.orange,
            ),
            const Divider(height: 16),
            _buildFeeRow(
              'Total Fees',
              '-\$${(breakdown.totalFeesCents / 100).toStringAsFixed(2)}',
              color: Colors.red,
              isBold: true,
            ),
            const SizedBox(height: 8),
            _buildFeeRow(
              'Net Amount',
              '\$${(breakdown.netAmountCents / 100).toStringAsFixed(2)}',
              color: AppTheme.success,
              isHeader: true,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFeeRow(
    String label,
    String amount, {
    Color? color,
    bool isHeader = false,
    bool isBold = false,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            label,
            style: TextStyle(
              fontSize: isHeader ? 14 : 13,
              fontWeight: isHeader || isBold ? FontWeight.w600 : FontWeight.normal,
              color: isHeader ? Colors.grey[800] : Colors.grey[600],
            ),
          ),
          Text(
            amount,
            style: TextStyle(
              fontSize: isHeader ? 16 : 14,
              fontWeight: isHeader || isBold ? FontWeight.bold : FontWeight.w500,
              color: color ?? Colors.grey[800],
            ),
          ),
        ],
      ),
    );
  }
}

/// Card showing comparison between different tiers
class _TierComparisonCard extends StatelessWidget {
  final RevenueShareTier currentTier;

  const _TierComparisonCard({required this.currentTier});

  @override
  Widget build(BuildContext context) {
    const sampleAmount = 10000; // $100

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
            Text(
              'On a \$100 transaction:',
              style: TextStyle(
                color: Colors.grey[600],
                fontSize: 13,
              ),
            ),
            const SizedBox(height: 16),
            ...RevenueShareTier.values.map((tier) {
              final breakdown = FeeBreakdown.calculate(
                grossAmountCents: sampleAmount,
                tier: tier,
              );
              final isCurrentTier = tier == currentTier;

              return Container(
                margin: const EdgeInsets.only(bottom: 8),
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: isCurrentTier
                      ? Color(tier.badgeColor).withOpacity(0.1)
                      : Colors.grey[50],
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(
                    color: isCurrentTier
                        ? Color(tier.badgeColor)
                        : Colors.transparent,
                    width: isCurrentTier ? 2 : 0,
                  ),
                ),
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
                        children: [
                          Row(
                            children: [
                              Text(
                                tier.displayName,
                                style: TextStyle(
                                  fontWeight: isCurrentTier
                                      ? FontWeight.bold
                                      : FontWeight.w500,
                                  fontSize: 13,
                                ),
                              ),
                              if (isCurrentTier) ...[
                                const SizedBox(width: 8),
                                Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 6,
                                    vertical: 2,
                                  ),
                                  decoration: BoxDecoration(
                                    color: Color(tier.badgeColor),
                                    borderRadius: BorderRadius.circular(4),
                                  ),
                                  child: const Text(
                                    'Current',
                                    style: TextStyle(
                                      color: Colors.white,
                                      fontSize: 10,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ),
                              ],
                            ],
                          ),
                          const SizedBox(height: 2),
                          Text(
                            'Fees: \$${(breakdown.totalFeesCents / 100).toStringAsFixed(2)} · Net: \$${(breakdown.netAmountCents / 100).toStringAsFixed(2)}',
                            style: TextStyle(
                              color: Colors.grey[600],
                              fontSize: 12,
                            ),
                          ),
                        ],
                      ),
                    ),
                    Text(
                      '${breakdown.revenueSharePercent.toStringAsFixed(0)}% + 2.9%',
                      style: TextStyle(
                        color: Color(tier.badgeColor),
                        fontWeight: FontWeight.bold,
                        fontSize: 12,
                      ),
                    ),
                  ],
                ),
              );
            }),
          ],
        ),
      ),
    );
  }
}
