import '../../domain/entities/partner_integration.dart';
import '../../domain/repositories/partner_integration_repository.dart';

/// Mock implementation of PartnerIntegrationRepository for development
class MockPartnerIntegrationRepository implements PartnerIntegrationRepository {
  PartnerIntegration _currentIntegration = const PartnerIntegration();

  /// Simulated delay for API calls
  final Duration delay;

  MockPartnerIntegrationRepository({
    this.delay = const Duration(milliseconds: 1500),
  });

  @override
  Future<PartnerIntegration> getIntegrationStatus() async {
    await Future.delayed(delay);
    return _currentIntegration;
  }

  @override
  Future<PartnerIntegration> connectWithOAuth() async {
    await Future.delayed(delay);

    // Simulate successful OAuth connection
    _currentIntegration = PartnerIntegration(
      partnerId: 'mock-partner-${DateTime.now().millisecondsSinceEpoch}',
      status: IntegrationStatus.connected,
      connectedAt: DateTime.now(),
    );

    return _currentIntegration;
  }

  @override
  Future<PartnerIntegration> saveManualToken({
    required String partnerId,
    required String apiToken,
  }) async {
    await Future.delayed(delay);

    // Validate inputs
    if (partnerId.isEmpty) {
      throw const PartnerIntegrationException(
        'Partner ID is required',
        code: 'invalid-partner-id',
      );
    }
    if (apiToken.isEmpty) {
      throw const PartnerIntegrationException(
        'API Token is required',
        code: 'invalid-token',
      );
    }

    // Simulate successful token save
    _currentIntegration = PartnerIntegration(
      partnerId: partnerId,
      status: IntegrationStatus.connected,
      connectedAt: DateTime.now(),
    );

    return _currentIntegration;
  }

  @override
  Future<void> disconnect() async {
    await Future.delayed(delay);
    _currentIntegration = const PartnerIntegration();
  }
}
