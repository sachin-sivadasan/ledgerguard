import 'env_config.dart';

/// Application configuration singleton
class AppConfig {
  static late EnvConfig _envConfig;

  static void init(EnvConfig config) {
    _envConfig = config;
  }

  static EnvConfig get current => _envConfig;
  static String get apiBaseUrl => _envConfig.apiBaseUrl;
  static bool get isDev => _envConfig.isDev;
  static bool get isProd => _envConfig.isProd;
}
