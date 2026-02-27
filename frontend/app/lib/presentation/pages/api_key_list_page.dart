import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../domain/entities/api_key.dart';
import '../blocs/api_key/api_key.dart';
import '../widgets/api_key_tile.dart';

class ApiKeyListPage extends StatefulWidget {
  const ApiKeyListPage({super.key});

  @override
  State<ApiKeyListPage> createState() => _ApiKeyListPageState();
}

class _ApiKeyListPageState extends State<ApiKeyListPage> {
  @override
  void initState() {
    super.initState();
    context.read<ApiKeyBloc>().add(const LoadApiKeysRequested());
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('API Keys'),
        actions: [
          IconButton(
            icon: const Icon(Icons.add),
            onPressed: () => _showCreateDialog(context),
            tooltip: 'Create API Key',
          ),
        ],
      ),
      body: BlocConsumer<ApiKeyBloc, ApiKeyState>(
        listener: (context, state) {
          if (state is ApiKeyCreated) {
            _showKeyCreatedDialog(context, state.fullKey, state.keyName);
          } else if (state is ApiKeyError && state.previousKeys != null) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text(state.message),
                backgroundColor: Colors.red,
              ),
            );
          }
        },
        builder: (context, state) {
          if (state is ApiKeyLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state is ApiKeyError && state.previousKeys == null) {
            return _buildErrorState(context, state.message);
          }

          if (state is ApiKeyEmpty) {
            return _buildEmptyState(context);
          }

          List<ApiKey> apiKeys = [];
          bool isCreating = false;
          String? revokingKeyId;

          if (state is ApiKeyLoaded) {
            apiKeys = state.apiKeys;
            isCreating = state.isCreating;
            revokingKeyId = state.revokingKeyId;
          } else if (state is ApiKeyCreated) {
            apiKeys = state.apiKeys;
          } else if (state is ApiKeyError && state.previousKeys != null) {
            apiKeys = state.previousKeys!;
          }

          return _buildKeyList(context, apiKeys, isCreating, revokingKeyId);
        },
      ),
    );
  }

  Widget _buildKeyList(
    BuildContext context,
    List<ApiKey> apiKeys,
    bool isCreating,
    String? revokingKeyId,
  ) {
    return RefreshIndicator(
      onRefresh: () async {
        context.read<ApiKeyBloc>().add(const LoadApiKeysRequested());
      },
      child: ListView(
        padding: const EdgeInsets.symmetric(vertical: 8),
        children: [
          if (isCreating)
            const Padding(
              padding: EdgeInsets.all(16),
              child: Center(child: CircularProgressIndicator()),
            ),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            child: Text(
              '${apiKeys.length} API key${apiKeys.length == 1 ? '' : 's'}',
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: Colors.grey[600],
                  ),
            ),
          ),
          ...apiKeys.map((apiKey) => ApiKeyTile(
                apiKey: apiKey,
                isRevoking: revokingKeyId == apiKey.id,
                onRevoke: () => _confirmRevoke(context, apiKey),
              )),
        ],
      ),
    );
  }

  Widget _buildEmptyState(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.key_off,
              size: 64,
              color: Colors.grey[400],
            ),
            const SizedBox(height: 16),
            Text(
              'No API Keys',
              style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 8),
            Text(
              'Create an API key to access the Revenue API',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Colors.grey[600],
                  ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: () => _showCreateDialog(context),
              icon: const Icon(Icons.add),
              label: const Text('Create API Key'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildErrorState(BuildContext context, String message) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.error_outline,
              size: 64,
              color: Colors.red[400],
            ),
            const SizedBox(height: 16),
            Text(
              'Error',
              style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 8),
            Text(
              message,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Colors.grey[600],
                  ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: () {
                context.read<ApiKeyBloc>().add(const LoadApiKeysRequested());
              },
              icon: const Icon(Icons.refresh),
              label: const Text('Retry'),
            ),
          ],
        ),
      ),
    );
  }

  void _showCreateDialog(BuildContext context) {
    final nameController = TextEditingController();
    final formKey = GlobalKey<FormState>();

    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Create API Key'),
        content: Form(
          key: formKey,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                'Give your API key a name to help you identify it later.',
                style: TextStyle(color: Colors.grey),
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: nameController,
                decoration: const InputDecoration(
                  labelText: 'Key Name',
                  hintText: 'e.g., Production Server',
                  border: OutlineInputBorder(),
                ),
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return 'Please enter a name';
                  }
                  if (value.length > 50) {
                    return 'Name must be 50 characters or less';
                  }
                  return null;
                },
                autofocus: true,
              ),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(dialogContext).pop(),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: () {
              if (formKey.currentState!.validate()) {
                Navigator.of(dialogContext).pop();
                context.read<ApiKeyBloc>().add(
                      CreateApiKeyRequested(nameController.text.trim()),
                    );
              }
            },
            child: const Text('Create'),
          ),
        ],
      ),
    );
  }

  void _showKeyCreatedDialog(BuildContext context, String fullKey, String keyName) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (dialogContext) => AlertDialog(
        title: Row(
          children: [
            Icon(Icons.check_circle, color: Colors.green[600]),
            const SizedBox(width: 8),
            const Text('API Key Created'),
          ],
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.amber[50],
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.amber[200]!),
              ),
              child: Row(
                children: [
                  Icon(Icons.warning, color: Colors.amber[700], size: 20),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      'Copy this key now. You won\'t be able to see it again!',
                      style: TextStyle(
                        color: Colors.amber[900],
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 16),
            Text(
              'Key Name: $keyName',
              style: const TextStyle(fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.grey[100],
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.grey[300]!),
              ),
              child: Row(
                children: [
                  Expanded(
                    child: SelectableText(
                      fullKey,
                      style: const TextStyle(
                        fontFamily: 'monospace',
                        fontSize: 13,
                      ),
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.copy),
                    onPressed: () {
                      Clipboard.setData(ClipboardData(text: fullKey));
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(
                          content: Text('API key copied to clipboard'),
                          backgroundColor: Colors.green,
                        ),
                      );
                    },
                    tooltip: 'Copy to clipboard',
                  ),
                ],
              ),
            ),
          ],
        ),
        actions: [
          ElevatedButton(
            onPressed: () {
              Navigator.of(dialogContext).pop();
              context.read<ApiKeyBloc>().add(const DismissKeyCreatedRequested());
            },
            child: const Text('Done'),
          ),
        ],
      ),
    );
  }

  void _confirmRevoke(BuildContext context, ApiKey apiKey) {
    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Revoke API Key'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Are you sure you want to revoke "${apiKey.name}"?'),
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.red[50],
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.red[200]!),
              ),
              child: Row(
                children: [
                  Icon(Icons.warning, color: Colors.red[700], size: 20),
                  const SizedBox(width: 8),
                  const Expanded(
                    child: Text(
                      'This action cannot be undone. Any applications using this key will stop working.',
                      style: TextStyle(color: Colors.red),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(dialogContext).pop(),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.red,
              foregroundColor: Colors.white,
            ),
            onPressed: () {
              Navigator.of(dialogContext).pop();
              context.read<ApiKeyBloc>().add(RevokeApiKeyRequested(apiKey.id));
            },
            child: const Text('Revoke'),
          ),
        ],
      ),
    );
  }
}
