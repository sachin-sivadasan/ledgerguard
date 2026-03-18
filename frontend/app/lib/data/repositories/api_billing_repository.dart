import '../../core/network/api_client.dart';
import '../../domain/entities/billing_status.dart';
import '../../domain/repositories/billing_repository.dart';

/// API implementation of BillingRepository using Razorpay backend endpoints.
class ApiBillingRepository implements BillingRepository {
  final ApiClient _apiClient;

  ApiBillingRepository({required ApiClient apiClient}) : _apiClient = apiClient;

  @override
  Future<BillingStatus> getBillingStatus() async {
    try {
      final response = await _apiClient.get('/api/v1/billing/status');

      if (response.statusCode != 200) {
        throw BillingException(
          'Failed to fetch billing status: ${response.statusCode}',
        );
      }

      return BillingStatus.fromJson(response.data as Map<String, dynamic>);
    } catch (e) {
      if (e is BillingException) rethrow;
      throw BillingException('Failed to fetch billing status: $e');
    }
  }

  @override
  Future<CheckoutResult> createCheckout(String plan) async {
    try {
      final response = await _apiClient.post(
        '/api/v1/billing/checkout',
        data: {'plan': plan},
      );

      if (response.statusCode != 200) {
        throw BillingException(
          'Failed to create checkout: ${response.statusCode}',
        );
      }

      return CheckoutResult.fromJson(response.data as Map<String, dynamic>);
    } catch (e) {
      if (e is BillingException) rethrow;
      throw BillingException('Failed to create checkout: $e');
    }
  }
}
