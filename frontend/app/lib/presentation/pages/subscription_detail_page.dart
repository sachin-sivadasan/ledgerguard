import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../../domain/entities/subscription.dart';
import '../blocs/subscription_detail/subscription_detail.dart';
import '../widgets/error_state_widget.dart';
import '../widgets/risk_badge.dart';

/// Page displaying subscription details
class SubscriptionDetailPage extends StatelessWidget {
  final String appId;
  final String subscriptionId;

  const SubscriptionDetailPage({
    super.key,
    required this.appId,
    required this.subscriptionId,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey[50],
      appBar: AppBar(
        title: const Text('Subscription Details'),
        actions: [
          BlocBuilder<SubscriptionDetailBloc, SubscriptionDetailState>(
            builder: (context, state) {
              final isRefreshing =
                  state is SubscriptionDetailLoaded && state.isRefreshing;
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
                        .read<SubscriptionDetailBloc>()
                        .add(const RefreshSubscriptionRequested()),
              );
            },
          ),
        ],
      ),
      body: BlocBuilder<SubscriptionDetailBloc, SubscriptionDetailState>(
        builder: (context, state) {
          if (state is SubscriptionDetailInitial) {
            context.read<SubscriptionDetailBloc>().add(
                  FetchSubscriptionRequested(
                    appId: appId,
                    subscriptionId: subscriptionId,
                  ),
                );
            return const Center(child: CircularProgressIndicator());
          }

          if (state is SubscriptionDetailLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state is SubscriptionDetailError) {
            return ErrorStateWidget(
              title: 'Failed to load subscription',
              message: state.message,
              onRetry: () => context.read<SubscriptionDetailBloc>().add(
                    FetchSubscriptionRequested(
                      appId: appId,
                      subscriptionId: subscriptionId,
                    ),
                  ),
            );
          }

          if (state is SubscriptionDetailLoaded) {
            return _buildContent(context, state.subscription);
          }

          return const SizedBox.shrink();
        },
      ),
    );
  }

  Widget _buildContent(BuildContext context, Subscription subscription) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildStoreCard(context, subscription),
          const SizedBox(height: 12),
          _buildStoreHealthButton(context, subscription),
          const SizedBox(height: 16),
          RiskStateIndicator(riskState: subscription.riskState),
          const SizedBox(height: 16),
          _buildBillingCard(context, subscription),
          const SizedBox(height: 16),
          _buildDatesCard(context, subscription),
        ],
      ),
    );
  }

  Widget _buildStoreCard(BuildContext context, Subscription subscription) {
    final storeName = subscription.myshopifyDomain
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
          // Store avatar
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
                _buildStatusBadge(subscription.status),
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

  Widget _buildBillingCard(BuildContext context, Subscription subscription) {
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
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Billing Details',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
          ),
          const SizedBox(height: 16),
          _buildInfoRow(
            context,
            'Plan',
            subscription.planName,
            Icons.card_membership,
          ),
          const Divider(height: 24),
          _buildInfoRow(
            context,
            'Price',
            subscription.formattedPrice,
            Icons.attach_money,
          ),
          const Divider(height: 24),
          _buildInfoRow(
            context,
            'Billing Interval',
            subscription.billingInterval.displayName,
            Icons.calendar_today,
          ),
          const Divider(height: 24),
          _buildInfoRow(
            context,
            'Monthly Recurring Revenue',
            '\$${(subscription.mrrCents / 100).toStringAsFixed(2)}/mo',
            Icons.trending_up,
            valueColor: Colors.blue,
          ),
        ],
      ),
    );
  }

  Widget _buildDatesCard(BuildContext context, Subscription subscription) {
    final dateFormat = DateFormat('MMM d, y');

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
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Timeline',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
          ),
          const SizedBox(height: 16),
          _buildInfoRow(
            context,
            'Created',
            dateFormat.format(subscription.createdAt),
            Icons.add_circle_outline,
          ),
          if (subscription.lastChargeDate != null) ...[
            const Divider(height: 24),
            _buildInfoRow(
              context,
              'Last Charge',
              dateFormat.format(subscription.lastChargeDate!),
              Icons.payment,
            ),
          ],
          if (subscription.expectedNextCharge != null) ...[
            const Divider(height: 24),
            _buildInfoRow(
              context,
              'Next Charge',
              dateFormat.format(subscription.expectedNextCharge!),
              Icons.event,
              valueColor: _isOverdue(subscription.expectedNextCharge!)
                  ? Colors.red
                  : Colors.green,
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildInfoRow(
    BuildContext context,
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
          child: Icon(icon, size: 20, color: Colors.grey[600]),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Text(
            label,
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: Colors.grey[600],
                ),
          ),
        ),
        Text(
          value,
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
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

  Widget _buildStoreHealthButton(
      BuildContext context, Subscription subscription) {
    return SizedBox(
      width: double.infinity,
      child: OutlinedButton.icon(
        onPressed: () {
          context.pushNamed(
            'store-health',
            pathParameters: {
              'appId': appId,
              'domain': subscription.myshopifyDomain,
            },
          );
        },
        icon: const Icon(Icons.health_and_safety),
        label: const Text('View Store Health'),
        style: OutlinedButton.styleFrom(
          padding: const EdgeInsets.symmetric(vertical: 12),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
        ),
      ),
    );
  }
}
