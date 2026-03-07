import 'package:flutter/material.dart';

import '../../../core/theme/app_theme.dart';
import '../../../domain/entities/chat_message.dart';

/// Displays structured data (risk, metrics, subscriptions, etc.) from the
/// most recent assistant message that has state data.
class DataPanel extends StatelessWidget {
  final ChatMessage message;

  const DataPanel({super.key, required this.message});

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        border: Border(
          left: BorderSide(
            color: Theme.of(context).dividerColor,
          ),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _buildHeader(context),
          const Divider(height: 1),
          Expanded(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  if (message.riskState != null)
                    _buildSection(context, 'Risk Summary', Icons.shield,
                        AppTheme.warning, message.riskState!),
                  if (message.metricsState != null)
                    _buildSection(context, 'Metrics', Icons.bar_chart,
                        AppTheme.primary, message.metricsState!),
                  if (message.subscriptionState != null)
                    _buildSection(context, 'Subscriptions', Icons.people,
                        AppTheme.accent, message.subscriptionState!),
                  if (message.earningsState != null)
                    _buildSection(context, 'Earnings', Icons.attach_money,
                        AppTheme.success, message.earningsState!),
                  if (message.storeHealthState != null)
                    _buildSection(context, 'Store Health', Icons.store,
                        AppTheme.secondary, message.storeHealthState!),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildHeader(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      child: Row(
        children: [
          Icon(Icons.analytics_outlined,
              size: 18, color: AppTheme.primary),
          const SizedBox(width: 8),
          Text(
            'Data Preview',
            style: Theme.of(context).textTheme.titleSmall?.copyWith(
                  fontWeight: FontWeight.w600,
                ),
          ),
        ],
      ),
    );
  }

  Widget _buildSection(
    BuildContext context,
    String title,
    IconData icon,
    Color color,
    Map<String, dynamic> data,
  ) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(8),
        side: BorderSide(color: Theme.of(context).dividerColor),
      ),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(icon, size: 16, color: color),
                const SizedBox(width: 6),
                Text(
                  title,
                  style: Theme.of(context).textTheme.labelLarge?.copyWith(
                        fontWeight: FontWeight.w600,
                        color: color,
                      ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            ..._buildDataRows(context, data),
          ],
        ),
      ),
    );
  }

  List<Widget> _buildDataRows(
      BuildContext context, Map<String, dynamic> data) {
    return data.entries.map((entry) {
      final value = entry.value;
      final label = _formatKey(entry.key);

      if (value is Map) {
        return Padding(
          padding: const EdgeInsets.only(bottom: 4),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(label,
                  style: Theme.of(context)
                      .textTheme
                      .labelMedium
                      ?.copyWith(fontWeight: FontWeight.w600)),
              const SizedBox(height: 2),
              ..._buildDataRows(context, value.cast<String, dynamic>()),
            ],
          ),
        );
      }

      if (value is List) {
        return Padding(
          padding: const EdgeInsets.only(bottom: 4),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('$label (${value.length} items)',
                  style: Theme.of(context)
                      .textTheme
                      .labelMedium
                      ?.copyWith(fontWeight: FontWeight.w600)),
            ],
          ),
        );
      }

      return Padding(
        padding: const EdgeInsets.only(bottom: 4),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Flexible(
              child: Text(label,
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: Theme.of(context).colorScheme.onSurfaceVariant,
                      )),
            ),
            const SizedBox(width: 8),
            Flexible(
              child: Text(
                '$value',
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      fontWeight: FontWeight.w600,
                    ),
                textAlign: TextAlign.end,
              ),
            ),
          ],
        ),
      );
    }).toList();
  }

  String _formatKey(String key) {
    // Convert snake_case to Title Case
    return key
        .split('_')
        .map((w) =>
            w.isEmpty ? '' : '${w[0].toUpperCase()}${w.substring(1)}')
        .join(' ');
  }
}
