import '../entities/partner_integration.dart';

/// Repository interface for partner integration operations
abstract class PartnerIntegrationRepository {
  /// Get current integration status
  Future<PartnerIntegration> getIntegrationStatus();

  /// Initiate OAuth connection with Shopify Partner
  Future<PartnerIntegration> connectWithOAuth();

  /// Save manual token (admin only)
  Future<PartnerIntegration> saveManualToken({
    required String partnerId,
    required String apiToken,
  });

  /// Disconnect partner integration
  Future<void> disconnect();
}

/// Exception for partner integration errors
class PartnerIntegrationException implements Exception {
  final String message;
  final String? code;

  const PartnerIntegrationException(this.message, {this.code});

  @override
  String toString() => message;
}
