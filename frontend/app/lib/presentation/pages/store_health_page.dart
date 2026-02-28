import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/store_health.dart';
import '../../domain/entities/subscription.dart';
import '../blocs/store_health/store_health.dart';
import '../widgets/error_state_widget.dart';
import '../widgets/risk_badge.dart';
import '../widgets/transaction_timeline.dart';

/// Page displaying store health details
class StoreHealthPage extends StatelessWidget {
  final String appId;
  final String domain;

  const StoreHealthPage({
    super.key,
    required this.appId,
    required this.domain,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey[50],
      appBar: AppBar(
        title: const Text('Store Health'),
        actions: [
          BlocBuilder<StoreHealthBloc, StoreHealthState>(
            builder: (context, state) {
              final isRefreshing =
                  state is StoreHealthLoaded && state.isRefreshing;
              return IconButton(
                icon: isRefreshing
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: Colors.white,
                        ),
                      )
                    : const Icon(Icons.refresh),
                onPressed: isRefreshing
                    ? null
                    : () => context
                        .read<StoreHealthBloc>()
                        .add(const RefreshStoreHealthRequested()),
              );
            },
          ),
        ],
      ),
      body: BlocBuilder<StoreHealthBloc, StoreHealthState>(
        builder: (context, state) {
          if (state is StoreHealthInitial) {
            context.read<StoreHealthBloc>().add(
                  LoadStoreHealthRequested(appId: appId, domain: domain),
                );
            return const Center(child: CircularProgressIndicator());
          }

          if (state is StoreHealthLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state is StoreHealthNotFound) {
            return ErrorStateWidget(
              title: 'Store Not Found',
              message: 'No subscription found for ${state.domain}',
              onRetry: () => context.read<StoreHealthBloc>().add(
                    LoadStoreHealthRequested(appId: appId, domain: domain),
                  ),
            );
          }

          if (state is StoreHealthError) {
            return ErrorStateWidget(
              title: 'Failed to load store health',
              message: state.message,
              onRetry: () => context.read<StoreHealthBloc>().add(
                    LoadStoreHealthRequested(appId: appId, domain: domain),
                  ),
            );
          }

          if (state is StoreHealthLoaded) {
            return _buildContent(context, state.storeHealth);
          }

          return const SizedBox.shrink();
        },
      ),
    );
  }

  Widget _buildContent(BuildContext context, StoreHealth health) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildStoreCard(context, health.subscription),
          const SizedBox(height: 16),
          RiskStateIndicator(riskState: health.subscription.riskState),
          const SizedBox(height: 16),
          _buildEarningsCard(context, health.earnings),
          const SizedBox(height: 16),
          _buildStatsCard(context, health),
          const SizedBox(height: 16),
          TransactionTimeline(transactions: health.transactions),
        ],
      ),
    );
  }

  Widget _buildStoreCard(BuildContext context, Subscription subscription) {
    final String storeName = (subscription.shopName?.isNotEmpty == true)
        ? subscription.shopName!
        : subscription.myshopifyDomain
            .replaceAll('.myshopify.com', '')
            .split(RegExp(r'[-_]'))
            .map((word) => word.isNotEmpty
                ? '${word[0].toUpperCase()}${word.substring(1)}'
                : '')
            .join(' ');

    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Row(
        children: [
          Container(
            width: 64,
            height: 64,
            decoration: BoxDecoration(
              color: Colors.blue.withOpacity(0.1),
              borderRadius: BorderRadius.circular(14),
            ),
            child: Center(
              child: Text(
                _getInitials(subscription.myshopifyDomain),
                style: const TextStyle(
                  color: Colors.blue,
                  fontWeight: FontWeight.bold,
                  fontSize: 22,
                ),
              ),
            ),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  storeName,
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                ),
                const SizedBox(height: 4),
                Text(
                  subscription.myshopifyDomain,
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: Colors.grey[600],
                      ),
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    _buildStatusBadge(subscription.status),
                    const SizedBox(width: 8),
                    Text(
                      subscription.formattedPrice,
                      style: TextStyle(
                        color: Colors.grey[700],
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                    Text(
                      '/${subscription.billingInterval == BillingInterval.annual ? 'yr' : 'mo'}',
                      style: TextStyle(
                        color: Colors.grey[500],
                        fontSize: 12,
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatusBadge(String status) {
    final isActive = status == 'ACTIVE';
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: isActive
            ? Colors.green.withOpacity(0.1)
            : Colors.grey.withOpacity(0.1),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(
        status,
        style: TextStyle(
          color: isActive ? Colors.green : Colors.grey[600],
          fontSize: 12,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }

  Widget _buildEarningsCard(BuildContext context, StoreEarnings earnings) {
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
                    color: AppTheme.success.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: const Icon(
                    Icons.account_balance_wallet,
                    color: AppTheme.success,
                    size: 20,
                  ),
                ),
                const SizedBox(width: 12),
                const Text(
                  'Store Earnings',
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 16,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                Expanded(
                  child: _buildEarningsTile(
                    'Pending',
                    earnings.formattedPending,
                    Colors.amber,
                    Icons.schedule,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: _buildEarningsTile(
                    'Available',
                    earnings.formattedAvailable,
                    AppTheme.success,
                    Icons.check_circle,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.grey.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Row(
                    children: [
                      Icon(Icons.paid, color: Colors.grey[600], size: 18),
                      const SizedBox(width: 8),
                      Text(
                        'Paid Out',
                        style: TextStyle(color: Colors.grey[600]),
                      ),
                    ],
                  ),
                  Text(
                    earnings.formattedPaidOut,
                    style: TextStyle(
                      fontWeight: FontWeight.w600,
                      color: Colors.grey[700],
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

  Widget _buildEarningsTile(
    String label,
    String amount,
    Color color,
    IconData icon,
  ) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: color.withOpacity(0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(icon, color: color, size: 16),
              const SizedBox(width: 6),
              Text(
                label,
                style: TextStyle(
                  color: color,
                  fontSize: 12,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            amount,
            style: TextStyle(
              fontSize: 20,
              fontWeight: FontWeight.bold,
              color: color,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatsCard(BuildContext context, StoreHealth health) {
    final dateFormat = DateFormat('MMM d, y');
    final sub = health.subscription;

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
            const Text(
              'Quick Stats',
              style: TextStyle(
                fontWeight: FontWeight.bold,
                fontSize: 16,
              ),
            ),
            const SizedBox(height: 16),
            _buildStatRow(
              'Total Transactions',
              '${health.transactions.length}',
              Icons.receipt,
            ),
            const Divider(height: 20),
            _buildStatRow(
              'Total Revenue',
              health.formattedTotalRevenue,
              Icons.attach_money,
            ),
            const Divider(height: 20),
            _buildStatRow(
              'MRR Contribution',
              '\$${(sub.mrrCents / 100).toStringAsFixed(2)}/mo',
              Icons.trending_up,
            ),
            if (sub.lastChargeDate != null) ...[
              const Divider(height: 20),
              _buildStatRow(
                'Last Charge',
                dateFormat.format(sub.lastChargeDate!),
                Icons.payment,
              ),
            ],
            if (sub.expectedNextCharge != null) ...[
              const Divider(height: 20),
              _buildStatRow(
                'Next Charge',
                dateFormat.format(sub.expectedNextCharge!),
                Icons.event,
                valueColor: _isOverdue(sub.expectedNextCharge!)
                    ? Colors.red
                    : Colors.green,
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildStatRow(
    String label,
    String value,
    IconData icon, {
    Color? valueColor,
  }) {
    return Row(
      children: [
        Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: Colors.grey[100],
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(icon, size: 18, color: Colors.grey[600]),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Text(
            label,
            style: TextStyle(color: Colors.grey[600]),
          ),
        ),
        Text(
          value,
          style: TextStyle(
            fontWeight: FontWeight.w600,
            color: valueColor,
          ),
        ),
      ],
    );
  }

  String _getInitials(String domain) {
    final storeName = domain.replaceAll('.myshopify.com', '');
    final parts = storeName.split(RegExp(r'[-_]'));
    if (parts.length >= 2) {
      return '${parts[0][0]}${parts[1][0]}'.toUpperCase();
    }
    return storeName.substring(0, 2).toUpperCase();
  }

  bool _isOverdue(DateTime date) {
    return date.isBefore(DateTime.now());
  }
}
