import 'package:url_launcher/url_launcher.dart';

import '../../core/network/api_client.dart';
import '../../domain/entities/partner_integration.dart';
import '../../domain/repositories/partner_integration_repository.dart';

/// API implementation of PartnerIntegrationRepository
class ApiPartnerIntegrationRepository implements PartnerIntegrationRepository {
  final ApiClient _apiClient;

  ApiPartnerIntegrationRepository({required ApiClient apiClient})
      : _apiClient = apiClient;

  @override
  Future<PartnerIntegration> getIntegrationStatus() async {
    try {
      final response = await _apiClient.get('/api/v1/integrations/shopify/status');

      if (response.statusCode == 200) {
        final data = response.data as Map<String, dynamic>;
        return PartnerIntegration(
          partnerId: data['partner_id'] as String?,
          status: data['connected'] == true
              ? IntegrationStatus.connected
              : IntegrationStatus.notConnected,
          connectedAt: data['connected_at'] != null
              ? DateTime.parse(data['connected_at'] as String)
              : null,
        );
      }

      return const PartnerIntegration();
    } catch (e) {
      // If 404, means not connected
      return const PartnerIntegration();
    }
  }

  @override
  Future<PartnerIntegration> connectWithOAuth() async {
    try {
      // Get OAuth URL from backend
      final response = await _apiClient.get('/api/v1/integrations/shopify/oauth');

      if (response.statusCode != 200) {
        throw const PartnerIntegrationException(
          'Failed to initiate OAuth connection',
          code: 'oauth-init-failed',
        );
      }

      final data = response.data as Map<String, dynamic>;
      final url = data['url'] as String?;

      if (url == null || url.isEmpty) {
        throw const PartnerIntegrationException(
          'Invalid OAuth URL received',
          code: 'invalid-oauth-url',
        );
      }

      // Open OAuth URL in browser
      final uri = Uri.parse(url);
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
      } else {
        throw const PartnerIntegrationException(
          'Could not open authentication page',
          code: 'launch-failed',
        );
      }

      // Return connecting status - the callback will complete the flow
      // In a real app, you'd use a deep link or polling to detect completion
      return const PartnerIntegration(
        status: IntegrationStatus.connecting,
      );
    } catch (e) {
      if (e is PartnerIntegrationException) rethrow;
      throw PartnerIntegrationException(
        'Failed to connect: ${e.toString()}',
        code: 'connection-failed',
      );
    }
  }

  @override
  Future<PartnerIntegration> saveManualToken({
    required String partnerId,
    required String apiToken,
  }) async {
    try {
      final response = await _apiClient.post(
        '/api/v1/integrations/shopify/token',
        data: {
          'partner_id': partnerId,
          'token': apiToken,
        },
      );

      if (response.statusCode == 200 || response.statusCode == 201) {
        return PartnerIntegration(
          partnerId: partnerId,
          status: IntegrationStatus.connected,
          connectedAt: DateTime.now(),
        );
      }

      throw const PartnerIntegrationException(
        'Failed to save token',
        code: 'save-failed',
      );
    } catch (e) {
      if (e is PartnerIntegrationException) rethrow;
      throw PartnerIntegrationException(
        'Failed to save token: ${e.toString()}',
        code: 'save-failed',
      );
    }
  }

  @override
  Future<void> disconnect() async {
    try {
      await _apiClient.delete('/api/v1/integrations/shopify/token');
    } catch (e) {
      throw PartnerIntegrationException(
        'Failed to disconnect: ${e.toString()}',
        code: 'disconnect-failed',
      );
    }
  }
}
