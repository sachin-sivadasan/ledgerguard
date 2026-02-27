import 'package:dio/dio.dart';

import '../../core/config/app_config.dart';
import '../../domain/entities/api_key.dart';
import '../../domain/repositories/api_key_repository.dart';
import '../../domain/repositories/auth_repository.dart';

/// API implementation of ApiKeyRepository
class ApiApiKeyRepository implements ApiKeyRepository {
  final Dio _dio;
  final AuthRepository _authRepository;

  ApiApiKeyRepository({
    Dio? dio,
    required AuthRepository authRepository,
  })  : _dio = dio ?? Dio(),
        _authRepository = authRepository;

  String get _baseUrl => AppConfig.apiBaseUrl;

  Future<Map<String, String>> get _headers async {
    final token = await _authRepository.getIdToken();
    return {
      'Authorization': 'Bearer $token',
      'Content-Type': 'application/json',
    };
  }

  @override
  Future<List<ApiKey>> getApiKeys() async {
    try {
      final response = await _dio.get(
        '$_baseUrl/api/v1/api-keys',
        options: Options(headers: await _headers),
      );

      final List<dynamic> data = response.data['api_keys'] ?? [];
      return data.map((json) => ApiKey.fromJson(json)).toList();
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        throw const ApiKeyUnauthorizedException();
      }
      throw ApiKeyException(e.message ?? 'Failed to fetch API keys');
    }
  }

  @override
  Future<ApiKeyCreationResult> createApiKey(String name) async {
    try {
      final response = await _dio.post(
        '$_baseUrl/api/v1/api-keys',
        data: {'name': name},
        options: Options(headers: await _headers),
      );

      return ApiKeyCreationResult.fromJson(response.data);
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        throw const ApiKeyUnauthorizedException();
      }
      if (e.response?.statusCode == 400) {
        final message = e.response?.data?['error']?['message'] ?? 'Invalid request';
        if (message.toString().toLowerCase().contains('limit')) {
          throw const ApiKeyLimitException();
        }
        throw ApiKeyException(message);
      }
      throw ApiKeyException(e.message ?? 'Failed to create API key');
    }
  }

  @override
  Future<void> revokeApiKey(String keyId) async {
    try {
      await _dio.delete(
        '$_baseUrl/v1/api-keys/$keyId',
        options: Options(headers: await _headers),
      );
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        throw const ApiKeyUnauthorizedException();
      }
      if (e.response?.statusCode == 404) {
        throw const ApiKeyNotFoundException();
      }
      throw ApiKeyException(e.message ?? 'Failed to revoke API key');
    }
  }
}
