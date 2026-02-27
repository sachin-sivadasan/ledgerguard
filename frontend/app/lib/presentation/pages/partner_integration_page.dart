import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../core/theme/app_theme.dart';
import '../blocs/partner_integration/partner_integration.dart';

/// Partner integration page for connecting Shopify Partner account
class PartnerIntegrationPage extends StatefulWidget {
  const PartnerIntegrationPage({super.key});

  @override
  State<PartnerIntegrationPage> createState() => _PartnerIntegrationPageState();
}

class _PartnerIntegrationPageState extends State<PartnerIntegrationPage> {
  final _formKey = GlobalKey<FormState>();
  final _partnerIdController = TextEditingController();
  final _apiTokenController = TextEditingController();

  @override
  void initState() {
    super.initState();
    // Check integration status on page load
    context
        .read<PartnerIntegrationBloc>()
        .add(const CheckIntegrationStatusRequested());
  }

  @override
  void dispose() {
    _partnerIdController.dispose();
    _apiTokenController.dispose();
    super.dispose();
  }

  void _onConnectWithOAuth() {
    context.read<PartnerIntegrationBloc>().add(const ConnectWithOAuthRequested());
  }

  void _onSaveManualToken() {
    if (_formKey.currentState?.validate() ?? false) {
      context.read<PartnerIntegrationBloc>().add(
            SaveManualTokenRequested(
              partnerId: _partnerIdController.text.trim(),
              apiToken: _apiTokenController.text.trim(),
            ),
          );
    }
  }

  void _onDisconnect() {
    context.read<PartnerIntegrationBloc>().add(const DisconnectRequested());
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Partner Integration'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.go('/dashboard'),
        ),
      ),
      body: BlocConsumer<PartnerIntegrationBloc, PartnerIntegrationState>(
        listener: (context, state) {
          if (state is PartnerIntegrationSuccess) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text(state.message),
                backgroundColor: AppTheme.success,
              ),
            );
            // Navigate to app selection after successful connection
            context.go('/app-selection');
          }
        },
        builder: (context, state) {
          return SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 600),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Header
                  Text(
                    'Connect Shopify Partner Account',
                    style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                          fontWeight: FontWeight.bold,
                        ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Link your Shopify Partner account to import app revenue data.',
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                          color: Colors.grey[600],
                        ),
                  ),
                  const SizedBox(height: 32),

                  // Error message
                  if (state is PartnerIntegrationError) ...[
                    _buildErrorBanner(state.message),
                    const SizedBox(height: 24),
                  ],

                  // Success/Connected state
                  if (state is PartnerIntegrationSuccess ||
                      state is PartnerIntegrationConnected) ...[
                    _buildConnectedCard(context, state),
                    const SizedBox(height: 24),
                  ],

                  // Loading state
                  if (state is PartnerIntegrationLoading) ...[
                    _buildLoadingCard(state.message),
                    const SizedBox(height: 24),
                  ],

                  // Not connected - show connect options
                  if (state is PartnerIntegrationNotConnected ||
                      state is PartnerIntegrationInitial ||
                      state is PartnerIntegrationError) ...[
                    // OAuth Connect Button
                    _buildOAuthSection(context, state),
                    const SizedBox(height: 32),

                    // Manual Token Form (for development, showing to all users)
                    _buildManualTokenSection(context, state),
                  ],
                ],
              ),
            ),
          );
        },
      ),
    );
  }

  Widget _buildErrorBanner(String message) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppTheme.danger.withOpacity(0.1),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: AppTheme.danger.withOpacity(0.3)),
      ),
      child: Row(
        children: [
          const Icon(Icons.error_outline, color: AppTheme.danger),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              message,
              style: const TextStyle(color: AppTheme.danger),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildConnectedCard(BuildContext context, PartnerIntegrationState state) {
    String? partnerId;
    if (state is PartnerIntegrationSuccess) {
      partnerId = state.integration.partnerId;
    } else if (state is PartnerIntegrationConnected) {
      partnerId = state.integration.partnerId;
    }

    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: AppTheme.success.withOpacity(0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppTheme.success.withOpacity(0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: AppTheme.success,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(Icons.check, color: Colors.white, size: 20),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Connected',
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.bold,
                            color: AppTheme.success,
                          ),
                    ),
                    if (partnerId != null)
                      Text(
                        'Partner ID: $partnerId',
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: Colors.grey[600],
                            ),
                      ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          OutlinedButton.icon(
            onPressed: _onDisconnect,
            icon: const Icon(Icons.link_off),
            label: const Text('Disconnect'),
            style: OutlinedButton.styleFrom(
              foregroundColor: AppTheme.danger,
              side: const BorderSide(color: AppTheme.danger),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLoadingCard(String? message) {
    return Container(
      padding: const EdgeInsets.all(32),
      decoration: BoxDecoration(
        color: Colors.grey[100],
        borderRadius: BorderRadius.circular(12),
      ),
      child: Column(
        children: [
          const CircularProgressIndicator(),
          if (message != null) ...[
            const SizedBox(height: 16),
            Text(
              message,
              style: TextStyle(color: Colors.grey[600]),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildOAuthSection(BuildContext context, PartnerIntegrationState state) {
    final isLoading = state is PartnerIntegrationLoading;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: AppTheme.primary.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: const Icon(
                    Icons.storefront,
                    color: AppTheme.primary,
                    size: 24,
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Connect with OAuth',
                        style: Theme.of(context).textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.bold,
                            ),
                      ),
                      Text(
                        'Recommended',
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: AppTheme.success,
                            ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Text(
              'Securely connect your Shopify Partner account using OAuth. This is the recommended method for production use.',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Colors.grey[600],
                  ),
            ),
            const SizedBox(height: 20),
            SizedBox(
              width: double.infinity,
              height: 48,
              child: ElevatedButton.icon(
                onPressed: isLoading ? null : _onConnectWithOAuth,
                icon: const Icon(Icons.link),
                label: const Text('Connect Shopify Partner'),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildManualTokenSection(BuildContext context, PartnerIntegrationState state) {
    final isLoading = state is PartnerIntegrationLoading;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: Colors.orange.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: const Icon(
                      Icons.key,
                      color: Colors.orange,
                      size: 24,
                    ),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Manual Token Entry',
                          style: Theme.of(context).textTheme.titleMedium?.copyWith(
                                fontWeight: FontWeight.bold,
                              ),
                        ),
                        Text(
                          'Admin Only',
                          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                color: Colors.orange,
                              ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              Text(
                'Manually enter your Partner API credentials. Use this for testing or when OAuth is not available.',
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: Colors.grey[600],
                    ),
              ),
              const SizedBox(height: 20),
              TextFormField(
                controller: _partnerIdController,
                decoration: const InputDecoration(
                  labelText: 'Partner ID',
                  hintText: 'Enter your Partner ID',
                  prefixIcon: Icon(Icons.badge_outlined),
                ),
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return 'Partner ID is required';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _apiTokenController,
                obscureText: true,
                decoration: const InputDecoration(
                  labelText: 'API Token',
                  hintText: 'Enter your API token',
                  prefixIcon: Icon(Icons.vpn_key_outlined),
                ),
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return 'API Token is required';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 20),
              SizedBox(
                width: double.infinity,
                height: 48,
                child: ElevatedButton(
                  onPressed: isLoading ? null : _onSaveManualToken,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.orange,
                  ),
                  child: const Text('Save Token'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
