import 'package:dio/dio.dart';

import '../config/app_config.dart';
import '../services/snackbar_service.dart';
import '../../domain/repositories/auth_repository.dart';

/// Centralized API client with interceptors for error handling and token refresh
class ApiClient {
  final Dio _dio;
  final AuthRepository _authRepository;
  final SnackbarService _snackbarService;

  ApiClient({
    required AuthRepository authRepository,
    SnackbarService? snackbarService,
    Dio? dio,
  })  : _authRepository = authRepository,
        _snackbarService = snackbarService ?? SnackbarService(),
        _dio = dio ??
            Dio(BaseOptions(
              baseUrl: AppConfig.apiBaseUrl,
              connectTimeout: const Duration(seconds: 30),
              receiveTimeout: const Duration(seconds: 30),
            )) {
    _dio.interceptors.addAll([
      _AuthInterceptor(authRepository: _authRepository),
      _ErrorInterceptor(
        snackbarService: _snackbarService,
        authRepository: _authRepository,
      ),
    ]);
  }

  Dio get dio => _dio;

  /// Perform a GET request
  Future<Response<T>> get<T>(
    String path, {
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) {
    return _dio.get<T>(
      path,
      queryParameters: queryParameters,
      options: options,
    );
  }

  /// Perform a POST request
  Future<Response<T>> post<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) {
    return _dio.post<T>(
      path,
      data: data,
      queryParameters: queryParameters,
      options: options,
    );
  }

  /// Perform a PUT request
  Future<Response<T>> put<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) {
    return _dio.put<T>(
      path,
      data: data,
      queryParameters: queryParameters,
      options: options,
    );
  }

  /// Perform a DELETE request
  Future<Response<T>> delete<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) {
    return _dio.delete<T>(
      path,
      data: data,
      queryParameters: queryParameters,
      options: options,
    );
  }

  /// Perform a PATCH request
  Future<Response<T>> patch<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) {
    return _dio.patch<T>(
      path,
      data: data,
      queryParameters: queryParameters,
      options: options,
    );
  }
}

/// Interceptor to add auth token to requests
class _AuthInterceptor extends Interceptor {
  final AuthRepository _authRepository;

  _AuthInterceptor({required AuthRepository authRepository})
      : _authRepository = authRepository;

  @override
  Future<void> onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    // Skip auth header if already present
    if (options.headers.containsKey('Authorization')) {
      return handler.next(options);
    }

    // Add auth token
    final token = await _authRepository.getIdToken();
    if (token != null) {
      options.headers['Authorization'] = 'Bearer $token';
    }

    handler.next(options);
  }
}

/// Interceptor for global error handling and token refresh
class _ErrorInterceptor extends Interceptor {
  final SnackbarService _snackbarService;
  final AuthRepository _authRepository;
  bool _isRefreshing = false;

  _ErrorInterceptor({
    required SnackbarService snackbarService,
    required AuthRepository authRepository,
  })  : _snackbarService = snackbarService,
        _authRepository = authRepository;

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    final statusCode = err.response?.statusCode;

    // Handle 401 - Attempt token refresh
    if (statusCode == 401 && !_isRefreshing) {
      _isRefreshing = true;
      try {
        // Try to refresh the token
        final newToken = await _authRepository.getIdToken();
        _isRefreshing = false;

        if (newToken != null) {
          // Retry the request with new token
          final options = err.requestOptions;
          options.headers['Authorization'] = 'Bearer $newToken';

          final dio = Dio(BaseOptions(baseUrl: options.baseUrl));
          final response = await dio.fetch(options);
          return handler.resolve(response);
        }
      } catch (_) {
        _isRefreshing = false;
      }

      // Token refresh failed - sign out
      await _authRepository.signOut();
      _snackbarService.showError('Session expired. Please sign in again.');
      return handler.next(err);
    }

    // Map error codes to user-friendly messages
    final message = _getErrorMessage(statusCode, err);

    // Show snackbar for non-recoverable errors
    if (_shouldShowSnackbar(statusCode, err)) {
      _snackbarService.showError(message);
    }

    handler.next(err);
  }

  String _getErrorMessage(int? statusCode, DioException err) {
    switch (statusCode) {
      case 400:
        return 'Invalid request. Please check your input.';
      case 401:
        return 'Session expired. Please sign in again.';
      case 403:
        return 'You don\'t have permission to perform this action.';
      case 404:
        return 'The requested resource was not found.';
      case 409:
        return 'This action conflicts with existing data.';
      case 422:
        return 'Invalid data. Please check your input.';
      case 429:
        return 'Too many requests. Please try again later.';
      case 500:
      case 502:
      case 503:
        return 'Server error. Please try again later.';
      default:
        if (err.type == DioExceptionType.connectionTimeout ||
            err.type == DioExceptionType.receiveTimeout) {
          return 'Connection timed out. Please check your internet.';
        }
        if (err.type == DioExceptionType.connectionError) {
          return 'Connection error. Please check your internet.';
        }
        return 'An unexpected error occurred.';
    }
  }

  bool _shouldShowSnackbar(int? statusCode, DioException err) {
    // Don't show snackbar for:
    // - 401 (handled separately with sign out)
    // - 404 (often expected for empty states)
    // - Cancel exceptions
    if (statusCode == 401 || statusCode == 404) return false;
    if (err.type == DioExceptionType.cancel) return false;
    return true;
  }
}

/// API exception with user-friendly message
class ApiException implements Exception {
  final String message;
  final int? statusCode;
  final String? code;

  const ApiException(this.message, {this.statusCode, this.code});

  @override
  String toString() => message;
}
