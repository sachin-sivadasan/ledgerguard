import '../entities/billing_status.dart';

/// Repository interface for billing operations.
abstract class BillingRepository {
  /// Get the current billing status for the authenticated user.
  Future<BillingStatus> getBillingStatus();

  /// Create a checkout session for the given plan.
  /// Returns a [CheckoutResult] with the Razorpay hosted checkout URL.
  Future<CheckoutResult> createCheckout(String plan);
}

/// Exception thrown when billing operations fail.
class BillingException implements Exception {
  final String message;
  const BillingException(this.message);

  @override
  String toString() => message;
}
