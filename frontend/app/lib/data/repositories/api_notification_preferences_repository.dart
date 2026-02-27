import 'package:dio/dio.dart';

import '../../core/config/app_config.dart';
import '../../domain/entities/notification_preferences.dart';
import '../../domain/repositories/auth_repository.dart';
import '../../domain/repositories/notification_preferences_repository.dart';

/// API implementation of NotificationPreferencesRepository
class ApiNotificationPreferencesRepository implements NotificationPreferencesRepository {
  final AuthRepository _authRepository;
  final Dio _dio;

  ApiNotificationPreferencesRepository({
    required AuthRepository authRepository,
    Dio? dio,
  })  : _authRepository = authRepository,
        _dio = dio ?? Dio(BaseOptions(baseUrl: AppConfig.apiBaseUrl));

  @override
  Future<NotificationPreferences> fetchPreferences() async {
    try {
      final token = await _authRepository.getIdToken();
      if (token == null) {
        throw const UnauthorizedNotificationPreferencesException();
      }

      final response = await _dio.get(
        '/api/v1/users/notification-preferences',
        options: Options(
          headers: {'Authorization': 'Bearer $token'},
        ),
      );

      if (response.statusCode == 200 && response.data != null) {
        return NotificationPreferences.fromJson(response.data as Map<String, dynamic>);
      }

      // Return default preferences if no data
      return NotificationPreferences.defaultPreferences;
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        throw const UnauthorizedNotificationPreferencesException();
      }
      throw LoadNotificationPreferencesException(
        e.response?.data?['message'] ?? 'Failed to load notification preferences',
      );
    }
  }

  @override
  Future<void> savePreferences(NotificationPreferences preferences) async {
    try {
      final token = await _authRepository.getIdToken();
      if (token == null) {
        throw const UnauthorizedNotificationPreferencesException();
      }

      final response = await _dio.put(
        '/api/v1/users/notification-preferences',
        data: preferences.toJson(),
        options: Options(
          headers: {'Authorization': 'Bearer $token'},
        ),
      );

      if (response.statusCode != 200 && response.statusCode != 204) {
        throw SaveNotificationPreferencesException(
          response.data?['message'] ?? 'Failed to save notification preferences',
        );
      }
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        throw const UnauthorizedNotificationPreferencesException();
      }
      throw SaveNotificationPreferencesException(
        e.response?.data?['message'] ?? 'Failed to save notification preferences',
      );
    }
  }
}
