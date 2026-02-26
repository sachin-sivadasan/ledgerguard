import 'package:dio/dio.dart';

import '../../core/config/app_config.dart';
import '../../domain/entities/user_profile.dart';
import '../../domain/repositories/user_profile_repository.dart';

/// API implementation of UserProfileRepository
class ApiUserProfileRepository implements UserProfileRepository {
  final Dio _dio;
  UserProfile? _cachedProfile;

  ApiUserProfileRepository({Dio? dio})
      : _dio = dio ?? Dio(BaseOptions(baseUrl: AppConfig.apiBaseUrl));

  @override
  UserProfile? get cachedProfile => _cachedProfile;

  @override
  void clearCache() {
    _cachedProfile = null;
  }

  @override
  Future<UserProfile?> fetchUserProfile(String authToken) async {
    try {
      final response = await _dio.get(
        '/api/v1/me',
        options: Options(
          headers: {'Authorization': 'Bearer $authToken'},
        ),
      );

      if (response.statusCode == 200) {
        final data = response.data as Map<String, dynamic>;
        _cachedProfile = _parseProfile(data);
        return _cachedProfile;
      }

      return null;
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        throw const UnauthorizedException();
      }
      if (e.response?.statusCode == 404) {
        throw const ProfileNotFoundException();
      }
      throw UserProfileException(
        e.message ?? 'Failed to fetch profile',
        code: 'network-error',
      );
    }
  }

  UserProfile _parseProfile(Map<String, dynamic> data) {
    return UserProfile(
      id: data['id'] as String,
      email: data['email'] as String,
      role: UserRole.fromString(data['role'] as String? ?? 'ADMIN'),
      planTier: PlanTier.fromString(data['plan_tier'] as String? ?? 'STARTER'),
      displayName: data['display_name'] as String?,
    );
  }
}
