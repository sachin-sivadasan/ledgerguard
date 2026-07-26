class AppConfig {
  static const String devApiUrl = 'http://localhost:8080';
  static const String stagingApiUrl =
      'https://ledgerspear-api-ineifpjrdq-uc.a.run.app';
  static const String prodApiUrl = 'https://api.ledgerspear.com';

  static String get apiBaseUrl => const String.fromEnvironment(
        'API_BASE_URL',
        defaultValue: devApiUrl,
      );
}
