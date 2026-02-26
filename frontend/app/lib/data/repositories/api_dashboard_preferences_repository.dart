import 'package:dio/dio.dart';

import '../../core/config/app_config.dart';
import '../../domain/entities/dashboard_preferences.dart';
import '../../domain/repositories/auth_repository.dart';
import '../../domain/repositories/dashboard_preferences_repository.dart';

/// API implementation of DashboardPreferencesRepository
class ApiDashboardPreferencesRepository
    implements DashboardPreferencesRepository {
  final Dio _dio;
  final AuthRepository _authRepository;

  ApiDashboardPreferencesRepository({
    Dio? dio,
    required AuthRepository authRepository,
  })  : _dio = dio ?? Dio(BaseOptions(baseUrl: AppConfig.apiBaseUrl)),
        _authRepository = authRepository;

  @override
  Future<DashboardPreferences> fetchPreferences() async {
    final token = await _authRepository.getIdToken();
    if (token == null) {
      throw const UnauthorizedPreferencesException();
    }

    try {
      final response = await _dio.get(
        '/api/v1/user/preferences/dashboard',
        options: Options(
          headers: {'Authorization': 'Bearer $token'},
        ),
      );

      if (response.statusCode == 200 && response.data != null) {
        return DashboardPreferences.fromJson(
            response.data as Map<String, dynamic>);
      }

      // No preferences stored yet, return defaults
      return DashboardPreferences.defaults();
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        throw const UnauthorizedPreferencesException();
      }
      if (e.response?.statusCode == 404) {
        // No preferences found, return defaults
        return DashboardPreferences.defaults();
      }
      throw FetchPreferencesException(
        e.message ?? 'Failed to fetch preferences',
      );
    }
  }

  @override
  Future<void> savePreferences(DashboardPreferences preferences) async {
    final token = await _authRepository.getIdToken();
    if (token == null) {
      throw const UnauthorizedPreferencesException();
    }

    try {
      await _dio.put(
        '/api/v1/user/preferences/dashboard',
        data: preferences.toJson(),
        options: Options(
          headers: {'Authorization': 'Bearer $token'},
        ),
      );
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        throw const UnauthorizedPreferencesException();
      }
      throw SavePreferencesException(
        e.message ?? 'Failed to save preferences',
      );
    }
  }
}
