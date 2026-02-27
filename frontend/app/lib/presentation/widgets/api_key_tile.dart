import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../domain/entities/api_key.dart';

/// A tile displaying an API key with copy and revoke actions
class ApiKeyTile extends StatelessWidget {
  final ApiKey apiKey;
  final VoidCallback? onRevoke;
  final bool isRevoking;

  const ApiKeyTile({
    super.key,
    required this.apiKey,
    this.onRevoke,
    this.isRevoking = false,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        apiKey.name,
                        style: Theme.of(context).textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.bold,
                            ),
                      ),
                      const SizedBox(height: 4),
                      Row(
                        children: [
                          Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 8,
                              vertical: 4,
                            ),
                            decoration: BoxDecoration(
                              color: Colors.grey[100],
                              borderRadius: BorderRadius.circular(4),
                              border: Border.all(color: Colors.grey[300]!),
                            ),
                            child: Text(
                              apiKey.keyPrefix,
                              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                    fontFamily: 'monospace',
                                    color: Colors.grey[700],
                                  ),
                            ),
                          ),
                          const SizedBox(width: 8),
                          IconButton(
                            icon: const Icon(Icons.copy, size: 16),
                            onPressed: () => _copyPrefix(context),
                            tooltip: 'Copy prefix',
                            visualDensity: VisualDensity.compact,
                            padding: EdgeInsets.zero,
                            constraints: const BoxConstraints(),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
                if (isRevoking)
                  const SizedBox(
                    width: 24,
                    height: 24,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                else
                  IconButton(
                    icon: const Icon(Icons.delete_outline, color: Colors.red),
                    onPressed: onRevoke,
                    tooltip: 'Revoke key',
                  ),
              ],
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                _buildInfoChip(
                  context,
                  icon: Icons.access_time,
                  label: 'Created ${apiKey.formattedCreatedAt}',
                ),
                const SizedBox(width: 12),
                _buildInfoChip(
                  context,
                  icon: Icons.schedule,
                  label: apiKey.formattedLastUsed,
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildInfoChip(BuildContext context, {required IconData icon, required String label}) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 14, color: Colors.grey[500]),
        const SizedBox(width: 4),
        Text(
          label,
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Colors.grey[600],
              ),
        ),
      ],
    );
  }

  void _copyPrefix(BuildContext context) {
    Clipboard.setData(ClipboardData(text: apiKey.keyPrefix));
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Key prefix copied to clipboard'),
        duration: Duration(seconds: 2),
      ),
    );
  }
}
