import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../../domain/entities/user_profile.dart';
import '../../blocs/partner_integration/partner_integration.dart';
import '../../blocs/role/role.dart';

/// Admin-only page for manual integration management
class ManualIntegrationPage extends StatefulWidget {
  const ManualIntegrationPage({super.key});

  @override
  State<ManualIntegrationPage> createState() => _ManualIntegrationPageState();
}

class _ManualIntegrationPageState extends State<ManualIntegrationPage> {
  final _formKey = GlobalKey<FormState>();
  final _partnerIdController = TextEditingController();
  final _apiTokenController = TextEditingController();

  @override
  void dispose() {
    _partnerIdController.dispose();
    _apiTokenController.dispose();
    super.dispose();
  }

  void _saveToken() {
    if (_formKey.currentState?.validate() ?? false) {
      context.read<PartnerIntegrationBloc>().add(
            SaveManualTokenRequested(
              partnerId: _partnerIdController.text.trim(),
              apiToken: _apiTokenController.text.trim(),
            ),
          );
    }
  }

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<RoleBloc, RoleState>(
      builder: (context, state) {
        // Check if user has admin access
        if (state is RoleLoaded && state.hasRole(UserRole.admin)) {
          return _buildContent(context);
        }

        // Show access denied for non-admins
        if (state is RoleLoaded) {
          return _buildAccessDenied(context);
        }

        // Show loading while role is being fetched
        return const Scaffold(
          body: Center(child: CircularProgressIndicator()),
        );
      },
    );
  }

  Widget _buildContent(BuildContext context) {
    return BlocConsumer<PartnerIntegrationBloc, PartnerIntegrationState>(
      listener: (context, state) {
        if (state is PartnerIntegrationSuccess) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(state.message),
              backgroundColor: Colors.green,
            ),
          );
          // Clear the form after success
          _partnerIdController.clear();
          _apiTokenController.clear();
        } else if (state is PartnerIntegrationError) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(state.message),
              backgroundColor: Colors.red,
            ),
          );
        }
      },
      builder: (context, state) {
        final isLoading = state is PartnerIntegrationLoading;

        return Scaffold(
          appBar: AppBar(
            title: const Text('Manual Integration'),
            leading: IconButton(
              icon: const Icon(Icons.arrow_back),
              onPressed: () => context.go('/settings'),
            ),
          ),
          body: Padding(
            padding: const EdgeInsets.all(24),
            child: Form(
              key: _formKey,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Partner API Token',
                    style: Theme.of(context).textTheme.titleLarge,
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Manually configure your Shopify Partner API token for testing.',
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                          color: Colors.grey[600],
                        ),
                  ),
                  const SizedBox(height: 24),
                  TextFormField(
                    controller: _partnerIdController,
                    enabled: !isLoading,
                    decoration: const InputDecoration(
                      labelText: 'Partner ID',
                      hintText: 'Enter your Partner ID',
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
                    enabled: !isLoading,
                    obscureText: true,
                    decoration: const InputDecoration(
                      labelText: 'API Token',
                      hintText: 'Enter your API token',
                    ),
                    validator: (value) {
                      if (value == null || value.trim().isEmpty) {
                        return 'API Token is required';
                      }
                      return null;
                    },
                  ),
                  const SizedBox(height: 24),
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed: isLoading ? null : _saveToken,
                      child: isLoading
                          ? const SizedBox(
                              height: 20,
                              width: 20,
                              child: CircularProgressIndicator(
                                strokeWidth: 2,
                              ),
                            )
                          : const Text('Save Token'),
                    ),
                  ),
                ],
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _buildAccessDenied(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Access Denied'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.go('/dashboard'),
        ),
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.lock_outlined,
              size: 64,
              color: Colors.grey[400],
            ),
            const SizedBox(height: 16),
            Text(
              'Admin Access Required',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            Text(
              'You do not have permission to access this page.',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Colors.grey[600],
                  ),
            ),
            const SizedBox(height: 24),
            ElevatedButton(
              onPressed: () => context.go('/dashboard'),
              child: const Text('Go to Dashboard'),
            ),
          ],
        ),
      ),
    );
  }
}
