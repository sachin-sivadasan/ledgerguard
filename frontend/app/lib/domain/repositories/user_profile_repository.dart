import '../entities/user_profile.dart';

/// Repository interface for user profile operations
abstract class UserProfileRepository {
  /// Fetch user profile from backend
  /// Returns null if not found
  Future<UserProfile?> fetchUserProfile(String authToken);

  /// Get cached user profile
  UserProfile? get cachedProfile;

  /// Clear cached profile
  void clearCache();
}

/// User profile exceptions
class UserProfileException implements Exception {
  final String message;
  final String? code;

  const UserProfileException(this.message, {this.code});

  @override
  String toString() => 'UserProfileException: $message';
}

class UnauthorizedException extends UserProfileException {
  const UnauthorizedException([String message = 'Unauthorized'])
      : super(message, code: 'unauthorized');
}

class ProfileNotFoundException extends UserProfileException {
  const ProfileNotFoundException([String message = 'Profile not found'])
      : super(message, code: 'not-found');
}
