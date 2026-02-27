import 'package:dio/dio.dart';

import '../../core/config/app_config.dart';
import '../../domain/entities/daily_insight.dart';
import '../../domain/repositories/app_repository.dart';
import '../../domain/repositories/auth_repository.dart';
import '../../domain/repositories/insight_repository.dart';

/// API implementation of InsightRepository
class ApiInsightRepository implements InsightRepository {
  final Dio _dio;
  final AuthRepository _authRepository;
  final AppRepository _appRepository;

  ApiInsightRepository({
    Dio? dio,
    required AuthRepository authRepository,
    required AppRepository appRepository,
  })  : _dio = dio ?? Dio(BaseOptions(baseUrl: AppConfig.apiBaseUrl)),
        _authRepository = authRepository,
        _appRepository = appRepository;

  @override
  Future<DailyInsight?> fetchDailyInsight() async {
    final selectedApp = await _appRepository.getSelectedApp();
    if (selectedApp == null) {
      throw const NoAppSelectedInsightException();
    }

    final token = await _authRepository.getIdToken();
    if (token == null) {
      throw const UnauthorizedInsightException();
    }

    try {
      final response = await _dio.get(
        '/api/v1/apps/${selectedApp.id}/insights/daily',
        options: Options(
          headers: {'Authorization': 'Bearer $token'},
        ),
      );

      if (response.statusCode == 200 && response.data != null) {
        return DailyInsight.fromJson(response.data as Map<String, dynamic>);
      }

      return null;
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        throw const UnauthorizedInsightException();
      }
      if (e.response?.statusCode == 403) {
        throw const ProRequiredInsightException();
      }
      if (e.response?.statusCode == 404) {
        return null;
      }
      throw InsightException(
        e.message ?? 'Failed to fetch insight',
        code: 'network-error',
      );
    }
  }
}
