import 'dart:io' show Platform;

/// Environment configuration for LedgerGuard
enum Environment { dev, staging, prod }

class EnvConfig {
  final Environment environment;
  final String _devApiBaseUrlAndroid;
  final String _devApiBaseUrlOther;
  final String _prodApiBaseUrl;
  final String firebaseProjectId;

  const EnvConfig._({
    required this.environment,
    required String devApiBaseUrlAndroid,
    required String devApiBaseUrlOther,
    required String prodApiBaseUrl,
    required this.firebaseProjectId,
  })  : _devApiBaseUrlAndroid = devApiBaseUrlAndroid,
        _devApiBaseUrlOther = devApiBaseUrlOther,
        _prodApiBaseUrl = prodApiBaseUrl;

  /// Returns the appropriate API base URL based on environment and platform
  String get apiBaseUrl {
    if (environment == Environment.prod || environment == Environment.staging) {
      return _prodApiBaseUrl;
    }
    // Dev environment - use 10.0.2.2 for Android emulator
    try {
      if (Platform.isAndroid) {
        return _devApiBaseUrlAndroid;
      }
    } catch (_) {
      // Platform not available (web), use default
    }
    return _devApiBaseUrlOther;
  }

  static const EnvConfig dev = EnvConfig._(
    environment: Environment.dev,
    devApiBaseUrlAndroid: 'http://10.0.2.2:8080',
    devApiBaseUrlOther: 'http://localhost:8080',
    prodApiBaseUrl: 'https://api.ledgerspear.com',
    firebaseProjectId: 'ledgerguard-dev',
  );

  static const EnvConfig staging = EnvConfig._(
    environment: Environment.staging,
    devApiBaseUrlAndroid: 'http://10.0.2.2:8080',
    devApiBaseUrlOther: 'http://localhost:8080',
    prodApiBaseUrl: 'https://api.ledgerspear.com',
    firebaseProjectId: 'ledgerguard-c7557',
  );

  static const EnvConfig prod = EnvConfig._(
    environment: Environment.prod,
    devApiBaseUrlAndroid: 'http://10.0.2.2:8080',
    devApiBaseUrlOther: 'http://localhost:8080',
    prodApiBaseUrl: 'https://api.ledgerspear.com',
    firebaseProjectId: 'ledgerguard-c7557',
  );

  bool get isDev => environment == Environment.dev;
  bool get isStaging => environment == Environment.staging;
  bool get isProd => environment == Environment.prod;
}
