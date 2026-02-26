import 'package:dio/dio.dart';

import '../../core/config/app_config.dart';
import '../../domain/entities/risk_summary.dart';
import '../../domain/repositories/app_repository.dart';
import '../../domain/repositories/auth_repository.dart';
import '../../domain/repositories/risk_repository.dart';

/// API implementation of RiskRepository
class ApiRiskRepository implements RiskRepository {
  final Dio _dio;
  final AuthRepository _authRepository;
  final AppRepository _appRepository;

  ApiRiskRepository({
    Dio? dio,
    required AuthRepository authRepository,
    required AppRepository appRepository,
  })  : _dio = dio ?? Dio(BaseOptions(baseUrl: AppConfig.apiBaseUrl)),
        _authRepository = authRepository,
        _appRepository = appRepository;

  @override
  Future<RiskSummary?> fetchRiskSummary() async {
    final selectedApp = await _appRepository.getSelectedApp();
    if (selectedApp == null) {
      throw const NoAppSelectedRiskException();
    }

    final token = await _authRepository.getIdToken();
    if (token == null) {
      throw const UnauthorizedRiskException();
    }

    try {
      final response = await _dio.get(
        '/api/v1/apps/${selectedApp.id}/metrics/latest',
        options: Options(
          headers: {'Authorization': 'Bearer $token'},
        ),
      );

      if (response.statusCode == 200 && response.data != null) {
        return _parseRiskSummary(response.data as Map<String, dynamic>);
      }

      return null;
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        throw const UnauthorizedRiskException();
      }
      if (e.response?.statusCode == 404) {
        return null;
      }
      throw RiskException(
        e.message ?? 'Failed to fetch risk summary',
        code: 'network-error',
      );
    }
  }

  RiskSummary _parseRiskSummary(Map<String, dynamic> data) {
    return RiskSummary(
      safeCount: (data['safe_count'] as num?)?.toInt() ?? 0,
      oneCycleMissedCount: (data['one_cycle_missed_count'] as num?)?.toInt() ?? 0,
      twoCyclesMissedCount: (data['two_cycles_missed_count'] as num?)?.toInt() ?? 0,
      churnedCount: (data['churned_count'] as num?)?.toInt() ?? 0,
      revenueAtRiskCents: (data['revenue_at_risk_cents'] as num?)?.toInt() ?? 0,
    );
  }
}
