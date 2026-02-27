import 'package:dio/dio.dart';

import '../../core/config/app_config.dart';
import '../../domain/entities/earnings_timeline.dart';
import '../../domain/repositories/app_repository.dart';
import '../../domain/repositories/auth_repository.dart';
import '../../domain/repositories/earnings_repository.dart';

/// API implementation of EarningsRepository
class ApiEarningsRepository implements EarningsRepository {
  final Dio _dio;
  final AuthRepository _authRepository;
  final AppRepository _appRepository;

  ApiEarningsRepository({
    Dio? dio,
    required AuthRepository authRepository,
    required AppRepository appRepository,
  })  : _dio = dio ?? Dio(BaseOptions(baseUrl: AppConfig.apiBaseUrl)),
        _authRepository = authRepository,
        _appRepository = appRepository;

  @override
  Future<EarningsTimeline> fetchMonthlyEarnings({
    required int year,
    required int month,
    required EarningsMode mode,
  }) async {
    // Get the selected app
    final selectedApp = await _appRepository.getSelectedApp();
    if (selectedApp == null) {
      throw const NoAppSelectedEarningsException();
    }

    // Get auth token
    final token = await _authRepository.getIdToken();
    if (token == null) {
      throw const UnauthorizedEarningsException();
    }

    try {
      // Extract numeric ID from full GID
      final appId = _extractNumericId(selectedApp.id);

      // Build query parameters
      final queryParams = <String, dynamic>{
        'year': year,
        'month': month,
        'mode': mode == EarningsMode.split ? 'split' : 'combined',
      };

      final response = await _dio.get(
        '/api/v1/apps/$appId/earnings',
        queryParameters: queryParams,
        options: Options(
          headers: {'Authorization': 'Bearer $token'},
        ),
      );

      if (response.statusCode == 200) {
        final data = response.data as Map<String, dynamic>;
        return EarningsTimeline.fromJson(data);
      }

      // Return empty timeline for no data
      return EarningsTimeline(
        month: '$year-${month.toString().padLeft(2, '0')}',
        earnings: [],
      );
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        throw const UnauthorizedEarningsException();
      }
      if (e.response?.statusCode == 400) {
        final message = e.response?.data?['error']?['message'] as String? ?? '';
        if (message.contains('invalid month')) {
          throw const InvalidMonthException();
        }
        if (message.contains('future')) {
          throw const FutureMonthException();
        }
      }
      throw EarningsException(
        e.message ?? 'Failed to fetch earnings',
        code: 'network-error',
      );
    }
  }

  /// Extracts numeric ID from Shopify GID
  /// e.g., "gid://partners/App/4599915" -> "4599915"
  String _extractNumericId(String gid) {
    final parts = gid.split('/');
    return parts.isNotEmpty ? parts.last : gid;
  }
}
